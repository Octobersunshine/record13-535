package lock

import (
	"context"
	"fmt"
	"sync"
)

type MemoryLocker struct {
	mu    sync.Mutex
	locks map[string]*sync.Mutex
}

func NewMemoryLocker() *MemoryLocker {
	return &MemoryLocker{
		locks: make(map[string]*sync.Mutex),
	}
}

type memoryUnlocker struct {
	mu  *sync.Mutex
	key string
	ml  *MemoryLocker
}

func (m *MemoryLocker) Lock(ctx context.Context, key string) (Unlocker, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	m.mu.Lock()
	mu, ok := m.locks[key]
	if !ok {
		mu = &sync.Mutex{}
		m.locks[key] = mu
	}
	m.mu.Unlock()

	mu.Lock()
	return &memoryUnlocker{mu: mu, key: key, ml: m}, nil
}

func (u *memoryUnlocker) Unlock(ctx context.Context) error {
	u.mu.Unlock()

	u.ml.mu.Lock()
	if mu, ok := u.ml.locks[u.key]; ok && mu == u.mu {
		delete(u.ml.locks, u.key)
	}
	u.ml.mu.Unlock()

	return nil
}

func TryLock(ctx context.Context, locker Locker, key string) (Unlocker, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTryLockTimeout)
	defer cancel()

	unlocker, err := locker.Lock(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for %s: %w", key, err)
	}
	return unlocker, nil
}
