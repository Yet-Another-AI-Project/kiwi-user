package application

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

type RBACApplication struct {
	logger logger.ILogger

	applicationService *service.ApplicationService
	rbacService        *service.RBACService
}

func NewRBACApplication(
	logger logger.ILogger,
	applicationService *service.ApplicationService,
	rbacService *service.RBACService,
) *RBACApplication {
	return &RBACApplication{
		logger:             logger,
		applicationService: applicationService,
		rbacService:        rbacService,
	}
}

func (r *RBACApplication) GetApplication(ctx context.Context, name string) (*aggregate.ApplicationAggregate, []*aggregate.RoleAggregate, *facade.Error) {
	applicationAggregate, err := r.applicationService.GetApplication(ctx, name)
	if err != nil {
		return nil, nil, facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return nil, nil, facade.ErrForbidden.Facade("application not found")
	}

	roleAggregates := make([]*aggregate.RoleAggregate, 0)
	for _, role := range applicationAggregate.Roles {
		roleAggregate, err := r.rbacService.GetRole(ctx, applicationAggregate.Application.Name, role.Name)
		if err != nil {
			return nil, nil, facade.ErrServerInternal.Wrap(err)
		}

		roleAggregates = append(roleAggregates, roleAggregate)
	}

	return applicationAggregate, roleAggregates, nil
}

func (r *RBACApplication) CreateApplication(ctx context.Context, name string) *facade.Error {

	applicationAggregate := &aggregate.ApplicationAggregate{
		Application: &entity.ApplicationEntity{
			Name: name,
		},
	}

	_, err := r.applicationService.CreateApplication(ctx, applicationAggregate)
	if err != nil {
		switch err {
		case service.ErrApplicationAlreadyExists:
			return facade.ErrForbidden.Facade("application already exists")
		default:
			return facade.ErrServerInternal.Wrap(err)
		}
	}

	return nil
}

func (r *RBACApplication) CreateRole(ctx context.Context, request *dto.CreateRoleRequest) *facade.Error {
	applicationAggregate, err := r.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return facade.ErrForbidden.Facade("application not found")
	}

	if _, err := r.rbacService.CreateRole(
		ctx,
		applicationAggregate,
		enum.ParseRoleType(request.RoleType),
		request.RoleName); err != nil {

		if xerror.Is(err, service.ErrRoleAlreadyExists) {
			return facade.ErrForbidden.Facade("role already exists")
		}

		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (r *RBACApplication) CreateRoleScope(ctx context.Context, request *dto.CreateScopeRequest) *facade.Error {
	applicationAggregate, err := r.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return facade.ErrForbidden.Facade("application not found")
	}

	if _, err := r.rbacService.CreateRoleScope(ctx, applicationAggregate, request.RoleName, request.ScopeName); err != nil {
		if xerror.Is(err, service.ErrScopeAlreadyExists) {
			return facade.ErrForbidden.Facade("scope already exists")
		}

		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (r *RBACApplication) UpdateApplicationDefaultRole(ctx context.Context, request *dto.UpdateDefaultRoleRequest) *facade.Error {
	applicationAggregate, err := r.applicationService.GetApplication(ctx, request.ApplicationName)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if applicationAggregate == nil {
		return facade.ErrForbidden.Facade("application not found")
	}

	if err := r.rbacService.SetDefaultRole(ctx, applicationAggregate, enum.ParseDefaultRoleType(request.Type), request.RoleName); err != nil {
		if xerror.Is(err, service.ErrRoleNotFound) {
			return facade.ErrForbidden.Facade("role not found")
		}

		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}

func (r *RBACApplication) UpdateUserPersonalRole(ctx context.Context, request *dto.SetUserRoleRequest) {

}
