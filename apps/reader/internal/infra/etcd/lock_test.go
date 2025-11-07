//go:build integration

package etcd_test

import (
	"context"
	"testing"

	"github.com/natsoman/youtube-chat-reader/apps/reader/internal/infra/etcd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func TestLocker_TryLock(t *testing.T) {
	t.Parallel()

	t.Run("successfully acquire lock once", func(t *testing.T) {
		t.Parallel()

		sess := newSession(t)
		locker, err := etcd.NewLocker(sess)
		require.NotNil(t, locker)
		require.NoError(t, err)

		// When acquiring lock for first time
		acquired, err := locker.TryLock(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
		assert.True(t, acquired)

		// And
		// When trying to acquire again with same session
		acquired, err = locker.TryLock(t.Context(), t.Name())

		// Then should not be acquired
		assert.NoError(t, err)
		assert.False(t, acquired)
	})

	t.Run("not acquired when lock is already acquired by others", func(t *testing.T) {
		t.Parallel()

		// Given
		otherLocker, err := etcd.NewLocker(newSession(t))
		require.NotNil(t, otherLocker)
		require.NoError(t, err)

		acquired, err := otherLocker.TryLock(t.Context(), t.Name())
		require.True(t, acquired)
		require.NoError(t, err)

		locker, err := etcd.NewLocker(newSession(t))
		require.NotNil(t, locker)
		require.NoError(t, err)

		// When
		acquired, err = locker.TryLock(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
		assert.False(t, acquired)
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		locker, err := etcd.NewLocker(newSession(t))
		require.NotNil(t, locker)
		require.NoError(t, err)

		// When
		acquired, err := locker.TryLock(ctx, t.Name())

		// Then
		assert.Error(t, err)
		assert.False(t, acquired)
	})
}

func TestLocker_Release(t *testing.T) {
	t.Parallel()

	t.Run("successfully release lock", func(t *testing.T) {
		t.Parallel()

		// Given
		locker, err := etcd.NewLocker(newSession(t))
		require.NoError(t, err)
		require.NotNil(t, locker)

		acquired, err := locker.TryLock(t.Context(), t.Name())
		require.NoError(t, err)
		require.True(t, acquired)

		// When
		err = locker.Release(t.Context(), t.Name())

		// Then
		assert.NoError(t, err)
	})

	t.Run("error returns when lock does not exists", func(t *testing.T) {
		t.Parallel()

		// Given
		locker, err := etcd.NewLocker(newSession(t))
		require.NoError(t, err)
		require.NotNil(t, locker)

		// When
		err = locker.Release(t.Context(), t.Name())

		// Then
		assert.EqualError(t, err, "lock not found")
	})

	t.Run("returns error when context is canceled", func(t *testing.T) {
		t.Parallel()

		// Given
		ctx, cancel := context.WithCancel(t.Context())
		cancel() // Cancel the context immediately

		locker, err := etcd.NewLocker(newSession(t))
		require.NotNil(t, locker)
		require.NoError(t, err)

		// When
		err = locker.Release(ctx, t.Name())

		// Then
		assert.Error(t, err)
	})
}

func newSession(t *testing.T) *concurrency.Session {
	t.Helper()

	sess, err := concurrency.NewSession(_client, concurrency.WithTTL(3))
	require.NotNil(t, sess)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sess.Close()
	})

	return sess
}
