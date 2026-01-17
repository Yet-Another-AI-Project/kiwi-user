package aggregate

import "kiwi-user/internal/domain/model/entity"

type OrganizationApplicationAggregate struct {
	OrganizationApplication *entity.OrganizationApplicationEntity
	Application             *entity.ApplicationEntity
}
