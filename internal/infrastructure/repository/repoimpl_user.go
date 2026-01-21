package repository

import (
	"context"
	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/application"
	"kiwi-user/internal/infrastructure/repository/ent/binding"
	"kiwi-user/internal/infrastructure/repository/ent/user"
	"kiwi-user/internal/infrastructure/repository/ent/wechatopenid"

	"github.com/bwmarrin/snowflake"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/google/uuid"
)

type userImpl struct {
	baseImpl
	node *snowflake.Node
}

func (u *userImpl) Find(ctx context.Context, id string) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	userDO, err := db.User.Get(ctx, id)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if userDO == nil {
		return nil, nil
	}

	applicationDO, err := userDO.QueryApplication().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDOs, err := userDO.QueryBindings().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleDO, err := userDO.QueryPersonalRole().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	openidDOs, err := userDO.QueryOpenIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	qyWechatUserIDDOs, err := userDO.QueryQyWechatUserIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.UserAggregate{
		User:            convertUserDOToEntity(userDO),
		Application:     convertApplicationDOToEntity(applicationDO),
		Bindings:        converBindingDOsToEntities(bindingDOs),
		PersonalRole:    convertRoleDOToEntity(roleDO),
		WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
		QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
	}, nil
}

func (u *userImpl) FindIn(ctx context.Context, ids []string) ([]*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	userDOs, err := db.User.Query().Where(user.IDIn(ids...)).All(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if userDOs == nil {
		return nil, nil
	}

	userAggregates := make([]*aggregate.UserAggregate, 0)

	for _, userDO := range userDOs {
		applicationDO, err := userDO.QueryApplication().Only(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		bindingDOs, err := userDO.QueryBindings().All(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		roleDO, err := userDO.QueryPersonalRole().Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, xerror.Wrap(err)
		}

		openidDOs, err := userDO.QueryOpenIds().All(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		qyWechatUserIDDOs, err := userDO.QueryQyWechatUserIds().All(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}

		userAggregates = append(userAggregates, &aggregate.UserAggregate{
			User:            convertUserDOToEntity(userDO),
			Application:     convertApplicationDOToEntity(applicationDO),
			Bindings:        converBindingDOsToEntities(bindingDOs),
			PersonalRole:    convertRoleDOToEntity(roleDO),
			WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
			QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
		})
	}

	return userAggregates, nil
}

func (u *userImpl) FindByName(ctx context.Context, applicationName string, name string) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	userDO, err := db.User.Query().Where(user.Name(name), user.HasApplicationWith(application.Name(applicationName))).Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if userDO == nil {
		return nil, nil
	}

	applicationDO, err := userDO.QueryApplication().Only(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDOs, err := userDO.QueryBindings().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleDO, err := userDO.QueryPersonalRole().Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	openidDOs, err := userDO.QueryOpenIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	qyWechatUserIDDOs, err := userDO.QueryQyWechatUserIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.UserAggregate{
		User:            convertUserDOToEntity(userDO),
		Application:     convertApplicationDOToEntity(applicationDO),
		Bindings:        converBindingDOsToEntities(bindingDOs),
		PersonalRole:    convertRoleDOToEntity(roleDO),
		WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
		QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
	}, nil
}

func (u *userImpl) FindByBindingForUpdate(ctx context.Context, applicationID uuid.UUID, bindingEntity *entity.BindingEntity) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	// advisory lock
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", applicationID.String()+bindingEntity.Type.String()+bindingEntity.Identity); err != nil {
		return nil, xerror.Wrap(err)
	}

	applicationDO, err := db.Application.Query().Where(application.ID(applicationID)).Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDO, err := db.Binding.Query().Where(
		binding.ApplicationIDEQ(applicationDO.ID),
		binding.TypeEQ(binding.Type(bindingEntity.Type)),
		binding.IdentityEQ(bindingEntity.Identity),
	).ForUpdate().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if bindingDO == nil {
		return nil, nil
	}

	userDO, err := bindingDO.QueryUser().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDOs, err := userDO.QueryBindings().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleDO, err := userDO.QueryPersonalRole().Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	openidDOs, err := userDO.QueryOpenIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	qyWechatUserIDDOs, err := userDO.QueryQyWechatUserIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.UserAggregate{
		User:            convertUserDOToEntity(userDO),
		Application:     convertApplicationDOToEntity(applicationDO),
		Bindings:        converBindingDOsToEntities(bindingDOs),
		PersonalRole:    convertRoleDOToEntity(roleDO),
		WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
		QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
	}, nil
}

func (u *userImpl) Update(ctx context.Context, user *aggregate.UserAggregate) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	// TODO: handle binding delete
	for _, bindingEntity := range user.Bindings {
		if bindingEntity.ID == uuid.Nil {
			bindingDO, err := db.Binding.Create().
				SetType(binding.Type(bindingEntity.Type)).
				SetIdentity(bindingEntity.Identity).
				SetEmail(bindingEntity.Email).
				SetVerified(bindingEntity.Verified).
				SetSalt(bindingEntity.Salt).
				SetUserID(user.User.ID).
				Save(ctx)

			if err != nil {
				return nil, xerror.Wrap(err)
			}
			bindingEntity.ID = bindingDO.ID
		} else {
			if _, err := db.Binding.UpdateOneID(bindingEntity.ID).
				SetType(binding.Type(bindingEntity.Type)).
				SetIdentity(bindingEntity.Identity).
				SetEmail(bindingEntity.Email).
				SetVerified(bindingEntity.Verified).
				SetSalt(bindingEntity.Salt).
				Save(ctx); err != nil {
				return nil, xerror.Wrap(err)
			}
		}
	}

	for _, WechatOpenIDEntity := range user.WechatOpenIDs {
		if WechatOpenIDEntity.ID == uuid.Nil {
			openIDDO, err := db.WechatOpenID.Create().
				SetOpenID(WechatOpenIDEntity.OpenID).
				SetPlatform(wechatopenid.Platform(WechatOpenIDEntity.Platform)).
				SetUserID(user.User.ID).
				Save(ctx)
			if err != nil {
				return nil, xerror.Wrap(err)
			}
			WechatOpenIDEntity.ID = openIDDO.ID
		}
	}

	for _, qyWechatUserIDEntity := range user.QyWechatUserIDs {
		if qyWechatUserIDEntity.ID == uuid.Nil {
			createQuery := db.QyWechatUserID.Create().
				SetUserID(user.User.ID)
			if qyWechatUserIDEntity.QyWechatUserID != "" {
				createQuery = createQuery.SetQyWechatUserID(qyWechatUserIDEntity.QyWechatUserID)
			}
			if qyWechatUserIDEntity.OpenID != "" {
				createQuery = createQuery.SetOpenID(qyWechatUserIDEntity.OpenID)
			}
			qyWechatUserIDDO, err := createQuery.Save(ctx)
			if err != nil {
				return nil, xerror.Wrap(err)
			}
			qyWechatUserIDEntity.ID = qyWechatUserIDDO.ID
		}
	}

	query := db.User.UpdateOneID(user.User.ID).
		SetDisplayName(user.User.DisplayName).
		SetAvatar(user.User.Avatar).
		SetDepartment(user.User.Department)

	if user.PersonalRole != nil {
		query = query.SetPersonalRoleID(user.PersonalRole.ID)
	}

	userDO, err := query.Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	openidDOs, err := userDO.QueryOpenIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	qyWechatUserIDDOs, err := userDO.QueryQyWechatUserIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.UserAggregate{
		User:            convertUserDOToEntity(userDO),
		Application:     user.Application,
		Bindings:        user.Bindings,
		PersonalRole:    user.PersonalRole,
		WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
		QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
	}, nil
}

func (u *userImpl) Create(ctx context.Context, user *aggregate.UserAggregate) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	createQuery := db.User.Create().
		SetID(u.node.Generate().String()).
		SetName(user.User.Name).
		SetDisplayName(user.User.DisplayName).
		SetAvatar(user.User.Avatar).
		SetReferralChannel(user.User.RefferalChannel).
		SetApplicationID(user.Application.ID)

	if user.PersonalRole != nil {
		createQuery = createQuery.SetPersonalRoleID(user.PersonalRole.ID)
	}

	userDO, err := createQuery.Save(ctx)

	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDOs := make([]*ent.Binding, 0)

	for _, bindingEntity := range user.Bindings {
		bindingDO, err := db.Binding.Create().
			SetApplicationID(user.Application.ID).
			SetType(binding.Type(bindingEntity.Type)).
			SetIdentity(bindingEntity.Identity).
			SetEmail(bindingEntity.Email).
			SetVerified(bindingEntity.Verified).
			SetSalt(bindingEntity.Salt).
			SetUser(userDO).
			Save(ctx)

		if err != nil {
			return nil, xerror.Wrap(err)
		}

		bindingDOs = append(bindingDOs, bindingDO)
	}

	openidDOs := make([]*ent.WechatOpenID, 0)

	for _, WechatOpenIDEntity := range user.WechatOpenIDs {
		openidDO, err := db.WechatOpenID.Create().
			SetOpenID(WechatOpenIDEntity.OpenID).
			SetPlatform(wechatopenid.Platform(WechatOpenIDEntity.Platform)).
			SetUser(userDO).
			Save(ctx)

		if err != nil {
			return nil, xerror.Wrap(err)
		}

		openidDOs = append(openidDOs, openidDO)
	}

	qyWechatUserIDDOs := make([]*ent.QyWechatUserID, 0)

	for _, qyWechatUserIDEntity := range user.QyWechatUserIDs {
		createQuery := db.QyWechatUserID.Create().
			SetUser(userDO)
		if qyWechatUserIDEntity.QyWechatUserID != "" {
			createQuery = createQuery.SetQyWechatUserID(qyWechatUserIDEntity.QyWechatUserID)
		}
		if qyWechatUserIDEntity.OpenID != "" {
			createQuery = createQuery.SetOpenID(qyWechatUserIDEntity.OpenID)
		}
		qyWechatUserIDDO, err := createQuery.Save(ctx)
		if err != nil {
			return nil, xerror.Wrap(err)
		}
		qyWechatUserIDDOs = append(qyWechatUserIDDOs, qyWechatUserIDDO)
	}

	return &aggregate.UserAggregate{
		User:            convertUserDOToEntity(userDO),
		Application:     user.Application,
		Bindings:        converBindingDOsToEntities(bindingDOs),
		WechatOpenIDs:   convertWechatOpenIDDOToEntities(openidDOs),
		QyWechatUserIDs: convertQyWechatUserIDDOToEntities(qyWechatUserIDDOs),
		PersonalRole:    user.PersonalRole,
	}, nil
}

func (u *userImpl) Delete(ctx context.Context, user *aggregate.UserAggregate) error {
	panic("not implemented")
}

func (u *userImpl) FindWechatOpenIDByUserAndPlatform(ctx context.Context, userID string, platform string) (*entity.WechatOpenIDEntity, error) {
	db := u.getEntClient(ctx)

	openidDO, err := db.WechatOpenID.Query().Where(wechatopenid.HasUserWith(user.ID(userID)), wechatopenid.PlatformEQ(wechatopenid.Platform(platform))).Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if openidDO == nil {
		return nil, nil
	}

	return convertWechatOpenIDDOToEntity(openidDO), nil
}

func (u *userImpl) FindByWechatOpenIDAndPlatformForUpdate(ctx context.Context, applicationID uuid.UUID, openID string, platform string) (*aggregate.UserAggregate, error) {
	db := u.getEntClient(ctx)

	// advisory lock
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", applicationID.String()+platform+openID); err != nil {
		return nil, xerror.Wrap(err)
	}

	applicationDO, err := db.Application.Query().Where(application.ID(applicationID)).Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	openidDO, err := db.WechatOpenID.Query().Where(
		wechatopenid.OpenIDEQ(openID),
		wechatopenid.PlatformEQ(wechatopenid.Platform(platform)),
	).ForUpdate().Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	if openidDO == nil {
		return nil, nil
	}

	userDO, err := openidDO.QueryUser().Only(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	bindingDOs, err := userDO.QueryBindings().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	openidDOs, err := userDO.QueryOpenIds().All(ctx)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	roleDO, err := userDO.QueryPersonalRole().Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, xerror.Wrap(err)
	}

	return &aggregate.UserAggregate{
		User:          convertUserDOToEntity(userDO),
		Application:   convertApplicationDOToEntity(applicationDO),
		Bindings:      converBindingDOsToEntities(bindingDOs),
		WechatOpenIDs: convertWechatOpenIDDOToEntities(openidDOs),
		PersonalRole:  convertRoleDOToEntity(roleDO),
	}, nil
}

func NewUserImpl(db *Client, cfg *config.Config) (contract.IUserRepository, error) {
	node, err := snowflake.NewNode(cfg.APIServer.Index)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &userImpl{
		baseImpl: baseImpl{
			db: db,
		},
		node: node,
	}, nil
}
