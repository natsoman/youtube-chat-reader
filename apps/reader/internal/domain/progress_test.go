package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewLiveStreamProgress(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name           string
		id             string
		chatID         string
		scheduledStart time.Time
		expectedError  error
	}{
		{
			name:          "empty id",
			expectedError: errors.New("id is empty"),
		},
		{
			name:          "empty chat id",
			id:            "id",
			expectedError: errors.New("chat id is empty"),
		},
		{
			name:          "zero scheduled start",
			id:            "id",
			chatID:        "chatId",
			expectedError: errors.New("scheduled start is zero"),
		},
		{
			name:           "success",
			id:             "id",
			chatID:         "chatId",
			scheduledStart: now,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			lsp, err := domain.NewLiveStreamProgress(tc.id, tc.chatID, tc.scheduledStart)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Nil(t, lsp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, lsp)
				assert.Equal(t, tc.id, lsp.ID())
				assert.Equal(t, tc.chatID, lsp.ChatID())
				assert.Equal(t, tc.scheduledStart, lsp.ScheduledStart())
				assert.Empty(t, lsp.NextPageToken())
				assert.Nil(t, lsp.FinishedAt())
				assert.Empty(t, lsp.FinishReason())
			}
		})
	}
}

func TestLiveStreamProgress_SetNextPageToken(t *testing.T) {
	t.Parallel()

	lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
	assert.NoError(t, err)

	// Initially empty
	assert.Empty(t, lsp.NextPageToken())

	// Set token
	lsp.SetNextPageToken("token123")
	assert.Equal(t, "token123", lsp.NextPageToken())

	// Update token
	lsp.SetNextPageToken("token456")
	assert.Equal(t, "token456", lsp.NextPageToken())
}

func TestLiveStreamProgress_Finish(t *testing.T) {
	t.Parallel()

	lsp, err := domain.NewLiveStreamProgress("id", "chatId", time.Now().UTC())
	assert.NoError(t, err)

	// Initially not finished
	assert.Nil(t, lsp.FinishedAt())
	assert.Empty(t, lsp.FinishReason())

	// Finish
	finishTime := time.Now().UTC()
	lsp.Finish(finishTime, "stream ended")

	assert.NotNil(t, lsp.FinishedAt())
	assert.Equal(t, finishTime, *lsp.FinishedAt())
	assert.Equal(t, "stream ended", lsp.FinishReason())
}
