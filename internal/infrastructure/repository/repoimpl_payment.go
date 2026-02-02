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
		SetUserID(paymentAggregate.Payment.UserID).
		SetService(paymentAggregate.Payment.Service).
		SetCurrency(paymentAggregate.Payment.Currency).
		SetStatus(payment.Status(paymentAggregate.Payment.Status.String())).
		SetAmount(paymentAggregate.Payment.Amount).
		SetDescription(paymentAggregate.Payment.Description).
		SetPaidAt(paymentAggregate.Payment.PaidAt).
		SetPaymentType(payment.PaymentType(paymentAggregate.Payment.PaymentType.String())).
		// channel info
		SetChannel(payment.Channel(paymentAggregate.Payment.ChannelInfo.Channel.String())).
		SetPaymentType(payment.PaymentType(paymentAggregate.Payment.PaymentType.String())).
		SetWechatTransactionID(paymentAggregate.Payment.ChannelInfo.WeChatTransactionID).
		SetWechatPlatform(paymentAggregate.Payment.ChannelInfo.WechatPlatform.String()).
		SetWechatOpenID(paymentAggregate.Payment.ChannelInfo.WeChatOpenID).
		SetStripeSubscriptionID(paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID).
		SetStripeSubscriptionStatus(paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus.String()).
		SetStripeInterval(paymentAggregate.Payment.ChannelInfo.StripeInterval.String()).
		SetStripeCurrentPeriodStart(paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodStart).
		SetStripeCurrentPeriodEnd(paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodEnd).
		SetStripeCustomerID(paymentAggregate.Payment.ChannelInfo.StripeCustomerID).
		SetStripeCustomerEmail(paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail).
		SetStripeCheckoutSessionID(paymentAggregate.Payment.ChannelInfo.StripeCheckoutSessionID).
		SetStripeCancelAtPeriodEnd(paymentAggregate.Payment.ChannelInfo.StripeCancelAtPeriodEnd).
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
		).
		// channel info
		SetWechatTransactionID(paymentAggregate.Payment.ChannelInfo.WeChatTransactionID).
		SetWechatOpenID(paymentAggregate.Payment.ChannelInfo.WeChatOpenID).
		SetWechatPlatform(paymentAggregate.Payment.ChannelInfo.WechatPlatform.String()).
		SetStripeSubscriptionID(paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID).
		SetStripeSubscriptionStatus(paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus.String()).
		SetStripeInterval(paymentAggregate.Payment.ChannelInfo.StripeInterval.String()).
		SetStripeCurrentPeriodStart(paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodStart).
		SetStripeCurrentPeriodEnd(paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodEnd).
		SetStripeCustomerID(paymentAggregate.Payment.ChannelInfo.StripeCustomerID).
		SetStripeCustomerEmail(paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail).
		SetStripeCheckoutSessionID(paymentAggregate.Payment.ChannelInfo.StripeCheckoutSessionID).
		SetStripeCancelAtPeriodEnd(paymentAggregate.Payment.ChannelInfo.StripeCancelAtPeriodEnd).
		// status
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

	paymentDO, err := db.Payment.Query().Where(payment.StripeSubscriptionID(subscriptionID)).Only(ctx)
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

	paymentDO, err := db.Payment.Query().Where(payment.StripeCheckoutSessionID(sessionID)).Only(ctx)
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

func (p *paymentImpl) FindActiveSubscriptionByUserIDAndService(ctx context.Context, userID string, service string) (*aggregate.PaymentAggregate, error) {
	db := p.getEntClient(ctx)

	paymentDO, err := db.Payment.Query().
		Where(
			payment.UserID(userID),
			payment.Service(service),
			payment.StripeSubscriptionStatusEQ(enum.SubscriptionStatusActive.String()),
		).
		Order(ent.Desc(payment.FieldCreatedAt)).
		First(ctx)
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
