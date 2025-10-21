package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewBan(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()

	testCases := []struct {
		name          string
		id            string
		authorID      string
		videoID       string
		banType       string
		duration      time.Duration
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
			name:          "unknown ban type",
			id:            "id",
			authorID:      "authorId",
			videoID:       "videoId",
			banType:       "unknown",
			expectedError: errors.New("unknown ban type 'unknown'"),
		},
		{
			name:          "temporary ban with zero duration",
			id:            "id",
			authorID:      "authorId",
			videoID:       "videoId",
			banType:       "temporary",
			duration:      0,
			publishedAt:   now,
			expectedError: errors.New("duration is zero"),
		},
		{
			name:        "temporary ban uppercase",
			id:          "id",
			authorID:    "authorId",
			videoID:     "videoId",
			banType:     "TEMPORARY",
			duration:    5 * time.Minute,
			publishedAt: now,
		},
		{
			name:        "temporary ban lowercase",
			id:          "id",
			authorID:    "authorId",
			videoID:     "videoId",
			banType:     "temporary",
			duration:    10 * time.Minute,
			publishedAt: now,
		},
		{
			name:        "permanent ban uppercase",
			id:          "id",
			authorID:    "authorId",
			videoID:     "videoId",
			banType:     "PERMANENT",
			publishedAt: now,
		},
		{
			name:        "permanent ban lowercase",
			id:          "id",
			authorID:    "authorId",
			videoID:     "videoId",
			banType:     "permanent",
			publishedAt: now,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ban, err := domain.NewBan(tc.id, tc.authorID, tc.videoID, tc.banType, tc.duration, tc.publishedAt)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Nil(t, ban)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, ban)
				assert.Equal(t, tc.id, ban.ID())
				assert.Equal(t, tc.authorID, ban.AuthorID())
				assert.Equal(t, tc.videoID, ban.VideoID())
				assert.Equal(t, tc.publishedAt, ban.PublishedAt())

				if tc.banType == "TEMPORARY" || tc.banType == "temporary" {
					assert.Equal(t, domain.Temporary, ban.BanType())
					assert.Equal(t, tc.duration, ban.Duration())
				} else {
					assert.Equal(t, domain.Permanent, ban.BanType())
				}
			}
		})
	}
}

func TestBanType_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		banType  domain.BanType
		expected string
	}{
		{
			name:     "temporary",
			banType:  domain.Temporary,
			expected: "temporary",
		},
		{
			name:     "permanent",
			banType:  domain.Permanent,
			expected: "permanent",
		},
		{
			name:     "unknown",
			banType:  domain.BanType(999),
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.banType.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}
