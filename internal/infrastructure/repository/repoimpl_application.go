package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/application"

	"github.com/futurxlab/golanggraph/xerror"
)

type applicationImpl struct {
	baseImpl
}

// Create implements contract.IApplicationRepository.
func (a *applicationImpl) Create(ctx context.Context, applicationAggregate *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error) {
	db := a.getEntClient(ctx)

	applicationDO, err := db.Application.Create().
		SetName(applicationAggregate.Application.Name).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var defaultPersonalRole *ent.Role
	var defaultOrgRole *ent.Role
	var defaultOrgAdminRole *ent.Role
	roles := make([]*ent.Role, 0)
	for _, role := range applicationAggregate.Roles {
		roleDO, err := db.Role.Create().
			SetName(role.Name).
			SetApplication(applicationDO).
			Save(ctx)

		if err != nil {
			return nil, xerror.Wrap(err)
		}

		if roleDO.Name == applicationAggregate.DefaultPersonalRole.Name {
			defaultPersonalRole = roleDO
			_, err := db.Application.UpdateOne(applicationDO).
				SetDefaultPersonalRole(roleDO).
				Save(ctx)

			if err != nil {
				return nil, xerror.Wrap(err)
			}
		}

		if roleDO.Name == applicationAggregate.DefaultOrgRole.Name {
			defaultOrgRole = roleDO
			_, err := db.Application.UpdateOne(applicationDO).
				SetDefaultOrgRole(roleDO).
				Save(ctx)

			if err != nil {
				return nil, xerror.Wrap(err)
			}
		}

		if roleDO.Name == applicationAggregate.DefaultOrgAdminRole.Name {
			defaultOrgAdminRole = roleDO
			_, err := db.Application.UpdateOne(applicationDO).
				SetDefaultOrgAdminRole(roleDO).
				Save(ctx)

			if err != nil {
				return nil, xerror.Wrap(err)
			}
		}

		roles = append(roles, roleDO)
	}

	return &aggregate.ApplicationAggregate{
		Application:         convertApplicationDOToEntity(applicationDO),
		Roles:               convertRolesDOToEntities(roles),
		DefaultPersonalRole: convertRoleDOToEntity(defaultPersonalRole),
		DefaultOrgRole:      convertRoleDOToEntity(defaultOrgRole),
		DefaultOrgAdminRole: convertRoleDOToEntity(defaultOrgAdminRole),
	}, nil
}

// Delete implements contract.IApplicationRepository.
func (a *applicationImpl) Delete(ctx context.Context, applicationAggregate *aggregate.ApplicationAggregate) error {
	panic("not implemented")
}

// FindAll implements contract.IApplicationRepository.
func (a *applicationImpl) FindAll(ctx context.Context, offset int, limit int) ([]*aggregate.ApplicationAggregate, error) {
	db := a.getEntClient(ctx)

	applications, err := db.Application.Query().
		WithRoles().
		WithDefaultPersonalRole().
		WithDefaultOrgRole().
		WithDefaultOrgAdminRole().
		Where(application.DeletedAtIsNil()).
		Offset(offset).
		Limit(limit).
		All(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var result []*aggregate.ApplicationAggregate

	for _, application := range applications {
		roles, err := application.QueryRoles().All(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		defaultPersonalRole, err := application.QueryDefaultPersonalRole().Only(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		defaultOrgRole, err := application.QueryDefaultOrgRole().Only(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		defaultOrgAdminRole, err := application.QueryDefaultOrgAdminRole().Only(ctx)

		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		result = append(result, &aggregate.ApplicationAggregate{
			Application:         convertApplicationDOToEntity(application),
			Roles:               convertRolesDOToEntities(roles),
			DefaultPersonalRole: convertRoleDOToEntity(defaultPersonalRole),
			DefaultOrgRole:      convertRoleDOToEntity(defaultOrgRole),
			DefaultOrgAdminRole: convertRoleDOToEntity(defaultOrgAdminRole),
		})
	}

	return result, nil
}

// FindByName implements contract.IApplicationRepository.
func (a *applicationImpl) FindByName(ctx context.Context, name string) (*aggregate.ApplicationAggregate, error) {
	db := a.getEntClient(ctx)

	application, err := db.Application.Query().
		Where(application.Name(name)).
		WithRoles().
		WithDefaultPersonalRole().
		WithDefaultOrgRole().
		WithDefaultOrgAdminRole().
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if application == nil {
		return nil, nil
	}

	roles, err := application.QueryRoles().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	defaultPersonalRole, err := application.QueryDefaultPersonalRole().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	defaultOrgRole, err := application.QueryDefaultOrgRole().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	defaultOrgAdminRole, err := application.QueryDefaultOrgAdminRole().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.ApplicationAggregate{
		Application:         convertApplicationDOToEntity(application),
		Roles:               convertRolesDOToEntities(roles),
		DefaultPersonalRole: convertRoleDOToEntity(defaultPersonalRole),
		DefaultOrgRole:      convertRoleDOToEntity(defaultOrgRole),
		DefaultOrgAdminRole: convertRoleDOToEntity(defaultOrgAdminRole),
	}, nil
}

// Update implements contract.IApplicationRepository.
// Only used to update default role here
func (a *applicationImpl) Update(ctx context.Context, applicationAggregate *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error) {
	db := a.getEntClient(ctx)

	query := db.Application.Update().Where(application.Name(applicationAggregate.Application.Name))

	if applicationAggregate.DefaultPersonalRole != nil {
		query = query.SetDefaultPersonalRoleID(applicationAggregate.DefaultPersonalRole.ID)
	}

	if applicationAggregate.DefaultOrgRole != nil {
		query = query.SetDefaultOrgRoleID(applicationAggregate.DefaultOrgRole.ID)
	}

	if applicationAggregate.DefaultOrgAdminRole != nil {
		query = query.SetDefaultOrgAdminRoleID(applicationAggregate.DefaultOrgAdminRole.ID)
	}

	_, err := query.Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.ApplicationAggregate{
		Application:         applicationAggregate.Application,
		Roles:               applicationAggregate.Roles,
		DefaultPersonalRole: applicationAggregate.DefaultPersonalRole,
		DefaultOrgRole:      applicationAggregate.DefaultOrgRole,
		DefaultOrgAdminRole: applicationAggregate.DefaultOrgAdminRole,
	}, nil
}

func NewApplicationImpl(db *Client) contract.IApplicationRepository {
	return &applicationImpl{
		baseImpl: baseImpl{
			db: db,
		},
	}
}
