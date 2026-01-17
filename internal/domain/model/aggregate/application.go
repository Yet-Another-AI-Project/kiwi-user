package aggregate

import "kiwi-user/internal/domain/model/entity"

type ApplicationAggregate struct {
	Application         *entity.ApplicationEntity
	Roles               []*entity.RoleEntity
	DefaultPersonalRole *entity.RoleEntity
	DefaultOrgRole      *entity.RoleEntity
	DefaultOrgAdminRole *entity.RoleEntity
}
