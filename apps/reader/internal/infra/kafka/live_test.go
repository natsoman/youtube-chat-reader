//go:generate mockgen -destination=mock_test.go -package=kafka_test -source=live.go
package kafka_test

import (
	"errors"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/kafka"
)

func TestLiveStreamFoundEventHandler_Handle(t *testing.T) {
	t.Parallel()

	t.Run("successfully consumes event and inserts live stream progress", func(t *testing.T) {
		t.Parallel()

		mockLiveStreamProgressRepo := NewMockLiveStreamProgressRepository(gomock.NewController(t))
		handler, err := kafka.NewLiveStreamFoundEventHandler(mockLiveStreamProgressRepo)
		require.NoError(t, err)

		// Given
		lsp, err := domain.NewLiveStreamProgress("a", "b", time.Date(2025, time.October, 20, 12, 0, 0, 0, time.UTC))
		require.NoError(t, err)
		mockLiveStreamProgressRepo.EXPECT().
			Insert(gomock.Any(), lsp)

		// When
		eventPayload := []byte(`{"videoId": "a","chatId": "b","scheduledStart": "2025-10-20T12:00:00Z"}`)
		err = handler.Handle(t.Context(), &sarama.ConsumerMessage{Value: eventPayload})

		// Then
		assert.NoError(t, err)
	})

	t.Run("fails to unmarshal event payload", func(t *testing.T) {
		t.Parallel()

		mockLiveStreamProgressRepo := NewMockLiveStreamProgressRepository(gomock.NewController(t))
		handler, err := kafka.NewLiveStreamFoundEventHandler(mockLiveStreamProgressRepo)
		require.NoError(t, err)

		// When
		err = handler.Handle(t.Context(), &sarama.ConsumerMessage{Value: []byte(`a`)})

		// Then
		assert.ErrorContains(t, err, "unmarshal event payload")
	})

	t.Run("fails to make new live stream progress from event", func(t *testing.T) {
		t.Parallel()

		mockLiveStreamProgressRepo := NewMockLiveStreamProgressRepository(gomock.NewController(t))
		handler, err := kafka.NewLiveStreamFoundEventHandler(mockLiveStreamProgressRepo)
		require.NoError(t, err)

		// When
		eventPayload := []byte(`{"videoId": "","chatId": "b","scheduledStart": "2025-10-20T12:00:00Z"}`)
		err = handler.Handle(t.Context(), &sarama.ConsumerMessage{Value: eventPayload})

		// Then
		assert.ErrorContains(t, err, "new live stream")
	})

	t.Run("fails to insert live stream progress", func(t *testing.T) {
		t.Parallel()

		mockLiveStreamProgressRepo := NewMockLiveStreamProgressRepository(gomock.NewController(t))
		handler, err := kafka.NewLiveStreamFoundEventHandler(mockLiveStreamProgressRepo)
		require.NoError(t, err)

		// Given
		mockLiveStreamProgressRepo.EXPECT().
			Insert(gomock.Any(), gomock.Any()).
			Return(errors.New("insert error"))

		// When
		eventPayload := []byte(`{"videoId": "a","chatId": "b","scheduledStart": "2025-10-20T12:00:00Z"}`)
		err = handler.Handle(t.Context(), &sarama.ConsumerMessage{Value: eventPayload})

		// Then
		assert.ErrorContains(t, err, "insert error")
	})
}
