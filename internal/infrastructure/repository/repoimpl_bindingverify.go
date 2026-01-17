package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/bindingverify"

	"github.com/futurxlab/golanggraph/xerror"
)

type bindingVerifyImpl struct {
	baseImpl
}

func (b *bindingVerifyImpl) Find(ctx context.Context, userID string, bindingType enum.BindingType, identity string) (*entity.BindingVerifyEntity, error) {
	db := b.getEntClient(ctx)

	bindingVerifyDO, err := db.BindingVerify.Query().
		Where(bindingverify.UserID(userID)).
		Where(bindingverify.TypeEQ(bindingverify.Type(bindingType))).
		Where(bindingverify.Identity(identity)).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if bindingVerifyDO == nil {
		return nil, nil
	}

	return convertBindingVerifyDOToEntity(bindingVerifyDO), nil
}

func (b *bindingVerifyImpl) Create(ctx context.Context, bindingVerifyEntity *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error) {
	db := b.getEntClient(ctx)

	bindingVerifyDO, err := db.BindingVerify.Create().
		SetUserID(bindingVerifyEntity.UserID).
		SetType(bindingverify.Type(bindingVerifyEntity.Type)).
		SetIdentity(bindingVerifyEntity.Identity).
		SetCode(bindingVerifyEntity.Code).
		SetExpiresAt(bindingVerifyEntity.ExpiresAt).
		Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return convertBindingVerifyDOToEntity(bindingVerifyDO), nil
}

func (b *bindingVerifyImpl) Delete(ctx context.Context, bindingVerifyEntity *entity.BindingVerifyEntity) error {

	db := b.getEntClient(ctx)

	err := db.BindingVerify.DeleteOneID(bindingVerifyEntity.ID).Exec(ctx)

	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (b *bindingVerifyImpl) Update(ctx context.Context, bindingVerify *entity.BindingVerifyEntity) (*entity.BindingVerifyEntity, error) {
	db := b.getEntClient(ctx)

	query := db.BindingVerify.UpdateOneID(bindingVerify.ID).
		SetCode(bindingVerify.Code).
		SetExpiresAt(bindingVerify.ExpiresAt)

	if bindingVerify.VerifiedAt.Unix() > 0 {
		query.SetVerifiedAt(bindingVerify.VerifiedAt)
	}

	bindingVerifyDO, err := query.Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return convertBindingVerifyDOToEntity(bindingVerifyDO), nil
}

func NewBindingVerifyImpl(db *Client) contract.IBindingVerifyRepository {
	return &bindingVerifyImpl{
		baseImpl{
			db: db,
		},
	}
}
