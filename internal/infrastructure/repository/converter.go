package repository

import (
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/repository/ent"
	"time"
)

func convertApplicationDOToEntity(application *ent.Application) *entity.ApplicationEntity {
	return &entity.ApplicationEntity{
		ID:   application.ID,
		Name: application.Name,
	}
}

func convertRolesDOToEntities(roles []*ent.Role) []*entity.RoleEntity {
	var entities []*entity.RoleEntity

	for _, role := range roles {
		entities = append(entities, convertRoleDOToEntity(role))
	}

	return entities
}

func convertRoleDOToEntity(role *ent.Role) *entity.RoleEntity {
	if role == nil {
		return nil
	}
	return &entity.RoleEntity{
		ID:   role.ID,
		Type: enum.ParseRoleType(role.Type.String()),
		Name: role.Name,
	}
}

func convertRoleDOsToEntities(roles []*ent.Role) []*entity.RoleEntity {
	var entities []*entity.RoleEntity

	for _, role := range roles {
		entities = append(entities, convertRoleDOToEntity(role))
	}

	return entities
}

func convertScopeDOsToEntities(scopes []*ent.Scope) []*entity.ScopeEntity {
	var entities []*entity.ScopeEntity

	for _, scope := range scopes {
		entities = append(entities, convertScopeDOToEntity(scope))
	}

	return entities
}

func convertScopeDOToEntity(scope *ent.Scope) *entity.ScopeEntity {
	return &entity.ScopeEntity{
		ID:   scope.ID,
		Name: scope.Name,
	}
}

func convertBindingDOToEntity(binding *ent.Binding) *entity.BindingEntity {
	return &entity.BindingEntity{
		ID:       binding.ID,
		Type:     enum.ParseBindingType(binding.Type.String()),
		Identity: binding.Identity,
		Email:    binding.Email,
		Verified: binding.Verified,
		Salt:     binding.Salt,
	}
}

func converBindingDOsToEntities(bindings []*ent.Binding) []*entity.BindingEntity {
	var entities []*entity.BindingEntity

	for _, binding := range bindings {
		entities = append(entities, convertBindingDOToEntity(binding))
	}

	return entities
}

func convertUserDOToEntity(user *ent.User) *entity.UserEntity {
	return &entity.UserEntity{
		ID:              user.ID,
		Name:            user.Name,
		DisplayName:     user.DisplayName,
		Avatar:          user.Avatar,
		RefferalChannel: user.ReferralChannel,
		Department:      user.Department,
	}
}

func convertDeviceDOToEntity(device *ent.Device) *entity.DeviceEntity {
	return &entity.DeviceEntity{
		ID:                    device.ID,
		DeviceType:            device.DeviceType,
		DeviceID:              device.DeviceID,
		OrganizationID:        device.OrganizationID,
		RefreshToken:          device.RefreshToken,
		RefreshTokenExpiresAt: device.RefreshTokenExpiresAt,
	}
}

func convertOrganizationDOToEntity(organization *ent.Organization) *entity.OrganizationEntity {
	if organization.ExpiresAt.Unix() < 0 {
		organization.ExpiresAt = time.Unix(0, 0)
	}

	if organization.RefreshAt.Unix() < 0 {
		organization.RefreshAt = time.Unix(0, 0)
	}

	return &entity.OrganizationEntity{
		ID:             organization.ID,
		Name:           organization.Name,
		Status:         enum.ParseOrganizationStatus(organization.Status.String()),
		PermissionCode: organization.PermissionCode,
		RefreshAt:      organization.RefreshAt,
		ExpiresAt:      organization.ExpiresAt,
		LogoImageURL:   organization.LogoURL,
	}
}

func convertBindingVerifyDOToEntity(bindingVerify *ent.BindingVerify) *entity.BindingVerifyEntity {
	if bindingVerify.ExpiresAt.Unix() < 0 {
		bindingVerify.ExpiresAt = time.Unix(0, 0)
	}
	bindingVerifyEntity := &entity.BindingVerifyEntity{
		ID:         bindingVerify.ID,
		CreatedAt:  bindingVerify.CreatedAt,
		UpdatedAt:  bindingVerify.UpdatedAt,
		Type:       enum.ParseBindingType(bindingVerify.Type.String()),
		Identity:   bindingVerify.Identity,
		Code:       bindingVerify.Code,
		ExpiresAt:  bindingVerify.ExpiresAt,
		VerifiedAt: bindingVerify.VerifiedAt,
	}

	return bindingVerifyEntity
}

func convertWechatOpenIDDOToEntity(wechatOpenID *ent.WechatOpenID) *entity.WechatOpenIDEntity {
	return &entity.WechatOpenIDEntity{
		ID:       wechatOpenID.ID,
		OpenID:   wechatOpenID.OpenID,
		Platform: enum.ParseWechatOpenIDPlatform(wechatOpenID.Platform.String()),
	}
}

func convertWechatOpenIDDOToEntities(openIDs []*ent.WechatOpenID) []*entity.WechatOpenIDEntity {
	var entities []*entity.WechatOpenIDEntity

	for _, openID := range openIDs {
		entities = append(entities, convertWechatOpenIDDOToEntity(openID))
	}

	return entities
}

func convertQyWechatUserIDDOToEntity(qyWechatUserID *ent.QyWechatUserID) *entity.QyWechatUserIDEntity {
	return &entity.QyWechatUserIDEntity{
		ID:             qyWechatUserID.ID,
		QyWechatUserID: qyWechatUserID.QyWechatUserID,
		OpenID:         qyWechatUserID.OpenID,
	}
}

func convertQyWechatUserIDDOToEntities(qyWechatUserIDs []*ent.QyWechatUserID) []*entity.QyWechatUserIDEntity {
	var entities []*entity.QyWechatUserIDEntity

	for _, qyWechatUserID := range qyWechatUserIDs {
		entities = append(entities, convertQyWechatUserIDDOToEntity(qyWechatUserID))
	}

	return entities
}

func convertPaymentDOToEntity(payment *ent.Payment) *entity.PaymentEntity {
	return &entity.PaymentEntity{
		OutTradeNo: payment.OutTradeNo,
		UserID:     payment.UserID,
		ChannelInfo: entity.PaymentChannelInfo{
			Channel:                  enum.ParsePaymentChannel(payment.Channel.String()),
			WechatPlatform:           enum.ParseWechatOpenIDPlatform(payment.WechatPlatform),
			WeChatTransactionID:      payment.WechatTransactionID,
			WeChatOpenID:             payment.WechatOpenID,
			StripeSubscriptionID:     payment.StripeSubscriptionID,
			StripeSubscriptionStatus: enum.ParseSubscriptionStatus(payment.StripeSubscriptionStatus),
			StripeInterval:           enum.ParseSubscriptionInterval(payment.StripeInterval),
			StripeCurrentPeriodStart: payment.StripeCurrentPeriodStart,
			StripeCurrentPeriodEnd:   payment.StripeCurrentPeriodEnd,
			StripeCustomerID:         payment.StripeCustomerID,
			StripeCustomerEmail:      payment.StripeCustomerEmail,
			StripeCheckoutSessionID:  payment.StripeCheckoutSessionID,
		},
		Service:     payment.Service,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		Description: payment.Description,
		Status:      enum.ParsePaymentStatus(payment.Status.String()),
		CreatedAt:   payment.CreatedAt,
		UpdatedAt:   payment.UpdatedAt,
		PaidAt:      payment.PaidAt,
		PaymentType: enum.ParsePaymentType(payment.PaymentType.String()),
	}
}

func convertOrganizationApplicationDoToEntity(organizationApplication *ent.OrganizationApplication) *entity.OrganizationApplicationEntity {
	return &entity.OrganizationApplicationEntity{
		ID:              organizationApplication.ID,
		ApplicationID:   organizationApplication.ApplicationID,
		Name:            organizationApplication.Name,
		Status:          organizationApplication.Status.String(),
		TrailDays:       organizationApplication.TrialDays,
		BrandShortName:  organizationApplication.BrandShortName,
		PrimaryBusiness: organizationApplication.PrimaryBusiness,
		UsageScenario:   organizationApplication.UsageScenario,
		DiscoveryWay:    organizationApplication.DiscoveryWay,
		ReferrerName:    organizationApplication.ReferrerName,
		ReviewStatus:    enum.OrganizationRequestStatus(organizationApplication.ReviewStatus),
		ReviewComment:   organizationApplication.ReviewComment,
		UserID:          organizationApplication.UserID,
		OrgRoleName:     organizationApplication.OrgRoleName,
	}
}
