package entity

import "time"

type StripeEventEntity struct {
	ID          int
	EventID     string
	EventType   string
	Processed   bool
	ProcessedAt time.Time
	CreatedAt   time.Time
}
