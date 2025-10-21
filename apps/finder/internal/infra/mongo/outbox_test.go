//go:build integration

package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"
)

var dropOutboxCollFunc = func() {
	cancelCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_ = _mongoDB.Collection("outbox").Drop(cancelCtx)
}

func TestOutboxRepository_InsertLiveStreamsFound(t *testing.T) {
	t.Run("successfully inserts live streams into outbox", func(t *testing.T) {
		t.Cleanup(dropOutboxCollFunc)

		// When
		err := _outboxRepo.InsertLiveStreamsFound(context.Background(), newLiveStreams(t))

		// Then
		assert.NoError(t, err)

		// Verify the outbox collection has the expected documents
		cursor, err := _mongoDB.Collection("outbox").Find(context.Background(), bson.M{"topic": "live_stream_found"})
		require.NoError(t, err)

		var outboxEvents []bson.M
		err = cursor.All(context.Background(), &outboxEvents)
		require.NoError(t, err)

		assert.Len(t, outboxEvents, 2)
		assert.Equal(t, "id1", outboxEvents[0]["key"])
		assert.Equal(t, "id2", outboxEvents[1]["key"])
		assert.False(t, outboxEvents[0]["published"].(bool))
		assert.False(t, outboxEvents[1]["published"].(bool))
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		liveStreams := newLiveStreams(t)

		// When
		err := _outboxRepo.InsertLiveStreamsFound(ctx, liveStreams)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestOutboxRepository_Pending(t *testing.T) {
	t.Run("returns empty slice when no pending events exist", func(t *testing.T) {
		// When
		events, err := _outboxRepo.Pending(context.Background())

		// Then
		assert.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("returns pending outbox events", func(t *testing.T) {
		t.Cleanup(dropOutboxCollFunc)

		// Given
		require.NoError(t, _outboxRepo.InsertLiveStreamsFound(context.Background(), newLiveStreams(t)))

		// When
		events, err := _outboxRepo.Pending(context.Background())

		// Then
		assert.NoError(t, err)
		assert.Len(t, events, 2)
		assert.Equal(t, "id1", events[0].Key)
		assert.Equal(t, "live_stream_found", events[0].Topic)
		assert.False(t, events[0].Published)
		assert.JSONEq(t, `{"videoId" : "id1","channelId" : "channelId1","chatId" : "chatId1","title" : "title1","thumbnailUrl" : "thumbUrl1","publishedAt" : "2025-01-01T10:00:00Z","scheduledStart" : "2025-01-01T10:00:00Z"}`, string(events[0].Payload))
		assert.Equal(t, "id2", events[1].Key)
		assert.Equal(t, "live_stream_found", events[1].Topic)
		assert.False(t, events[1].Published)
		assert.JSONEq(t, `{"videoId" : "id2","channelId" : "channelId2","chatId" : "chatId2","title" : "title2","thumbnailUrl" : "thumbUrl2","publishedAt" : "2025-01-01T10:00:00Z","scheduledStart" : "2025-01-01T10:00:00Z"}`, string(events[1].Payload))
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		// When
		events, err := _outboxRepo.Pending(ctx)

		// Then
		assert.Nil(t, events)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestOutboxRepository_MarkAsPublished(t *testing.T) {
	t.Run("successfully marks events as published", func(t *testing.T) {
		t.Cleanup(dropOutboxCollFunc)

		// Given
		liveStreams := newLiveStreams(t)
		require.NoError(t, _outboxRepo.InsertLiveStreamsFound(context.Background(), liveStreams))

		events, err := _outboxRepo.Pending(context.Background())
		require.NoError(t, err)
		require.Len(t, events, 2)

		// When
		err = _outboxRepo.MarkAsPublished(context.Background(), events)

		// Then
		assert.NoError(t, err)
		events, err = _outboxRepo.Pending(context.Background())
		require.NoError(t, err)
		require.Len(t, events, 0)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		// Given
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context immediately

		// When
		err := _outboxRepo.MarkAsPublished(ctx, []kafka.OutboxEvent{
			{ID: primitive.NewObjectID().Hex()},
		})

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
