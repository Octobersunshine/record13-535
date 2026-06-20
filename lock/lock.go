package lock

import (
	"context"
	"fmt"
	"time"
)

type Locker interface {
	Lock(ctx context.Context, key string) (Unlocker, error)
}

type Unlocker interface {
	Unlock(ctx context.Context) error
}

const (
	lockKeyPrefix         = "ticket:lock:"
	defaultTryLockTimeout = 3 * time.Second
)

func SlotLockKey(slotID string) string {
	return fmt.Sprintf("%s%s", lockKeyPrefix, slotID)
}

func ReservationLockKey(reservationID string) string {
	return fmt.Sprintf("%sres:%s", lockKeyPrefix, reservationID)
}
