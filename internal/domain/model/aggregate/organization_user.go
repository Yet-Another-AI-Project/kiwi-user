package aggregate

import "kiwi-user/internal/domain/model/entity"

type OrganizationUserAggregate struct {
	Organization     *entity.OrganizationEntity
	Application      *entity.ApplicationEntity
	User             *entity.UserEntity
	OrganizationRole *entity.RoleEntity
}
