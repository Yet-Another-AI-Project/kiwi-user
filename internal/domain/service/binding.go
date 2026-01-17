package service

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"time"

	"github.com/futurxlab/golanggraph/xerror"
)

type BindingService struct {
	userRepository          contract.IUserRepository
	bindingVerifyRepository contract.IBindingVerifyRepository
}

func NewBindingService(
	bindingVerifyRepository contract.IBindingVerifyRepository,
) *BindingService {
	return &BindingService{
		bindingVerifyRepository: bindingVerifyRepository,
	}
}

func (b *BindingService) CreateBindingVerify(ctx context.Context, bindingVerify *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error) {
	return b.bindingVerifyRepository.Create(ctx, bindingVerify)
}

func (b *BindingService) UpdateBindingVerify(ctx context.Context, bindingVerify *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error) {
	return b.bindingVerifyRepository.Update(ctx, bindingVerify)
}

func (b *BindingService) BindingCodeVerified(ctx context.Context, userAggregate *aggregate.UserAggregate, bindingVerify *entity.BindingVerifyEntity) error {
	err := b.bindingVerifyRepository.WithTransaction(ctx, func(ctx context.Context) error {
		// update binding verify
		bindingVerify.VerifiedAt = time.Now()
		if _, err := b.bindingVerifyRepository.Update(ctx, bindingVerify); err != nil {
			return xerror.Wrap(err)
		}

		// create user binding
		userAggregate.Bindings = append(userAggregate.Bindings, &entity.BindingEntity{
			Type:     bindingVerify.Type,
			Identity: bindingVerify.Identity,
			Verified: true,
		})

		_, err := b.userRepository.Update(ctx, userAggregate)

		if err != nil {
			return xerror.Wrap(err)
		}

		return nil
	})

	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
