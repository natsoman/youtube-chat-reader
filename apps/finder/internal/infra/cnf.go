package infra

import (
	"github.com/kelseyhightower/envconfig"
)

type Conf struct {
	LogLevel string `default:"debug" split_words:"true"`
	OTEL     OTEL
	MongoDB  MongoDB
	Kafka    Kafka
	YouTube  YouTube
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

type YouTube struct {
	Host     string   `default:"https://www.googleapis.com"`
	APIKey   string   `required:"true" split_words:"true"`
	Channels []string `required:"true"`
}

type OTEL struct {
	// nolint:lll
	CollectorGRPCAddr string  `default:"otel-collector-opentelemetry-collector.observability.svc.cluster.local:4317" split_words:"true"`
	SampleRate        float64 `default:"1" split_words:"true"`
}

func NewConf() (*Conf, error) {
	cnf := &Conf{}
	if err := envconfig.Process("", cnf); err != nil {
		return nil, err
	}

	return cnf, nil
}
