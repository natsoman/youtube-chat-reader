package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/domain"
)

func TestNewChatMessages(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		nextPageToken string
	}{
		{
			name:          "with token",
			nextPageToken: "token123",
		},
		{
			name:          "with empty token",
			nextPageToken: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cm := domain.NewChatMessages(tc.nextPageToken)
			assert.NotNil(t, cm)
			assert.Equal(t, tc.nextPageToken, cm.NextPageToken())
			assert.Empty(t, cm.TextMessages())
			assert.Empty(t, cm.Donates())
			assert.Empty(t, cm.Bans())
			assert.Empty(t, cm.Authors())
		})
	}
}

func TestChatMessages_AddTextMessage(t *testing.T) {
	t.Parallel()

	cm := domain.NewChatMessages("token")
	now := time.Now().UTC()

	tm1, err := domain.NewTextMessage("tm1", "videoId", "author1", "Hello", now)
	require.NoError(t, err)
	cm.AddTextMessage(tm1)

	tm2, err := domain.NewTextMessage("tm2", "videoId", "author2", "World", now)
	require.NoError(t, err)
	cm.AddTextMessage(tm2)

	messages := cm.TextMessages()
	assert.Len(t, messages, 2)

	// Verify messages exist by ID
	found := false

	for _, msg := range messages {
		if msg.ID() == "tm1" {
			found = true

			assert.Equal(t, "Hello", msg.Text())
		}
	}

	assert.True(t, found, "tm1 should be found")

	// Adding duplicate should not add (only adds if not exists)
	tm1Updated, err := domain.NewTextMessage("tm1", "videoId", "author1", "Updated", now)
	require.NoError(t, err)
	cm.AddTextMessage(tm1Updated)

	// Still 2, not added
	messages = cm.TextMessages()
	assert.Len(t, messages, 2)

	// Original message should still be there
	for _, msg := range messages {
		if msg.ID() == "tm1" {
			assert.Equal(t, "Hello", msg.Text()) // Original text
		}
	}
}

func TestChatMessages_AddBan(t *testing.T) {
	t.Parallel()

	cm := domain.NewChatMessages("token")
	now := time.Now().UTC()

	ban1, err := domain.NewBan("ban1", "author1", "videoId", "permanent", 0, now)
	require.NoError(t, err)
	cm.AddBan(ban1)

	ban2, err := domain.NewBan("ban2", "author2", "videoId", "temporary", 5*time.Minute, now)
	require.NoError(t, err)
	cm.AddBan(ban2)

	bans := cm.Bans()
	assert.Len(t, bans, 2)

	// Verify bans exist by ID
	found := false

	for _, ban := range bans {
		if ban.ID() == "ban1" {
			found = true

			assert.Equal(t, domain.Permanent, ban.BanType())
		}
	}

	assert.True(t, found, "ban1 should be found")

	// Adding duplicate should not add (only adds if not exists)
	ban1Updated, err := domain.NewBan("ban1", "author1", "videoId", "temporary", 10*time.Minute, now)
	require.NoError(t, err)
	cm.AddBan(ban1Updated)

	// Still 2, not added
	bans = cm.Bans()
	assert.Len(t, bans, 2)

	// Original ban should still be there
	for _, ban := range bans {
		if ban.ID() == "ban1" {
			assert.Equal(t, domain.Permanent, ban.BanType()) // Original type
		}
	}
}

func TestChatMessages_AddDonate(t *testing.T) {
	t.Parallel()

	cm := domain.NewChatMessages("token")
	now := time.Now().UTC()

	donate1, err := domain.NewDonate("d1", "author1", "videoId", "Great!", "$10", 10000000, "USD", now)
	require.NoError(t, err)
	cm.AddDonate(donate1)

	donate2, err := domain.NewDonate("d2", "author2", "videoId", "Amazing!", "$5", 5000000, "USD", now)
	require.NoError(t, err)
	cm.AddDonate(donate2)

	donates := cm.Donates()
	assert.Len(t, donates, 2)

	// Verify donates exist by ID
	found := false

	for _, donate := range donates {
		if donate.ID() == "d1" {
			found = true

			assert.Equal(t, "$10", donate.Amount())
		}
	}

	assert.True(t, found, "d1 should be found")

	// Adding duplicate should not add (only adds if not exists)
	donate1Updated, err := domain.NewDonate("d1", "author1", "videoId", "Updated!", "$20", 20000000, "USD", now)
	require.NoError(t, err)
	cm.AddDonate(donate1Updated)

	// Still 2, not added
	donates = cm.Donates()
	assert.Len(t, donates, 2)

	// Original donate should still be there
	for _, donate := range donates {
		if donate.ID() == "d1" {
			assert.Equal(t, "$10", donate.Amount()) // Original amount
		}
	}
}

func TestChatMessages_AddAuthor(t *testing.T) {
	t.Parallel()

	cm := domain.NewChatMessages("token")

	author1, err := domain.NewAuthor("author1", "Alice", "https://example.com/1.jpg", true)
	require.NoError(t, err)
	cm.AddAuthor(author1)

	author2, err := domain.NewAuthor("author2", "Bob", "https://example.com/2.jpg", false)
	require.NoError(t, err)
	cm.AddAuthor(author2)

	authors := cm.Authors()
	assert.Len(t, authors, 2)

	// Verify authors exist by ID
	found := false

	for _, author := range authors {
		if author.ID() == "author1" {
			found = true

			assert.Equal(t, "Alice", author.Name())
		}
	}

	assert.True(t, found, "author1 should be found")

	// Adding duplicate should not add (only adds if not exists)
	author1Updated, err := domain.NewAuthor("author1", "Alice Updated", "https://example.com/updated.jpg", false)
	require.NoError(t, err)
	cm.AddAuthor(author1Updated)

	// Still 2, not added
	authors = cm.Authors()
	assert.Len(t, authors, 2)

	// Original author should still be there
	for _, author := range authors {
		if author.ID() == "author1" {
			assert.Equal(t, "Alice", author.Name()) // Original name
		}
	}
}
