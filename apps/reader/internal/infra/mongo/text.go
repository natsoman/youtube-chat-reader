package mongo

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type TextMessageRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewTextMessageRepository(db *mongo.Database) (*TextMessageRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const liveStreamProgressCollName = "texts"

	return &TextMessageRepository{
		readColl: db.Collection(liveStreamProgressCollName, options.Collection().
			SetReadPreference(readpref.SecondaryPreferred()).
			SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(liveStreamProgressCollName, options.Collection().
			SetWriteConcern(writeconcern.Majority()),
		),
	}, nil
}

func (r *TextMessageRepository) Insert(ctx context.Context, tms []domain.TextMessage) error {
	if len(tms) == 0 {
		return nil
	}

	docs := make([]interface{}, len(tms))
	for i, tm := range tms {
		docs[i] = newTextMessageDoc(&tm)
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

type textMessageDoc struct {
	ID          string    `bson:"_id"`
	VideoID     string    `bson:"videoId"`
	AuthorID    string    `bson:"authorId"`
	Text        string    `bson:"text"`
	PublishedAt time.Time `bson:"publishedAt"`
}

func newTextMessageDoc(tm *domain.TextMessage) textMessageDoc {
	return textMessageDoc{
		ID:          tm.ID(),
		VideoID:     tm.VideoID(),
		AuthorID:    tm.AuthorID(),
		Text:        tm.Text(),
		PublishedAt: tm.PublishedAt(),
	}
}
