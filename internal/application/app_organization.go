package application

import (
	"context"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/client/alibaba/oss"
	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/google/uuid"
	"github.com/posthog/posthog-go"
)

type OrganizationApplication struct {
	applicationReadRepository  contract.IApplicationReadRepository
	userReadRepository         contract.IUserReadRepository
	roleReadRepository         contract.IRoleReadRepository
	organizationReadRepository contract.IOrganizationReadRepository
	organizationService        *service.OrganizationService
	posthogClient              posthog.Client
	logger                     logger.ILogger
	config                     *config.Config
	ossClient                  *oss.AliyunOss
}

func (o *OrganizationApplication) CreateOrganizationRequest(
	ctx context.Context,
	userID string,
	requestType string,
	orgName string) *facade.Error {

	userAggregate, err := o.userReadRepository.Find(ctx, userID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	if err := o.organizationService.CreateRequest(ctx, &entity.OrganizationRequestEntity{
		UserID:           userID,
		Type:             enum.ParseOrganizationRequestType(requestType),
		OrganizationName: orgName,
		ApplicationID:    userAggregate.Application.ID,
	}); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (o *OrganizationApplication) CreateOrganization(
	ctx context.Context,
	request *dto.CreateOrganizationRequst) (*aggregate.OrganizationAggregate, *facade.Error) {

	applicationAggregate, err := o.applicationReadRepository.FindByName(ctx, request.Application)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return nil, facade.ErrForbidden.Facade("application not found")
	}

	organizationAggregate, err := o.organizationReadRepository.FindByName(ctx, applicationAggregate.Application.Name, request.Name)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if organizationAggregate != nil {
		return nil, facade.ErrForbidden.Facade("organization already exists")
	}

	organizationStatus := enum.ParseOrganizationStatus(request.Version)

	if organizationStatus == enum.OrganizationStatusUnknown {
		return nil, facade.ErrBadRequest.Facade("invalid organization status")
	}

	if request.ExpiresAt <= time.Now().Unix() {
		return nil, facade.ErrBadRequest.Facade("invalid expires_at")
	}

	newOrganizationAggregate := &aggregate.OrganizationAggregate{
		Organization: &entity.OrganizationEntity{
			Name:         request.Name,
			Status:       organizationStatus,
			ExpiresAt:    time.Unix(request.ExpiresAt, 0),
			LogoImageURL: "",
		},
		Application: applicationAggregate.Application,
	}

	organization, err := o.organizationService.CreateOrganization(ctx, newOrganizationAggregate)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return organization, nil
}

func (o *OrganizationApplication) UpdateOrganization(
	ctx context.Context,
	request *dto.UpdateOrganizationRequest) (*aggregate.OrganizationAggregate, *facade.Error) {

	organizationID, err := uuid.Parse(request.ID)

	if err != nil {
		return nil, facade.ErrBadRequest.Facade("invalid organization id")
	}

	organizationAggregate, err := o.organizationReadRepository.Find(ctx, organizationID)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if organizationAggregate == nil {
		return nil, facade.ErrForbidden.Facade("organization not found")
	}

	if request.Name != "" {

		existOrganizationAggregate, err := o.organizationReadRepository.FindByName(ctx, organizationAggregate.Application.Name, request.Name)

		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}

		if existOrganizationAggregate != nil &&
			existOrganizationAggregate.Organization.ID != organizationAggregate.Organization.ID {
			return nil, facade.ErrForbidden.Facade("organization already exists")
		}

		organizationAggregate.Organization.Name = request.Name
	}

	if request.Version != "" {
		organizationStatus := enum.ParseOrganizationStatus(request.Version)
		if organizationStatus == enum.OrganizationStatusUnknown {
			return nil, facade.ErrBadRequest.Facade("invalid organization status")
		}

		organizationAggregate.Organization.Status = organizationStatus
	}

	if request.ExpiresAt != 0 {
		if request.ExpiresAt <= time.Now().Unix() {
			return nil, facade.ErrBadRequest.Facade("invalid expires_at")
		}

		organizationAggregate.Organization.ExpiresAt = time.Unix(request.ExpiresAt, 0)
	}

	if request.PermissionCode != "" {
		organizationAggregate.Organization.PermissionCode = request.PermissionCode
	}

	organizationAggregate, err = o.organizationService.UpdateOrganization(ctx, organizationAggregate)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return organizationAggregate, nil
}

func (o *OrganizationApplication) CreateOrganizationUser(
	ctx context.Context,
	request *dto.CreateOrganizationUserRequest) *facade.Error {

	if request.UserID == "" {
		return facade.ErrBadRequest.Facade("user_id is required")
	}

	userAggregate, err := o.userReadRepository.Find(ctx, request.UserID)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	organizationID, err := uuid.Parse(request.OrganizationID)
	if err != nil {
		return facade.ErrBadRequest.Facade("invalid organization id")
	}

	organizationAggregate, err := o.organizationReadRepository.Find(ctx, organizationID)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if organizationAggregate == nil {
		return facade.ErrForbidden.Facade("organization not found")
	}

	roleAggregate, err := o.roleReadRepository.FindByName(ctx, userAggregate.Application.Name, request.Role)

	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if roleAggregate == nil {
		return facade.ErrForbidden.Facade("role not found")
	}

	if roleAggregate.Role.Type != enum.RoleTypeOrganization {
		return facade.ErrForbidden.Facade("role is not organization role")
	}

	if _, err := o.organizationService.UpsertOrganizationUser(ctx, organizationAggregate, userAggregate, roleAggregate); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if err := o.posthogClient.Enqueue(posthog.Capture{
		DistinctId: userAggregate.User.ID,
		Event:      "org_user_join",
		Properties: map[string]interface{}{
			"$set": map[string]interface{}{
				"org_id": organizationID.String(),
			},
		},
	}); err != nil {
		o.logger.Errorf(ctx, "posthog event failed: %w", err)
	}

	return nil
}

func (o *OrganizationApplication) GetOrganizationInfos(
	ctx context.Context,
	request dto.GetOrganizationInfosRequest) ([]*aggregate.OrganizationAggregate, *facade.Error) {

	orgIDs := make([]uuid.UUID, 0, len(request.OrgIDs))

	for _, orgID := range request.OrgIDs {
		id, err := uuid.Parse(orgID)
		if err != nil {
			return nil, facade.ErrBadRequest.Facade("invalid organization id")
		}

		orgIDs = append(orgIDs, id)
	}

	organizationAggregates, err := o.organizationReadRepository.FindIn(ctx, orgIDs)

	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return organizationAggregates, nil
}

func (o *OrganizationApplication) PageOrganizationInfos(
	ctx context.Context,
	pageNum int,
	pageSize int) ([]*aggregate.OrganizationAggregate, int, *facade.Error) {

	organizationAggregates, total, err := o.organizationReadRepository.PageFind(ctx, (pageNum-1)*pageSize, pageSize)

	if err != nil {
		return nil, 0, facade.ErrServerInternal.Wrap(err)
	}

	return organizationAggregates, total, nil
}

func (o *OrganizationApplication) GetOrganizationAggregate(ctx context.Context, orgID uuid.UUID) (*aggregate.OrganizationAggregate, *facade.Error) {
	if orgID == uuid.Nil {
		return nil, facade.ErrForbidden.Facade("invalid orgID, orgID is null")
	}
	organizationAggregate, err := o.organizationReadRepository.Find(ctx, orgID)
	if err != nil {
		return nil, facade.ErrForbidden.Wrap(err)
	}
	return organizationAggregate, nil
}

func NewOrganizationApplication(
	applicationReadRepository contract.IApplicationReadRepository,
	userReadRepository contract.IUserReadRepository,
	roleReadRepository contract.IRoleReadRepository,
	organizationReadRepository contract.IOrganizationReadRepository,
	organizationService *service.OrganizationService,
	posthogClient posthog.Client,
	logger logger.ILogger,
	ossClient *oss.AliyunOss,
	config *config.Config,
) *OrganizationApplication {
	return &OrganizationApplication{
		applicationReadRepository:  applicationReadRepository,
		userReadRepository:         userReadRepository,
		roleReadRepository:         roleReadRepository,
		organizationReadRepository: organizationReadRepository,
		organizationService:        organizationService,
		posthogClient:              posthogClient,
		logger:                     logger,
		ossClient:                  ossClient,
		config:                     config,
	}
}
