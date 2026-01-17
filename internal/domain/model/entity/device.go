package entity

import (
	"time"

	"github.com/google/uuid"
)

type DeviceEntity struct {
	ID                    int64
	DeviceType            string
	DeviceID              string
	OrganizationID        uuid.UUID
	RefreshToken          string
	RefreshTokenExpiresAt time.Time
}
