package entity

import (
	"kiwi-user/internal/domain/model/enum"

	"github.com/google/uuid"
)

type RoleEntity struct {
	ID   uuid.UUID
	Type enum.RoleType
	Name string
}
