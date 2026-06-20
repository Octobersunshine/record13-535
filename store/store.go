package store

import (
	"context"
	"fmt"
	"sync"
	"ticket-reservation/lock"
	"ticket-reservation/model"
	"time"
)

var timeNow = time.Now

type Store struct {
	mu           sync.RWMutex
	slots        map[string]*model.TimeSlot
	reservations map[string]*model.Reservation
	locker       lock.Locker
}

func New(locker lock.Locker) *Store {
	return &Store{
		slots:        make(map[string]*model.TimeSlot),
		reservations: make(map[string]*model.Reservation),
		locker:       locker,
	}
}

func (s *Store) CreateSlot(slot *model.TimeSlot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := slotKey(slot.Date, slot.StartTime, slot.EndTime)
	for _, existing := range s.slots {
		ek := slotKey(existing.Date, existing.StartTime, existing.EndTime)
		if ek == key {
			return fmt.Errorf("time slot already exists: %s %s-%s", slot.Date, slot.StartTime, slot.EndTime)
		}
		if existing.Date == slot.Date && overlapping(existing, slot) {
			return fmt.Errorf("time slot overlaps with existing slot: %s %s-%s", existing.Date, existing.StartTime, existing.EndTime)
		}
	}

	s.slots[slot.ID] = slot
	return nil
}

func (s *Store) GetSlot(slotID string) (*model.TimeSlot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slot, ok := s.slots[slotID]
	if !ok {
		return nil, false
	}
	cp := *slot
	return &cp, true
}

func (s *Store) ListSlots() []model.TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]model.TimeSlot, 0, len(s.slots))
	for _, slot := range s.slots {
		result = append(result, *slot)
	}
	return result
}

func (s *Store) Reserve(slotID string, quantity int, userID string) (*model.Reservation, error) {
	ctx := context.Background()
	lockKey := lock.SlotLockKey(slotID)

	unlocker, err := lock.TryLock(ctx, s.locker, lockKey)
	if err != nil {
		return nil, fmt.Errorf("reservation failed: %w", err)
	}
	defer unlocker.Unlock(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	slot, ok := s.slots[slotID]
	if !ok {
		return nil, fmt.Errorf("time slot not found: %s", slotID)
	}

	remaining := slot.Total - slot.Reserved
	if quantity > remaining {
		return nil, fmt.Errorf("insufficient quota: requested %d, remaining %d", quantity, remaining)
	}

	slot.Reserved += quantity

	reservation := &model.Reservation{
		ID:        fmt.Sprintf("res-%d", timeNow().UnixNano()),
		SlotID:    slotID,
		Quantity:  quantity,
		UserID:    userID,
		CreatedAt: timeNow(),
	}
	s.reservations[reservation.ID] = reservation
	return reservation, nil
}

func (s *Store) CancelReservation(reservationID string) error {
	ctx := context.Background()

	s.mu.RLock()
	res, ok := s.reservations[reservationID]
	if !ok {
		s.mu.RUnlock()
		return fmt.Errorf("reservation not found: %s", reservationID)
	}
	slotID := res.SlotID
	s.mu.RUnlock()

	lockKey := lock.SlotLockKey(slotID)
	unlocker, err := lock.TryLock(ctx, s.locker, lockKey)
	if err != nil {
		return fmt.Errorf("cancel failed: %w", err)
	}
	defer unlocker.Unlock(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()

	res, ok = s.reservations[reservationID]
	if !ok {
		return fmt.Errorf("reservation not found: %s", reservationID)
	}

	slot, ok := s.slots[res.SlotID]
	if !ok {
		return fmt.Errorf("time slot not found: %s", res.SlotID)
	}

	slot.Reserved -= res.Quantity
	if slot.Reserved < 0 {
		slot.Reserved = 0
	}

	delete(s.reservations, reservationID)
	return nil
}

func (s *Store) GetReservationsBySlot(slotID string) []model.Reservation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []model.Reservation
	for _, res := range s.reservations {
		if res.SlotID == slotID {
			result = append(result, *res)
		}
	}
	return result
}

func slotKey(date, start, end string) string {
	return fmt.Sprintf("%s|%s|%s", date, start, end)
}

func overlapping(a, b *model.TimeSlot) bool {
	if a.Date != b.Date {
		return false
	}
	return a.StartTime < b.EndTime && b.StartTime < a.EndTime
}
