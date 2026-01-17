package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"

	"github.com/google/uuid"
)

type IUserReadRepository interface {
	Find(ctx context.Context, id string) (*aggregate.UserAggregate, error)
	FindIn(ctx context.Context, ids []string) ([]*aggregate.UserAggregate, error)
	FindByName(ctx context.Context, application string, name string) (*aggregate.UserAggregate, error)
	FindByBindingForUpdate(ctx context.Context, applicationID uuid.UUID, binding *entity.BindingEntity) (*aggregate.UserAggregate, error)
	FindWechatOpenIDByUserAndPlatform(ctx context.Context, userID string, platform string) (*entity.WechatOpenIDEntity, error)
	FindByWechatOpenIDAndPlatformForUpdate(ctx context.Context, applicationID uuid.UUID, openID string, platform string) (*aggregate.UserAggregate, error)
}

type IUserWriteRepository interface {
	Update(ctx context.Context, user *aggregate.UserAggregate) (*aggregate.UserAggregate, error)
	Create(ctx context.Context, user *aggregate.UserAggregate) (*aggregate.UserAggregate, error)
}

type IUserRepository interface {
	ITransaction
	IUserReadRepository
	IUserWriteRepository
}
