package application

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"net/http"
	"time"

	"github.com/avast/retry-go"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

type PaymentApplication struct {
	userReadRepository    contract.IUserReadRepository
	paymentReadRepository contract.IPaymentReadRepository
	paymentService        *service.PaymentService
	stripePaymentService  *service.StripePaymentService

	logger logger.ILogger
}

func NewPaymentApplication(
	userReadRepository contract.IUserReadRepository,
	paymentReadRepository contract.IPaymentReadRepository,
	paymentService *service.PaymentService,
	stripePaymentService *service.StripePaymentService,
	logger logger.ILogger,
) *PaymentApplication {

	return &PaymentApplication{
		userReadRepository:    userReadRepository,
		paymentReadRepository: paymentReadRepository,
		paymentService:        paymentService,
		stripePaymentService:  stripePaymentService,
		logger:                logger,
	}
}

func (p *PaymentApplication) CreatePayment(ctx context.Context, encrypt string) (*dto.PaymentResponse, *facade.Error) {
	if p.paymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("payment service not enabled"))
	}

	paymentEntity, prepayResponse, err := p.paymentService.CreatePayment(ctx, encrypt)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.PaymentResponse{
		OutTradeNo: paymentEntity.OutTradeNo,
		Channel:    string(paymentEntity.ChannelInfo.Channel),
		WeChatPayDetails: dto.WeChatPayResponse{
			AppID:     prepayResponse.Wechat.AppID,
			TimeStamp: prepayResponse.Wechat.TimeStamp,
			NonceStr:  prepayResponse.Wechat.NonceStr,
			Package:   prepayResponse.Wechat.Package,
			SignType:  prepayResponse.Wechat.SignType,
			PaySign:   prepayResponse.Wechat.PaySign,
		},
	}, nil
}

func (p *PaymentApplication) GetPaymentStatus(ctx context.Context, OutTradeNo string) (*dto.QueryPaymentStatusResponse, *facade.Error) {
	if p.paymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("payment service not enabled"))
	}

	paymentAggregate, err := p.paymentReadRepository.FindByOutTradeNo(ctx, OutTradeNo)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	if paymentAggregate.Payment.Status == enum.PaymentStatusNotPay {
		paymentAggregate, err = p.paymentService.GetPaymentStatus(ctx, paymentAggregate)
		if err != nil {
			return nil, facade.ErrServerInternal.Wrap(err)
		}
	}

	return &dto.QueryPaymentStatusResponse{
		TradeState:     string(paymentAggregate.Payment.Status),
		TradeStateDesc: paymentAggregate.Payment.Status.String(),
		SuccessTime:    paymentAggregate.Payment.PaidAt.String(),
		TransactionID:  paymentAggregate.Payment.ChannelInfo.TransactionID,
		Channel:        string(paymentAggregate.Payment.ChannelInfo.Channel),
	}, nil
}

func (p *PaymentApplication) HandleWechatPaymentCallback(ctx context.Context, req *http.Request, w http.ResponseWriter) (*dto.PaymentNotifyResponse, *facade.Error) {
	payment, err := p.paymentService.HandlePaymentNotify(ctx, req)

	if err != nil {
		p.logger.Errorf(ctx, "HandleWechatPaymentCallback failed: %w", err)

		w.WriteHeader(facade.ErrBadRequest.Code)
		return &dto.PaymentNotifyResponse{
			Code: "FAIL",
			Msg:  "Payment notification failed",
		}, facade.ErrBadRequest.Wrap(err)
	}
	p.logger.Infof(ctx, "Handle Wechat Payment Callback success")

	w.WriteHeader(http.StatusOK)
	successResponse := &dto.PaymentNotifyResponse{
		Code: "SUCCESS",
		Msg:  "",
	}

	newCtx := context.WithoutCancel(ctx)

	if err := retry.Do(
		func() error {
			return p.paymentService.SendNotification(payment)

		},
		retry.Delay(2*time.Second),
		retry.Attempts(10),
		retry.DelayType(retry.BackOffDelay),
		retry.OnRetry(func(n uint, err error) {
			p.logger.Errorf(newCtx, "SendNotification failed after all retries: %w", err)
		}),
	); err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return successResponse, nil
}

func (p *PaymentApplication) CreateStripeCheckoutSession(ctx context.Context, encrypt string) (*dto.StripeCheckoutResponse, *facade.Error) {
	if p.stripePaymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("stripe payment service not enabled"))
	}

	response, err := p.stripePaymentService.CreateCheckoutSession(ctx, encrypt)
	if err != nil {
		p.logger.Errorf(ctx, "CreateStripeCheckoutSession failed: %v", err)
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.StripeCheckoutResponse{
		CheckoutURL: response.CheckoutURL,
		SessionID:   response.SessionID,
		OutTradeNo:  response.OutTradeNo,
	}, nil
}

func (p *PaymentApplication) HandleStripeWebhook(ctx context.Context, req *http.Request, w http.ResponseWriter) (*dto.StripeWebhookResponse, *facade.Error) {
	if p.stripePaymentService == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return &dto.StripeWebhookResponse{
			Received: false,
			Error:    "stripe payment service not enabled",
		}, facade.ErrServerInternal.Wrap(xerror.New("stripe payment service not enabled"))
	}

	err := p.stripePaymentService.HandleWebhook(ctx, req)
	if err != nil {
		p.logger.Errorf(ctx, "HandleStripeWebhook failed: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return &dto.StripeWebhookResponse{
			Received: false,
			Error:    err.Error(),
		}, facade.ErrBadRequest.Wrap(err)
	}

	w.WriteHeader(http.StatusOK)
	return &dto.StripeWebhookResponse{
		Received: true,
	}, nil
}
