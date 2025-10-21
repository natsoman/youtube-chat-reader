package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/dnwe/otelsarama"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"google.golang.org/api/option"
	apiyoutube "google.golang.org/api/youtube/v3"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/app"
	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra"
	inframongo "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/mongo"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/mongo/otel"
	infraotel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/otel"
	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/youtube"
	youtubeotel "github.com/natsoman/youtube-chat-reader/apps/finder/internal/infra/youtube/otel"
	"github.com/natsoman/youtube-chat-reader/pkg/kafka"
	pkgmongo "github.com/natsoman/youtube-chat-reader/pkg/mongo"
	pkgmongootel "github.com/natsoman/youtube-chat-reader/pkg/mongo/otel"
	"github.com/natsoman/youtube-chat-reader/pkg/otel"
)

const _serviceName = "finder"

var _version string

func main() {
	exitCode := 1

	defer func() { os.Exit(exitCode) }()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cnf, err := infra.NewConf()
	if err != nil {
		fmt.Printf("Failed to create configuration: %v", err)
		return
	}

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

	outboxRepo, err := inframongo.NewOutboxRepository(
		mongoClient.Database(cnf.MongoDB.Database),
		cnf.Kafka.Topics.LiveStreamFoundV1,
	)
	if err != nil {
		log.Error("Failed to create outbox repository", "err", err)
		return
	}

	instOutboxRepo, err := mongootel.NewInstrumentedOutboxRepository(outboxRepo)
	if err != nil {
		log.Error("Failed to create instrumented outbox repository", "err", err)
		return
	}

	liveStreamRepo, err := inframongo.NewLiveStreamRepository(mongoClient.Database(cnf.MongoDB.Database))
	if err != nil {
		log.Error("Failed to create live repository", "err", err)
		return
	}

	instLiveStreamRepo, err := mongootel.NewInstrumentedLiveStreamRepository(liveStreamRepo)
	if err != nil {
		log.Error("Failed to create instrumented live repository", "err", err)
		return
	}

	transactor, err := pkgmongo.NewTransactor(mongoClient)
	if err != nil {
		log.Error("Failed to create transactor", "err", err)
		return
	}

	instTransactor, err := pkgmongootel.NewInstrumentedTransactor(transactor)
	if err != nil {
		log.Error("Failed to create instrumented transactor", "err", err)
		return
	}

	youtubeSvc, err := apiyoutube.NewService(ctx, option.WithAPIKey(cnf.YouTube.APIKey))
	if err != nil {
		log.Error("Failed to create YouTube service", "err", err)
		return
	}

	youtubeClient, err := youtube.NewClient(youtubeSvc.Videos, youtubeSvc.Search)
	if err != nil {
		log.Error("Failed to create YouTube client", "err", err)
		return
	}

	instYoutubeClient, err := youtubeotel.NewInstrumentedClient(youtubeClient)
	if err != nil {
		log.Error("Failed to create instrumented YouTube client", "err", err)
		return
	}

	liveStreamFinder, err := app.NewLiveStreamFinder(
		instYoutubeClient,
		instLiveStreamRepo,
		instOutboxRepo,
		instTransactor,
	)
	if err != nil {
		log.Error("Failed to create live stream finder", "err", err)
		return
	}

	instLiveStreamFinder, err := infraotel.NewInstrumentedLiveStreamFinder(liveStreamFinder)
	if err != nil {
		log.Error("Failed to create instrumented live stream finder", "err", err)
		return
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true

	syncProducer, err := sarama.NewSyncProducer(cnf.Kafka.Brokers, saramaConfig)
	if err != nil {
		log.Error("Failed to create sync producer", "err", err)
		return
	}

	defer func() {
		if err = syncProducer.Close(); err != nil {
			log.Error("Failed to close sync producer", "err", err)
		}
	}()

	instSyncProducer := otelsarama.WrapSyncProducer(saramaConfig, syncProducer)

	outboxSyncProducer, err := kafka.NewOutboxSyncProducer(instSyncProducer, outboxRepo)
	if err = instLiveStreamFinder.Find(ctx, cnf.YouTube.Channels); err != nil {
		log.Error("Failed to find live streams", "err", err)
		return
	}

	if err = outboxSyncProducer.ProducePending(ctx); err != nil {
		log.Error("Failed to produce outbox events", "err", err)
		return
	}

	exitCode = 0
}
