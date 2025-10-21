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

var dropDonatesCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("donates").Drop(cancelCtx)
}

func TestDonateRepository_Insert(t *testing.T) {
	t.Run("successfully inserts new donates", func(t *testing.T) {
		t.Cleanup(dropDonatesCollFunc)

		// Given
		now := time.Now().UTC()
		donate1, err := domain.NewDonate("donate1", "author1", "video1", "Great stream!", "$10.00", 10000000, "USD", now)
		require.NoError(t, err)
		donate2, err := domain.NewDonate("donate2", "author2", "video1", "Keep it up!", "$5.00", 5000000, "USD", now)
		require.NoError(t, err)

		// When
		err = _donateRepo.Insert(t.Context(), []domain.Donate{*donate1, *donate2})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("donates")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("successfully ignores duplicates", func(t *testing.T) {
		t.Cleanup(dropDonatesCollFunc)

		// Given
		now := time.Now().UTC()
		donate1, err := domain.NewDonate("donate1", "author1", "video1", "Great stream!", "$10.00", 10000000, "USD", now)
		require.NoError(t, err)
		donate2, err := domain.NewDonate("donate2", "author2", "video1", "Keep it up!", "$5.00", 5000000, "USD", now)
		require.NoError(t, err)
		require.NoError(t, _donateRepo.Insert(t.Context(), []domain.Donate{*donate1, *donate2}))

		// When - try to insert the same donates again
		err = _donateRepo.Insert(t.Context(), []domain.Donate{*donate1})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("donates")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		// When
		err := _donateRepo.Insert(t.Context(), []domain.Donate{})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		donate, err := domain.NewDonate("donate1", "author1", "video1", "Great stream!", "$10.00", 10000000, "USD", time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _donateRepo.Insert(ctx, []domain.Donate{*donate})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
