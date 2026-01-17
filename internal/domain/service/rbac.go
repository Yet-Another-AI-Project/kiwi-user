package service

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"

	"github.com/futurxlab/golanggraph/xerror"
)

type RBACService struct {
	applicationRepository contract.IApplicationRepository
	roleRepository        contract.IRoleRepository
}

func NewRBACService(
	applicationRepository contract.IApplicationRepository,
	roleRepository contract.IRoleRepository) *RBACService {
	return &RBACService{
		applicationRepository: applicationRepository,
		roleRepository:        roleRepository,
	}
}

func (r *RBACService) GetRole(ctx context.Context, application string, role string) (*aggregate.RoleAggregate, error) {
	roleAggregate, err := r.roleRepository.FindByName(ctx, application, role)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if roleAggregate == nil {
		return nil, xerror.Wrap(ErrRoleNotFound)
	}

	return roleAggregate, nil
}

func (r *RBACService) CreateRole(ctx context.Context, application *aggregate.ApplicationAggregate, roleType enum.RoleType, role string) (*aggregate.RoleAggregate, error) {
	roleAggregate, err := r.roleRepository.FindByName(ctx, application.Application.Name, role)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if roleAggregate != nil {
		return nil, xerror.Wrap(ErrRoleAlreadyExists)
	}

	newRoleAggregate := &aggregate.RoleAggregate{
		Role: &entity.RoleEntity{
			Type: roleType,
			Name: role,
		},
		Application: application.Application,
	}

	newRoleAggregate, err = r.roleRepository.Create(ctx, newRoleAggregate)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return newRoleAggregate, nil
}

func (r *RBACService) CreateRoleScope(ctx context.Context, application *aggregate.ApplicationAggregate, role string, newScope string) (*aggregate.RoleAggregate, error) {

	roleAggregate, err := r.roleRepository.FindByName(ctx, application.Application.Name, role)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if roleAggregate == nil {
		return nil, xerror.Wrap(ErrRoleNotFound)
	}

	for _, scope := range roleAggregate.Scopes {
		if scope.Name == newScope {
			return nil, xerror.Wrap(ErrScopeAlreadyExists)
		}
	}

	roleAggregate.Scopes = append(roleAggregate.Scopes, &entity.ScopeEntity{
		Name: newScope,
	})

	roleAggregate, err = r.roleRepository.Update(ctx, roleAggregate)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return roleAggregate, nil
}

func (r *RBACService) DeleteRoleScope(ctx context.Context, application *aggregate.ApplicationAggregate, role string, deleteScope string) (*aggregate.RoleAggregate, error) {
	roleAggregate, err := r.roleRepository.FindByName(ctx, application.Application.Name, role)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if roleAggregate == nil {
		return nil, xerror.Wrap(ErrRoleNotFound)
	}

	newScopeEntities := make([]*entity.ScopeEntity, 0)
	for _, scope := range roleAggregate.Scopes {
		if scope.Name != deleteScope {
			newScopeEntities = append(newScopeEntities, scope)
		}
	}

	roleAggregate.Scopes = newScopeEntities
	roleAggregate, err = r.roleRepository.Update(ctx, roleAggregate)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return roleAggregate, nil
}

func (r *RBACService) SetDefaultRole(ctx context.Context, application *aggregate.ApplicationAggregate, defaultRoleType enum.DefaultRoleType, roleName string) error {

	for _, role := range application.Roles {
		if role.Name == roleName {
			switch defaultRoleType {
			case enum.DefaultRoleTypePersonal:
				if role.Type != enum.RoleTypePersonal {
					return xerror.Wrap(ErrRoleNotFound)
				}
				application.DefaultPersonalRole = role
			case enum.DefaultRoleTypeOrganization:
				if role.Type != enum.RoleTypeOrganization {
					return xerror.Wrap(ErrRoleNotFound)
				}
				application.DefaultOrgRole = role
			case enum.DefaultRoleTypeOrganizationAdmin:
				if role.Type != enum.RoleTypeOrganization {
					return xerror.Wrap(ErrRoleNotFound)
				}
				application.DefaultOrgAdminRole = role
			}

			if _, err := r.applicationRepository.Update(ctx, application); err != nil {
				return xerror.Wrap(err)
			}

			return nil
		}
	}

	return xerror.Wrap(ErrRoleNotFound)
}
