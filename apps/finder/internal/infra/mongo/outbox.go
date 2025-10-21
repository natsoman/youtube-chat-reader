package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type OutboxRepository struct {
	writeColl *mongo.Collection
	readColl  *mongo.Collection

	topic string
}

func NewOutboxRepository(db *mongo.Database, topic string) (*OutboxRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	if topic == "" {
		return nil, errors.New("topic is empty")
	}

	const outboxCollName = "outbox"

	return &OutboxRepository{
		readColl: db.Collection(outboxCollName,
			options.Collection().
				SetReadPreference(readpref.PrimaryPreferred()).
				SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(outboxCollName,
			options.Collection().
				SetWriteConcern(writeconcern.Majority()),
		),
		topic: topic,
	}, nil
}

func (r *OutboxRepository) InsertLiveStreamsFound(ctx context.Context, liveStreams []domain.LiveStream) error {
	type eventPayload struct {
		VideoID        string    `json:"videoId"`
		ChannelID      string    `json:"channelId"`
		ChatID         string    `json:"chatId"`
		Title          string    `json:"title"`
		ThumbnailURL   string    `json:"thumbnailUrl"`
		PublishedAt    time.Time `json:"publishedAt"`
		ScheduledStart time.Time `json:"scheduledStart"`
	}

	docs := make([]interface{}, len(liveStreams))
	for i, liveStream := range liveStreams {
		payload, err := json.Marshal(eventPayload{
			VideoID:        liveStream.ID(),
			ChannelID:      liveStream.ChannelID(),
			ChatID:         liveStream.ChatID(),
			Title:          liveStream.Title(),
			ThumbnailURL:   liveStream.ThumbnailURL(),
			PublishedAt:    liveStream.PublishedAt(),
			ScheduledStart: liveStream.ScheduledStart(),
		})
		if err != nil {
			return fmt.Errorf("marshal: %v", err)
		}

		docs[i] = outboxEvent{
			ID:      primitive.NewObjectID(),
			Topic:   r.topic,
			Key:     liveStream.ID(),
			Payload: payload,
		}
	}

	_, err := r.writeColl.InsertMany(ctx, docs)
	if err != nil {
		return err
	}

	return nil
}

func (r *OutboxRepository) Pending(ctx context.Context) ([]kafka.OutboxEvent, error) {
	cur, err := r.readColl.Find(ctx, bson.M{
		"$or": []bson.M{
			{"published": nil},
			{"published": false},
		},
	})
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = cur.Close(ctx)
	}()

	var pendingEvents []outboxEvent
	if err = cur.All(ctx, &pendingEvents); err != nil {
		return nil, err
	}

	kafkaOutboxEvents := make([]kafka.OutboxEvent, len(pendingEvents))
	for i, pendingEvent := range pendingEvents {
		kafkaOutboxEvents[i] = kafka.OutboxEvent{
			ID:        pendingEvent.ID.Hex(),
			Key:       pendingEvent.Key,
			Topic:     pendingEvent.Topic,
			Payload:   pendingEvent.Payload,
			Published: pendingEvent.Published,
		}
	}

	return kafkaOutboxEvents, err
}

func (r *OutboxRepository) MarkAsPublished(ctx context.Context, events []kafka.OutboxEvent) error {
	ids := make([]primitive.ObjectID, len(events))
	for i, e := range events {
		id, err := primitive.ObjectIDFromHex(e.ID)
		if err != nil {
			return err
		}

		ids[i] = id
	}

	filter := bson.M{"_id": bson.M{"$in": ids}}
	update := bson.M{"$set": bson.M{"published": true}}

	_, err := r.writeColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	return nil
}

type outboxEvent struct {
	ID        primitive.ObjectID `bson:"_id"`
	Topic     string             `bson:"topic"`
	Key       string             `bson:"key"`
	Payload   []byte             `bson:"payload"`
	Published bool               `bson:"published"`
}
