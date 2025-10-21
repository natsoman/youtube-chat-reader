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
)

var (
	mongoC  *mongodb.MongoDBContainer
	mongoDB *mongo.Database
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	mongoC, err = mongodb.Run(
		ctx,
		"mongo:6.0",
		mongodb.WithReplicaSet("rs0"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections").WithStartupTimeout(time.Second*10),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = testcontainers.TerminateContainer(mongoC)
	}()

	mongoURI, err := mongoC.ConnectionString(ctx)
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

	mongoDB = client.Database("test")
	defer func() { _ = mongoDB.Drop(ctx) }()

	os.Exit(m.Run())
}
