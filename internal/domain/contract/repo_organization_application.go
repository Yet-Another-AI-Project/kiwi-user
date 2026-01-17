package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"

	"github.com/google/uuid"
)

type IOrganizationApplicationReadRepository interface {
	FindByUserID(ctx context.Context, request *entity.OrganizationApplicationEntity) ([]*aggregate.OrganizationApplicationAggregate, error)
	FindByID(ctx context.Context, ID uuid.UUID) (*aggregate.OrganizationApplicationAggregate, error)
	PageFind(ctx context.Context, request *entity.OrganizationApplicationEntity, offset int, limit int) ([]*aggregate.OrganizationApplicationAggregate, int, error)
}

type IOrganizationApplicationWriteRepository interface {
	Create(ctx context.Context, orgAppEntity *entity.OrganizationApplicationEntity, appEntity *entity.ApplicationEntity) (*aggregate.OrganizationApplicationAggregate, error)
	Update(ctx context.Context, request *entity.OrganizationApplicationEntity) (*aggregate.OrganizationApplicationAggregate, error)
}

type IOrganizationApplicationRepository interface {
	ITransaction
	IOrganizationApplicationReadRepository
	IOrganizationApplicationWriteRepository
}
