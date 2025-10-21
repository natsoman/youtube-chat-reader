//go:build integration

package redis_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	infraredis "github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/redis"
)

func TestNewLocker(t *testing.T) {
	t.Run("successfully creates new locker", func(t *testing.T) {
		locker, err := infraredis.NewLocker(_client)
		assert.NoError(t, err)
		assert.NotNil(t, locker)
	})

	t.Run("returns error when client is nil", func(t *testing.T) {
		locker, err := infraredis.NewLocker(nil)
		assert.EqualError(t, err, "redis client is nil")
		assert.Nil(t, locker)
	})
}

func TestLocker_Lock(t *testing.T) {
	t.Parallel()

	t.Run("successfully acquires lock", func(t *testing.T) {
		t.Parallel()

		// When
		locked, err := _locker.Lock(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
		assert.True(t, locked)

		// Verify the key exists in Redis
		exists, err := _client.Exists(t.Context(), t.Name()).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("fails to acquire lock when already locked", func(t *testing.T) {
		t.Parallel()

		// Given - first lock
		locked, err := _locker.Lock(t.Context(), t.Name())
		require.NoError(t, err)
		require.True(t, locked)

		// When - try to lock again
		locked, err = _locker.Lock(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
		assert.False(t, locked)
	})

	t.Run("sets TTL from context deadline", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		// When
		locked, err := _locker.Lock(ctx, t.Name())

		// Then
		assert.NoError(t, err)
		assert.True(t, locked)

		// Verify TTL is set
		ttl, err := _client.TTL(t.Context(), t.Name()).Result()
		assert.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, 2*time.Second)
	})

	t.Run("returns error when context deadline already exceeded", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
		defer cancel()

		// When
		locked, err := _locker.Lock(ctx, t.Name())

		// Then
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.False(t, locked)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		// When
		locked, err := _locker.Lock(ctx, t.Name())

		// Then
		assert.Error(t, err)
		assert.False(t, locked)
	})
}

func TestLocker_Release(t *testing.T) {
	t.Parallel()

	t.Run("successfully releases lock", func(t *testing.T) {
		t.Parallel()

		// Given - acquire lock first
		locked, err := _locker.Lock(t.Context(), t.Name())
		require.NoError(t, err)
		require.True(t, locked)

		// When
		err = _locker.Release(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)

		// Verify the key no longer exists
		exists, err := _client.Exists(t.Context(), t.Name()).Result()
		assert.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("successfully releases non-existent lock", func(t *testing.T) {
		t.Parallel()

		// When
		err := _locker.Release(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel immediately

		// When
		err := _locker.Release(ctx, t.Name())

		// Then
		assert.Error(t, err)
	})
}

func TestLocker_ConcurrentLocking(t *testing.T) {
	t.Parallel()

	t.Run("only one goroutine acquires lock", func(t *testing.T) {
		t.Parallel()

		const numGoroutines = 10
		lockAcquired := make(chan bool, numGoroutines)

		// When multiple goroutines try to acquire the same lock
		for i := 0; i < numGoroutines; i++ {
			go func() {
				locked, err := _locker.Lock(t.Context(), "concurrent-key")
				if err == nil && locked {
					lockAcquired <- true
				} else {
					lockAcquired <- false
				}
			}()
		}

		// Then
		acquiredCount := 0
		for i := 0; i < numGoroutines; i++ {
			if <-lockAcquired {
				acquiredCount++
			}
		}

		// Only one should have acquired the lock
		assert.Equal(t, 1, acquiredCount)
	})
}
