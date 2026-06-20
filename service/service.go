package service

import (
	"fmt"
	"ticket-reservation/model"
	"ticket-reservation/store"
	"time"
)

type ReservationService struct {
	store *store.Store
}

func NewReservationService(s *store.Store) *ReservationService {
	return &ReservationService{store: s}
}

type CreateSlotRequest struct {
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Total     int    `json:"total"`
}

type ReserveRequest struct {
	SlotID   string `json:"slot_id"`
	Quantity int    `json:"quantity"`
	UserID   string `json:"user_id"`
}

type CancelRequest struct {
	ReservationID string `json:"reservation_id"`
}

func (svc *ReservationService) CreateSlot(req CreateSlotRequest) (*model.TimeSlot, error) {
	if err := validateTimeFormat(req.Date, req.StartTime, req.EndTime); err != nil {
		return nil, err
	}
	if req.Total <= 0 {
		return nil, fmt.Errorf("total must be positive")
	}

	slot := &model.TimeSlot{
		ID:        fmt.Sprintf("slot-%s-%s-%s", req.Date, req.StartTime, req.EndTime),
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Total:     req.Total,
		Reserved:  0,
		CreatedAt: time.Now(),
	}

	if err := svc.store.CreateSlot(slot); err != nil {
		return nil, err
	}
	return slot, nil
}

func (svc *ReservationService) GetSlot(slotID string) (*model.TimeSlot, error) {
	slot, ok := svc.store.GetSlot(slotID)
	if !ok {
		return nil, fmt.Errorf("time slot not found: %s", slotID)
	}
	return slot, nil
}

func (svc *ReservationService) ListSlots() []model.TimeSlot {
	return svc.store.ListSlots()
}

func (svc *ReservationService) Reserve(req ReserveRequest) (*model.Reservation, error) {
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if req.SlotID == "" {
		return nil, fmt.Errorf("slot_id is required")
	}
	if req.UserID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	return svc.store.Reserve(req.SlotID, req.Quantity, req.UserID)
}

func (svc *ReservationService) Cancel(reservationID string) error {
	if reservationID == "" {
		return fmt.Errorf("reservation_id is required")
	}
	return svc.store.CancelReservation(reservationID)
}

func validateTimeFormat(date, start, end string) error {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %s", date)
	}
	if _, err := time.Parse("15:04", start); err != nil {
		return fmt.Errorf("invalid start_time format, expected HH:MM: %s", start)
	}
	if _, err := time.Parse("15:04", end); err != nil {
		return fmt.Errorf("invalid end_time format, expected HH:MM: %s", end)
	}
	if start >= end {
		return fmt.Errorf("start_time must be before end_time")
	}
	return nil
}
