package youtube

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type Ticker interface {
	Start(d time.Duration) (<-chan time.Time, func())
}

type StreamChatMessagesGRPCClient struct {
	log             *slog.Logger
	steamListTicker Ticker
	recvTicker      Ticker
	apiKeys         []string
	grpcClient      V3DataLiveChatMessageServiceClient
}

func NewStreamChatMessagesGRPCClient(grpcClient V3DataLiveChatMessageServiceClient, steamListTicker, recvTicker Ticker,
	apiKeys []string) (*StreamChatMessagesGRPCClient, error) {
	if steamListTicker == nil {
		return nil, errors.New("steam list ticker is nil")
	}

	if recvTicker == nil {
		return nil, errors.New("recv ticker is nil")
	}

	if len(apiKeys) == 0 {
		return nil, errors.New("API keys are empty")
	}

	if grpcClient == nil {
		return nil, errors.New("V3DataLiveChatMessageServiceClient is nil")
	}

	return &StreamChatMessagesGRPCClient{
		log:             slog.Default().With("cmp", "youtube.grpc_client"),
		steamListTicker: steamListTicker,
		recvTicker:      recvTicker,
		apiKeys:         apiKeys,
		grpcClient:      grpcClient,
	}, nil
}

func (c *StreamChatMessagesGRPCClient) StreamChatMessages(ctx context.Context, lsp *domain.LiveStreamProgress) (
	<-chan domain.ChatMessages, <-chan error) {
	cmChan := make(chan domain.ChatMessages)
	errChan := make(chan error)

	nextPageToken := lsp.NextPageToken()

	go func() {
		defer close(cmChan)

		l := c.log.With("ls_id", lsp.ID())

		l.DebugContext(ctx, "YouTube streaming is starting")
		defer l.DebugContext(ctx, "YouTube streaming stopped")

		streamThrottle, streamThrottleStop := c.steamListTicker.Start(time.Second * 2)
		defer streamThrottleStop()

		for {
			select {
			case <-streamThrottle: // Call StreamList
				maxResults := uint32(2000)
				liveChatId := lsp.ChatID()

				streamList, sErr := c.grpcClient.StreamList(
					metadata.NewOutgoingContext(ctx, metadata.Pairs("x-goog-api-key", c.apiKey())),
					&LiveChatMessageListRequest{
						LiveChatId: &liveChatId,
						MaxResults: &maxResults,
						PageToken:  &nextPageToken,
						Part:       []string{"id", "snippet", "authorDetails"},
					})
				if sErr != nil {
					l.ErrorContext(ctx, "StreamList", "npt", nextPageToken, "err", sErr.Error())

					// Send error to the consumer and stop execution.
					// cmChan will be closed, and it will inform the consumer.
					errChan <- parseGRPCError(ctx, l, sErr)

					return
				}

				l.DebugContext(ctx, "StreamList", "npt", nextPageToken)

				func() {
					recvThrottle, recvThrottleStop := c.recvTicker.Start(time.Second * 2)
					defer recvThrottleStop()

					for {
						select {
						case <-recvThrottle: // Receive messages
							cm, err := func() (*domain.ChatMessages, error) {
								resp, err := streamList.Recv()
								if err != nil {
									l.ErrorContext(ctx, "StreamList.Recv", "npt", nextPageToken, "err", err.Error())
									return nil, parseGRPCError(ctx, l, err)
								}

								l.DebugContext(ctx, "StreamList.Recv", "npt", nextPageToken, "num_of_items", len(resp.Items))

								cm, err := chatMessagesFromResp(lsp.ID(), resp)
								if err != nil {
									return nil, err
								}

								return cm, nil
							}()

							if ctx.Err() != nil {
								return
							}

							if err != nil {
								if err == io.EOF {
									// Stop stream receiving.
									return
								}

								// Send error to the consumer and stop stream receiving. Let them decide
								// if they want to stop execution through context cancellation.
								errChan <- err

								return
							}

							if cm.NextPageToken() != "" {
								nextPageToken = cm.NextPageToken()
							}

							cmChan <- *cm
						case <-ctx.Done():
							return
						}
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	return cmChan, errChan
}

func (c *StreamChatMessagesGRPCClient) apiKey() string {
	// nolint:gosec
	return c.apiKeys[rand.Intn(len(c.apiKeys))]
}

func chatMessagesFromResp(liveStreamID string, resp *LiveChatMessageListResponse) (*domain.ChatMessages, error) {
	cm := domain.NewChatMessages(resp.GetNextPageToken())

	for _, item := range resp.Items {
		publishedAt, err := time.Parse(time.RFC3339, item.Snippet.GetPublishedAt())
		if err != nil {
			return nil, fmt.Errorf("parse published at: %v", err)
		}

		switch item.Snippet.GetType() {
		case LiveChatMessageSnippet_TypeWrapper_TEXT_MESSAGE_EVENT:
			msg, err := domain.NewTextMessage(
				item.GetId(),
				liveStreamID,
				item.Snippet.GetAuthorChannelId(),
				item.Snippet.GetDisplayMessage(),
				publishedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("new text messages: %v", err)
			}

			cm.AddTextMessage(msg)
		case LiveChatMessageSnippet_TypeWrapper_USER_BANNED_EVENT:
			ban, err := domain.NewBan(
				item.GetId(),
				item.Snippet.GetUserBannedDetails().GetBannedUserDetails().GetChannelId(),
				liveStreamID,
				item.Snippet.GetUserBannedDetails().GetBanType().String(),
				time.Duration(item.Snippet.GetUserBannedDetails().GetBanDurationSeconds())*time.Second,
				publishedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("new ban: %v", err)
			}

			cm.AddBan(ban)
		case LiveChatMessageSnippet_TypeWrapper_SUPER_CHAT_EVENT:
			dnt, err := domain.NewDonate(
				item.GetId(),
				item.Snippet.GetAuthorChannelId(),
				liveStreamID,
				item.Snippet.GetSuperChatDetails().GetUserComment(),
				item.Snippet.GetSuperChatDetails().GetAmountDisplayString(),
				uint(item.Snippet.GetSuperChatDetails().GetAmountMicros()),
				item.Snippet.GetSuperChatDetails().GetCurrency(),
				publishedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("new donate: %v", err)
			}

			cm.AddDonate(dnt)
		}

		a, err := domain.NewAuthor(
			item.AuthorDetails.GetChannelId(),
			item.AuthorDetails.GetDisplayName(),
			item.AuthorDetails.GetProfileImageUrl(),
			item.AuthorDetails.GetIsVerified(),
		)
		if err != nil {
			return nil, fmt.Errorf("new author: %v", err)
		}

		cm.AddAuthor(a)
	}

	return cm, nil
}

func parseGRPCError(ctx context.Context, l *slog.Logger, err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound, codes.PermissionDenied:
		return domain.ErrChatNotFound
	case codes.FailedPrecondition:
		return domain.ErrChatOffline
	case codes.ResourceExhausted:
		return domain.ErrUnavailableLiveStream
	case codes.Canceled:
		return nil
	default:
		return err
	}
}
