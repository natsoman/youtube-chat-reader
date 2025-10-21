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

type DonateRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewDonateRepository(db *mongo.Database) (*DonateRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const DonatesCollName = "donates"

	return &DonateRepository{
		readColl: db.Collection(DonatesCollName, options.Collection().
			SetReadPreference(readpref.SecondaryPreferred()).
			SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(DonatesCollName, options.Collection().
			SetWriteConcern(writeconcern.Majority()),
		),
	}, nil
}

func (r *DonateRepository) Insert(ctx context.Context, dd []domain.Donate) error {
	if len(dd) == 0 {
		return nil
	}

	docs := make([]interface{}, len(dd))
	for i, b := range dd {
		docs[i] = newDonateDoc(&b)
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

type donateDoc struct {
	ID           string    `bson:"_id"`
	AuthorID     string    `bson:"authorId"`
	VideoID      string    `bson:"videoId"`
	Comment      string    `bson:"comment"`
	Amount       string    `bson:"amount"`
	AmountMicros uint      `bson:"amountMicros"`
	Currency     string    `bson:"currency"`
	PublishedAt  time.Time `bson:"publishedAt"`
}

func newDonateDoc(b *domain.Donate) donateDoc {
	return donateDoc{
		ID:           b.ID(),
		AuthorID:     b.AuthorID(),
		VideoID:      b.VideoID(),
		Comment:      b.Comment(),
		Amount:       b.Amount(),
		AmountMicros: b.AmountMicros(),
		Currency:     b.Currency(),
		PublishedAt:  b.PublishedAt(),
	}
}
