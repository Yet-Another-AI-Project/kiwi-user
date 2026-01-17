package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/utils"
	"strings"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/oss"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

type UserApplication struct {
	logger logger.ILogger

	roleReadRepository             contract.IRoleReadRepository
	userReadRepository             contract.IUserReadRepository
	organizationReadRepository     contract.IOrganizationReadRepository
	organizationUserReadRepository contract.IOrganizationUserReadRepository
	applicationReadRepository      contract.IApplicationReadRepository

	organizationService *service.OrganizationService
	userService         *service.UserService
	loginService        *service.LoginService
	posthogClient       posthog.Client
	ossClient           *oss.AliyunOss

	config *config.Config
}

func (u *UserApplication) GetOrganizationUserInfos(ctx context.Context, organizationID uuid.UUID) ([]*dto.PublicUserInfo, *facade.Error) {

	organizationAggregate, err := u.organizationReadRepository.Find(ctx, organizationID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if organizationAggregate == nil {
		return nil, facade.ErrForbidden.Facade("organization not found")
	}

	organizationUsers, err := u.organizationUserReadRepository.FindByOrganizationID(ctx, organizationID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	userInfos := make([]*dto.PublicUserInfo, 0)

	for _, organizationUser := range organizationUsers {
		userInfos = append(userInfos, &dto.PublicUserInfo{
			UserID:      organizationUser.User.ID,
			Application: organizationUser.Application.Name,
			Name:        organizationUser.User.Name,
			DisplayName: organizationUser.User.DisplayName,
			Avatar:      organizationUser.User.Avatar,
		})
	}

	return userInfos, nil
}

func (u *UserApplication) DeleteOrganizationUser(ctx context.Context, userID string, organizationID uuid.UUID) *facade.Error {

	organizationAggregate, err := u.organizationReadRepository.Find(ctx, organizationID)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if organizationAggregate == nil {
		return facade.ErrForbidden.Facade("organization not found")
	}

	organizationUserAggregate, err := u.organizationUserReadRepository.Find(ctx, userID, organizationID)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if organizationUserAggregate == nil {
		return facade.ErrForbidden.Facade("organization user not found")
	}

	if err = u.organizationService.DeleteOrganizationUser(ctx, organizationUserAggregate); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (u *UserApplication) GetDetailUserInfo(ctx context.Context, userID, organizationID string) (*dto.UserInfo, *facade.Error) {

	userAggregate, err := u.userReadRepository.Find(ctx, userID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden.Facade("user not found")
	}

	return getUserInfo(
		ctx,
		organizationID,
		userAggregate,
		u.roleReadRepository,
		u.organizationUserReadRepository)
}

func (u *UserApplication) GetPublicUserInfos(ctx context.Context, userIDs []string) ([]*dto.PublicUserInfo, *facade.Error) {

	userAggregate, err := u.userReadRepository.FindIn(ctx, userIDs)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	userInfos := make([]*dto.PublicUserInfo, 0)
	for _, user := range userAggregate {
		userInfos = append(userInfos, &dto.PublicUserInfo{
			UserID:      user.User.ID,
			Application: user.Application.Name,
			Name:        user.User.Name,
			DisplayName: user.User.DisplayName,
			Avatar:      user.User.Avatar,
			Username:    user.User.Name,
			Department:  user.User.Department,
		})
	}

	return userInfos, nil
}

func (u *UserApplication) UpdateUserRole(ctx context.Context, request dto.UpdateUserRoleRequest) *facade.Error {

	if request.UserID == "" {
		return facade.ErrBadRequest.Facade("user_id is required")
	}

	userAggregate, err := u.userReadRepository.Find(ctx, request.UserID)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	roleAggregate, err := u.roleReadRepository.FindByName(ctx, userAggregate.Application.Name, request.Role)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if roleAggregate == nil {
		return facade.ErrForbidden.Facade("role not found")
	}

	if roleAggregate.Role.Type != enum.RoleTypePersonal {
		return facade.ErrForbidden.Facade("role is not personal role")
	}

	if _, err := u.userService.UpdateUserRole(ctx, userAggregate, roleAggregate); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (u *UserApplication) UpdateUserInfo(ctx context.Context, userID string, request dto.UpdateUserInfoRequest) (*dto.UserInfo, *facade.Error) {
	if userID == "" {
		return nil, facade.ErrBadRequest.Facade("user_id is required")
	}
	userAggregate, err := u.userReadRepository.Find(ctx, userID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden.Facade("user not found")
	}

	// 判断是否为URL
	if request.Avatar != "" && !(strings.HasPrefix(request.Avatar, "http://") || strings.HasPrefix(request.Avatar, "https://")) {
		// base64 转 url
		reader, err := base64.StdEncoding.DecodeString(request.Avatar)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}
		key := fmt.Sprintf("user/user_avatar/%s", userAggregate.User.ID)
		if err := u.ossClient.PutObject(u.config.OSS.BucketName, key, bytes.NewReader(reader)); err != nil {
			u.logger.Warnf(ctx, "ossClient PutObject error: %w", err)
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		request.Avatar = fmt.Sprintf("https://%s/%s", u.config.OSS.CDN, key)
	}

	// update user info
	userAggregate, err = u.userService.UpdateUserInfo(ctx, userAggregate, request.DisplayName, request.Avatar, request.Department)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return getUserInfo(ctx, "", userAggregate, u.roleReadRepository, u.organizationUserReadRepository)
}

func (u *UserApplication) ChangePassword(ctx context.Context, userID string, request dto.ChangePasswordRequest) (*dto.OperationResponse, *facade.Error) {
	if userID == "" {
		return nil, facade.ErrBadRequest.Facade("user_id is required")
	}

	if request.OldPassword == request.NewPassword {
		return nil, facade.ErrBadRequest.Facade("新密码不能与旧密码相同")
	}

	userAggregate, err := u.userReadRepository.Find(ctx, userID)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return nil, facade.ErrForbidden.Facade("user not found")
	}

	passwordBinding := findPasswordBinding(userAggregate.Bindings)
	if passwordBinding == nil || passwordBinding.Salt == "" {
		return nil, facade.ErrForbidden.Facade("旧密码不正确")
	}

	hashedOldPassword, err := utils.EncodePassword(request.OldPassword, passwordBinding.Salt)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if hashedOldPassword != passwordBinding.Identity {
		return nil, facade.ErrForbidden.Facade("旧密码不正确")
	}

	userAggregate, err = u.userService.UpdatePassword(ctx, userAggregate, request.NewPassword)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.OperationResponse{Success: true}, nil
}

func (u *UserApplication) CreateUserWithPassword(ctx context.Context, request dto.CreateUserWithPasswordRequest) (*dto.UserInfo, *facade.Error) {

	applicationAggregate, err := u.applicationReadRepository.FindByName(ctx, request.Application)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return nil, facade.ErrForbidden.Facade("application not found")
	}

	var roleAggregate *aggregate.RoleAggregate
	if applicationAggregate.DefaultPersonalRole != nil {

		roleAggregate, err = u.roleReadRepository.FindByName(ctx, applicationAggregate.Application.Name, applicationAggregate.DefaultPersonalRole.Name)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		if roleAggregate == nil {
			return nil, facade.ErrForbidden.Facade("role not found")
		}
	}

	userAggregate, err := u.userService.UserRegisterWithPassword(ctx, applicationAggregate, roleAggregate, request.Name, request.Password, true)
	if err != nil {
		if xerror.Is(err, service.ErrUserAlreadyExists) {
			return nil, facade.ErrForbidden.Facade("username already been registered")
		}
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if request.OrganizationID != "" {

		organizationID, err := uuid.Parse(request.OrganizationID)
		if err != nil {
			return nil, facade.ErrBadRequest.Facade("invalid organization id")
		}

		organizationAggregate, err := u.organizationReadRepository.Find(ctx, organizationID)

		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		if organizationAggregate == nil {
			return nil, facade.ErrForbidden.Facade("organization not found")
		}

		roleAggregate, err := u.roleReadRepository.FindByName(ctx, userAggregate.Application.Name, request.Role)

		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		if roleAggregate == nil {
			return nil, facade.ErrForbidden.Facade("role not found")
		}

		if roleAggregate.Role.Type != enum.RoleTypeOrganization {
			return nil, facade.ErrForbidden.Facade("role is not organization role")
		}

		if _, err := u.organizationService.UpsertOrganizationUser(ctx, organizationAggregate, userAggregate, roleAggregate); err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		if err := u.posthogClient.Enqueue(posthog.Capture{
			DistinctId: userAggregate.User.ID,
			Event:      "org_user_join",
			Properties: map[string]interface{}{
				"$set": map[string]interface{}{
					"org_id": organizationID.String(),
				},
			},
		}); err != nil {
			u.logger.Errorf(ctx, "posthog event failed: %w", err)
		}
	}

	return getUserInfo(ctx, "", userAggregate, u.roleReadRepository, u.organizationUserReadRepository)
}

func (u *UserApplication) BindingPhone(ctx context.Context, userID string, phone string, mini_program_code string) *facade.Error {
	if phone == "" {
		if mini_program_code == "" {
			return facade.ErrBadRequest.Facade("mini_program_code is required")
		}

		var err error
		phone, err = u.loginService.GetPhoneFromMiniProgramCode(mini_program_code, u.config.Wechat.MiniProgramID, u.config.Wechat.MiniProgramSecret)
		if err != nil {
			return facade.ErrServerInternal.Wrap(err)
		}
	}

	userAggregate, err := u.userReadRepository.Find(ctx, userID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePhone && binding.Identity == phone {
			return facade.ErrForbidden.Facade("phone already bound")
		}
	}

	userAggregate.Bindings = append(userAggregate.Bindings, &entity.BindingEntity{
		Type:     enum.BindingTypePhone,
		Identity: phone,
		Verified: true,
	})

	if _, err := u.userService.Update(ctx, userAggregate); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (u *UserApplication) BindingPhoneWithVerifyCode(ctx context.Context, userID string, phone string, verifyCode string) *facade.Error {
	if phone == "" {
		return facade.ErrBadRequest.Facade("phone is required")
	}
	if verifyCode == "" {
		return facade.ErrBadRequest.Facade("verify_code is required")
	}

	userAggregate, err := u.userReadRepository.Find(ctx, userID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePhone && binding.Identity == phone {
			return facade.ErrForbidden.Facade("phone already bound")
		}
	}

	// todo: 2025-04-24 临时取消短信验证，待实名认证通过后再恢复测试
	//// check verifyCode
	//isSuccess, err := u.smsClient.CheckVerifyCode(phone, "注册验证码", verifyCode)
	//
	//if err != nil {
	//	return facade.ErrServerInternal.Wrap(err)
	//}
	//
	//if !isSuccess {
	//	return facade.ErrForbidden.Facade("verify_code is invalid")
	//}

	userAggregate.Bindings = append(userAggregate.Bindings, &entity.BindingEntity{
		Type:     enum.BindingTypePhone,
		Identity: phone,
		Verified: true,
	})

	if _, err := u.userService.Update(ctx, userAggregate); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func findPasswordBinding(bindings []*entity.BindingEntity) *entity.BindingEntity {
	for _, binding := range bindings {
		if binding.Type == enum.BindingTypePassword {
			return binding
		}
	}
	return nil
}

func NewUserApplication(
	logger logger.ILogger,
	roleReadRepository contract.IRoleReadRepository,
	userReadRepository contract.IUserReadRepository,
	organizationReadRepository contract.IOrganizationReadRepository,
	organizationUserReadRepository contract.IOrganizationUserReadRepository,
	applicationReadRepository contract.IApplicationReadRepository,

	organizationService *service.OrganizationService,
	userService *service.UserService,
	loginService *service.LoginService,
	posthogClient posthog.Client,
	ossClient *oss.AliyunOss,
	config *config.Config,
) *UserApplication {
	return &UserApplication{
		logger:                         logger,
		roleReadRepository:             roleReadRepository,
		userReadRepository:             userReadRepository,
		organizationReadRepository:     organizationReadRepository,
		organizationUserReadRepository: organizationUserReadRepository,
		applicationReadRepository:      applicationReadRepository,
		organizationService:            organizationService,
		userService:                    userService,
		loginService:                   loginService,
		posthogClient:                  posthogClient,
		ossClient:                      ossClient,
		config:                         config,
	}
}
