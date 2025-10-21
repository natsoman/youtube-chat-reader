package kafka_test

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/pkg/kafka"
)

func TestOutboxSyncProducer_ProducePending(t *testing.T) {
	t.Parallel()

	t.Run("successfully produce events", func(t *testing.T) {
		outboxSyncProducer, deps := setupTest(t)

		// Given
		outboxEvents := []kafka.OutboxEvent{{
			ID:      "1",
			Topic:   "topic1",
			Key:     "k",
			Payload: []byte(`{"a":"b"}`),
		}, {
			ID:      "2",
			Topic:   "topic1",
			Key:     "k",
			Payload: []byte(`{"a":"b"}`),
		}}
		deps.mockOutboxRepo.EXPECT().
			Pending(t.Context()).
			Return(outboxEvents, nil)
		producerMessages := []*sarama.ProducerMessage{
			{
				Topic: outboxEvents[0].Topic,
				Key:   sarama.StringEncoder(outboxEvents[0].Key),
				Value: sarama.StringEncoder(outboxEvents[0].Payload),
			},
			{
				Topic: outboxEvents[1].Topic,
				Key:   sarama.StringEncoder(outboxEvents[1].Key),
				Value: sarama.StringEncoder(outboxEvents[1].Payload),
			},
		}
		deps.mockSyncProducer.EXPECT().
			SendMessages(producerMessages)
		deps.mockOutboxRepo.EXPECT().
			MarkAsPublished(t.Context(), outboxEvents)

		// When
		err := outboxSyncProducer.ProducePending(t.Context())

		// Then
		assert.NoError(t, err)
	})

	t.Run("failed to get pending events", func(t *testing.T) {
		outboxSyncProducer, deps := setupTest(t)

		// Given
		deps.mockOutboxRepo.EXPECT().
			Pending(gomock.Any()).
			Return(nil, errors.New("error"))

		// When
		err := outboxSyncProducer.ProducePending(t.Context())

		// Then
		assert.EqualError(t, err, "pending events: error")
	})

	t.Run("zero pending events", func(t *testing.T) {
		outboxSyncProducer, deps := setupTest(t)

		// Given
		deps.mockOutboxRepo.EXPECT().
			Pending(gomock.Any()).
			Return(make([]kafka.OutboxEvent, 0), nil)

		// When
		err := outboxSyncProducer.ProducePending(t.Context())

		// Then
		assert.NoError(t, err)
	})

	t.Run("failed to send messages", func(t *testing.T) {
		outboxSyncProducer, deps := setupTest(t)

		// Given
		outboxEvents := []kafka.OutboxEvent{{
			ID:      "1",
			Topic:   "topic1",
			Key:     "k",
			Payload: []byte(`{"a":"b"}`),
		}}
		deps.mockOutboxRepo.EXPECT().
			Pending(gomock.Any()).
			Return(outboxEvents, nil)
		deps.mockSyncProducer.EXPECT().
			SendMessages(gomock.Any()).
			Return(errors.New("error"))

		// When
		err := outboxSyncProducer.ProducePending(t.Context())

		// Then
		assert.EqualError(t, err, "send messages: error")
	})

	t.Run("failed to mark messages as published", func(t *testing.T) {
		outboxSyncProducer, deps := setupTest(t)

		// Given
		outboxEvents := []kafka.OutboxEvent{{
			ID:      "1",
			Topic:   "topic1",
			Key:     "k",
			Payload: []byte(`{"a":"b"}`),
		}}
		deps.mockOutboxRepo.EXPECT().
			Pending(gomock.Any()).
			Return(outboxEvents, nil)
		deps.mockSyncProducer.EXPECT().
			SendMessages(gomock.Any())
		deps.mockOutboxRepo.EXPECT().
			MarkAsPublished(gomock.Any(), gomock.Any()).
			Return(errors.New("error"))

		// When
		err := outboxSyncProducer.ProducePending(t.Context())

		// Then
		assert.EqualError(t, err, "mark event as published: error")
	})
}

type testDeps struct {
	mockOutboxRepo   *MockOutboxRepository
	mockSyncProducer *MockSyncProducer
}

func setupTest(t *testing.T) (*kafka.OutboxSyncProducer, *testDeps) {
	t.Helper()
	t.Parallel()

	ctrl := gomock.NewController(t)
	deps := &testDeps{
		mockOutboxRepo:   NewMockOutboxRepository(ctrl),
		mockSyncProducer: NewMockSyncProducer(ctrl),
	}

	outboxSyncProducer, err := kafka.NewOutboxSyncProducer(
		deps.mockSyncProducer,
		deps.mockOutboxRepo,
	)
	require.NoError(t, err)

	return outboxSyncProducer, deps
}
