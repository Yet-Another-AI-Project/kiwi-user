package application

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/jwt"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

func getUserInfo(
	ctx context.Context,
	organizationID string,
	userAggregate *aggregate.UserAggregate,
	roleReadRepository contract.IRoleReadRepository,
	organizationUserReadRepository contract.IOrganizationUserReadRepository) (*dto.UserInfo, *facade.Error) {

	userInfo := &dto.UserInfo{
		UserID:      userAggregate.User.ID,
		Name:        userAggregate.User.Name,
		Avatar:      userAggregate.User.Avatar,
		Application: userAggregate.Application.Name,
		Department:  userAggregate.User.Department,
	}

	// find phone if have
	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePhone {
			userInfo.Phone = binding.Identity
		}
	}

	// NOTE: override name if displayname is not empty, for UI back compatibility
	// TODO: return both name and displayname in response
	if userAggregate.User.DisplayName != "" {
		userInfo.Name = userAggregate.User.DisplayName
	}

	if userAggregate.PersonalRole != nil {
		roleAggregate, err := roleReadRepository.FindByName(ctx, userAggregate.Application.Name, userAggregate.PersonalRole.Name)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		scopes := make([]string, 0)
		for _, scope := range roleAggregate.Scopes {
			scopes = append(scopes, scope.Name)
		}

		userInfo.PersonalRole = userAggregate.PersonalRole.Name
		userInfo.PersonalScopes = scopes
	}

	// get organizations info
	orgs, err := organizationUserReadRepository.FindAll(ctx, userAggregate.User.ID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	for _, org := range orgs {
		orgRoleAggregate, err := roleReadRepository.FindByName(ctx, org.Application.Name, org.OrganizationRole.Name)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		orgScopes := make([]string, 0)
		for _, scope := range orgRoleAggregate.Scopes {
			orgScopes = append(orgScopes, scope.Name)
		}

		orgInfo := &dto.OrganizationUser{
			ID:                 org.Organization.ID.String(),
			Name:               org.Organization.Name,
			Status:             org.Organization.Status.String(),
			PermissionCode:     org.Organization.PermissionCode,
			RefreshAt:          org.Organization.RefreshAt.Unix(),
			ExpiresAt:          org.Organization.ExpiresAt.Unix(),
			OrganizationRole:   org.OrganizationRole.Name,
			OrganizationScopes: orgScopes,
		}

		userInfo.Orgs = append(userInfo.Orgs, orgInfo)
		userInfo.CurrentOrgID = organizationID
	}

	return userInfo, nil
}

func generateLoginResult(
	ctx context.Context,
	user *aggregate.UserAggregate,
	deviceEntity *entity.DeviceEntity,
	rbacService *service.RBACService,
	jwthelper *jwt.JWTHelper) (*dto.LoginResponse, error) {

	roleName := ""
	scopes := make([]string, 0)

	if user.PersonalRole != nil {
		roleName = user.PersonalRole.Name

		roleAggregate, err := rbacService.GetRole(ctx, user.Application.Name, user.PersonalRole.Name)

		if err != nil {
			return nil, xerror.Wrap(err)
		}

		for _, scope := range roleAggregate.Scopes {
			scopes = append(scopes, scope.Name)
		}
	}

	orgnizationID := ""

	if deviceEntity.OrganizationID != uuid.Nil {
		orgnizationID = deviceEntity.OrganizationID.String()
	}

	// generate jwt token
	up := jwthelper.NewAccessPayload(
		user.User.ID,
		roleName,
		scopes,
		user.Application.Name,
		deviceEntity.DeviceType,
		deviceEntity.DeviceID,
		orgnizationID)

	accessToken, err := jwthelper.GenerateRSA256JWT(up)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	result := &dto.LoginResponse{
		Type:                 "Bearer",
		AccessToken:          accessToken.String(),
		AccessTokenExpiresAt: up.Expire,
		UserID:               user.User.ID,
	}

	if deviceEntity != nil {
		result.DeviceID = deviceEntity.DeviceID
		result.DeviceType = deviceEntity.DeviceType
		result.RefreshToken = deviceEntity.RefreshToken
		result.RefreshTokenExpiresAt = deviceEntity.RefreshTokenExpiresAt.Unix()
	}

	return result, nil
}
