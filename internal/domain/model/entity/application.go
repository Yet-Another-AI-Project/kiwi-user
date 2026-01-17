package entity

import "github.com/google/uuid"

type ApplicationEntity struct {
	ID   uuid.UUID
	Name string
}
