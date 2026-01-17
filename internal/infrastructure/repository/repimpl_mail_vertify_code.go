package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/mailvertifycode"
)

type mailVertifyCodeImpl struct {
	baseImpl
}

func (m *mailVertifyCodeImpl) Find(ctx context.Context, email string, codetype enum.VertificationCodeType) (*entity.MailVertifyCodeEntity, error) {
	mailVertifyCode, err := m.db.MailVertifyCode.Query().Where(mailvertifycode.EmailEQ(email), mailvertifycode.TypeEQ(mailvertifycode.Type(codetype.String()))).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &entity.MailVertifyCodeEntity{
		ID:        mailVertifyCode.ID,
		Email:     mailVertifyCode.Email,
		Code:      mailVertifyCode.Code,
		Type:      enum.ParseVertificationCodeType(mailVertifyCode.Type.String()),
		ExpiresAt: mailVertifyCode.ExpiresAt,
	}, nil
}

func (m *mailVertifyCodeImpl) Create(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error {
	_, err := m.db.MailVertifyCode.Create().
		SetEmail(mailVertifyCode.Email).
		SetCode(mailVertifyCode.Code).
		SetType(mailvertifycode.Type(mailVertifyCode.Type.String())).
		SetExpiresAt(mailVertifyCode.ExpiresAt).
		Save(ctx)
	return err
}

func (m *mailVertifyCodeImpl) Update(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error {
	_, err := m.db.MailVertifyCode.UpdateOneID(mailVertifyCode.ID).
		SetEmail(mailVertifyCode.Email).
		SetCode(mailVertifyCode.Code).
		SetType(mailvertifycode.Type(mailVertifyCode.Type.String())).
		SetExpiresAt(mailVertifyCode.ExpiresAt).
		Save(ctx)
	return err
}

func (m *mailVertifyCodeImpl) Delete(ctx context.Context, mailVertifyCode *entity.MailVertifyCodeEntity) error {
	return m.db.MailVertifyCode.DeleteOneID(mailVertifyCode.ID).Exec(ctx)
}

func NewMailVertifyCodeImpl(client *Client) contract.IMailVertifyCodeRepository {
	return &mailVertifyCodeImpl{
		baseImpl: baseImpl{db: client},
	}
}
