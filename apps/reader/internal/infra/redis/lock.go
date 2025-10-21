package redis

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker struct {
	log    *slog.Logger
	client *redis.ClusterClient
}

func NewLocker(client *redis.ClusterClient) (*Locker, error) {
	if client == nil {
		return nil, errors.New("redis client is nil")
	}

	return &Locker{log: slog.Default().With("cmp", "locker"), client: client}, nil
}

func (l *Locker) Lock(ctx context.Context, key string) (bool, error) {
	var ttl time.Duration

	deadline, ok := ctx.Deadline()
	if ok {
		ttl = time.Until(deadline)
		if ttl <= time.Millisecond {
			return false, context.DeadlineExceeded
		}
	}

	ok, err := l.client.SetNX(ctx, key, nil, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("set nx command: %v", err)
	}

	return ok, nil
}

func (l *Locker) Release(ctx context.Context, key string) error {
	if _, err := l.client.Del(ctx, key).Result(); err != nil {
		return fmt.Errorf("del command: %v", err)
	}

	return nil
}
