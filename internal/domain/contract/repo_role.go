package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
)

type IRoleReadRepository interface {
	FindByName(ctx context.Context, application string, name string) (*aggregate.RoleAggregate, error)
}

type IRoleWriteRepository interface {
	Create(ctx context.Context, role *aggregate.RoleAggregate) (*aggregate.RoleAggregate, error)
	Update(ctx context.Context, role *aggregate.RoleAggregate) (*aggregate.RoleAggregate, error)
	Delete(ctx context.Context, role *aggregate.RoleAggregate) error
}

type IRoleRepository interface {
	ITransaction
	IRoleReadRepository
	IRoleWriteRepository
}
