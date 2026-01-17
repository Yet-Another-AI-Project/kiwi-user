package admin

import (
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/facade/dto"
)

func convertApplicateionAggregateToDTO(applicationAggregate *aggregate.ApplicationAggregate, roleAggregates []*aggregate.RoleAggregate) *dto.Application {
	application := &dto.Application{
		Name: applicationAggregate.Application.Name,
	}

	roles := make([]*dto.Role, 0)
	for _, role := range roleAggregates {
		scopes := make([]string, 0)
		for _, scope := range role.Scopes {
			scopes = append(scopes, scope.Name)
		}

		roles = append(roles, &dto.Role{
			Name:   role.Role.Name,
			Type:   role.Role.Type.String(),
			Scopes: scopes,
		})

		if applicationAggregate.DefaultPersonalRole != nil && applicationAggregate.DefaultPersonalRole.Name == role.Role.Name {
			application.DefaultPersonalRole = role.Role.Name
		}

		if applicationAggregate.DefaultOrgRole != nil && applicationAggregate.DefaultOrgRole.Name == role.Role.Name {
			application.DefaultOrganizationRole = role.Role.Name
		}

		if applicationAggregate.DefaultOrgAdminRole != nil && applicationAggregate.DefaultOrgAdminRole.Name == role.Role.Name {
			application.DefaultOrganizationAdminRole = role.Role.Name
		}
	}

	application.Roles = roles

	return application
}

func convertOrganizationAggregateToDTO(organizationAggregate *aggregate.OrganizationAggregate) *dto.OrganizationV2 {
	organization := &dto.OrganizationV2{
		ID:           organizationAggregate.Organization.ID.String(),
		Application:  organizationAggregate.Application.Name,
		Name:         organizationAggregate.Organization.Name,
		Version:      organizationAggregate.Organization.Status.String(),
		RefreshAt:    organizationAggregate.Organization.RefreshAt.Unix(),
		ExpiresAt:    organizationAggregate.Organization.ExpiresAt.Unix(),
		LogoImageURL: organizationAggregate.Organization.LogoImageURL,
	}

	return organization
}

func convertOrganizationApplicationAggregateToDTO(organizationApplicationAggregate *aggregate.OrganizationApplicationAggregate) *dto.OrganizationApplication {
	organizationApplication := &dto.OrganizationApplication{
		ID:              organizationApplicationAggregate.OrganizationApplication.ID.String(),
		ApplicationID:   organizationApplicationAggregate.OrganizationApplication.ApplicationID.String(),
		Name:            organizationApplicationAggregate.OrganizationApplication.Name,
		Status:          organizationApplicationAggregate.OrganizationApplication.Status,
		TrailDays:       organizationApplicationAggregate.OrganizationApplication.TrailDays,
		BrandShortName:  organizationApplicationAggregate.OrganizationApplication.BrandShortName,
		PrimaryBusiness: organizationApplicationAggregate.OrganizationApplication.PrimaryBusiness,
		UsageScenario:   organizationApplicationAggregate.OrganizationApplication.UsageScenario,
		ReferrerName:    organizationApplicationAggregate.OrganizationApplication.ReferrerName,
		DiscoveryWay:    organizationApplicationAggregate.OrganizationApplication.DiscoveryWay,
		ReviewStatus:    organizationApplicationAggregate.OrganizationApplication.ReviewStatus,
		ReviewComment:   organizationApplicationAggregate.OrganizationApplication.ReviewComment,
		UserID:          organizationApplicationAggregate.OrganizationApplication.UserID,
		OrgRoleName:     organizationApplicationAggregate.OrganizationApplication.OrgRoleName,
	}

	return organizationApplication
}
