package entity

import (
	"kiwi-user/internal/domain/model/enum"

	"github.com/google/uuid"
)

type OrganizationRequestEntity struct {
	ID               uuid.UUID
	Type             enum.OrganizationRequestType
	Status           enum.OrganizationRequestStatus
	ApplicationID    uuid.UUID
	UserID           string
	OrganizationID   uuid.UUID
	OrganizationName string
}
