package aggregate

import "kiwi-user/internal/domain/model/entity"

type RoleAggregate struct {
	Role        *entity.RoleEntity
	Application *entity.ApplicationEntity
	Scopes      []*entity.ScopeEntity
}
