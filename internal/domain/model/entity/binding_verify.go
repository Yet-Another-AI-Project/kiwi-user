package entity

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"github.com/google/uuid"
)

type BindingVerifyEntity struct {
	ID         uuid.UUID        `json:"uuid.UUID"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
	UserID     string           `json:"user_id"`
	Type       enum.BindingType `json:"type"`
	Identity   string           `json:"identity"`
	Code       string           `json:"code"`
	ExpiresAt  time.Time        `json:"expires_at"`
	VerifiedAt time.Time        `json:"verified_at"`
}
