package api

import (
	"kiwi-user/internal/facade/dto"

	"github.com/Yet-Another-AI-Project/kiwi-lib/server/facade"
	"github.com/gin-gonic/gin"
)

// CreatePayment godoc
// @Summary CreatePayment
// @Tags Payment
// @Description CreatePayment
// @Accept  json
// @Produce  json
// @Param  request body dto.PaymentRequest true "create payment request"
// @Success 200 {object}  facade.BaseResponse{data=dto.PaymentResponse}
//
// @Router /v1/payments [post]
func (c *Controller) CreatePayment(ctx *gin.Context) (*dto.PaymentResponse, *facade.Error) {
	var request dto.PaymentRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.paymentApplication.CreatePayment(ctx.Request.Context(), request.Encrypt)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// QueryPaymentStatus godoc
// @Summary QueryPaymentStatus
// @Tags Payment
// @Description QueryPaymentStatus
// @Accept  json
// @Produce  json
// @Param out-trade-no path string true "Out Trade No"
// @Success 200 {object}  facade.BaseResponse{data=dto.QueryPaymentStatusResponse}
//
// @Router /v1/payments/{out-trade-no}/status [get]
func (c *Controller) QueryPaymentStatus(ctx *gin.Context) (*dto.QueryPaymentStatusResponse, *facade.Error) {
	outTradeNo := ctx.Param("out-trade-no")

	response, err := c.paymentApplication.GetPaymentStatus(ctx.Request.Context(), outTradeNo)
	if err != nil {
		return nil, err
	}

	return response, nil

}

// WechatPaymentCallback godoc
// @Summary WechatPaymentCallback
// @Tags Payment
// @Description Handle Wechat payment notification callbacks
// @Accept  json
// @Produce  json
// @Success 200
// @Failure 400 {object} facade.BaseResponse{data=dto.PaymentNotifyResponse}
// @Router /v1/payments/wechat/notify [post]
func (c *Controller) WechatPaymentCallback(ctx *gin.Context) (*dto.PaymentNotifyResponse, *facade.Error) {
	response, _ := c.paymentApplication.HandleWechatPaymentCallback(ctx.Request.Context(), ctx.Request, ctx.Writer)

	return response, nil
}

// CreateStripeCheckoutSession godoc
// @Summary CreateStripeCheckoutSession
// @Tags Payment
// @Description Create a Stripe Checkout Session for subscription payment
// @Accept  json
// @Produce  json
// @Param  request body dto.StripeCheckoutRequest true "create stripe checkout session request"
// @Success 200 {object}  facade.BaseResponse{data=dto.StripeCheckoutResponse}
// @Router /v1/payments/stripe/checkout [post]
func (c *Controller) CreateStripeCheckoutSession(ctx *gin.Context) (*dto.StripeCheckoutResponse, *facade.Error) {
	var request dto.StripeCheckoutRequest
	if err := ctx.BindJSON(&request); err != nil {
		return nil, facade.ErrBadRequest.Wrap(err)
	}

	response, err := c.paymentApplication.CreateStripeCheckoutSession(ctx.Request.Context(), request.Encrypt)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StripeWebhook godoc
// @Summary StripeWebhook
// @Tags Payment
// @Description Handle Stripe webhook events
// @Accept  json
// @Produce  json
// @Success 200 {object}  facade.BaseResponse{data=dto.StripeWebhookResponse}
// @Failure 400 {object} facade.BaseResponse{data=dto.StripeWebhookResponse}
// @Router /v1/payments/stripe/webhook [post]
func (c *Controller) StripeWebhook(ctx *gin.Context) (*dto.StripeWebhookResponse, *facade.Error) {
	response, _ := c.paymentApplication.HandleStripeWebhook(ctx.Request.Context(), ctx.Request, ctx.Writer)

	return response, nil
}
