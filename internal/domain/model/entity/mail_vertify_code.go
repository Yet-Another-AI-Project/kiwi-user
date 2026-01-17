package entity

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"github.com/google/uuid"
)

type MailVertifyCodeEntity struct {
	ID        uuid.UUID
	Email     string
	Code      string
	Type      enum.VertificationCodeType
	ExpiresAt time.Time
}
