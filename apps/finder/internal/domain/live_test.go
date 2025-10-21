package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/finder/internal/domain"
)

func TestNewLiveStream(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		id             string
		title          string
		channelID      string
		channelTitle   string
		thumbnailURL   string
		chatID         string
		publishedAt    time.Time
		scheduledStart time.Time

		expectedError error
	}{
		{
			expectedError: errors.New("id is empty"),
		},
		{
			id:            "id",
			expectedError: errors.New("title is empty"),
		},
		{
			id:            "id",
			title:         "title",
			expectedError: errors.New("channel id is empty"),
		},
		{
			id:            "id",
			title:         "title",
			channelID:     "channelId",
			expectedError: errors.New("channel title is empty"),
		},
		{
			id:            "id",
			title:         "title",
			channelID:     "channelId",
			channelTitle:  "channelTitle",
			expectedError: errors.New("thumbnail URL is empty"),
		},
		{
			id:            "id",
			title:         "title",
			channelID:     "channelId",
			channelTitle:  "channelTitle",
			thumbnailURL:  "thumbnailUrl",
			expectedError: errors.New("chat id is empty"),
		},
		{
			id:            "id",
			title:         "title",
			channelID:     "channelId",
			channelTitle:  "channelTitle",
			thumbnailURL:  "thumbnailUrl",
			chatID:        "chatId",
			expectedError: errors.New("published at is zero"),
		},
		{
			id:            "id",
			title:         "title",
			channelID:     "channelId",
			channelTitle:  "channelTitle",
			thumbnailURL:  "thumbnailUrl",
			chatID:        "chatId",
			publishedAt:   time.Now(),
			expectedError: errors.New("scheduled start is zero"),
		},
		{
			id:             "id",
			title:          "title",
			channelID:      "channelId",
			channelTitle:   "channelTitle",
			thumbnailURL:   "thumbnailUrl",
			chatID:         "chatId",
			publishedAt:    time.Now(),
			scheduledStart: time.Now(),
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			liveStream, err := domain.NewLiveStream(
				tc.id,
				tc.title,
				tc.channelID,
				tc.channelTitle,
				tc.thumbnailURL,
				tc.chatID,
				tc.publishedAt,
				tc.scheduledStart,
			)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, liveStream)
			}
		})
	}
}
