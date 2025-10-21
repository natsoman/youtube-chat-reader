//go:generate mockgen -destination=mock_sarama_test.go -package=kafka_test github.com/IBM/sarama SyncProducer,ConsumerGroupSession,ConsumerGroupClaim
//go:generate mockgen -destination=mock_outbox_test.go -package=kafka_test -source=outbox.go

package kafka_test
