package aggregate

import "kiwi-user/internal/domain/model/entity"

type UserAggregate struct {
	User            *entity.UserEntity
	Application     *entity.ApplicationEntity
	Bindings        []*entity.BindingEntity
	PersonalRole    *entity.RoleEntity
	WechatOpenIDs   []*entity.WechatOpenIDEntity
	QyWechatUserIDs []*entity.QyWechatUserIDEntity
}
