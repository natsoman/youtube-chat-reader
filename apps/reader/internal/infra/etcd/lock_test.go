//go:build integration

package etcd_test

import (
	"testing"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/etcd"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocker_LockAndRelease(t *testing.T) {
	locker, err := etcd.NewLocker(_client)
	require.NotNil(t, locker)
	require.NoError(t, err)
	defer locker.Close()

	// Acquire lock for first time
	acquired, err := locker.Lock(t.Context(), "test-key")
	require.NoError(t, err)
	assert.True(t, acquired)

	// Try to acquire again (should fail)
	acquired, err = locker.Lock(t.Context(), "test-key")
	require.NoError(t, err)
	assert.False(t, acquired)

	// Release
	require.NoError(t, locker.Release(t.Context(), "test-key"))

	// And try lock again
	acquired, err = locker.Lock(t.Context(), "test-key")
	require.NoError(t, err)
	assert.True(t, acquired)
}
