package main

import (
	"context"
	"errors"
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

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/kafka"
	inframongo "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo"
	mongootel "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/mongo/otel"
	pkgkafka "github.com/natsoman/youtube-chat-reader/pkg/kafka"
	"github.com/natsoman/youtube-chat-reader/pkg/otel"
)

const _serviceName = "reader-consumer"

var _version string

func main() {
	exitCode := 1

	defer func() { os.Exit(exitCode) }()

	cnf, err := infra.NewConsumerConf()
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

	saramaConf := sarama.NewConfig()
	saramaConf.Version = sarama.V4_0_0_0
	saramaConf.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	saramaConf.Consumer.Offsets.Initial = sarama.OffsetNewest
	saramaConf.Consumer.Offsets.AutoCommit.Enable = false

	consumerGroup, err := sarama.NewConsumerGroup(cnf.Kafka.Brokers, _serviceName, saramaConf)
	if err != nil {
		log.Error("Failed to construct consumer group", "err", err)
		return
	}

	defer func() {
		if err = consumerGroup.Close(); err != nil {
			log.Error("Failed to close consumer group", "err", err)
		}
	}()

	liveMessageHandler, err := kafka.NewLiveStreamFoundEventHandler(instLiveStreamProgressRepo)
	if err != nil {
		log.Error("Failed to construct live message handler", "err", err)
		return
	}

	consumerGroupHandler, err := pkgkafka.NewConsumerGroupHandler(
		log,
		map[string]pkgkafka.MessageHandler{
			cnf.Kafka.Topics.LiveStreamFoundV1: liveMessageHandler.Handle,
		},
		time.Second*5,
	)
	if err != nil {
		log.Error("Failed to construct consumer group handler", "err", err)
		return
	}

	for {
		if err = consumerGroup.Consume(
			ctx,
			[]string{cnf.Kafka.Topics.LiveStreamFoundV1},
			otelsarama.WrapConsumerGroupHandler(consumerGroupHandler),
		); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error("Failed to consume", "err", err)
			}

			return
		}
	}
}
