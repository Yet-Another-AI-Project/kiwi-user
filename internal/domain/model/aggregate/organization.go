package aggregate

import "kiwi-user/internal/domain/model/entity"

type OrganizationAggregate struct {
	Organization *entity.OrganizationEntity
	Application  *entity.ApplicationEntity
}
