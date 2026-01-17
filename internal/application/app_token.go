package application

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/jwt"
	"time"

	"github.com/futurxlab/golanggraph/logger"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

type TokenApplication struct {
	deviceService *service.DeviceService
	rbacService   *service.RBACService

	rsa       *jwt.RSA
	jwthelper *jwt.JWTHelper

	deviceReadRepository           contract.IDeviceReadRepository
	userReadRepository             contract.IUserReadRepository
	roleReadRepository             contract.IRoleReadRepository
	organizationUserReadRepository contract.IOrganizationUserReadRepository

	posthogClient posthog.Client
	logger        logger.ILogger
}

func NewTokenApplication(
	rsa *jwt.RSA,
	jwthelper *jwt.JWTHelper,
	deviceService *service.DeviceService,
	rbacService *service.RBACService,
	deviceReadRepository contract.IDeviceReadRepository,
	userReadRepository contract.IUserReadRepository,
	organizationUserReadRepository contract.IOrganizationUserReadRepository,
	roleReadRepository contract.IRoleReadRepository,
	posthogClient posthog.Client,
	logger logger.ILogger) *TokenApplication {
	return &TokenApplication{
		rsa:                            rsa,
		jwthelper:                      jwthelper,
		deviceService:                  deviceService,
		rbacService:                    rbacService,
		deviceReadRepository:           deviceReadRepository,
		userReadRepository:             userReadRepository,
		roleReadRepository:             roleReadRepository,
		organizationUserReadRepository: organizationUserReadRepository,
		posthogClient:                  posthogClient,
		logger:                         logger,
	}
}

func (t *TokenApplication) GetPublicKey(ctx context.Context) (string, *facade.Error) {
	publicKey, err := t.rsa.GetPublicKey()
	if err != nil {
		return "", facade.ErrServerInternal.Wrap(err)
	}

	return publicKey, nil
}

func (t *TokenApplication) VerifyAccessToken(ctx context.Context, token string) (*dto.UserInfo, *facade.Error) {
	jwtToken, err := t.jwthelper.VerifyRS256JWT(token)
	if err != nil {
		if xerror.Is(err, jwt.ErrInvalidJWTToken) {
			return nil, facade.ErrForbidden
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	payload, err := t.jwthelper.DecodeAccessPayload(jwtToken.Payload)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if payload.Expire < time.Now().Unix() {
		return nil, facade.ErrForbidden
	}

	// get user info
	userAggregate, err := t.userReadRepository.Find(ctx, payload.UserID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden
	}

	return getUserInfo(ctx, payload.OrganizationID, userAggregate, t.roleReadRepository, t.organizationUserReadRepository)
}

func (t *TokenApplication) RefreshAccessToken(ctx context.Context, request dto.RefreshAccessTokenRequest) (*dto.RefreshAccessTokenResponse, *facade.Error) {
	// find user
	if request.UserID == "" {
		return nil, facade.ErrBadRequest.Facade("invalid user id")
	}

	userAggregate, err := t.userReadRepository.Find(ctx, request.UserID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden.Facade("user not found")
	}

	// find device and refresh token
	deviceAggregate, err := t.deviceReadRepository.FindByDevice(ctx, request.UserID, request.Device.DeviceType, request.Device.DeviceID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if deviceAggregate == nil {
		return nil, facade.ErrForbidden.Facade("device not found")
	}

	if deviceAggregate.Device.RefreshTokenExpiresAt.Before(time.Now()) ||
		deviceAggregate.Device.RefreshToken != request.RefreshToken {
		return nil, facade.ErrForbidden.Facade("invalid refresh token")
	}

	// check organization
	if deviceAggregate.Device.OrganizationID != uuid.Nil {
		orgs, err := t.organizationUserReadRepository.FindAll(ctx, userAggregate.User.ID)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		found := false
		for _, org := range orgs {
			if org.Organization.ID == deviceAggregate.Device.OrganizationID {
				found = true
				break
			}
		}

		if !found {
			return nil, facade.ErrForbidden.Facade("organization id not found")
		}
	}

	// generate new refresh token
	// deviceAggregate, err = t.deviceService.RegenerateRefreshToken(ctx, deviceAggregate, false)
	// if err != nil {
	// 	return nil, facade.ErrServerInternal.Wrap(err)
	// }

	// genereate new access token
	result, err := generateLoginResult(ctx, userAggregate, deviceAggregate.Device, t.rbacService, t.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.RefreshAccessTokenResponse{
		LoginResponse: *result,
	}, nil
}

func (t *TokenApplication) Logout(ctx context.Context, request dto.LogoutRequest) *facade.Error {
	// find device and refresh token
	deviceAggregate, err := t.deviceReadRepository.FindByDevice(ctx, request.UserID, request.Device.DeviceType, request.Device.DeviceID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if deviceAggregate == nil {
		return facade.ErrForbidden.Facade("device not found")
	}

	if deviceAggregate.Device.RefreshTokenExpiresAt.Before(time.Now()) ||
		deviceAggregate.Device.RefreshToken != request.RefreshToken {
		return facade.ErrForbidden.Facade("invalid refresh token")
	}

	// 获取用户信息
	userAggregate, err := t.userReadRepository.Find(ctx, request.UserID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	// 使 refresh token 失效
	deviceAggregate.Device.RefreshTokenExpiresAt = time.Now()
	_, err = t.deviceService.UpdateDevice(ctx, deviceAggregate)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	// 记录登出事件
	if err = t.posthogClient.Enqueue(posthog.Capture{
		DistinctId: request.UserID,
		Event:      "logout",
		Properties: map[string]interface{}{
			"device_type": request.Device.DeviceType,
			"device_id":   request.Device.DeviceID,
			"name":        userAggregate.User.Name,
			"display":     userAggregate.User.DisplayName,
			"application": userAggregate.Application.Name,
		},
	}); err != nil {
		t.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return nil
}
