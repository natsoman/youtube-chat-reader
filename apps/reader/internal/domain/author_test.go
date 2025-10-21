package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewAuthor(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name            string
		id              string
		authorName      string
		profileImageURL string
		isVerified      bool
		expectedError   error
	}{
		{
			name:          "empty id",
			expectedError: errors.New("id is empty"),
		},
		{
			name:          "empty profile image url",
			id:            "id",
			authorName:    "name",
			expectedError: errors.New("profile image url is empty"),
		},
		{
			name:            "success",
			id:              "id",
			authorName:      "name",
			profileImageURL: "https://example.com/image.jpg",
			isVerified:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			author, err := domain.NewAuthor(tc.id, tc.authorName, tc.profileImageURL, tc.isVerified)
			if tc.expectedError != nil {
				assert.EqualError(t, err, tc.expectedError.Error())
				assert.Nil(t, author)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, author)
				assert.Equal(t, tc.id, author.ID())
				assert.Equal(t, tc.authorName, author.Name())
				assert.Equal(t, tc.profileImageURL, author.ProfileImageURL())
				assert.Equal(t, tc.isVerified, author.IsVerified())
			}
		})
	}
}
