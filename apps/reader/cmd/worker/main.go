package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/app"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/google"
	inframongo "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo/otel"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/redis"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/youtube"
	"github.com/natsoman/youtube-chat-reader/pkg/otel"
)

const _serviceName = "reader-worker"

var _version string

func main() {
	exitCode := 1

	defer func() { os.Exit(exitCode) }()

	cnf, err := infra.NewWorkerConf()
	if err != nil {
		fmt.Printf("Failed to create configuration: %v", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	telemetry, err := otel.Configure(
		ctx,
		_serviceName,
		cnf.OTEL.CollectorGRPCAddr,
		otel.WithLogLevel(cnf.LogLevel),
		otel.WithServiceVersion(_version),
	)
	if err != nil {
		fmt.Printf("Failed to configure OTEL: %v", err)
		return
	}

	defer telemetry.Shutdown()

	log := slog.Default()

	log.Info("Starting...")
	defer log.Info("Stopped")

	mongoClientOpts := options.Client().
		SetMonitor(otelmongo.NewMonitor()).
		ApplyURI(cnf.MongoDB.URI).
		SetAppName(_serviceName)

	mongoClient, err := mongo.Connect(ctx, mongoClientOpts)
	if err != nil {
		log.Error("Failed to connect to Mongo", "err", err)
		return
	}

	defer func() {
		timeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		if err = mongoClient.Disconnect(timeCtx); err != nil {
			log.Error("Failed to disconnect from Mongo", "err", err)
			return
		}

		log.Debug("Disconnected from Mongo")
	}()

	conn, err := grpc.NewClient(
		cnf.YouTube.GRPCTarget,
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS13})),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Error("Failed to create GRPC connection", "err", err)
		return
	}

	grpcClient, err := youtube.NewStreamChatMessagesGRPCClient(
		youtube.NewV3DataLiveChatMessageServiceClient(conn),
		&google.Ticker{},
		&google.Ticker{},
		cnf.YouTube.APIKeys,
	)
	if err != nil {
		log.Error("Failed to create YouTube GRPC client", "err", err)
		return
	}

	redisClusterClient := goredis.NewClusterClient(&goredis.ClusterOptions{
		Addrs:      cnf.Redis.Addr,
		ClientName: _serviceName,
		MaxRetries: 3,
	})

	if err = redisotel.InstrumentTracing(redisClusterClient); err != nil {
		log.Error("Failed to instrument Redis with trace", "err", err)
		return
	}

	if err = redisotel.InstrumentMetrics(redisClusterClient); err != nil {
		log.Error("Failed to instrument Redis with metrics", "err", err)
		return
	}

	locker, err := redis.NewLocker(redisClusterClient)
	if err != nil {
		log.Error("Failed to create Redis locker", "err", err)
		return
	}

	liveStreamProgressRepo, err := inframongo.NewLiveStreamProgressRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create live stream progress repository", "err", err)
		return
	}

	instLiveStreamProgressRepo, err := mongootel.NewInstrumentedLiveStreamProgressRepository(liveStreamProgressRepo)
	if err != nil {
		log.Error("Failed to create instrumented live stream progress repository", "err", err)
		return
	}

	banRepo, err := inframongo.NewBanRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create ban repository", "err", err)
		return
	}

	instBanRepo, err := mongootel.NewInstrumentedBanRepository(banRepo)
	if err != nil {
		log.Error("Failed to create instrumented ban repository", "err", err)
		return
	}

	textMessageRepo, err := inframongo.NewTextMessageRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create text message repository", "err", err)
		return
	}

	instTextMessageRepo, err := mongootel.NewInstrumentedTextMessageRepository(textMessageRepo)
	if err != nil {
		log.Error("Failed to create instrumented text message repository", "err", err)
		return
	}

	donateRepo, err := inframongo.NewDonateRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create donate repository", "err", err)
		return
	}

	instDonateRepo, err := mongootel.NewInstrumentedDonateRepository(donateRepo)
	if err != nil {
		log.Error("Failed to create instrumented donate repository", "err", err)
		return
	}

	authorRepo, err := inframongo.NewAuthorRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create author repository", "err", err)
		return
	}

	instAuthorRepo, err := mongootel.NewInstrumentedAuthorRepository(authorRepo)
	if err != nil {
		log.Error("Failed to create instrumented author repository", "err", err)
		return
	}

	liveStreamReader, err := app.NewLiveStreamReader(
		&google.Clock{},
		&google.Ticker{},
		locker,
		grpcClient,
		instLiveStreamProgressRepo,
		instBanRepo,
		instTextMessageRepo,
		instDonateRepo,
		instAuthorRepo,
		app.WithMaxRetryInterval(cnf.MaxRetryInterval),
		app.WithAdvanceStart(cnf.AdvanceStart),
	)
	if err != nil {
		log.Error("Failed to create live stream reader", "err", err)
		return
	}

	liveStreamReader.Read(ctx)
}
