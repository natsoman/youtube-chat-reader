package kafka_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"
)

func TestConsumerGroupHandler_ConsumeClaim(t *testing.T) {
	t.Parallel()

	const (
		testTopicName = "test.topic.two"
		timeout       = time.Second
	)

	newConsumerMessage := func() *sarama.ConsumerMessage {
		return &sarama.ConsumerMessage{
			Topic: testTopicName,
			Key:   []byte("key"),
			Value: []byte("val"),
		}
	}

	t.Run("message is handled successfully and gets marked", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// Given
		consumerMessage := newConsumerMessage()

		mockConsumerGroupSess := NewMockConsumerGroupSession(ctrl)
		mockConsumerGroupSess.EXPECT().
			Context().AnyTimes().
			Return(t.Context())
		mockConsumerGroupSess.EXPECT().
			MarkMessage(consumerMessage, "")

		consumerMessagesChan := make(chan *sarama.ConsumerMessage)
		mockConsumerGroupClaim := NewMockConsumerGroupClaim(ctrl)
		mockConsumerGroupClaim.EXPECT().
			Messages().
			Return(consumerMessagesChan)

		consumerGroupHandler, _ := kafka.NewConsumerGroupHandler(
			slog.Default(),
			map[string]kafka.MessageHandler{
				testTopicName: func(ctx context.Context, m *sarama.ConsumerMessage) error {
					return nil
				},
			},
			timeout,
		)

		// When
		go func() {
			defer close(consumerMessagesChan)

			consumerMessagesChan <- consumerMessage
		}()

		err := consumerGroupHandler.ConsumeClaim(mockConsumerGroupSess, mockConsumerGroupClaim)

		// Then
		assert.NoError(t, err)
	})

	t.Run("message is handled unsuccessfully and is not get marker as processed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// Given
		consumerMessage := newConsumerMessage()
		mockConsumerGroupSess := NewMockConsumerGroupSession(ctrl)
		consumerMessagesChan := make(chan *sarama.ConsumerMessage)

		mockConsumerGroupSess.EXPECT().
			Context().
			AnyTimes().
			Return(t.Context())

		mockConsumerGroupClaim := NewMockConsumerGroupClaim(ctrl)
		mockConsumerGroupClaim.EXPECT().
			Messages().
			Return(consumerMessagesChan)

		consumerGroupHandler, _ := kafka.NewConsumerGroupHandler(
			slog.Default(),
			map[string]kafka.MessageHandler{
				testTopicName: func(ctx context.Context, m *sarama.ConsumerMessage) error {
					time.Sleep(time.Nanosecond)
					return ctx.Err()
				},
			},
			time.Nanosecond,
		)

		// When
		go func() {
			defer close(consumerMessagesChan)

			consumerMessagesChan <- consumerMessage
		}()

		err := consumerGroupHandler.ConsumeClaim(mockConsumerGroupSess, mockConsumerGroupClaim)

		// Then
		assert.EqualError(t, err, "handle message: context deadline exceeded")
	})

	t.Run("handler is missing", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)

		// Given
		consumerMessage := newConsumerMessage()
		consumerMessage.Topic = "unhandled.topic"

		mockConsumerGroupSess := NewMockConsumerGroupSession(ctrl)
		mockConsumerGroupSess.EXPECT().
			Context().
			AnyTimes().
			Return(t.Context())
		mockConsumerGroupSess.EXPECT().
			MarkMessage(consumerMessage, "")

		consumerMessagesChan := make(chan *sarama.ConsumerMessage)
		mockConsumerGroupClaim := NewMockConsumerGroupClaim(ctrl)
		mockConsumerGroupClaim.EXPECT().
			Messages().
			Return(consumerMessagesChan)

		consumerGroupHandler, _ := kafka.NewConsumerGroupHandler(
			slog.Default(),
			map[string]kafka.MessageHandler{},
			timeout,
		)

		// When
		go func() {
			defer close(consumerMessagesChan)

			consumerMessagesChan <- consumerMessage
		}()

		err := consumerGroupHandler.ConsumeClaim(mockConsumerGroupSess, mockConsumerGroupClaim)

		// Then
		assert.NoError(t, err)
	})
}
