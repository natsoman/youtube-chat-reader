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

var dropBansCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("bans").Drop(cancelCtx)
}

func TestBanRepository_Insert(t *testing.T) {
	t.Run("successfully inserts new bans", func(t *testing.T) {
		t.Cleanup(dropBansCollFunc)

		// Given
		now := time.Now().UTC()
		ban1, err := domain.NewBan("ban1", "author1", "video1", "permanent", 0, now)
		require.NoError(t, err)
		ban2, err := domain.NewBan("ban2", "author2", "video1", "temporary", 5*time.Minute, now)
		require.NoError(t, err)

		// When
		err = _banRepo.Insert(t.Context(), []domain.Ban{*ban1, *ban2})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("bans")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("successfully ignores duplicates", func(t *testing.T) {
		t.Cleanup(dropBansCollFunc)

		// Given
		now := time.Now().UTC()
		ban1, err := domain.NewBan("ban1", "author1", "video1", "permanent", 0, now)
		require.NoError(t, err)
		ban2, err := domain.NewBan("ban2", "author2", "video1", "temporary", 5*time.Minute, now)
		require.NoError(t, err)
		require.NoError(t, _banRepo.Insert(t.Context(), []domain.Ban{*ban1, *ban2}))

		// When - try to insert the same bans again
		err = _banRepo.Insert(t.Context(), []domain.Ban{*ban1})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("bans")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		// When
		err := _banRepo.Insert(t.Context(), []domain.Ban{})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		ban, err := domain.NewBan("ban1", "author1", "video1", "permanent", 0, time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _banRepo.Insert(ctx, []domain.Ban{*ban})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
