package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
)

type RedisLocker struct {
	rs *redsync.Redsync
}

type RedisConfig struct {
	Address  string
	Password string
	DB       int
}

func NewRedisLocker(cfg RedisConfig) (*RedisLocker, error) {
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	pool := goredis.NewPool(client)
	rs := redsync.New(pool)

	return &RedisLocker{rs: rs}, nil
}

type redisUnlocker struct {
	mu *redsync.Mutex
}

func (r *RedisLocker) Lock(ctx context.Context, key string) (Unlocker, error) {
	mu := r.rs.NewMutex(key,
		redsync.WithExpiry(8*time.Second),
		redsync.WithTries(3),
		redsync.WithRetryDelay(500*time.Millisecond),
	)

	if err := mu.LockContext(ctx); err != nil {
		return nil, fmt.Errorf("redis lock failed for %s: %w", key, err)
	}

	return &redisUnlocker{mu: mu}, nil
}

func (u *redisUnlocker) Unlock(ctx context.Context) error {
	ok, err := u.mu.UnlockContext(ctx)
	if err != nil {
		return fmt.Errorf("redis unlock failed: %w", err)
	}
	if !ok {
		return fmt.Errorf("redis unlock returned false, lock may have expired")
	}
	return nil
}
