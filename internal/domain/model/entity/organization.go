package entity

import (
	"kiwi-user/internal/domain/model/enum"
	"time"

	"github.com/google/uuid"
)

type OrganizationEntity struct {
	ID             uuid.UUID
	Name           string
	Status         enum.OrganizationStatus
	PermissionCode string
	RefreshAt      time.Time
	ExpiresAt      time.Time
	LogoImageURL   string
}
