package app

import (
	"context"
	"fmt"
	"slices"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type Transactor interface {
	// Atomic executes all operations of Outbox and Repository, or none at all.
	Atomic(ctx context.Context, fn func(txnCtx context.Context) error) error
}

type Outbox interface {
	InsertLiveStreamsFound(ctx context.Context, liveStreams []domain.LiveStream) error
}

type Repository interface {
	// Insert persists the given domain.LiveStream(s). Existing live streams are skipped.
	Insert(ctx context.Context, liveStreams []domain.LiveStream) error
	// Existing returns which of the given live streams ids are already persisted.
	Existing(ctx context.Context, liveStreamIDs []string) ([]string, error)
}

type YoutubeClient interface {
	// SearchUpcomingLiveStream returns identifiers of upcoming live streams for the specified channel.
	SearchUpcomingLiveStream(ctx context.Context, channelID string) ([]string, error)
	// ListLiveStreams returns live streams associated with the given videoIDs
	// that have chat enabled and have not finished.
	ListLiveStreams(ctx context.Context, videoIDs []string) ([]domain.LiveStream, error)
}

type LiveStreamFinder struct {
	youtubeClient YoutubeClient
	repo          Repository
	outbox        Outbox
	txn           Transactor
}

func NewLiveStreamFinder(
	youtubeClient YoutubeClient,
	repo Repository,
	outbox Outbox,
	txn Transactor,
) (*LiveStreamFinder, error) {
	if youtubeClient == nil {
		return nil, fmt.Errorf("youtube client is nil")
	}

	if repo == nil {
		return nil, fmt.Errorf("repository is nil")
	}

	if outbox == nil {
		return nil, fmt.Errorf("outbox writer is nil")
	}

	if txn == nil {
		return nil, fmt.Errorf("transactor is nil")
	}

	return &LiveStreamFinder{youtubeClient: youtubeClient,
		repo:   repo,
		outbox: outbox,
		txn:    txn,
	}, nil
}

// Find discovers upcoming domain.LiveStream(s) for the provided YouTube channels.
// Live streams are persisted to the Repository along with the Outbox.
func (f *LiveStreamFinder) Find(ctx context.Context, channelIDs []string) error {
	var foundedLiveStreamIDs []string

	for _, channelID := range channelIDs {
		channelFoundedLiveStreamIDs, err := f.youtubeClient.SearchUpcomingLiveStream(ctx, channelID)
		if err != nil {
			return fmt.Errorf("search upcoming live streams: %v", err)
		}

		foundedLiveStreamIDs = append(foundedLiveStreamIDs, channelFoundedLiveStreamIDs...)
	}

	if len(foundedLiveStreamIDs) == 0 {
		return nil
	}

	existingLiveStreamIDs, err := f.repo.Existing(ctx, foundedLiveStreamIDs)
	if err != nil {
		return fmt.Errorf("list existing live streams: %v", err)
	}

	var newLiveStreamIDs []string

	for _, id := range foundedLiveStreamIDs {
		if !slices.Contains(existingLiveStreamIDs, id) {
			newLiveStreamIDs = append(newLiveStreamIDs, id)
		}
	}

	if len(newLiveStreamIDs) == 0 {
		return nil
	}

	liveStreams, err := f.youtubeClient.ListLiveStreams(ctx, newLiveStreamIDs)
	if err != nil {
		return fmt.Errorf("list youtube live streams: %v", err)
	}

	return f.txn.Atomic(ctx, func(txCtx context.Context) error {
		if err = f.repo.Insert(txCtx, liveStreams); err != nil {
			return fmt.Errorf("insert to repo: %v", err)
		}

		if err = f.outbox.InsertLiveStreamsFound(txCtx, liveStreams); err != nil {
			return fmt.Errorf("insert to outbox: %v", err)
		}

		return nil
	})
}
