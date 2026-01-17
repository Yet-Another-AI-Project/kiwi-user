package entity

import (
	"kiwi-user/internal/domain/model/enum"

	"github.com/google/uuid"
)

type OrganizationApplicationEntity struct {
	ID              uuid.UUID
	ApplicationID   uuid.UUID
	Name            string
	Status          string
	TrailDays       int32
	BrandShortName  string
	PrimaryBusiness string
	UsageScenario   string
	ReferrerName    string
	DiscoveryWay    string
	ReviewStatus    enum.OrganizationRequestStatus
	ReviewComment   string
	UserID          string
	OrgRoleName     string
}
