package aggregate

import "kiwi-user/internal/domain/model/entity"

type PaymentAggregate struct {
	Payment *entity.PaymentEntity
}
