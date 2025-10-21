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

type BanRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewBanRepository(db *mongo.Database) (*BanRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const bansCollName = "bans"

	return &BanRepository{
		readColl: db.Collection(bansCollName, options.Collection().
			SetReadPreference(readpref.SecondaryPreferred()).
			SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(bansCollName, options.Collection().
			SetWriteConcern(writeconcern.Majority()),
		),
	}, nil
}

func (r *BanRepository) Insert(ctx context.Context, bb []domain.Ban) error {
	if len(bb) == 0 {
		return nil
	}

	docs := make([]interface{}, len(bb))
	for i, b := range bb {
		docs[i] = newBanDoc(&b)
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

type banDoc struct {
	ID          string        `bson:"_id"`
	VideoID     string        `bson:"videoId"`
	AuthorID    string        `bson:"authorId"`
	Type        string        `bson:"type"`
	Duration    time.Duration `bson:"duration,omitempty"`
	PublishedAt time.Time     `bson:"publishedAt"`
}

func newBanDoc(b *domain.Ban) banDoc {
	return banDoc{
		ID:          b.ID(),
		VideoID:     b.VideoID(),
		AuthorID:    b.AuthorID(),
		Type:        b.BanType().String(),
		Duration:    b.Duration(),
		PublishedAt: b.PublishedAt(),
	}
}
