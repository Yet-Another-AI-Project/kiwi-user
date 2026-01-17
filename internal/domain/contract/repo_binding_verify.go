package contract

import (
	"context"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
)

type IBindingVerifyReadRepository interface {
	Find(ctx context.Context, userID string, bindingType enum.BindingType, identity string) (*entity.BindingVerifyEntity, error)
}

type IBindingVerifyWriteRepository interface {
	Create(ctx context.Context, bindingVerifyEntity *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error)
	Update(ctx context.Context, bindingVerifyEntity *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error)
	Delete(ctx context.Context, bindingVerifyEntity *entity.BindingVerifyEntity) error
}

type IBindingVerifyRepository interface {
	ITransaction
	IBindingVerifyReadRepository
	IBindingVerifyWriteRepository
}
