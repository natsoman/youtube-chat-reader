package etcd

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"go.etcd.io/etcd/client/v3/concurrency"
)

// Locker implements distributed locking using etcd
type Locker struct {
	session *concurrency.Session
	mu      sync.Mutex
	mutexes map[string]*concurrency.Mutex
}

// NewLocker creates a new distributed lock that relies on Etcd lease
func NewLocker(session *concurrency.Session) (*Locker, error) {
	if session == nil {
		return nil, errors.New("session is nil")
	}

	return &Locker{
		session: session,
		mutexes: make(map[string]*concurrency.Mutex),
	}, nil
}

func (l *Locker) TryLock(ctx context.Context, key string) (bool, error) {
	timeCtx, cancel := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancel()

	l.mu.Lock()
	defer l.mu.Unlock()

	mutex := concurrency.NewMutex(l.session, "/locks/"+key)

	// Try to acquire the lock
	if err := mutex.TryLock(timeCtx); err != nil {
		if errors.Is(err, concurrency.ErrLocked) {
			return false, nil
		}

		return false, fmt.Errorf("try lock: %v", err)
	}

	storedMutex, exists := l.mutexes[key]
	if exists {
		if mutex.Key() == storedMutex.Key() {
			// Locker has been already acquired by this session
			l.mutexes[key] = mutex
			return false, nil
		}
	}

	// Store the mutex for later release
	l.mutexes[key] = mutex

	return true, nil
}

func (l *Locker) Release(ctx context.Context, key string) error {
	timeCtx, cancel := context.WithTimeout(ctx, time.Millisecond*50)
	defer cancel()

	l.mu.Lock()
	defer l.mu.Unlock()

	mutex, exists := l.mutexes[key]
	if !exists {
		return errors.New("lock not found")
	}

	if err := mutex.Unlock(timeCtx); err != nil {
		return fmt.Errorf("release lock: %v", err)
	}

	delete(l.mutexes, key)

	return nil
}
