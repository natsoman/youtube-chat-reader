//go:generate mockgen -destination=mock_youtube_test.go -package=youtube_test -source=stream_list_grpc.pb.go -exclude_interfaces V3DataLiveChatMessageServiceServer,UnsafeV3DataLiveChatMessageServiceServer
//go:generate mockgen -destination=mock_test.go -package=youtube_test -source=grpc.go
package youtube_test

import (
	"context"
	"errors"
	"io"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/google"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/youtube"
)

func TestNewGRPCClient(t *testing.T) {
	t.Parallel()

	t.Run("successfully creates client with valid inputs", func(t *testing.T) {
		_, deps := setupTest(t)

		// Given
		apiKeys := []string{"api-key-1", "api-key-2"}
		ticker := &google.Ticker{}

		// When
		client, err := youtube.NewStreamChatMessagesGRPCClient(deps.dataLiveChatMessageServiceClient, ticker, ticker, apiKeys)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("returns error when api keys are empty", func(t *testing.T) {
		_, deps := setupTest(t)

		// Given
		apiKeys := []string{}
		ticker := &google.Ticker{}

		// When
		client, err := youtube.NewStreamChatMessagesGRPCClient(deps.dataLiveChatMessageServiceClient, ticker, ticker, apiKeys)

		// Then
		assert.EqualError(t, err, "API keys are empty")
		assert.Nil(t, client)
	})

	t.Run("returns error when grpc client is nil", func(t *testing.T) {
		// Given
		apiKeys := []string{"api-key-1"}
		ticker := &google.Ticker{}

		// When
		client, err := youtube.NewStreamChatMessagesGRPCClient(nil, ticker, ticker, apiKeys)

		// Then
		assert.EqualError(t, err, "V3DataLiveChatMessageServiceClient is nil")
		assert.Nil(t, client)
	})

	t.Run("returns error when stream list ticker is nil", func(t *testing.T) {
		_, deps := setupTest(t)

		// Given
		apiKeys := []string{"api-key-1"}
		ticker := &google.Ticker{}

		// When
		client, err := youtube.NewStreamChatMessagesGRPCClient(deps.dataLiveChatMessageServiceClient, nil, ticker, apiKeys)

		// Then
		assert.EqualError(t, err, "steam list ticker is nil")
		assert.Nil(t, client)
	})

	t.Run("returns error when recv ticker is nil", func(t *testing.T) {
		_, deps := setupTest(t)

		// Given
		apiKeys := []string{"api-key-1"}
		ticker := &google.Ticker{}

		// When
		client, err := youtube.NewStreamChatMessagesGRPCClient(deps.dataLiveChatMessageServiceClient, ticker, nil, apiKeys)

		// Then
		assert.EqualError(t, err, "recv ticker is nil")
		assert.Nil(t, client)
	})
}

func TestGRPCClient_StreamChatMessages(t *testing.T) {
	t.Parallel()

	t.Run("makes correct request and handles successfully multiple message types in one response", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		const publishedAt = "2023-01-01T12:00:00Z"

		lsp := newLiveStreamProgress(t)
		resp := &mockServerStreamingClient{
			responses: []*youtube.LiveChatMessageListResponse{
				{
					NextPageToken: strPtr("next-token-multi"),
					Items: []*youtube.LiveChatMessage{
						{
							Id: strPtr("text-msg-1"),
							Snippet: &youtube.LiveChatMessageSnippet{
								Type:            youtube.LiveChatMessageSnippet_TypeWrapper_TEXT_MESSAGE_EVENT.Enum(),
								PublishedAt:     strPtr(publishedAt),
								AuthorChannelId: strPtr("author-1"),
								DisplayMessage:  strPtr("Hello"),
							},
							AuthorDetails: &youtube.LiveChatMessageAuthorDetails{
								ChannelId:       strPtr("author-1"),
								DisplayName:     strPtr("User 1"),
								ProfileImageUrl: strPtr("https://example.com/user1.jpg"),
							},
						},
						{
							Id: strPtr("sc-msg-1"),
							Snippet: &youtube.LiveChatMessageSnippet{
								Type:            youtube.LiveChatMessageSnippet_TypeWrapper_SUPER_CHAT_EVENT.Enum(),
								PublishedAt:     strPtr(publishedAt),
								AuthorChannelId: strPtr("author-2"),
								DisplayedContent: &youtube.LiveChatMessageSnippet_SuperChatDetails{
									SuperChatDetails: &youtube.LiveChatSuperChatDetails{
										UserComment:         strPtr("Donate!"),
										AmountDisplayString: strPtr("$10.00"),
										AmountMicros:        uint64Ptr(10000000),
										Currency:            strPtr("USD"),
									},
								},
							},
							AuthorDetails: &youtube.LiveChatMessageAuthorDetails{
								ChannelId:       strPtr("author-2"),
								DisplayName:     strPtr("Donor"),
								ProfileImageUrl: strPtr("https://example.com/donor2.jpg"),
							},
						},
						{
							Id: strPtr("ban-msg-1"),
							Snippet: &youtube.LiveChatMessageSnippet{
								Type:        youtube.LiveChatMessageSnippet_TypeWrapper_USER_BANNED_EVENT.Enum(),
								PublishedAt: strPtr(publishedAt),
								DisplayedContent: &youtube.LiveChatMessageSnippet_UserBannedDetails{
									UserBannedDetails: &youtube.LiveChatUserBannedMessageDetails{
										BannedUserDetails: &youtube.ChannelProfileDetails{
											ChannelId: strPtr("banned-channel"),
										},
										BanType:            youtube.LiveChatUserBannedMessageDetails_BanTypeWrapper_PERMANENT.Enum(),
										BanDurationSeconds: uint64Ptr(0),
									},
								},
							},
							AuthorDetails: &youtube.LiveChatMessageAuthorDetails{
								ChannelId:       strPtr("author-3"),
								DisplayName:     strPtr("Moderator"),
								ProfileImageUrl: strPtr("https://example.com/mod.jpg"),
							},
						},
					},
				},
			},
		}
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})

		recvThrottle := make(chan time.Time)
		deps.recvTicker.EXPECT().
			Start(gomock.Any()).
			Return(recvThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(
				metadata.NewOutgoingContext(t.Context(), metadata.Pairs("x-goog-api-key", "test-api-key-1")),
				gomock.Cond(func(req any) bool {
					r, ok := req.(*youtube.LiveChatMessageListRequest)
					if !ok {
						return false
					}

					return r.GetLiveChatId() == lsp.ChatID() &&
						r.GetMaxResults() == 2000 &&
						r.GetPageToken() == "" &&
						slices.Equal(r.Part, []string{"id", "snippet", "authorDetails"})
				})).
			Return(resp, nil)

		// When
		go func() {
			streamListThrottle <- time.Now()

			recvThrottle <- time.Now()
		}()

		msgChan, _ := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case msg := <-msgChan:
			assert.NotNil(t, msg)
			assert.Len(t, msg.TextMessages(), 1)
			assert.Len(t, msg.Donates(), 1)
			assert.Len(t, msg.Bans(), 1)
			assert.Len(t, msg.Authors(), 3)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for message")
		}
	})

	t.Run("returns error when stream list fails", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("grpc connection error"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.Error(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("returns chat not found error on not found gRPC error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.NotFound, "chat not found"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.ErrorIs(t, err, domain.ErrChatNotFound)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("returns chat offline error on failed precondition gRPC error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.FailedPrecondition, "chat offline"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.ErrorIs(t, err, domain.ErrChatOffline)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("returns unavailable live stream error on resource exhausted gRPC error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})

		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.ResourceExhausted, "quota exceeded"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.ErrorIs(t, err, domain.ErrUnavailableLiveStream)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("handles stream receive error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		resp := &mockServerStreamingClient{errors: []error{status.Error(codes.Internal, "internal error")}}
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})

		recvThrottle := make(chan time.Time)
		deps.recvTicker.EXPECT().
			Start(gomock.Any()).
			Return(recvThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(resp, nil)

		// When
		go func() {
			streamListThrottle <- time.Now()

			recvThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.Error(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("handles stream receive end of file and continues", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})

		recvThrottle := make(chan time.Time)
		deps.recvTicker.EXPECT().
			Start(gomock.Any()).
			Return(recvThrottle, func() {}).
			Times(2)

		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(&mockServerStreamingClient{errors: []error{errors.New("grpc connection error")}}, nil).
			After(
				deps.dataLiveChatMessageServiceClient.EXPECT().
					StreamList(gomock.Any(), gomock.Any()).
					Return(&mockServerStreamingClient{errors: []error{io.EOF}}, nil),
			)

		// When
		go func() {
			streamListThrottle <- time.Now()

			recvThrottle <- time.Now()

			streamListThrottle <- time.Now()

			recvThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.Error(t, err)
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("error is returned when an item has invalid published at", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		invalidPublishedAt := "invalid-timestamp"
		nextPageToken := "next-token-invalid"

		streamClient := &mockServerStreamingClient{
			responses: []*youtube.LiveChatMessageListResponse{
				{
					NextPageToken: &nextPageToken,
					Items: []*youtube.LiveChatMessage{
						{
							Snippet: &youtube.LiveChatMessageSnippet{
								PublishedAt: &invalidPublishedAt,
							},
						},
					},
				},
			},
		}
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})

		recvThrottle := make(chan time.Time)
		deps.recvTicker.EXPECT().
			Start(gomock.Any()).
			Return(recvThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(streamClient, nil)

		// When
		go func() {
			streamListThrottle <- time.Now()

			recvThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "parse published at")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})

	t.Run("returns nil for Canceled error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, status.Error(codes.Canceled, "canceled"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		msgChan, _ := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case _, ok := <-msgChan:
			assert.False(t, ok, "channel should be closed")
		case <-time.After(time.Second):
			// Expected - context canceled returns nil which doesn't send on errChan
		}
	})

	t.Run("returns original error for non-gRPC error", func(t *testing.T) {
		client, deps := setupTest(t)

		// Given
		streamListThrottle := make(chan time.Time)
		deps.streamListTicker.EXPECT().
			Start(gomock.Any()).
			Return(streamListThrottle, func() {})
		deps.dataLiveChatMessageServiceClient.EXPECT().
			StreamList(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("regular error"))

		// When
		go func() {
			streamListThrottle <- time.Now()
		}()

		_, errChan := client.StreamChatMessages(t.Context(), newLiveStreamProgress(t))

		// Then
		select {
		case err := <-errChan:
			assert.EqualError(t, err, "regular error")
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for error")
		}
	})
}

type mockServerStreamingClient struct {
	responses []*youtube.LiveChatMessageListResponse
	errors    []error
	index     int
}

func (m *mockServerStreamingClient) Recv() (*youtube.LiveChatMessageListResponse, error) {
	if m.index >= len(m.responses) {
		if m.index < len(m.errors) {
			err := m.errors[m.index]
			m.index++

			return nil, err
		}

		return nil, io.EOF
	}

	resp := m.responses[m.index]

	var err error
	if m.index < len(m.errors) {
		err = m.errors[m.index]
	}

	m.index++

	return resp, err
}

func (m *mockServerStreamingClient) Header() (metadata.MD, error) {
	return nil, nil
}

func (m *mockServerStreamingClient) Trailer() metadata.MD {
	return nil
}

func (m *mockServerStreamingClient) CloseSend() error {
	return nil
}

func (m *mockServerStreamingClient) Context() context.Context {
	return context.Background()
}

func (m *mockServerStreamingClient) SendMsg(msg interface{}) error {
	return nil
}

func (m *mockServerStreamingClient) RecvMsg(msg interface{}) error {
	return nil
}

type testDeps struct {
	dataLiveChatMessageServiceClient *MockV3DataLiveChatMessageServiceClient
	streamListTicker                 *MockTicker
	recvTicker                       *MockTicker
}

func setupTest(t *testing.T) (*youtube.StreamChatMessagesGRPCClient, *testDeps) {
	t.Helper()
	t.Parallel()

	ctrl := gomock.NewController(t)
	deps := &testDeps{
		dataLiveChatMessageServiceClient: NewMockV3DataLiveChatMessageServiceClient(ctrl),
		streamListTicker:                 NewMockTicker(ctrl),
		recvTicker:                       NewMockTicker(ctrl),
	}

	client, err := youtube.NewStreamChatMessagesGRPCClient(
		deps.dataLiveChatMessageServiceClient,
		deps.streamListTicker,
		deps.recvTicker,
		[]string{"test-api-key-1"},
	)
	require.NoError(t, err)

	return client, deps
}

func newLiveStreamProgress(t *testing.T) *domain.LiveStreamProgress {
	t.Helper()

	lsp, err := domain.NewLiveStreamProgress("live-stream-1", "chat-1", time.Now())
	require.NoError(t, err)

	return lsp
}

func strPtr(s string) *string {
	return &s
}

func uint64Ptr(u uint64) *uint64 {
	return &u
}
