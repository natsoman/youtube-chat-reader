//go:build integration

package mongo_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
	inframongo "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/mongo"
)

var (
	_mongoC  *mongodb.MongoDBContainer
	_mongoDB *mongo.Database

	_liveStreamRepo *inframongo.LiveStreamRepository
	_outboxRepo     *inframongo.OutboxRepository
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	_mongoC, err = mongodb.Run(
		ctx,
		"mongo:8.0",
		mongodb.WithReplicaSet("rs0"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").WithStartupTimeout(time.Second*15),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = testcontainers.TerminateContainer(_mongoC)
	}()

	mongoURI, err := _mongoC.ConnectionString(ctx)
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

	liveStreamRepo, err := inframongo.NewLiveStreamRepository(_mongoDB)
	if err != nil {
		log.Fatal(err)
	}

	outboxRepo, err := inframongo.NewOutboxRepository(_mongoDB, "live_stream_found")
	if err != nil {
		log.Fatal(err)
	}

	_liveStreamRepo = liveStreamRepo
	_outboxRepo = outboxRepo

	os.Exit(m.Run())
}

func newLiveStreams(t *testing.T) []domain.LiveStream {
	t.Helper()

	tm := time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC)

	ls1, err := domain.NewLiveStream("id1", "title1", "channelId1", "chanTitle1", "thumbUrl1", "chatId1", tm, tm)
	require.NoError(t, err)

	ls2, err := domain.NewLiveStream("id2", "title2", "channelId2", "chanTitle2", "thumbUrl2", "chatId2", tm, tm)
	require.NoError(t, err)

	return []domain.LiveStream{*ls1, *ls2}
}
