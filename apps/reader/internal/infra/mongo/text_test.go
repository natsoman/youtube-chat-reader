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

var dropTextsCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("texts").Drop(cancelCtx)
}

func TestTextMessageRepository_Insert(t *testing.T) {
	t.Run("successfully inserts new text messages", func(t *testing.T) {
		t.Cleanup(dropTextsCollFunc)

		// Given
		now := time.Now().UTC()
		text1, err := domain.NewTextMessage("text1", "video1", "author1", "Hello world!", now)
		require.NoError(t, err)
		text2, err := domain.NewTextMessage("text2", "video1", "author2", "Great content!", now)
		require.NoError(t, err)

		// When
		err = _textMessageRepo.Insert(t.Context(), []domain.TextMessage{*text1, *text2})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("texts")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("successfully ignores duplicates", func(t *testing.T) {
		t.Cleanup(dropTextsCollFunc)

		// Given
		now := time.Now().UTC()
		text1, err := domain.NewTextMessage("text1", "video1", "author1", "Hello world!", now)
		require.NoError(t, err)
		text2, err := domain.NewTextMessage("text2", "video1", "author2", "Great content!", now)
		require.NoError(t, err)
		require.NoError(t, _textMessageRepo.Insert(t.Context(), []domain.TextMessage{*text1, *text2}))

		// When - try to insert the same text messages again
		err = _textMessageRepo.Insert(t.Context(), []domain.TextMessage{*text1})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("texts")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("handles empty slice", func(t *testing.T) {
		// When
		err := _textMessageRepo.Insert(t.Context(), []domain.TextMessage{})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		text, err := domain.NewTextMessage("text1", "video1", "author1", "Hello world!", time.Now().UTC())
		require.NoError(t, err)

		// When
		err = _textMessageRepo.Insert(ctx, []domain.TextMessage{*text})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
