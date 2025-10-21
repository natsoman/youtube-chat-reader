package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewTextMessage(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name          string
		id            string
		videoID       string
		authorID      string
		text          string
		publishedAt   time.Time
		expectedError error
	}{
		{
			name:          "empty id",
			expectedError: errors.New("id is empty"),
		},
		{
			name:          "empty video id",
			id:            "id",
			expectedError: errors.New("video id is empty"),
		},
		{
			name:          "empty author id",
			id:            "id",
			videoID:       "videoId",
			expectedError: errors.New("author id is empty"),
		},
		{
			name:          "zero published at",
			id:            "id",
			videoID:       "videoId",
			authorID:      "authorId",
			expectedError: errors.New("published at is zero"),
		},
		{
			name:        "success",
			id:          "id",
			videoID:     "videoId",
			authorID:    "authorId",
			text:        "Hello world!",
			publishedAt: now,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			textMessage, err := domain.NewTextMessage(tc.id, tc.videoID, tc.authorID, tc.text, tc.publishedAt)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Nil(t, textMessage)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, textMessage)
				assert.Equal(t, tc.id, textMessage.ID())
				assert.Equal(t, tc.videoID, textMessage.VideoID())
				assert.Equal(t, tc.authorID, textMessage.AuthorID())
				assert.Equal(t, tc.text, textMessage.Text())
				assert.Equal(t, tc.publishedAt, textMessage.PublishedAt())
			}
		})
	}
}
