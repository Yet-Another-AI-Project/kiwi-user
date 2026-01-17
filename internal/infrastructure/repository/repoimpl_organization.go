package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/application"
	"kiwi-user/internal/infrastructure/repository/ent/organization"

	"github.com/google/uuid"
)

type organizationImpl struct {
	baseImpl
}

func (u *organizationImpl) Find(ctx context.Context, id uuid.UUID) (*aggregate.OrganizationAggregate, error) {
	db := u.getEntClient(ctx)

	organizationDO, err := db.Organization.Get(ctx, id)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if organizationDO == nil {
		return nil, nil
	}

	applicationDO, err := organizationDO.QueryApplication().Only(ctx)
	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationAggregate{
		Organization: convertOrganizationDOToEntity(organizationDO),
		Application:  convertApplicationDOToEntity(applicationDO),
	}, nil
}

func (u *organizationImpl) PageFind(ctx context.Context, offset int, limit int) ([]*aggregate.OrganizationAggregate, int, error) {
	db := u.getEntClient(ctx)

	organizationDOs, err := db.Organization.Query().
		Offset(offset).
		Limit(limit).
		Order(ent.Desc(organization.FieldCreatedAt)).
		All(ctx)

	if err != nil {
		return nil, 0, err
	}

	organizationAggregates := make([]*aggregate.OrganizationAggregate, 0)
	for _, organizationDO := range organizationDOs {
		applicationDO, err := organizationDO.QueryApplication().Only(ctx)
		if err != nil {
			return nil, 0, err
		}

		organizationAggregates = append(organizationAggregates, &aggregate.OrganizationAggregate{
			Organization: convertOrganizationDOToEntity(organizationDO),
			Application:  convertApplicationDOToEntity(applicationDO),
		})
	}

	total, err := db.Organization.Query().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return organizationAggregates, total, nil
}

func (u *organizationImpl) FindIn(ctx context.Context, ids []uuid.UUID) ([]*aggregate.OrganizationAggregate, error) {
	db := u.getEntClient(ctx)

	organizationDOs, err := db.Organization.Query().
		Where(organization.IDIn(ids...)).
		All(ctx)

	if err != nil {
		return nil, err
	}

	organizationAggregates := make([]*aggregate.OrganizationAggregate, 0)
	for _, organizationDO := range organizationDOs {
		applicationDO, err := organizationDO.QueryApplication().Only(ctx)
		if err != nil {
			return nil, err
		}

		organizationAggregates = append(organizationAggregates, &aggregate.OrganizationAggregate{
			Organization: convertOrganizationDOToEntity(organizationDO),
			Application:  convertApplicationDOToEntity(applicationDO),
		})
	}

	return organizationAggregates, nil
}

func (u *organizationImpl) FindByName(ctx context.Context, applicationName string, orgName string) (*aggregate.OrganizationAggregate, error) {
	db := u.getEntClient(ctx)

	organizationDO, err := db.Organization.Query().
		Where(organization.Name(orgName), organization.HasApplicationWith(application.Name(applicationName))).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if organizationDO == nil {
		return nil, nil
	}

	applicationDO, err := organizationDO.QueryApplication().Only(ctx)
	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationAggregate{
		Organization: convertOrganizationDOToEntity(organizationDO),
		Application:  convertApplicationDOToEntity(applicationDO),
	}, nil
}

func (u *organizationImpl) Create(ctx context.Context, organizationAggregate *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error) {
	db := u.getEntClient(ctx)

	organizationDO, err := db.Organization.Create().
		SetName(organizationAggregate.Organization.Name).
		SetStatus(organization.Status(organizationAggregate.Organization.Status)).
		SetExpiresAt(organizationAggregate.Organization.ExpiresAt).
		SetApplicationID(organizationAggregate.Application.ID).
		SetLogoURL(organizationAggregate.Organization.LogoImageURL).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationAggregate{
		Organization: convertOrganizationDOToEntity(organizationDO),
		Application:  organizationAggregate.Application,
	}, nil
}

func (u *organizationImpl) Update(ctx context.Context, organizationAggregate *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error) {
	db := u.getEntClient(ctx)

	organizationDO, err := db.Organization.UpdateOneID(organizationAggregate.Organization.ID).
		SetName(organizationAggregate.Organization.Name).
		SetStatus(organization.Status(organizationAggregate.Organization.Status)).
		SetExpiresAt(organizationAggregate.Organization.ExpiresAt).
		SetPermissionCode(organizationAggregate.Organization.PermissionCode).
		SetLogoURL(organizationAggregate.Organization.LogoImageURL).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &aggregate.OrganizationAggregate{
		Organization: convertOrganizationDOToEntity(organizationDO),
		Application:  organizationAggregate.Application,
	}, nil
}

func NewOrganizationImpl(db *Client) contract.IOrganizationRepository {
	return &organizationImpl{
		baseImpl{
			db: db,
		},
	}
}
