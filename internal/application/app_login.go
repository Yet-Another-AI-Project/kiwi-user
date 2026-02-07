package application

import (
	"context"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/captcha"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/volcengine/msgsms"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"

	"kiwi-user/internal/infrastructure/jwt"
)

type LoginApplication struct {
	loginService             *service.LoginService
	vertificationCodeService *service.VertificationCodeService
	applicationService       *service.ApplicationService
	deviceService            *service.DeviceService
	rbacService              *service.RBACService

	deviceReadRepository           contract.IDeviceReadRepository
	userReadRepository             contract.IUserReadRepository
	organizationUserReadRepository contract.IOrganizationUserReadRepository

	config        *config.Config
	logger        logger.ILogger
	jwthelper     *jwt.JWTHelper
	posthogClient posthog.Client
	smsClient     msgsms.SmsClient
	captchaClient captcha.CaptchaClient
}

func NewLoginApplication(
	config *config.Config,
	logger logger.ILogger,
	loginService *service.LoginService,
	applicationService *service.ApplicationService,
	deviceService *service.DeviceService,
	rbacService *service.RBACService,
	deviceReadRepository contract.IDeviceReadRepository,
	userReadRepository contract.IUserReadRepository,
	organizationUserReadRepository contract.IOrganizationUserReadRepository,
	jwthelper *jwt.JWTHelper,
	posthogClient posthog.Client,
	smsClient msgsms.SmsClient,
	vertificationCodeService *service.VertificationCodeService,
	captchaClient captcha.CaptchaClient,
) *LoginApplication {
	return &LoginApplication{
		config:                         config,
		logger:                         logger,
		loginService:                   loginService,
		applicationService:             applicationService,
		deviceService:                  deviceService,
		rbacService:                    rbacService,
		deviceReadRepository:           deviceReadRepository,
		userReadRepository:             userReadRepository,
		organizationUserReadRepository: organizationUserReadRepository,
		jwthelper:                      jwthelper,
		posthogClient:                  posthogClient,
		smsClient:                      smsClient,
		vertificationCodeService:       vertificationCodeService,
		captchaClient:                  captchaClient,
	}
}

func (l *LoginApplication) WechatWebLogin(ctx context.Context, request dto.WechatWebLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// get application aggergate
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}

		return nil, facade.ErrServerInternal.Wrap(err)
	}

	var referralChannel entity.UserRefferalChannel

	if request.ReferralChannel != nil {
		referralChannel = entity.UserRefferalChannel{
			Type: request.ReferralChannel.Type,
			ID:   request.ReferralChannel.ID,
			Name: request.ReferralChannel.Name,
		}
	}

	// login and get user aggregate
	user, err := l.loginService.WechatWebLogin(
		ctx,
		application,
		referralChannel,
		request.Code,
		request.Platform)
	if err != nil {
		if xerror.Is(err, service.ErrWechatInvalidScope) {
			return nil, facade.ErrForbidden.Wrap(err)
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// get refreshtoken
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "$set",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":        user.User.Name,
				"display":     user.User.DisplayName,
				"application": user.Application.Name,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	if err := l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "login",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"type":     "wechat",
				"platform": request.Platform,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return result, nil
}

func (l *LoginApplication) WechatMiniProgramLogin(ctx context.Context, request dto.WechatMiniProgramLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// get application aggergate
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}

		return nil, facade.ErrServerInternal.Wrap(err)
	}

	var referralChannel entity.UserRefferalChannel

	if request.ReferralChannel != nil {
		referralChannel = entity.UserRefferalChannel{
			Type: request.ReferralChannel.Type,
			ID:   request.ReferralChannel.ID,
			Name: request.ReferralChannel.Name,
		}
	}

	// login and get user aggregate
	user, err := l.loginService.WechatMiniprogramLogin(ctx, application, referralChannel, request.Code, request.MiniProgramPhoneCode)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// get refreshtoken
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return result, nil
}

func (l *LoginApplication) QyWechatLogin(ctx context.Context, request dto.QyWechatLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// get application aggregate
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}

		return nil, facade.ErrServerInternal.Wrap(err)
	}

	var referralChannel entity.UserRefferalChannel

	if request.ReferralChannel != nil {
		referralChannel = entity.UserRefferalChannel{
			Type: request.ReferralChannel.Type,
			ID:   request.ReferralChannel.ID,
			Name: request.ReferralChannel.Name,
		}
	}

	// login and get user aggregate
	user, err := l.loginService.QyWechatLogin(ctx, application, referralChannel, request.Code)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// get refreshtoken
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return result, nil
}

func (l *LoginApplication) PasswordLogin(ctx context.Context, request dto.PasswordLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// get application aggergate
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// login and get user aggregate
	user, err := l.loginService.PasswordLogin(ctx, application, request.Name, request.Password)
	if err != nil {

		if xerror.Is(err, service.ErrUserNotFound) {
			return nil, facade.ErrForbidden.Facade("user not found")
		}

		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// get refreshtoken
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if err := l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "$set",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":        user.User.Name,
				"display":     user.User.DisplayName,
				"application": user.Application.Name,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	if err := l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "login",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"type":     "namepass",
				"platform": "",
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return result, nil
}

func (l *LoginApplication) OrganizationLogin(ctx context.Context, request dto.OrganizationLoginRequest) (*dto.LoginResponse, *facade.Error) {
	if request.UserID == "" {
		return nil, facade.ErrBadRequest.Facade("invalid user id")
	}

	organizationID, err := uuid.Parse(request.OrganizationID)
	if err != nil {
		return nil, facade.ErrBadRequest.Facade("invalid organization id")
	}

	// find user
	userAggregate, err := l.userReadRepository.Find(ctx, request.UserID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden.Facade("user not found")
	}

	// check organization
	orgs, err := l.organizationUserReadRepository.FindAll(ctx, userAggregate.User.ID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	found := false
	for _, org := range orgs {
		if org.Organization.ID.String() == request.OrganizationID {
			found = true
			break
		}
	}

	if !found {
		return nil, facade.ErrForbidden.Facade("user not in organization")
	}

	// find device and refresh token
	deviceAggregate, err := l.deviceReadRepository.FindByDevice(ctx, request.UserID, request.Device.DeviceType, request.Device.DeviceID)
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

	// update organization id
	deviceAggregate.Device.OrganizationID = organizationID
	deviceAggregate, err = l.deviceService.UpdateDevice(ctx, deviceAggregate)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// genereate new access token
	result, err := generateLoginResult(ctx, userAggregate, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return result, nil
}

func (l *LoginApplication) PhoneLogin(ctx context.Context, request dto.PhoneLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// 1. 获取应用
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 2. 验证手机验证码
	verified, err := l.smsClient.CheckVerifyCode(request.Phone, request.VerifyCode)
	if err != nil {
		l.logger.Errorf(ctx, "phone login error: %w", err)
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if !verified {
		l.logger.Errorf(ctx, "phone login code verified fail")
		return nil, facade.ErrForbidden.Facade("invalid verification code")
	}

	// 3. 查找或创建用户
	user, err := l.loginService.PhoneLogin(ctx, application, request.Phone)
	if err != nil {
		l.logger.Errorf(ctx, "phone login or create account error: %w", err)
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 4. 更新或创建设备
	l.logger.Debugf(ctx, "phone login UpsertDevice")
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 5. 生成登录结果
	l.logger.Debugf(ctx, "phone login generateLoginResult")
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 6. 记录登录事件
	l.logger.Debugf(ctx, "phone login Enqueue")
	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "$set",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":        user.User.Name,
				"display":     user.User.DisplayName,
				"application": user.Application.Name,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "login",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"type":     "phone",
				"platform": "app",
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return result, nil
}

// SendSmsVerifyCode sends a verification code to the specified phone
func (l *LoginApplication) SendPhoneVerifyCode(ctx context.Context, phone string) *facade.Error {
	if phone == "" {
		return facade.ErrBadRequest.Facade("phone is required")
	}

	result, err := l.smsClient.SendVerifyCode(phone, l.config.Sms.VerifyTemplateID)
	if err != nil {
		return facade.ErrForbidden.Wrap(err)
	}

	if result == nil {
		return facade.ErrServerInternal.Facade("sms no response")
	}

	if result.ResponseMetadata.Error != nil {
		l.logger.Errorf(ctx, "SendVerifyCode ResponseMetadata Error: %w", result.ResponseMetadata.Error)
		return facade.ErrServerInternal.Facade(result.ResponseMetadata.Error.Message)

	}

	return nil
}

// SendEmailVerificationCode sends a verification code to the specified email
func (l *LoginApplication) SendEmailVerificationCode(ctx context.Context, request dto.SendEmailVerificationCodeRequest) *facade.Error {
	// Send verification code
	err := l.vertificationCodeService.SendEmailVerificationCode(ctx, request.Email, request.CodeType)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

// EmailLogin handles email verification code login
func (l *LoginApplication) EmailLogin(ctx context.Context, request dto.EmailLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// 1. Get application
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 2. Verify email code
	verified, err := l.vertificationCodeService.VerifyEmailCode(ctx, request.Email, request.VerifyCode, enum.VertificationCodeTypeLogin)
	if err != nil {
		l.logger.Debugf(ctx, "email login code verified fail: %w", err)
		return nil, facade.ErrForbidden.Facade(err.Error())
	}

	if !verified {
		l.logger.Debugf(ctx, "email login code verified fail")
		return nil, facade.ErrForbidden.Facade("email verification failed")
	}

	// 3. Find or create user
	user, err := l.loginService.EmailLogin(ctx, application, request.Email)
	if err != nil {
		l.logger.Debugf(ctx, "email login or create account error: %w", err)
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 4. Update or create device
	l.logger.Debugf(ctx, "email login UpsertDevice")
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 5. Generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// 6. Record login event
	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "$set",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":        user.User.Name,
				"display":     user.User.DisplayName,
				"application": user.Application.Name,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "login",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"type":     "email",
				"platform": "app",
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return result, nil
}

func (l *LoginApplication) GoogleWebLogin(ctx context.Context, request dto.GoogleWebLoginRequest) (*dto.LoginResponse, *facade.Error) {
	// get application aggregate
	application, err := l.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		if xerror.Is(err, service.ErrApplicationNotFound) {
			return nil, facade.ErrForbidden.Facade("application not found")
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	var referralChannel entity.UserRefferalChannel

	if request.ReferralChannel != nil {
		referralChannel = entity.UserRefferalChannel{
			Type: request.ReferralChannel.Type,
			ID:   request.ReferralChannel.ID,
			Name: request.ReferralChannel.Name,
		}
	}

	// login and get user aggregate
	user, err := l.loginService.GoogleWebLogin(
		ctx,
		application,
		referralChannel,
		request.Code,
		request.RedirectURI)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// get refreshtoken
	deviceAggregate, err := l.deviceService.UpsertDevice(
		ctx,
		user.User.ID,
		request.Device.DeviceType,
		request.Device.DeviceID,
		uuid.Nil)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// generate login result
	result, err := generateLoginResult(ctx, user, deviceAggregate.Device, l.rbacService, l.jwthelper)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	// record login event
	if err = l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "$set",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"name":        user.User.Name,
				"display":     user.User.DisplayName,
				"application": user.Application.Name,
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	if err := l.posthogClient.Enqueue(posthog.Capture{
		DistinctId: user.User.ID,
		Event:      "login",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"type":     "google",
				"platform": "web",
			},
		},
	}); err != nil {
		l.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return result, nil
}
