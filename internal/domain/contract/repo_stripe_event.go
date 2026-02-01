package contract

import (
	"context"
	"kiwi-user/internal/domain/model/entity"
)

type IStripeEventReadRepository interface {
	FindByEventID(ctx context.Context, eventID string) (*entity.StripeEventEntity, error)
	ExistsByEventID(ctx context.Context, eventID string) (bool, error)
}

type IStripeEventWriteRepository interface {
	Create(ctx context.Context, event *entity.StripeEventEntity) (*entity.StripeEventEntity, error)
}

type IStripeEventRepository interface {
	ITransaction
	IStripeEventReadRepository
	IStripeEventWriteRepository
}
