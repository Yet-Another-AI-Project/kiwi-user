package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"

	"github.com/google/uuid"
)

type IOrganizationUserReadRepository interface {
	Find(ctx context.Context, userID string, organizationID uuid.UUID) (*aggregate.OrganizationUserAggregate, error)
	FindAll(ctx context.Context, userID string) ([]*aggregate.OrganizationUserAggregate, error)
	FindByOrganizationID(ctx context.Context, organizationID uuid.UUID) ([]*aggregate.OrganizationUserAggregate, error)
}

type IOrganizationUserWriteRepository interface {
	Create(ctx context.Context, organizationUser *aggregate.OrganizationUserAggregate) (*aggregate.OrganizationUserAggregate, error)
	Update(ctx context.Context, organizationUser *aggregate.OrganizationUserAggregate) (*aggregate.OrganizationUserAggregate, error)
	Delete(ctx context.Context, organizationUser *aggregate.OrganizationUserAggregate) error
}

type IOrganizationUserRepository interface {
	ITransaction
	IOrganizationUserReadRepository
	IOrganizationUserWriteRepository
}
