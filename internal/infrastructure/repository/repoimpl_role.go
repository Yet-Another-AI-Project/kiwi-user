package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/application"
	"kiwi-user/internal/infrastructure/repository/ent/role"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type roleImpl struct {
	baseImpl
}

func (r *roleImpl) FindByName(ctx context.Context, applicationName string, name string) (*aggregate.RoleAggregate, error) {
	db := r.getEntClient(ctx)

	roleDO, err := db.Role.Query().Where(role.Name(name), role.HasApplicationWith(application.Name(applicationName))).Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if roleDO == nil {
		return nil, nil
	}

	applicationDO, err := roleDO.QueryApplication().Only(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	scopeDOs, err := roleDO.QueryScopes().All(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleAggregate := &aggregate.RoleAggregate{
		Role:        convertRoleDOToEntity(roleDO),
		Application: convertApplicationDOToEntity(applicationDO),
		Scopes:      convertScopeDOsToEntities(scopeDOs),
	}

	return roleAggregate, nil
}

func (r *roleImpl) Create(ctx context.Context, roleAggregate *aggregate.RoleAggregate) (*aggregate.RoleAggregate, error) {
	db := r.getEntClient(ctx)

	roleDO, err := db.Role.Create().
		SetName(roleAggregate.Role.Name).
		SetApplicationID(roleAggregate.Application.ID).
		SetType(role.Type(roleAggregate.Role.Type)).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	scopeDOs := make([]*ent.Scope, 0)
	for _, scope := range roleAggregate.Scopes {
		scopeDO, err := db.Scope.Create().
			SetName(scope.Name).
			SetRole(roleDO).
			Save(ctx)

		if err != nil {
			return nil, xerror.Wrap(err)
		}

		scopeDOs = append(scopeDOs, scopeDO)
	}

	roleAggregate.Role = convertRoleDOToEntity(roleDO)
	roleAggregate.Scopes = convertScopeDOsToEntities(scopeDOs)

	return roleAggregate, nil
}

func (r *roleImpl) Update(ctx context.Context, roleAggregate *aggregate.RoleAggregate) (*aggregate.RoleAggregate, error) {
	db := r.getEntClient(ctx)

	scopes := make([]*entity.ScopeEntity, 0)
	for _, scope := range roleAggregate.Scopes {
		if scope.ID == uuid.Nil {
			scopeDO, err := db.Scope.Create().
				SetName(scope.Name).
				SetRoleID(roleAggregate.Role.ID).
				Save(ctx)

			if err != nil {
				return nil, xerror.Wrap(err)
			}

			scopes = append(scopes, convertScopeDOToEntity(scopeDO))
		} else {
			scopes = append(scopes, scope)
		}
	}

	roleAggregate.Scopes = scopes

	return roleAggregate, nil
}

func (r *roleImpl) Delete(ctx context.Context, roleAggregate *aggregate.RoleAggregate) error {
	panic("not implemented")
}

func NewRoleImpl(db *Client) contract.IRoleRepository {
	return &roleImpl{
		baseImpl{
			db: db,
		},
	}
}
