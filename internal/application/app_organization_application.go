package application

import (
	"context"
	"encoding/json"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/volcengine/msgsms"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	libutils "github.com/Yet-Another-AI-Project/kiwi-lib/tools/utils"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/google/uuid"
)

type OrganizationApplicationApplication struct {
	userReadRepository             contract.IUserReadRepository
	roleReadRepository             contract.IRoleReadRepository
	organizationApplicationService *service.OrganizationApplicationService
	organizationService            *service.OrganizationService
	applicationService             *service.ApplicationService
	logger                         logger.ILogger
	smsClient                      msgsms.SmsClient
	config                         *config.Config
}

func (o *OrganizationApplicationApplication) CreateOrganizationApplication(
	ctx context.Context,
	userID string,
	request *dto.CreateOrganizationApplicationRequest) *facade.Error {

	userAggregate, err := o.userReadRepository.Find(ctx, userID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	// 查询是否绑定过手机，没有则返回错误
	phoneExist := false
	for _, binding := range userAggregate.Bindings {
		if binding.Type == enum.BindingTypePhone && binding.Verified == true && binding.Identity != "" {
			phoneExist = true
			break
		}
	}
	if !phoneExist {
		return facade.ErrForbidden.Facade("user_id does not bind any phone")
	}

	applicationAggregate, err := o.applicationService.GetApplication(ctx, request.Application)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return facade.ErrForbidden.Facade("application not found")
	}

	if request.Name == "" || request.BrandShortName == "" || request.UsageScenario == "" || request.PrimaryBusiness == "" || request.OrgRoleName == "" {
		return facade.ErrBadRequest.Facade("bad request, missing body")
	}

	if err := o.organizationApplicationService.CreateOrganizationApplication(ctx, &entity.OrganizationApplicationEntity{
		UserID:          userID,
		Name:            request.Name,
		BrandShortName:  request.BrandShortName,
		PrimaryBusiness: request.PrimaryBusiness,
		ReferrerName:    request.ReferrerName,
		UsageScenario:   request.UsageScenario,
		DiscoveryWay:    request.DiscoveryWay,
		TrailDays:       7,
		OrgRoleName:     request.OrgRoleName,
		//Contract:     request.Contract,
	}, applicationAggregate.Application); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	// todo: 异步通知子电（邮箱）,本期不做
	return nil
}

func (o *OrganizationApplicationApplication) ReviewOrganizationApplication(
	ctx context.Context,
	request *dto.UpdateOrganizationApplicationRequest) *facade.Error {

	ID, err := uuid.Parse(request.ID)

	if err != nil {
		return facade.ErrBadRequest.Facade("invalid id")
	}

	// 审核通过则创建企业，绑定用户到企业上；失败则更新为失败
	oaa, phone, err := o.organizationApplicationService.UpdateOrganizationApplicationWithTx(ctx,
		&entity.OrganizationApplicationEntity{
			ID:            ID,
			ReviewStatus:  request.ReviewStatus,
			ReviewComment: request.ReviewComment,
		})
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}
	if phone == "" {
		return facade.ErrForbidden.Facade("user does not bind any phone")
	}

	// 异步通知用户审核结果(短信)
	if oaa != nil {
		libutils.SafeGo(ctx, o.logger, func() {
			phones := make([]string, 0)
			phones = append(phones, phone)
			param := dto.SendSmsTemplateParam{
				Code: request.ReviewStatus.String(),
			}
			pj, _ := json.Marshal(param)
			result, _, err := o.smsClient.SendSms(phones, o.config.Sms.SmsTemplateId, string(pj))
			if err != nil {
				o.logger.Errorf(ctx, "send msg error: %w", err)
				return
			}
			if result == nil {
				o.logger.Errorf(ctx, "send msg error: %w", err)
				return
			}
			if result.ResponseMetadata.Error != nil {
				o.logger.Errorf(ctx, "send msg error: %w", err)
				return
			}
			if result != nil && result.ResponseMetadata.Error == nil && result.Result != nil {
				o.logger.Errorf(ctx, "send msg error: %w", err)
				return
			}
			o.logger.Infof(ctx, "send success: %s", phone)
			return
		})
	}

	return nil
}

func (o *OrganizationApplicationApplication) GetUserOrganizationApplications(
	ctx context.Context, userID string, name string) ([]*aggregate.OrganizationApplicationAggregate, *facade.Error) {

	organizationAggregates, err := o.organizationApplicationService.GetOrganizationApplicationAggregates(ctx, userID, name)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return organizationAggregates, nil
}

func (o *OrganizationApplicationApplication) PageOrganizationApplicationInfos(
	ctx context.Context,
	name string,
	pageNum int,
	pageSize int) ([]*aggregate.OrganizationApplicationAggregate, int, *facade.Error) {

	organizationAggregates, total, err := o.organizationApplicationService.PageFind(ctx, name, (pageNum-1)*pageSize, pageSize)

	if err != nil {
		return nil, 0, facade.ErrServerInternal.Wrap(err)
	}

	return organizationAggregates, total, nil
}

func NewOrganizationRequestApplication(
	userReadRepository contract.IUserReadRepository,
	roleReadRepository contract.IRoleReadRepository,
	organizationApplicationService *service.OrganizationApplicationService,
	organizationService *service.OrganizationService,
	applicationService *service.ApplicationService,
	logger logger.ILogger,
	smsClient msgsms.SmsClient,
	config *config.Config,
) *OrganizationApplicationApplication {
	return &OrganizationApplicationApplication{
		userReadRepository:             userReadRepository,
		roleReadRepository:             roleReadRepository,
		organizationApplicationService: organizationApplicationService,
		organizationService:            organizationService,
		applicationService:             applicationService,
		logger:                         logger,
		smsClient:                      smsClient,
		config:                         config,
	}
}
