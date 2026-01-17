package application

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/utils"
	"time"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
)

const (
	defaultBindingVerifyExpireSecond = 60 * 10
)

type BindingApplication struct {
	userReadRepository          contract.IUserReadRepository
	bindingVerifyReadRepository contract.IBindingVerifyReadRepository
	bindingService              *service.BindingService
}

func NewBindingApplication(
	userReadRepository contract.IUserReadRepository,
	bindingVerifyReadRepository contract.IBindingVerifyReadRepository,
	bindingService *service.BindingService,
) *BindingApplication {
	return &BindingApplication{
		userReadRepository:          userReadRepository,
		bindingVerifyReadRepository: bindingVerifyReadRepository,
		bindingService:              bindingService,
	}
}

func (b *BindingApplication) SendPhoneVerifyCode(ctx context.Context, userID string, request *dto.SendPhoneVerifyCodeRequest) *facade.Error {

	bindingVerifyEntity, err := b.bindingVerifyReadRepository.Find(ctx, userID, enum.BindingTypePhone, request.Phone)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	code := utils.GenerateVerificationCode()

	if bindingVerifyEntity == nil {
		bindingVerifyEntity = &entity.BindingVerifyEntity{
			UserID:    userID,
			Type:      enum.BindingTypePhone,
			Identity:  request.Phone,
			Code:      code,
			ExpiresAt: time.Now().Add(defaultBindingVerifyExpireSecond),
		}

		bindingVerifyEntity, err = b.bindingService.CreateBindingVerify(ctx, bindingVerifyEntity)
		if err != nil {
			return facade.ErrServerInternal.Wrap(err)
		}
	} else {
		if bindingVerifyEntity.UpdatedAt.Add(time.Second * 60).After(time.Now()) {
			return facade.ErrForbidden.Facade("request too frequent")
		}

		if bindingVerifyEntity.VerifiedAt.Unix() > 0 {
			return facade.ErrForbidden.Facade("phone already verified")
		}

		bindingVerifyEntity.Code = code
		bindingVerifyEntity.ExpiresAt = time.Now().Add(defaultBindingVerifyExpireSecond)

		bindingVerifyEntity, err = b.bindingService.UpdateBindingVerify(ctx, bindingVerifyEntity)
		if err != nil {
			return facade.ErrServerInternal.Wrap(err)
		}
	}

	// TODO: send sms through huaweicloud sms service

	return nil
}

func (b *BindingApplication) VerifyPhoneCode(ctx context.Context, userID string, request *dto.VerifyPhoneCodeRequest) *facade.Error {

	userAggregate, err := b.userReadRepository.Find(ctx, userID)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if userAggregate == nil {
		return facade.ErrForbidden.Facade("user not found")
	}

	bindingVerifyEntity, err := b.bindingVerifyReadRepository.Find(ctx, userID, enum.BindingTypePhone, request.Phone)
	if err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	if bindingVerifyEntity == nil {
		return facade.ErrForbidden.Facade("verify code not found")
	}

	if bindingVerifyEntity.VerifiedAt.Unix() > 0 {
		return facade.ErrForbidden.Facade("phone already verified")
	}

	if bindingVerifyEntity.ExpiresAt.Before(time.Now()) {
		return facade.ErrForbidden.Facade("verify code expired")
	}

	if bindingVerifyEntity.Code != request.Code {
		return facade.ErrForbidden.Facade("verify code not match")
	}

	if err := b.bindingService.BindingCodeVerified(ctx, userAggregate, bindingVerifyEntity); err != nil {
		return facade.ErrServerInternal.Wrap(err)
	}

	return nil
}
