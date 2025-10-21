//go:build integration

package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

var dropLiveStreamProgressCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("liveStreamProgress").Drop(cancelCtx)
}

func TestLiveStreamProgressRepository_Insert(t *testing.T) {
	t.Run("successfully inserts new live stream progress", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _liveStreamProgressRepo.Insert(t.Context(), lsp)

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("liveStreamProgress")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("ignores duplicates", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp))

		// When
		err = _liveStreamProgressRepo.Insert(t.Context(), lsp)

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _liveStreamProgressRepo.Insert(ctx, lsp)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestLiveStreamProgressRepository_Started(t *testing.T) {
	t.Run("returns empty slice when no live streams exist", func(t *testing.T) {
		// When
		started, err := _liveStreamProgressRepo.Started(t.Context(), time.Minute)

		// Then
		assert.NoError(t, err)
		assert.Empty(t, started)
	})

	t.Run("returns only live streams that start within duration", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		now := time.Now().UTC()
		lsp1, err := domain.NewLiveStreamProgress("videoId1", "chatId1", now.Add(-time.Minute))
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp1))

		lsp2, err := domain.NewLiveStreamProgress("videoId2", "chatId2", now.Add(30*time.Second))
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp2))

		lsp3, err := domain.NewLiveStreamProgress("videoId3", "chatId3", now.Add(2*time.Hour))
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp3))

		// When
		started, err := _liveStreamProgressRepo.Started(t.Context(), time.Minute)

		// Then
		assert.NoError(t, err)
		assert.Len(t, started, 2)
		ids := []string{started[0].ID(), started[1].ID()}
		assert.ElementsMatch(t, []string{"videoId1", "videoId2"}, ids)
	})

	t.Run("excludes finished live streams", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		now := time.Now().UTC()
		lsp1, err := domain.NewLiveStreamProgress("videoId1", "chatId1", now.Add(-time.Minute))
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp1))

		lsp2, err := domain.NewLiveStreamProgress("videoId2", "chatId2", now.Add(-time.Minute))
		require.NoError(t, err)
		lsp2.Finish(now, "test reason")
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp2))

		// When
		started, err := _liveStreamProgressRepo.Started(t.Context(), time.Minute)

		// Then
		assert.NoError(t, err)
		assert.Len(t, started, 1)
		assert.Equal(t, "videoId1", started[0].ID())
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		// When
		started, err := _liveStreamProgressRepo.Started(ctx, time.Minute)

		// Then
		assert.Nil(t, started)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestLiveStreamProgressRepository_Save(t *testing.T) {
	t.Run("successfully updates existing live stream progress", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp))

		// When
		lsp.SetNextPageToken("newToken")
		err = _liveStreamProgressRepo.Upsert(t.Context(), lsp)

		// Then
		assert.NoError(t, err)

		// Verify the update
		started, err := _liveStreamProgressRepo.Started(t.Context(), time.Hour)
		require.NoError(t, err)
		require.Len(t, started, 1)
		assert.Equal(t, "newToken", started[0].NextPageToken())
	})

	t.Run("successfully marks live stream as finished", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)
		require.NoError(t, _liveStreamProgressRepo.Insert(t.Context(), lsp))

		// When
		finishTime := time.Now().UTC()
		lsp.Finish(finishTime, "stream ended")
		err = _liveStreamProgressRepo.Upsert(t.Context(), lsp)

		// Then
		assert.NoError(t, err)

		// Verify the live stream is no longer returned by Started
		started, err := _liveStreamProgressRepo.Started(t.Context(), time.Hour)
		require.NoError(t, err)
		assert.Empty(t, started)
	})

	t.Run("successfully upserts non-existing live stream progress", func(t *testing.T) {
		t.Cleanup(dropLiveStreamProgressCollFunc)

		// Given
		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)
		lsp.SetNextPageToken("token")

		// When
		err = _liveStreamProgressRepo.Upsert(t.Context(), lsp)

		// Then
		assert.NoError(t, err)

		// Verify the upsert
		collection := _mongoDB.Collection("liveStreamProgress")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		lsp, err := domain.NewLiveStreamProgress("videoId1", "chatId1", time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _liveStreamProgressRepo.Upsert(ctx, lsp)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
