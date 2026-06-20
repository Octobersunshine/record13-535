package model

import "time"

type TimeSlot struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	Total     int       `json:"total"`
	Reserved  int       `json:"reserved"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *TimeSlot) Remaining() int {
	return s.Total - s.Reserved
}

type Reservation struct {
	ID        string    `json:"id"`
	SlotID    string    `json:"slot_id"`
	Quantity  int       `json:"quantity"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}
