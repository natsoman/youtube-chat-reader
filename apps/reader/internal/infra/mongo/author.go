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

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

type AuthorRepository struct {
	readColl  *mongo.Collection
	writeColl *mongo.Collection
}

func NewAuthorRepository(db *mongo.Database) (*AuthorRepository, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	const authorsCollName = "authors"

	return &AuthorRepository{
		readColl: db.Collection(authorsCollName, options.Collection().
			SetReadPreference(readpref.SecondaryPreferred()).
			SetReadConcern(readconcern.Majority()),
		),
		writeColl: db.Collection(authorsCollName, options.Collection().
			SetWriteConcern(writeconcern.Majority()),
		),
	}, nil
}

func (r *AuthorRepository) Upsert(ctx context.Context, aa []domain.Author) error {
	if len(aa) == 0 {
		return nil
	}

	models := make([]mongo.WriteModel, 0, len(aa))
	now := time.Now().UTC()

	for _, a := range aa {
		doc := newAuthorDoc(&a)
		doc.UpdatedAt = now

		model := mongo.NewReplaceOneModel().
			SetFilter(bson.M{"_id": a.ID()}).
			SetReplacement(doc).
			SetUpsert(true)

		models = append(models, model)
	}

	_, err := r.writeColl.BulkWrite(ctx, models)

	return err
}

type authorDoc struct {
	ID              string    `bson:"_id"`
	Name            string    `bson:"name"`
	ProfileImageURL string    `bson:"profileImageUrl"`
	IsVerified      bool      `bson:"isVerified"`
	UpdatedAt       time.Time `bson:"updatedAt"`
}

func newAuthorDoc(a *domain.Author) authorDoc {
	return authorDoc{
		ID:              a.ID(),
		Name:            a.Name(),
		ProfileImageURL: a.ProfileImageURL(),
		IsVerified:      a.IsVerified(),
		UpdatedAt:       time.Now().UTC(),
	}
}
