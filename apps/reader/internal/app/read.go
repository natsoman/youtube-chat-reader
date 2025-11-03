package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type Clock interface {
	Now() time.Time
}

type Ticker interface {
	Start(d time.Duration) (<-chan time.Time, func())
}

type LiveStreamProgressRepository interface {
	Upsert(ctx context.Context, lsp *domain.LiveStreamProgress) error
	// Started returns the progress of live streams that have already started or
	// will start within the provided startsWithin duration.
	Started(ctx context.Context, startsWithin time.Duration) ([]domain.LiveStreamProgress, error)
}

type ChatMessageStreamer interface {
	// StreamChatMessages streams chat messages and errors through the returned channels.
	// It stops when the chat message channel is closed and can also be stopped via context cancellation.
	//
	// If the chat ID does not exist, domain.ErrChatNotFound must be sent through the error channel.
	// domain.ErrChatOffline must be sent if the chat has gone offline.
	// domain.ErrUnavailableLiveStream must be sent if there are insufficient resources to read messages.
	StreamChatMessages(ctx context.Context, lsp *domain.LiveStreamProgress) (<-chan domain.ChatMessages, <-chan error)
}

type Locker interface {
	Lock(ctx context.Context, key string) (bool, error)
	Release(ctx context.Context, key string) error
}

type BanRepository interface {
	// Insert adds the provided bans to the repository, ignoring duplicates.
	Insert(ctx context.Context, bb []domain.Ban) error
}

type TextMessageRepository interface {
	// Insert adds the provided text messages to the repository, ignoring duplicates.
	Insert(ctx context.Context, tms []domain.TextMessage) error
}

type DonateRepository interface {
	// Insert adds the provided donates to the repository, ignoring duplicates.
	Insert(ctx context.Context, dd []domain.Donate) error
}

type AuthorRepository interface {
	Upsert(ctx context.Context, aa []domain.Author) error
}

type LiveStreamReader struct {
	log             *slog.Logger
	clock           Clock
	ticker          Ticker
	locker          Locker
	cmStreamer      ChatMessageStreamer
	progressRepo    LiveStreamProgressRepository
	banRepo         BanRepository
	textMessageRepo TextMessageRepository
	donateRepo      DonateRepository
	authorRepo      AuthorRepository
	retryInterval   time.Duration
	advanceStart    time.Duration
	wg              sync.WaitGroup
}

func NewLiveStreamReader(
	clock Clock,
	ticker Ticker,
	locker Locker,
	cmStreamer ChatMessageStreamer,
	progressRepo LiveStreamProgressRepository,
	banRepo BanRepository,
	textMessageRepo TextMessageRepository,
	donateRepo DonateRepository,
	authorRepo AuthorRepository,
	opts ...Option,
) (*LiveStreamReader, error) {
	if clock == nil {
		return nil, errors.New("clock is nil")
	}

	if ticker == nil {
		return nil, errors.New("ticker is nil")
	}

	if locker == nil {
		return nil, errors.New("locker is nil")
	}

	if cmStreamer == nil {
		return nil, errors.New("chat message streamer is nil")
	}

	if progressRepo == nil {
		return nil, errors.New("live stream progress repository is nil")
	}

	if banRepo == nil {
		return nil, errors.New("ban repository is nil")
	}

	if textMessageRepo == nil {
		return nil, errors.New("text message repository is nil")
	}

	if donateRepo == nil {
		return nil, errors.New("donate repository is nil")
	}

	if authorRepo == nil {
		return nil, errors.New("author repository is nil")
	}

	lsr := &LiveStreamReader{
		log:             slog.Default().With("cmp", "chat_reader"),
		clock:           clock,
		ticker:          ticker,
		locker:          locker,
		cmStreamer:      cmStreamer,
		banRepo:         banRepo,
		textMessageRepo: textMessageRepo,
		donateRepo:      donateRepo,
		authorRepo:      authorRepo,
		progressRepo:    progressRepo,
		retryInterval:   time.Second * 10,
		advanceStart:    time.Minute,
	}

	for _, opt := range opts {
		if err := opt(lsr); err != nil {
			return nil, err
		}
	}

	return lsr, nil
}

// Read continuously reads domain.ChatMessages of started or upcoming domain.LiveStreamProgress until they finish.
// It can be stopped via context cancellation.
func (lsr *LiveStreamReader) Read(ctx context.Context) {
	defer lsr.log.InfoContext(ctx, "Reading stopped")

	readStartedLiveStreams := func() {
		liveStreamsProgress, err := lsr.progressRepo.Started(ctx, lsr.advanceStart)
		if err != nil {
			lsr.log.ErrorContext(ctx, "Failed to fetch started live streams progress", "err", err)
			return
		}

		for _, lsp := range liveStreamsProgress {
			lsr.wg.Add(1)

			l := lsr.log.With("ls_id", lsp.ID())

			go func() {
				defer lsr.wg.Done()

				if !lsr.tryLock(ctx, l, lsp.ID()) {
					return
				}

				defer lsr.release(ctx, l, lsp.ID())

				if rErr := lsr.readLiveStream(ctx, l, &lsp); rErr != nil && !errors.Is(rErr, context.Canceled) {
					l.ErrorContext(ctx, "Failed to read live stream", "err", rErr)
				}
			}()
		}
	}

	t, stop := lsr.ticker.Start(lsr.retryInterval)
	defer stop()

	// Run once immediately
	readStartedLiveStreams()

	// Then keep retrying
	for {
		select {
		case <-t:
			readStartedLiveStreams()
		case <-ctx.Done():
			lsr.log.InfoContext(ctx, "Stopping read...")
			lsr.wg.Wait()

			return
		}
	}
}

func (lsr *LiveStreamReader) readLiveStream(ctx context.Context, l *slog.Logger, lsp *domain.LiveStreamProgress) error {
	streamCtx, streamCancel := context.WithCancel(ctx)
	defer streamCancel()

	l.InfoContext(ctx, "Start live stream reading")

	cmChan, errChan := lsr.cmStreamer.StreamChatMessages(streamCtx, lsp)

	for {
		exit, err := func() (bool, error) {
			select {
			case cm, ok := <-cmChan:
				if !ok {
					l.DebugContext(ctx, "Streaming channel closed")
					return true, nil
				}

				if cm.NextPageToken() != "" {
					lsp.SetNextPageToken(cm.NextPageToken())
				} else {
					lsp.Finish(lsr.clock.Now(), "empty next page token")
				}

				if err := lsr.store(ctx, lsp, &cm); err != nil {
					return false, err
				}

				l.InfoContext(ctx, "Chat stored",
					"npt", cm.NextPageToken(),
					"txt", len(cm.TextMessages()),
					"dnt", len(cm.Donates()),
					"ban", len(cm.Bans()),
					"auth", len(cm.Authors()),
				)

				if lsp.IsFinished() {
					return true, nil
				}
			case err, ok := <-errChan:
				if !ok {
					l.DebugContext(ctx, "Error channel closed")
					return true, nil
				}

				l.ErrorContext(ctx, "Error channel", "err", err)

				if !oneOf(err, domain.ErrUnavailableLiveStream, domain.ErrChatOffline, domain.ErrChatNotFound) {
					return false, err
				}

				lsp.Finish(lsr.clock.Now(), err.Error())

				if err = lsr.progressRepo.Upsert(ctx, lsp); err != nil {
					return false, fmt.Errorf("upsert live stream progress: %v", err)
				}

				return true, nil
			case <-ctx.Done():
				return false, ctx.Err()
			}

			return false, nil
		}()
		if err != nil {
			return err
		}

		if exit {
			return nil
		}
	}
}

func (lsr *LiveStreamReader) store(ctx context.Context, lsp *domain.LiveStreamProgress, cm *domain.ChatMessages) error {
	g, _ := errgroup.WithContext(ctx)
	g.SetLimit(4)

	if len(cm.Authors()) > 0 {
		g.Go(func() error {
			if err := lsr.authorRepo.Upsert(ctx, cm.Authors()); err != nil {
				return fmt.Errorf("insert to authors repo: %v", err)
			}

			return nil
		})
	}

	if len(cm.Bans()) > 0 {
		g.Go(func() error {
			if err := lsr.banRepo.Insert(ctx, cm.Bans()); err != nil {
				return fmt.Errorf("insert to ban repo: %v", err)
			}

			return nil
		})
	}

	if len(cm.TextMessages()) > 0 {
		g.Go(func() error {
			if err := lsr.textMessageRepo.Insert(ctx, cm.TextMessages()); err != nil {
				return fmt.Errorf("insert to text messages repo: %v", err)
			}

			return nil
		})
	}

	if len(cm.Donates()) > 0 {
		g.Go(func() error {
			if err := lsr.donateRepo.Insert(ctx, cm.Donates()); err != nil {
				return fmt.Errorf("insert to donates repo: %v", err)
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	// Store the next page token only after chat messages have been successfully persisted,
	// ensuring that no messages are lost.
	if err := lsr.progressRepo.Upsert(ctx, lsp); err != nil {
		return fmt.Errorf("upsert live stream progress: %v", err)
	}

	return nil
}

// tryLock attempts to acquire lock and returns true if succeeds, in any other case it returns false.
func (lsr *LiveStreamReader) tryLock(ctx context.Context, l *slog.Logger, liveStreamID string) bool {
	timeCtx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()

	ok, err := lsr.locker.Lock(timeCtx, liveStreamID)
	if err != nil {
		l.ErrorContext(ctx, "Failed to acquire lock", "err", err)

		return false
	}

	if !ok {
		l.DebugContext(ctx, "Locked")

		return false
	}

	l.DebugContext(ctx, "Lock acquired")

	return true
}

func (lsr *LiveStreamReader) release(ctx context.Context, log *slog.Logger, liveStreamID string) {
	timeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Millisecond*100)
	defer cancel()

	err := lsr.locker.Release(timeCtx, liveStreamID)
	if err != nil {
		log.ErrorContext(timeCtx, "Release lock failed", "err", err)
		return
	}

	log.DebugContext(timeCtx, "Release lock")
}

func oneOf(err error, errs ...error) bool {
	for _, e := range errs {
		if errors.Is(err, e) {
			return true
		}
	}

	return false
}
