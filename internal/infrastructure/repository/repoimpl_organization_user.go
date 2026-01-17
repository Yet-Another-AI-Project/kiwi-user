package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/organization"
	"kiwi-user/internal/infrastructure/repository/ent/organizationuser"
	"kiwi-user/internal/infrastructure/repository/ent/user"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type organizationUserImpl struct {
	baseImpl
}

func (u *organizationUserImpl) Find(ctx context.Context, userID string, organizationID uuid.UUID) (*aggregate.OrganizationUserAggregate, error) {
	db := u.getEntClient(ctx)

	organizationUserDO, err := db.OrganizationUser.Query().
		Where(
			organizationuser.HasOrganizationWith(organization.ID(organizationID)),
			organizationuser.HasUserWith(user.ID(userID))).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if organizationUserDO == nil {
		return nil, nil
	}

	organizationDO, err := organizationUserDO.QueryOrganization().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	userDO, err := organizationUserDO.QueryUser().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleDO, err := organizationUserDO.QueryRole().Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	applicationDO, err := userDO.QueryApplication().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	organization := &aggregate.OrganizationUserAggregate{
		Organization:     convertOrganizationDOToEntity(organizationDO),
		Application:      convertApplicationDOToEntity(applicationDO),
		User:             convertUserDOToEntity(userDO),
		OrganizationRole: convertRoleDOToEntity(roleDO),
	}

	return organization, nil
}

func (u *organizationUserImpl) FindAll(ctx context.Context, userID string) ([]*aggregate.OrganizationUserAggregate, error) {
	db := u.getEntClient(ctx)

	userDO, err := db.User.Query().Where(user.ID(userID)).Only(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	applicationDO, err := userDO.QueryApplication().Only(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	organizationUsers, err := userDO.QueryOrganizationUsers().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	organizations := make([]*aggregate.OrganizationUserAggregate, 0)

	for _, ou := range organizationUsers {
		organizationDO, err := ou.QueryOrganization().Only(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		roleDO, err := ou.QueryRole().Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		organization := &aggregate.OrganizationUserAggregate{
			Organization:     convertOrganizationDOToEntity(organizationDO),
			Application:      convertApplicationDOToEntity(applicationDO),
			User:             convertUserDOToEntity(userDO),
			OrganizationRole: convertRoleDOToEntity(roleDO),
		}

		organizations = append(organizations, organization)
	}

	return organizations, nil
}

func (u *organizationUserImpl) FindByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*aggregate.OrganizationUserAggregate, error) {
	db := u.getEntClient(ctx)

	organizationUsers, err := db.OrganizationUser.Query().Where(organizationuser.HasOrganizationWith(organization.ID(organizationID))).All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	organizationUserAggregates := make([]*aggregate.OrganizationUserAggregate, 0)

	for _, ou := range organizationUsers {

		organizationDO, err := ou.QueryOrganization().Only(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		applicationDO, err := organizationDO.QueryApplication().Only(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		userDO, err := ou.QueryUser().Only(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		roleDO, err := ou.QueryRole().Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		organizationUserAggregates = append(organizationUserAggregates, &aggregate.OrganizationUserAggregate{
			Organization:     convertOrganizationDOToEntity(organizationDO),
			Application:      convertApplicationDOToEntity(applicationDO),
			User:             convertUserDOToEntity(userDO),
			OrganizationRole: convertRoleDOToEntity(roleDO),
		})
	}

	return organizationUserAggregates, nil
}

func (u *organizationUserImpl) Update(
	ctx context.Context,
	organizationUserAggregate *aggregate.OrganizationUserAggregate) (*aggregate.OrganizationUserAggregate, error) {
	db := u.getEntClient(ctx)

	_, err := db.OrganizationUser.Update().
		Where(
			organizationuser.HasOrganizationWith(organization.ID(organizationUserAggregate.Organization.ID)),
			organizationuser.HasUserWith(user.ID(organizationUserAggregate.User.ID))).
		SetRoleID(organizationUserAggregate.OrganizationRole.ID).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return organizationUserAggregate, nil
}

func (u *organizationUserImpl) Create(
	ctx context.Context,
	organizationUserAggregate *aggregate.OrganizationUserAggregate) (*aggregate.OrganizationUserAggregate, error) {
	db := u.getEntClient(ctx)

	_, err := db.OrganizationUser.Create().
		SetOrganizationID(organizationUserAggregate.Organization.ID).
		SetUserID(organizationUserAggregate.User.ID).
		SetRoleID(organizationUserAggregate.OrganizationRole.ID).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return organizationUserAggregate, nil
}

func (u *organizationUserImpl) Delete(
	ctx context.Context,
	organizationUserAggregate *aggregate.OrganizationUserAggregate) error {
	db := u.getEntClient(ctx)

	_, err := db.OrganizationUser.Delete().
		Where(
			organizationuser.HasOrganizationWith(organization.ID(organizationUserAggregate.Organization.ID)),
			organizationuser.HasUserWith(user.ID(organizationUserAggregate.User.ID))).
		Exec(ctx)

	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func NewOrganizationUserImpl(db *Client) contract.IOrganizationUserRepository {
	return &organizationUserImpl{
		baseImpl{
			db: db,
		},
	}
}
