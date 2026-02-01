package entity

import "time"

type StripeEventEntity struct {
	ID             int
	EventID        string
	EventType      string
	SubscriptionID string
	UserID         string
	Processed      bool
	ProcessedAt    time.Time
	CreatedAt      time.Time
}
