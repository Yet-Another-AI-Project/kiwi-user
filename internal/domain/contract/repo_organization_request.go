package contract

import (
	"context"
	"kiwi-user/internal/domain/model/entity"
)

type IOrganizationRequestReadRepository interface {
	Find(ctx context.Context, request *entity.OrganizationRequestEntity) (*entity.OrganizationRequestEntity, error)
}

type IOrganizationRequestWriteRepository interface {
	Create(ctx context.Context, request *entity.OrganizationRequestEntity) (*entity.OrganizationRequestEntity, error)
}

type IOrganizationRequestRepository interface {
	ITransaction
	IOrganizationRequestReadRepository
	IOrganizationRequestWriteRepository
}
