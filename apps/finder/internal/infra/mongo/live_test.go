//go:build integration

package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

var dropLiveStreamsCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("liveStreams").Drop(cancelCtx)
}

func TestLiveStreamRepository_Insert(t *testing.T) {
	t.Run("successfully inserts new live streams", func(t *testing.T) {
		t.Cleanup(dropLiveStreamsCollFunc)

		// Given
		liveStreams := newLiveStreams(t)

		// When
		err := _liveStreamRepo.Insert(t.Context(), liveStreams)

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("liveStreams")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("successfully ignores duplicates", func(t *testing.T) {
		t.Cleanup(dropLiveStreamsCollFunc)

		// Given
		liveStreams := newLiveStreams(t)
		require.NoError(t, _liveStreamRepo.Insert(t.Context(), liveStreams))

		// When
		err := _liveStreamRepo.Insert(t.Context(), liveStreams[:1])

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("liveStreams")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		// When
		err := _liveStreamRepo.Insert(ctx, newLiveStreams(t))

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestLiveStreamRepository_Existing(t *testing.T) {
	t.Run("returns empty slice when no matching IDs exist", func(t *testing.T) {
		// When
		existing, err := _liveStreamRepo.Existing(t.Context(), []string{"nonexistent1", "nonexistent2"})

		// Then
		assert.NoError(t, err)
		assert.Empty(t, existing)
	})

	t.Run("returns only existing stream IDs", func(t *testing.T) {
		t.Cleanup(dropLiveStreamsCollFunc)

		// Given
		liveStreams := newLiveStreams(t)
		require.NoError(t, _liveStreamRepo.Insert(t.Context(), liveStreams))

		// When
		existing, err := _liveStreamRepo.Existing(t.Context(), []string{liveStreams[0].ID(), liveStreams[1].ID(), "4"})

		// Then
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{liveStreams[0].ID(), liveStreams[1].ID()}, existing)
	})

	t.Run("handles empty input slice", func(t *testing.T) {
		// When
		existing, err := _liveStreamRepo.Existing(t.Context(), []string{})

		// Then
		assert.NoError(t, err)
		assert.Empty(t, existing)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		// When
		existing, err := _liveStreamRepo.Existing(ctx, []string{"test1"})

		// Then
		assert.Nil(t, existing)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
