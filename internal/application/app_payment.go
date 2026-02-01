package application

import (
	"context"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/enum"
	"kiwi-user/internal/domain/service"
	"kiwi-user/internal/facade/dto"
	"kiwi-user/internal/infrastructure/payment/stripe"
	"net/http"
	"time"

	"github.com/avast/retry-go"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	libutils "github.com/Yet-Another-AI-Project/kiwi-lib/tools/utils"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
)

type PaymentApplication struct {
	userReadRepository    contract.IUserReadRepository
	paymentReadRepository contract.IPaymentReadRepository
	stripeClient          *stripe.StripeClient
	paymentService        *service.WechatPaymentService
	stripePaymentService  *service.StripePaymentService

	logger logger.ILogger
}

func NewPaymentApplication(
	userReadRepository contract.IUserReadRepository,
	paymentReadRepository contract.IPaymentReadRepository,
	paymentService *service.WechatPaymentService,
	stripePaymentService *service.StripePaymentService,
	stripeClient *stripe.StripeClient,
	logger logger.ILogger,
) *PaymentApplication {

	return &PaymentApplication{
		userReadRepository:    userReadRepository,
		paymentReadRepository: paymentReadRepository,
		paymentService:        paymentService,
		stripePaymentService:  stripePaymentService,
		stripeClient:          stripeClient,
		logger:                logger,
	}
}

func (p *PaymentApplication) CreateWechatPayment(ctx context.Context, encrypt string) (*dto.WechatPaymentResponse, *facade.Error) {
	if p.paymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("payment service not enabled"))
	}

	paymentEntity, prepayResponse, err := p.paymentService.CreatePayment(ctx, encrypt)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.WechatPaymentResponse{
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
		TransactionID:  paymentAggregate.Payment.ChannelInfo.WeChatTransactionID,
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

func (p *PaymentApplication) HandleStripeWebhook(ctx context.Context, body []byte, sig string) (*dto.StripeWebhookResponse, *facade.Error) {
	if p.stripePaymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("stripe payment service not enabled"))
	}

	event, err := p.stripeClient.VerifyWebhookSignature(body, sig)
	if err != nil {
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	paymentAggregate, err := p.stripePaymentService.HandleWebhook(ctx, event)
	if err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	notifyCtx := context.WithoutCancel(ctx)
	libutils.SafeGo(notifyCtx, p.logger, func() {
		// send notification to business service
		if err := retry.Do(
			func() error {

				if paymentAggregate == nil {
					return nil
				}

				return p.stripePaymentService.SendNotification(notifyCtx, paymentAggregate, event)

			},
			retry.Delay(2*time.Second),
			retry.Attempts(10),
			retry.DelayType(retry.BackOffDelay),
			retry.OnRetry(func(n uint, err error) {
				p.logger.Warnf(notifyCtx, "SendNotification failed and will retry, attempt: %d, error: %w", n, err)
			}),
		); err != nil {
			p.logger.Errorf(notifyCtx, "SendNotification failed: %w", err)
		}
	})

	return &dto.StripeWebhookResponse{
		Received: true,
	}, nil
}

func (p *PaymentApplication) CancelStripeSubscription(ctx context.Context, encrypt string) (*dto.StripeCancelSubscriptionResponse, *facade.Error) {
	if p.stripePaymentService == nil {
		return nil, facade.ErrServerInternal.Wrap(xerror.New("stripe payment service not enabled"))
	}

	paymentAggregate, err := p.stripePaymentService.CancelSubscription(ctx, encrypt)
	if err != nil {
		p.logger.Errorf(ctx, "CancelStripeSubscription failed: %v", err)
		return nil, facade.ErrServerInternal.Wrap(err)
	}

	return &dto.StripeCancelSubscriptionResponse{
		Success:        true,
		Message:        "",
		SubscriptionID: paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID,
	}, nil
}
