package api

import (
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/facade/dto"
)

func convertOrganizationAggregateToDTO(organizationAggregate *aggregate.OrganizationAggregate) *dto.Organization {
	organization := &dto.Organization{
		ID:             organizationAggregate.Organization.ID.String(),
		Application:    organizationAggregate.Application.Name,
		Name:           organizationAggregate.Organization.Name,
		Status:         organizationAggregate.Organization.Status.String(),
		PermissionCode: organizationAggregate.Organization.PermissionCode,
		RefreshAt:      organizationAggregate.Organization.RefreshAt.Unix(),
		ExpiresAt:      organizationAggregate.Organization.ExpiresAt.Unix(),
		LogoImageURL:   organizationAggregate.Organization.LogoImageURL,
	}

	return organization
}

func convertOrganizationAggregateToBasicInfos(organizationAggregate *aggregate.OrganizationAggregate) *dto.BasicInfos {
	if organizationAggregate == nil {
		return &dto.BasicInfos{}
	}
	basicInfos := &dto.BasicInfos{
		LogoImageURL: organizationAggregate.Organization.LogoImageURL,
	}

	return basicInfos
}
