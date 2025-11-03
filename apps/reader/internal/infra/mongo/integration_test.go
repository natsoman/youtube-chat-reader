//go:build integration

package mongo_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	inframongo "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo"
)

var (
	_mongoDB *mongo.Database

	_liveStreamProgressRepo *inframongo.LiveStreamProgressRepository
	_authorRepo             *inframongo.AuthorRepository
	_textMessageRepo        *inframongo.TextMessageRepository
	_banRepo                *inframongo.BanRepository
	_donateRepo             *inframongo.DonateRepository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	mongoContainer, err := mongodb.Run(
		ctx,
		"mongo:8.0",
		mongodb.WithReplicaSet("rs0"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").WithStartupTimeout(time.Second*15),
		),
	)
	defer func() {
		_ = testcontainers.TerminateContainer(mongoContainer)
	}()
	if err != nil {
		log.Fatal(err)
	}

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatal(err)
	}

	client, err := mongo.Connect(ctx,
		options.Client().ApplyURI(mongoURI),
		options.Client().SetDirect(true),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = client.Disconnect(ctx) }()

	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}

	_mongoDB = client.Database("testDb")
	defer func() { _ = _mongoDB.Drop(ctx) }()

	liveStreamProgressRepo, err := inframongo.NewLiveStreamProgressRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	authorRepo, err := inframongo.NewAuthorRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	textMessageRepo, err := inframongo.NewTextMessageRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	banRepo, err := inframongo.NewBanRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	donateRepo, err := inframongo.NewDonateRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	_liveStreamProgressRepo = liveStreamProgressRepo
	_authorRepo = authorRepo
	_textMessageRepo = textMessageRepo
	_banRepo = banRepo
	_donateRepo = donateRepo

	os.Exit(m.Run())
}
