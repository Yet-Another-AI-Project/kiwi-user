package contract

import (
	"context"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/enum"
	"time"
)

type IPaymentReadRepository interface {
	FindByOutTradeNo(ctx context.Context, outTradeNo string) (*aggregate.PaymentAggregate, error)
	FindPendingPayments(ctx context.Context, status enum.PaymentStatus, createdBefore time.Time) ([]*aggregate.PaymentAggregate, error)
	FindBySubscriptionID(ctx context.Context, subscriptionID string) (*aggregate.PaymentAggregate, error)
	FindByCheckoutSessionID(ctx context.Context, sessionID string) (*aggregate.PaymentAggregate, error)
}

type IPaymentWriteRepository interface {
	Create(ctx context.Context, payment *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error)
	Update(ctx context.Context, payment *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error)
}

type IPaymentRepository interface {
	ITransaction
	IPaymentReadRepository
	IPaymentWriteRepository
}
