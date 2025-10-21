package infra

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type WorkerConf struct {
	LogLevel string `default:"debug" split_words:"true"`
	OTEL     OTEL
	MongoDB  MongoDB
	Redis    Redis
	YouTube  YouTube

	MaxRetryInterval time.Duration `default:"10s" split_words:"true"`
	AdvanceStart     time.Duration `default:"30m" split_words:"true"`
}

type ConsumerConf struct {
	LogLevel string `default:"debug" split_words:"true"`
	OTEL     OTEL
	MongoDB  MongoDB
	Kafka    Kafka
}

type MongoDB struct {
	// nolint:lll
	URI      string `default:"mongodb://mongodb-0.replica-set.mongo.svc.cluster.local:27017,mongodb-1.replica-set.mongo.svc.cluster.local:27017,mongodb-2.replica-set.mongo.svc.cluster.local:27017/admin?replicaSet=rs0"`
	Database string `default:"youtube-chat-reader"`
}

type Kafka struct {
	// nolint:lll
	Brokers []string `default:"youtube-chat-reader-dual-role-0.youtube-chat-reader-kafka-brokers.kafka.svc.cluster.local:9092,youtube-chat-reader-dual-role-1.youtube-chat-reader-kafka-brokers.kafka.svc.cluster.local:9092,youtube-chat-reader-dual-role-2.youtube-chat-reader-kafka-brokers.kafka.svc.cluster.local:9092"`
	Topics  struct {
		LiveStreamFoundV1 string `default:"live_stream.found.v1" split_words:"true"`
	}
}

type Redis struct {
	// nolint:lll
	Addr []string `default:"redis-cluster-0.redis-cluster-headless.redis.svc.cluster.local:6379,redis-cluster-1.redis-cluster-headless.redis.svc.cluster.local:6379,redis-cluster-2.redis-cluster-headless.redis.svc.cluster.local:6379,redis-cluster-3.redis-cluster-headless.redis.svc.cluster.local:6379,redis-cluster-4.redis-cluster-headless.redis.svc.cluster.local:6379,redis-cluster-5.redis-cluster-headless.redis.svc.cluster.local:6379"`
}

type OTEL struct {
	// nolint:lll
	CollectorGRPCAddr string  `default:"otel-collector-opentelemetry-collector.observability.svc.cluster.local:4317" split_words:"true"`
	SampleRate        float64 `default:"1.0" split_words:"true"`
}

type YouTube struct {
	GRPCTarget string   `default:"dns:///youtube.googleapis.com:443" split_words:"true"`
	APIKeys    []string `required:"true" split_words:"true"`
}

func NewWorkerConf() (*WorkerConf, error) {
	cnf := &WorkerConf{}
	if err := envconfig.Process("", cnf); err != nil {
		return nil, err
	}

	return cnf, nil
}

func NewConsumerConf() (*ConsumerConf, error) {
	cnf := &ConsumerConf{}
	if err := envconfig.Process("", cnf); err != nil {
		return nil, err
	}

	return cnf, nil
}
