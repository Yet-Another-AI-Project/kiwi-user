package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"

	"github.com/google/uuid"
)

type IOrganizationReadRepository interface {
	Find(ctx context.Context, organizationID uuid.UUID) (*aggregate.OrganizationAggregate, error)
	PageFind(ctx context.Context, offset, limit int) ([]*aggregate.OrganizationAggregate, int, error)
	FindIn(ctx context.Context, organizationIDs []uuid.UUID) ([]*aggregate.OrganizationAggregate, error)
	FindByName(ctx context.Context, applicationName string, orgName string) (*aggregate.OrganizationAggregate, error)
}

type IOrganizationWriteRepository interface {
	Create(ctx context.Context, organization *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error)
	Update(ctx context.Context, organization *aggregate.OrganizationAggregate) (*aggregate.OrganizationAggregate, error)
}

type IOrganizationRepository interface {
	ITransaction
	IOrganizationReadRepository
	IOrganizationWriteRepository
}
