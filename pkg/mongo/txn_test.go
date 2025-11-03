//go:build integration

package mongo_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	pkgmongo "github.com/natsoman/youtube-chat-reader/pkg/mongo"
)

func TestTransactor_Atomic(t *testing.T) {
	t.Parallel()

	t.Run("both operations are applied successfully", func(t *testing.T) {
		collection := _mongoDB.Collection("transactorAtomic1")

		txn, err := pkgmongo.NewTransactor(_mongoDB.Client())
		require.NotNil(t, txn)
		require.NoError(t, err)

		// When two successfully inserts are executed
		err = txn.Atomic(t.Context(), func(ctx context.Context) error {
			_, err = collection.InsertOne(ctx, bson.D{{"key", "value"}})
			if err != nil {
				return err
			}

			_, err = collection.InsertOne(ctx, bson.D{{"key", "value"}})
			if err != nil {
				return err
			}

			return nil
		})

		// Then two documents must be existed
		assert.NoError(t, err)
		count, err := collection.CountDocuments(t.Context(), bson.D{{"key", "value"}})
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("none operation is applied when at least one fails", func(t *testing.T) {
		collection := _mongoDB.Collection("transactorAtomic2")
		require.NotNil(t, collection)

		txn, err := pkgmongo.NewTransactor(_mongoDB.Client())
		require.NotNil(t, txn)
		require.NoError(t, err)

		// Given a unique index
		_, err = collection.InsertOne(t.Context(), bson.D{{"uniqueKey", "value"}})
		require.NoError(t, err)
		_, err = collection.Indexes().CreateOne(t.Context(), mongo.IndexModel{
			Keys:    bson.D{{"uniqueKey", 1}},
			Options: options.Index().SetUnique(true),
		})
		require.NoError(t, err)

		// When one of the operations conflicts due to unique index
		err = txn.Atomic(t.Context(), func(ctx context.Context) error {
			_, err = collection.InsertOne(ctx, bson.D{{"key", "value"}})
			if err != nil {
				return err
			}

			_, err = collection.InsertOne(ctx, bson.D{{"uniqueKey", "value"}})
			if err != nil {
				return err
			}

			return nil
		})

		// Then an error is returned and no documents were written
		assert.Error(t, err)
		count, err := collection.CountDocuments(t.Context(), bson.D{{"key", "value"}})
		require.NoError(t, err)
		assert.Zero(t, count)
	})
}
