package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
)

type IApplicationReadRepository interface {
	FindByName(ctx context.Context, name string) (*aggregate.ApplicationAggregate, error)
	FindAll(ctx context.Context, offset, limit int) ([]*aggregate.ApplicationAggregate, error)
}

type IApplicationWriteRepository interface {
	Create(ctx context.Context, application *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error)
	Update(ctx context.Context, application *aggregate.ApplicationAggregate) (*aggregate.ApplicationAggregate, error)
	Delete(ctx context.Context, application *aggregate.ApplicationAggregate) error
}

type IApplicationRepository interface {
	ITransaction
	IApplicationReadRepository
	IApplicationWriteRepository
}
