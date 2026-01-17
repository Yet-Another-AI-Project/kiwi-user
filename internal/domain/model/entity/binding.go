package entity

import (
	"kiwi-user/internal/domain/model/enum"

	"github.com/google/uuid"
)

type BindingEntity struct {
	ID            uuid.UUID
	ApplicationID uuid.UUID
	Type          enum.BindingType
	Identity      string
	Verified      bool
	Salt          string
}
