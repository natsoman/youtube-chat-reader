package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type LiveStreamProgressRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewLiveStreamProgressRepository(db *mongo.Database) (*LiveStreamProgressRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const liveStreamProgressCollName = "liveStreamProgress"

	return &LiveStreamProgressRepository{
		readColl: db.Collection(liveStreamProgressCollName, options.Collection().
			SetReadPreference(readpref.SecondaryPreferred()).
			SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(liveStreamProgressCollName, options.Collection().
			SetWriteConcern(writeconcern.Majority()),
		),
	}, nil
}

func (r *LiveStreamProgressRepository) Insert(ctx context.Context, lsp *domain.LiveStreamProgress) error {
	_, err := r.writeColl.InsertOne(ctx, newLiveStreamProgressDoc(lsp))
	if err != nil {
		var me mongo.WriteException
		if errors.As(err, &me) {
			for _, we := range me.WriteErrors {
				if we.Code == 11000 { // duplicate key error
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

func (r *LiveStreamProgressRepository) Started(ctx context.Context, startsWithin time.Duration) (
	[]domain.LiveStreamProgress, error) {
	filter := bson.D{
		{Key: "finishedAt", Value: nil},
		{Key: "scheduledStart", Value: bson.D{
			{Key: "$lte", Value: time.Now().UTC().Add(startsWithin)},
		}},
	}

	cur, err := r.readColl.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var docs []liveStreamProgressDoc
	if err = cur.All(ctx, &docs); err != nil {
		return nil, err
	}

	pp := make([]domain.LiveStreamProgress, len(docs))
	for i, doc := range docs {
		p, err := doc.toDomain()
		if err != nil {
			return nil, fmt.Errorf("new live stream progress from doc: %v", err)
		}

		pp[i] = *p
	}

	return pp, nil
}

func (r *LiveStreamProgressRepository) Upsert(ctx context.Context, lsp *domain.LiveStreamProgress) error {
	updatedDoc := bson.M{"$set": newLiveStreamProgressDoc(lsp)}

	_, err := r.writeColl.UpdateOne(
		ctx,
		bson.M{"_id": lsp.ID()},
		updatedDoc,
		options.Update().SetUpsert(true),
	)

	return err
}

type liveStreamProgressDoc struct {
	VideoID        string     `bson:"_id"`
	ChatID         string     `bson:"chatId"`
	ScheduledStart time.Time  `bson:"scheduledStart"`
	NextPageToken  string     `bson:"nextPageToken,omitempty"`
	FinishedAt     *time.Time `bson:"finishedAt,omitempty"`
	FinishReason   string     `bson:"finishReason,omitempty"`
}

func newLiveStreamProgressDoc(lsp *domain.LiveStreamProgress) liveStreamProgressDoc {
	return liveStreamProgressDoc{
		VideoID:        lsp.ID(),
		ChatID:         lsp.ChatID(),
		ScheduledStart: lsp.ScheduledStart(),
		NextPageToken:  lsp.NextPageToken(),
		FinishedAt:     lsp.FinishedAt(),
		FinishReason:   lsp.FinishReason(),
	}
}

func (doc liveStreamProgressDoc) toDomain() (*domain.LiveStreamProgress, error) {
	lsp, err := domain.NewLiveStreamProgress(doc.VideoID, doc.ChatID, doc.ScheduledStart)
	if err != nil {
		return nil, err
	}

	lsp.SetNextPageToken(doc.NextPageToken)

	if doc.FinishedAt != nil && doc.FinishReason != "" {
		lsp.Finish(*doc.FinishedAt, doc.FinishReason)
	}

	return lsp, nil
}
