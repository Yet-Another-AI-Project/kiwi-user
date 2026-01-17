package aggregate

import "kiwi-user/internal/domain/model/entity"

type DeviceAggregate struct {
	Device *entity.DeviceEntity
	User   *entity.UserEntity
}
