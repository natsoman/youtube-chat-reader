package etcd

import (
	"context"
	"errors"
	"fmt"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// Locker implements distributed locking using etcd
type Locker struct {
	client  *clientv3.Client
	session *concurrency.Session
	mu      sync.Mutex
	locks   map[string]*concurrency.Mutex
}

// NewLocker creates a new distributed locker instance
func NewLocker(client *clientv3.Client) (*Locker, error) {
	session, err := concurrency.NewSession(client, concurrency.WithTTL(3))
	if err != nil {
		return nil, fmt.Errorf("create session: %v", err)
	}

	return &Locker{
		client:  client,
		session: session,
		locks:   make(map[string]*concurrency.Mutex),
	}, nil
}

// Lock attempts to acquire a lock for the given key. If lock is already acquired it
// waits til context timeout. Returns true if the lock was acquired, false if already locked,
// and an error if something went wrong
func (l *Locker) Lock(ctx context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, exists := l.locks[key]; exists {
		return false, nil
	}

	mutex := concurrency.NewMutex(l.session, "/locks/"+key)

	// Try to acquire the lock
	if err := mutex.Lock(ctx); err != nil {
		if err == context.DeadlineExceeded {
			return false, nil
		}

		return false, fmt.Errorf("lock: %v", err)
	}

	// Store the mutex for later release
	l.locks[key] = mutex

	return true, nil
}

// Release releases the lock for the given key
func (l *Locker) Release(ctx context.Context, key string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	mutex, exists := l.locks[key]
	if !exists {
		return errors.New("lock not found")
	}

	if err := mutex.Unlock(ctx); err != nil {
		return fmt.Errorf("release lock: %v", err)
	}

	delete(l.locks, key)

	return nil
}

func (l *Locker) Close() error {
	return l.session.Close()
}
