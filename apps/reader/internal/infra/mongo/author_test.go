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

var dropAuthorsCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("authors").Drop(cancelCtx)
}

func TestAuthorRepository_Upsert(t *testing.T) {
	t.Run("successfully inserts new authors", func(t *testing.T) {
		t.Cleanup(dropAuthorsCollFunc)

		// Given
		author1, err := domain.NewAuthor("author1", "Author One", "https://example.com/1.jpg", true)
		require.NoError(t, err)
		author2, err := domain.NewAuthor("author2", "Author Two", "https://example.com/2.jpg", false)
		require.NoError(t, err)

		// When
		err = _authorRepo.Upsert(t.Context(), []domain.Author{*author1, *author2})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("authors")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("successfully updates existing authors", func(t *testing.T) {
		t.Cleanup(dropAuthorsCollFunc)

		// Given
		author1, err := domain.NewAuthor("author1", "Author One", "https://example.com/1.jpg", true)
		require.NoError(t, err)
		require.NoError(t, _authorRepo.Upsert(t.Context(), []domain.Author{*author1}))

		// When - update the same author with different data
		updatedAuthor, err := domain.NewAuthor("author1", "Updated Name", "https://example.com/updated.jpg", false)
		require.NoError(t, err)
		err = _authorRepo.Upsert(t.Context(), []domain.Author{*updatedAuthor})

		// Then
		assert.NoError(t, err)
		collection := _mongoDB.Collection("authors")
		count, err := collection.CountDocuments(t.Context(), bson.M{})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)

		// Verify the author was updated
		var doc bson.M
		err = collection.FindOne(t.Context(), bson.M{"_id": "author1"}).Decode(&doc)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", doc["name"])
		assert.Equal(t, "https://example.com/updated.jpg", doc["profileImageUrl"])
		assert.False(t, doc["isVerified"].(bool))
	})

	t.Run("handles empty slice", func(t *testing.T) {
		// When
		err := _authorRepo.Upsert(t.Context(), []domain.Author{})

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		author, err := domain.NewAuthor("author1", "Author One", "https://example.com/1.jpg", true)
		require.NoError(t, err)

		// When
		err = _authorRepo.Upsert(ctx, []domain.Author{*author})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
