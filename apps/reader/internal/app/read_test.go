//go:generate mockgen -destination=mock_test.go -package=app_test -source=read.go
package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/app"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewLiveStreamReader(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, reader)
	})

	t.Run("nil clock", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			nil, // nil clock
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "clock is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil ticker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			nil, // nil ticker
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "ticker is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil locker", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			nil, // nil locker
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "locker is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil chat message streamer", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			nil, // nil chat message streamer
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "chat message streamer is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil progress repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			nil, // nil progress repository
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "progress repository is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil ban repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			nil, // nil ban repository
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "ban repository is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil text message repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			nil, // nil text message repository
			NewMockDonateRepository(ctrl),
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "text message repository is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil donate repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			nil, // nil donate repository
			NewMockAuthorRepository(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "donate repository is nil")
		assert.Nil(t, reader)
	})

	t.Run("nil author repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		reader, err := app.NewLiveStreamReader(
			NewMockClock(ctrl),
			NewMockTicker(ctrl),
			NewMockLocker(ctrl),
			NewMockChatMessageStreamer(ctrl),
			NewMockLiveStreamProgressRepository(ctrl),
			NewMockBanRepository(ctrl),
			NewMockTextMessageRepository(ctrl),
			NewMockDonateRepository(ctrl),
			nil, // nil author repository
		)

		// Then
		assert.ErrorContains(t, err, "author repository is nil")
		assert.Nil(t, reader)
	})
}

func TestLiveStreamReader_Read(t *testing.T) {
	t.Parallel()

	const timeout = time.Millisecond * 10

	t.Run("successfully store chat messages and updates next page token", func(t *testing.T) {
		reader, deps := setupTest(t, app.WithAdvanceStart(time.Minute))

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		lspWithUpdatedNextPageToken := lsp
		lspWithUpdatedNextPageToken.SetNextPageToken("nextPageToken")

		cm := newChatMessages(t, "nextPageToken")
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(time.Second*10).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), time.Minute).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), "id").
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), lsp).
				Return(cmChan, nil),
			deps.progressRepo.EXPECT().
				Upsert(gomock.Any(), lspWithUpdatedNextPageToken).
				After(deps.donateRepo.EXPECT().
					Insert(gomock.Any(), cm.Donates())).
				After(deps.textRepo.EXPECT().
					Insert(gomock.Any(), cm.TextMessages())).
				After(deps.banRepo.EXPECT().
					Insert(gomock.Any(), cm.Bans())).
				After(deps.authorRepo.EXPECT().
					Upsert(gomock.Any(), cm.Authors())),
			deps.locker.EXPECT().
				Release(gomock.Any(), "id").
				DoAndReturn(func(_ context.Context, _ string) error {
					cancel()
					return nil
				}),
		)

		// When
		go func() {
			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	t.Run("successfully store chat messages and finishes progress on empty next page token", func(t *testing.T) {
		reader, deps := setupTest(t, app.WithRetryInterval(time.Second*10), app.WithAdvanceStart(time.Minute))

		ctx, cancel := context.WithCancel(t.Context())

		// Given
		now := time.Now()
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		finished := lsp
		finished.Finish(now, "empty next page token")

		cm := newChatMessages(t, "")
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(time.Second*10).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), time.Minute).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), "id").
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), lsp).
				Return(cmChan, nil),
			deps.clock.EXPECT().
				Now().
				Return(now),
			deps.progressRepo.EXPECT().
				Upsert(gomock.Any(), finished).
				After(deps.donateRepo.EXPECT().
					Insert(gomock.Any(), cm.Donates())).
				After(deps.textRepo.EXPECT().
					Insert(gomock.Any(), cm.TextMessages())).
				After(deps.banRepo.EXPECT().
					Insert(gomock.Any(), cm.Bans())).
				After(deps.authorRepo.EXPECT().
					Upsert(gomock.Any(), cm.Authors())),
			deps.locker.EXPECT().
				Release(gomock.Any(), "id").
				DoAndReturn(func(_ context.Context, _ string) error {
					cancel()
					return nil
				}),
		)

		// When
		go func() {
			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	//nolint:dupl
	t.Run("handles error when upserting authors fails", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)
		deps.authorRepo.EXPECT().
			Upsert(gomock.Any(), gomock.Any()).
			Return(errors.New("error"))

		// When
		go func() {
			cm := *domain.NewChatMessages("nextPageToken")
			cm.AddAuthor(&domain.Author{})

			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	//nolint:dupl
	t.Run("handles error when inserting donates fails", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)
		deps.donateRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(errors.New("error"))

		// When
		go func() {
			cm := *domain.NewChatMessages("nextPageToken")
			cm.AddDonate(&domain.Donate{})

			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	//nolint:dupl
	t.Run("handles error when inserting text messages fails", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)
		deps.textRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(errors.New("error"))

		// When
		go func() {
			cm := *domain.NewChatMessages("nextPageToken")
			cm.AddTextMessage(&domain.TextMessage{})

			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	//nolint:dupl
	t.Run("handles error when inserting bans fails", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)
		deps.banRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(errors.New("error"))

		// When
		go func() {
			cm := *domain.NewChatMessages("nextPageToken")
			cm.AddBan(&domain.Ban{})

			cmChan <- cm
		}()

		reader.Read(ctx)
	})

	t.Run("handles error on live stream progress save", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.progressRepo.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				Return(errors.New("error")),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			cmChan <- *domain.NewChatMessages("nextPageToken")
		}()

		reader.Read(ctx)
	})

	t.Run("handles chat message streaming error", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		tickChan := make(chan time.Time)
		errChan := make(chan error)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(nil, errChan),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			errChan <- fmt.Errorf("error")
		}()

		reader.Read(ctx)
	})

	t.Run("successfully finishes when live stream becomes unavailable", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		now := time.Now().UTC()
		tickChan := make(chan time.Time)
		errChan := make(chan error)
		lspFinished := lsp
		lspFinished.Finish(now, domain.ErrUnavailableLiveStream.Error())

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(nil, errChan),
			deps.clock.EXPECT().
				Now().
				Return(now),
			deps.progressRepo.EXPECT().
				Upsert(gomock.Any(), lspFinished),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			errChan <- domain.ErrUnavailableLiveStream
		}()

		reader.Read(ctx)
	})

	t.Run("unsuccessfully finishes when live stream becomes unavailable", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		now := time.Now().UTC()
		tickChan := make(chan time.Time)
		errChan := make(chan error)
		lspFinished := lsp
		lspFinished.Finish(now, domain.ErrUnavailableLiveStream.Error())

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(nil, errChan),
			deps.clock.EXPECT().
				Now().
				Return(now),
			deps.progressRepo.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				Return(errors.New("error")),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			errChan <- domain.ErrUnavailableLiveStream
		}()

		reader.Read(ctx)
	})

	t.Run("handles error when fetching started progress fails", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{}, errors.New("error")),
		)

		// When
		reader.Read(ctx)
	})

	t.Run("handles case when no live streams are in progress", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		reader, deps := setupTest(t)

		// Given
		tickChan := make(chan time.Time)
		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{}, nil),
		)

		// When
		reader.Read(ctx)
	})

	t.Run("error lock cannot be acquired", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		tickChan := make(chan time.Time)
		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(false, errors.New("lock failed")),
		)

		// When
		reader.Read(ctx)
	})

	t.Run("lock is already acquired acquired", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
		require.NoError(t, err)

		tickChan := make(chan time.Time)
		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{*lsp}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(false, nil),
		)

		// When
		reader.Read(ctx)
	})

	t.Run("streaming error channel is closed successfully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		reader, deps := setupTest(t)

		// Given
		tickChan := make(chan time.Time)
		errChan := make(chan error)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(nil, errChan),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			close(errChan)
		}()

		reader.Read(ctx)
	})

	t.Run("streaming chat message channel is closed successfully", func(t *testing.T) {
		reader, deps := setupTest(t)

		ctx, cancel := context.WithTimeout(t.Context(), timeout)
		defer cancel()

		// Given
		tickChan := make(chan time.Time)
		cmChan := make(chan domain.ChatMessages)

		gomock.InOrder(
			deps.ticker.EXPECT().
				Start(gomock.Any()).
				Return(tickChan, func() {}),
			deps.progressRepo.EXPECT().
				Started(gomock.Any(), gomock.Any()).
				Return([]domain.LiveStreamProgress{{}}, nil),
			deps.locker.EXPECT().
				Lock(gomock.Any(), gomock.Any()).
				Return(true, nil),
			deps.cmStreamer.EXPECT().
				StreamChatMessages(gomock.Any(), gomock.Any()).
				Return(cmChan, nil),
			deps.locker.EXPECT().
				Release(gomock.Any(), gomock.Any()),
		)

		// When
		go func() {
			close(cmChan)
		}()

		reader.Read(ctx)
	})
}

type testDeps struct {
	clock        *MockClock
	ticker       *MockTicker
	locker       *MockLocker
	cmStreamer   *MockChatMessageStreamer
	banRepo      *MockBanRepository
	textRepo     *MockTextMessageRepository
	donateRepo   *MockDonateRepository
	authorRepo   *MockAuthorRepository
	progressRepo *MockLiveStreamProgressRepository
}

func setupTest(t *testing.T, o ...app.Option) (*app.LiveStreamReader, *testDeps) {
	t.Helper()

	ctrl := gomock.NewController(t)
	deps := &testDeps{
		clock:        NewMockClock(ctrl),
		ticker:       NewMockTicker(ctrl),
		locker:       NewMockLocker(ctrl),
		cmStreamer:   NewMockChatMessageStreamer(ctrl),
		banRepo:      NewMockBanRepository(ctrl),
		textRepo:     NewMockTextMessageRepository(ctrl),
		donateRepo:   NewMockDonateRepository(ctrl),
		authorRepo:   NewMockAuthorRepository(ctrl),
		progressRepo: NewMockLiveStreamProgressRepository(ctrl),
	}

	reader, err := app.NewLiveStreamReader(
		deps.clock,
		deps.ticker,
		deps.locker,
		deps.cmStreamer,
		deps.progressRepo,
		deps.banRepo,
		deps.textRepo,
		deps.donateRepo,
		deps.authorRepo,
		o...,
	)
	require.NoError(t, err)
	require.NotNil(t, reader)

	return reader, deps
}

func newChatMessages(t *testing.T, nextPageToken string) domain.ChatMessages {
	cm := domain.NewChatMessages(nextPageToken)

	author, err := domain.NewAuthor("id", "authorName", "profileImageUrl", true)
	require.NoError(t, err)

	cm.AddAuthor(author)

	ban, err := domain.NewBan("id", "authorId", "videoId", domain.Temporary.String(), time.Hour, time.Now().UTC())
	require.NoError(t, err)

	cm.AddBan(ban)

	textMsg, err := domain.NewTextMessage("id", "videoId", "authorId", "text", time.Now().UTC())
	require.NoError(t, err)

	cm.AddTextMessage(textMsg)

	donate, err := domain.NewDonate("id", "authorId", "videoId", "comment", "amount", 123, "euro", time.Now().UTC())
	require.NoError(t, err)

	cm.AddDonate(donate)

	return *cm
}
