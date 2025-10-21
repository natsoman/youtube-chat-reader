package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

type LiveStreamRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewLiveStreamRepository(db *mongo.Database) (*LiveStreamRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const liveStreamsCollName = "liveStreams"

	return &LiveStreamRepository{
		readColl: db.Collection(liveStreamsCollName,
			options.Collection().
				SetReadPreference(readpref.PrimaryPreferred()).
				SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(liveStreamsCollName,
			options.Collection().
				SetWriteConcern(writeconcern.Majority())),
	}, nil
}

func (r *LiveStreamRepository) Insert(ctx context.Context, liveStreams []domain.LiveStream) error {
	docs := []interface{}{}
	for _, ls := range liveStreams {
		docs = append(docs, newLiveStreamDoc(&ls))
	}

	_, err := r.writeColl.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	if err != nil {
		var me mongo.BulkWriteException
		if errors.As(err, &me) {
			for _, e := range me.WriteErrors {
				if e.Code == 11000 { // duplicate key error
					continue
				}

				return err
			}
		} else {
			return err
		}
	}

	return nil
}

func (r *LiveStreamRepository) Existing(ctx context.Context, liveStreamIDs []string) ([]string, error) {
	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: liveStreamIDs}}}}

	values, err := r.readColl.Distinct(ctx, "_id", filter)
	if err != nil {
		return nil, err
	}

	existingIDs := make([]string, len(values))
	for i, id := range values {
		existingIDs[i] = id.(string)
	}

	return existingIDs, nil
}

type liveStreamDoc struct {
	ID             string    `bson:"_id"`
	ChannelID      string    `bson:"channelId"`
	Title          string    `bson:"title"`
	ThumbnailURL   string    `bson:"thumbnailUrl"`
	PublishedAt    time.Time `bson:"publishedAt"`
	ChatID         string    `bson:"chatId"`
	ScheduledStart time.Time `bson:"scheduledStart"`
}

func newLiveStreamDoc(ls *domain.LiveStream) liveStreamDoc {
	return liveStreamDoc{
		ID:             ls.ID(),
		ChannelID:      ls.ChannelID(),
		Title:          ls.Title(),
		ThumbnailURL:   ls.ThumbnailURL(),
		PublishedAt:    ls.PublishedAt(),
		ChatID:         ls.ChatID(),
		ScheduledStart: ls.ScheduledStart(),
	}
}
