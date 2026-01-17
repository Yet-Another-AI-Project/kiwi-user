package contract

import (
	"context"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
)

type IMailVertifyCodeReadRepository interface {
	Find(ctx context.Context, email string, codetype enum.VertificationCodeType) (*entity.MailVertifyCodeEntity, error)
}

type IMailVertifyCodeWriteRepository interface {
	Create(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error
	Update(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error
	Delete(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error
}

type IMailVertifyCodeRepository interface {
	ITransaction
	IMailVertifyCodeReadRepository
	IMailVertifyCodeWriteRepository
}
