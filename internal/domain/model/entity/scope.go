package entity

import "github.com/google/uuid"

type ScopeEntity struct {
	ID          uuid.UUID
	Name        string
	HiddenInJWT bool
}
