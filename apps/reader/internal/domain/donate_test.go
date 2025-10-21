package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewDonate(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name          string
		id            string
		authorID      string
		videoID       string
		comment       string
		amount        string
		amountMicros  uint
		currency      string
		publishedAt   time.Time
		expectedError error
	}{
		{
			name:          "empty id",
			expectedError: errors.New("id is empty"),
		},
		{
			name:          "empty author id",
			id:            "id",
			expectedError: errors.New("author id is empty"),
		},
		{
			name:          "empty video id",
			id:            "id",
			authorID:      "authorId",
			expectedError: errors.New("video id is empty"),
		},
		{
			name:          "empty amount",
			id:            "id",
			authorID:      "authorId",
			videoID:       "videoId",
			expectedError: errors.New("amount is empty"),
		},
		{
			name:          "empty currency",
			id:            "id",
			authorID:      "authorId",
			videoID:       "videoId",
			amount:        "$10.00",
			expectedError: errors.New("currency is empty"),
		},
		{
			name:          "zero published at",
			id:            "id",
			authorID:      "authorId",
			videoID:       "videoId",
			amount:        "$10.00",
			currency:      "USD",
			expectedError: errors.New("published at is zero"),
		},
		{
			name:         "success",
			id:           "id",
			authorID:     "authorId",
			videoID:      "videoId",
			comment:      "Great stream!",
			amount:       "$10.00",
			amountMicros: 10000000,
			currency:     "USD",
			publishedAt:  now,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			donate, err := domain.NewDonate(
				tc.id,
				tc.authorID,
				tc.videoID,
				tc.comment,
				tc.amount,
				tc.amountMicros,
				tc.currency,
				tc.publishedAt,
			)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Nil(t, donate)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, donate)
				assert.Equal(t, tc.id, donate.ID())
				assert.Equal(t, tc.authorID, donate.AuthorID())
				assert.Equal(t, tc.videoID, donate.VideoID())
				assert.Equal(t, tc.comment, donate.Comment())
				assert.Equal(t, tc.amount, donate.Amount())
				assert.Equal(t, tc.amountMicros, donate.AmountMicros())
				assert.Equal(t, tc.currency, donate.Currency())
				assert.Equal(t, tc.publishedAt, donate.PublishedAt())
			}
		})
	}
}
