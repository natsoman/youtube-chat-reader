//go:generate mockgen -destination=mock_test.go -package=app_test -source=find.go
package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/app"
	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

func TestNewLiveStreamFinder(t *testing.T) {
	t.Parallel()

	t.Run("successful creation", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		finder, err := app.NewLiveStreamFinder(
			NewMockYoutubeClient(ctrl),
			NewMockRepository(ctrl),
			NewMockOutbox(ctrl),
			NewMockTransactor(ctrl),
		)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, finder)
	})

	t.Run("nil youtube client", func(t *testing.T) {
		t.Parallel()

		// When
		finder, err := app.NewLiveStreamFinder(nil, nil, nil, nil)

		// Then
		assert.ErrorContains(t, err, "youtube client is nil")
		assert.Nil(t, finder)
	})

	t.Run("nil repository", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		finder, err := app.NewLiveStreamFinder(
			NewMockYoutubeClient(ctrl),
			nil,
			NewMockOutbox(ctrl),
			NewMockTransactor(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "repository is nil")
		assert.Nil(t, finder)
	})

	t.Run("nil outbox writer", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		finder, err := app.NewLiveStreamFinder(
			NewMockYoutubeClient(ctrl),
			NewMockRepository(ctrl),
			nil,
			NewMockTransactor(ctrl),
		)

		// Then
		assert.ErrorContains(t, err, "outbox writer is nil")
		assert.Nil(t, finder)
	})

	t.Run("nil transactor", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// When
		finder, err := app.NewLiveStreamFinder(
			NewMockYoutubeClient(ctrl),
			NewMockRepository(ctrl),
			NewMockOutbox(ctrl),
			nil,
		)

		// Then
		assert.ErrorContains(t, err, "transactor is nil")
		assert.Nil(t, finder)
	})
}

func TestLiveStreamFinder_Find(t *testing.T) {
	t.Parallel()

	t.Run("returns no error when no channel IDs provided", func(t *testing.T) {
		finder, _ := setupTest(t)

		// When
		err := finder.Find(t.Context(), []string{})

		// Then
		assert.NoError(t, err)
	})

	t.Run("finds and stores new live streams successfully for multiple channels", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		liveStreams := newLiveStreams(t)
		gomock.InOrder(
			deps.youtubeClient.EXPECT().
				SearchUpcomingLiveStream(t.Context(), liveStreams[0].ChannelID()).
				Return([]string{liveStreams[0].ID()}, nil),
			deps.youtubeClient.EXPECT().
				SearchUpcomingLiveStream(t.Context(), liveStreams[1].ChannelID()).
				Return([]string{liveStreams[1].ID()}, nil),
			deps.repo.EXPECT().
				Existing(t.Context(), []string{liveStreams[0].ID(), liveStreams[1].ID()}),
			deps.youtubeClient.EXPECT().
				ListLiveStreams(t.Context(), []string{liveStreams[0].ID(), liveStreams[1].ID()}).
				Return(liveStreams, nil),
			deps.txn.EXPECT().
				Atomic(t.Context(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}),
			deps.repo.EXPECT().
				Insert(t.Context(), liveStreams),
			deps.outbox.EXPECT().
				InsertLiveStreamsFound(t.Context(), liveStreams),
		)

		// When
		err := finder.Find(t.Context(), []string{liveStreams[0].ChannelID(), liveStreams[1].ChannelID()})

		// Then
		assert.NoError(t, err)
	})

	t.Run("skips processing when all live streams already exist", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		channelID := "channel1"
		videoID := "video1"

		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), channelID).
			Return([]string{videoID}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{videoID}).
			Return([]string{videoID}, nil)

		// When
		err := finder.Find(t.Context(), []string{channelID})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when searching for upcoming live streams fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), "channel1").
			Return([]string{}, errors.New("search error"))

		// When
		err := finder.Find(t.Context(), []string{"channel1"})

		// Then
		assert.EqualError(t, err, "search upcoming live streams: search error")
	})

	t.Run("completes successfully when no new live streams found", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), "channelId1").
			Return([]string{"id1", "id2"}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{"id1", "id2"}).
			Return([]string{"id1", "id2"}, nil)

		// When
		err := finder.Find(t.Context(), []string{"channelId1"})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when transaction fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		liveStream := newLiveStreams(t)[0]

		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), liveStream.ChannelID()).
			Return([]string{liveStream.ID()}, nil)
		deps.youtubeClient.EXPECT().
			ListLiveStreams(t.Context(), []string{liveStream.ID()}).
			Return([]domain.LiveStream{liveStream}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{liveStream.ID()})
		deps.txn.EXPECT().
			Atomic(t.Context(), gomock.Any()).
			Return(errors.New("transaction failed"))

		// When
		err := finder.Find(t.Context(), []string{liveStream.ChannelID()})

		// Then
		assert.EqualError(t, err, "transaction failed")
	})

	t.Run("returns error when checking existing live streams fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		channelID := "channel1"
		videoID := "video1"

		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), channelID).
			Return([]string{videoID}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{videoID}).
			Return(nil, errors.New("database error"))

		// When
		err := finder.Find(t.Context(), []string{channelID})

		// Then
		assert.EqualError(t, err, "list existing live streams: database error")
	})

	t.Run("returns error when listing youtube live streams fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		channelID := "channel1"
		videoID := "video1"

		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), channelID).
			Return([]string{videoID}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{videoID}).
			Return([]string{}, nil)
		deps.youtubeClient.EXPECT().
			ListLiveStreams(t.Context(), []string{videoID}).
			Return(nil, errors.New("youtube api error"))

		// When
		err := finder.Find(t.Context(), []string{channelID})

		// Then
		assert.EqualError(t, err, "list youtube live streams: youtube api error")
	})

	t.Run("returns error when inserting into repository fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		liveStream := newLiveStreams(t)[0]
		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), liveStream.ChannelID()).
			Return([]string{liveStream.ID()}, nil)
		deps.youtubeClient.EXPECT().
			ListLiveStreams(t.Context(), []string{liveStream.ID()}).
			Return([]domain.LiveStream{liveStream}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{liveStream.ID()})
		deps.repo.EXPECT().
			Insert(t.Context(), []domain.LiveStream{liveStream}).
			Return(errors.New("insert error"))
		deps.txn.EXPECT().
			Atomic(t.Context(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		// When
		err := finder.Find(t.Context(), []string{liveStream.ChannelID()})

		// Then
		assert.EqualError(t, err, "insert to repo: insert error")
	})

	t.Run("returns error when inserting into outbox fails", func(t *testing.T) {
		finder, deps := setupTest(t)

		// Given
		liveStream := newLiveStreams(t)[0]
		deps.youtubeClient.EXPECT().
			SearchUpcomingLiveStream(t.Context(), liveStream.ChannelID()).
			Return([]string{liveStream.ID()}, nil)
		deps.youtubeClient.EXPECT().
			ListLiveStreams(t.Context(), []string{liveStream.ID()}).
			Return([]domain.LiveStream{liveStream}, nil)
		deps.repo.EXPECT().
			Existing(t.Context(), []string{liveStream.ID()})
		deps.repo.EXPECT().
			Insert(t.Context(), []domain.LiveStream{liveStream})
		deps.outbox.EXPECT().
			InsertLiveStreamsFound(t.Context(), []domain.LiveStream{liveStream}).
			Return(errors.New("outbox error"))
		deps.txn.EXPECT().
			Atomic(t.Context(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
				return fn(ctx)
			})

		// When
		err := finder.Find(t.Context(), []string{liveStream.ChannelID()})

		// Then
		assert.EqualError(t, err, "insert to outbox: outbox error")
	})
}

type testDeps struct {
	youtubeClient *MockYoutubeClient
	repo          *MockRepository
	outbox        *MockOutbox
	txn           *MockTransactor
}

func setupTest(t *testing.T) (*app.LiveStreamFinder, *testDeps) {
	t.Helper()
	t.Parallel()

	ctrl := gomock.NewController(t)
	deps := &testDeps{
		youtubeClient: NewMockYoutubeClient(ctrl),
		repo:          NewMockRepository(ctrl),
		outbox:        NewMockOutbox(ctrl),
		txn:           NewMockTransactor(ctrl),
	}

	finder, err := app.NewLiveStreamFinder(
		deps.youtubeClient,
		deps.repo,
		deps.outbox,
		deps.txn,
	)
	require.NoError(t, err)

	return finder, deps
}

func newLiveStreams(t *testing.T) []domain.LiveStream {
	t.Helper()

	now := time.Now()

	ls1, err := domain.NewLiveStream("id1", "title1", "channelId1", "chanTitle1", "thumbUrl1", "chatId1", now, now)
	require.NoError(t, err)

	ls2, err := domain.NewLiveStream("id2", "title2", "channelId1", "chanTitle2", "thumbUrl2", "chatId2", now, now)
	require.NoError(t, err)

	return []domain.LiveStream{*ls1, *ls2}
}
