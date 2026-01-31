package repository

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/infrastructure/repository/ent"
	"kiwi-user/internal/infrastructure/repository/ent/payment"
	"time"
)

type paymentImpl struct {
	baseImpl
}

func (p *paymentImpl) FindByOutTradeNo(ctx context.Context, outTradeNo string) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDO, err := db.Payment.Query().Where(payment.OutTradeNo(outTradeNo)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if paymentDO == nil {
		return nil, nil
	}

	return &aggregate.PaymentAggregate{
		Payment: convertPaymentDOToEntity(paymentDO),
	}, nil
}

func (p *paymentImpl) Create(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDO, err := db.Payment.Create().
		SetOutTradeNo(paymentAggregate.Payment.OutTradeNo).
		SetChannel(payment.Channel(paymentAggregate.Payment.ChannelInfo.Channel.String())).
		SetPlatform(payment.Platform(paymentAggregate.Payment.ChannelInfo.Platform.String())).
		SetService(paymentAggregate.Payment.Service).
		SetUserID(paymentAggregate.Payment.UserID).
		SetOpenID(paymentAggregate.Payment.ChannelInfo.OpenID).
		SetAmount(paymentAggregate.Payment.Amount).
		SetCurrency(paymentAggregate.Payment.Currency).
		SetStatus(payment.Status(paymentAggregate.Payment.Status.String())).
		SetDescription(paymentAggregate.Payment.Description).
		SetCreatedAt(paymentAggregate.Payment.CreatedAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &aggregate.PaymentAggregate{
		Payment: convertPaymentDOToEntity(paymentDO),
	}, nil
}

func (p *paymentImpl) Update(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	_, err := db.Payment.Update().
		Where(
			payment.OutTradeNo(paymentAggregate.Payment.OutTradeNo),
			payment.StatusNEQ(payment.StatusNOTPAY),
		).
		SetTransactionID(paymentAggregate.Payment.ChannelInfo.TransactionID).
		SetStatus(payment.Status(paymentAggregate.Payment.Status.String())).
		SetPaidAt(paymentAggregate.Payment.PaidAt).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return paymentAggregate, nil
}

func (p *paymentImpl) FindPendingPayments(ctx context.Context, status enum.PaymentStatus, createdBefore time.Time) ([]*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDOs, err := db.Payment.Query().
		Where(
			payment.StatusEQ(payment.Status(status.String())),
			payment.CreatedAtLT(createdBefore),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	var payments []*aggregate.PaymentAggregate

	for _, paymentDO := range paymentDOs {
		payments = append(payments, &aggregate.PaymentAggregate{
			Payment: convertPaymentDOToEntity(paymentDO),
		})
	}

	return payments, nil
}

func (p *paymentImpl) FindBySubscriptionID(ctx context.Context, subscriptionID string) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDO, err := db.Payment.Query().Where(payment.SubscriptionID(subscriptionID)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if paymentDO == nil {
		return nil, nil
	}

	return &aggregate.PaymentAggregate{
		Payment: convertPaymentDOToEntity(paymentDO),
	}, nil
}

func (p *paymentImpl) FindByCheckoutSessionID(ctx context.Context, sessionID string) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDO, err := db.Payment.Query().Where(payment.CheckoutSessionID(sessionID)).Only(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if paymentDO == nil {
		return nil, nil
	}

	return &aggregate.PaymentAggregate{
		Payment: convertPaymentDOToEntity(paymentDO),
	}, nil
}

func NewPaymentImpl(db *Client) contract.IPaymentRepository {
	return &paymentImpl{
		baseImpl{
			db: db,
		},
	}
}
