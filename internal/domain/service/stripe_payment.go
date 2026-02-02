package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"kiwi-user/config"
	"kiwi-user/internal/domain/contract"
	"kiwi-user/internal/domain/model/aggregate"
	"kiwi-user/internal/domain/model/entity"
	"kiwi-user/internal/domain/model/enum"
	stripeclient "kiwi-user/internal/infrastructure/payment/stripe"
	"kiwi-user/internal/infrastructure/utils"
	"kiwi-user/internal/infrastructure/utils/aes"

	"github.com/Yet-Another-AI-Project/kiwi-lib/tools/xhttp"
	"github.com/futurxlab/golanggraph/logger"
	"github.com/futurxlab/golanggraph/xerror"
	"github.com/stripe/stripe-go/v84"
)

type StripePaymentService struct {
	httpClient            *xhttp.Client
	paymentRepository     contract.IPaymentRepository
	stripeEventRepository contract.IStripeEventRepository
	userRepository        contract.IUserRepository
	stripeClient          *stripeclient.StripeClient
	AESKey                string
	ServiceNotifyURL      map[string]interface{}
	logger                logger.ILogger
}

func NewStripePaymentService(
	config *config.Config,
	paymentRepository contract.IPaymentRepository,
	stripeEventRepository contract.IStripeEventRepository,
	userRepository contract.IUserRepository,
	httpClient *xhttp.Client,
	logger logger.ILogger,
) (*StripePaymentService, error) {
	service := &StripePaymentService{
		paymentRepository:     paymentRepository,
		stripeEventRepository: stripeEventRepository,
		userRepository:        userRepository,
		httpClient:            httpClient,
		logger:                logger,
	}

	if config.Payment == nil || config.Payment.StripeAPIKey == "" {
		return nil, nil
	}

	service.stripeClient = stripeclient.NewStripeClient(
		config.Payment.StripeAPIKey,
		config.Payment.StripeWebhookSecret,
		config.Payment.StripeSuccessURL,
		config.Payment.StripeCancelURL,
		config.Payment.StripeMonthlyPriceID,
		config.Payment.StripeYearlyPriceID,
	)

	service.AESKey = config.Payment.AESEncryptKey
	service.ServiceNotifyURL = config.Payment.ServiceNotifyURL

	return service, nil
}

type StripeCheckoutRequest struct {
	UserID        string `json:"user_id"`
	Service       string `json:"service"`
	Interval      string `json:"interval"`
	CustomerEmail string `json:"customer_email"`
	Description   string `json:"description"`
}

type StripeCheckoutResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	OutTradeNo  string `json:"out_trade_no"`
}

type StripeCancelSubscriptionRequest struct {
	UserID  string `json:"user_id"`
	Service string `json:"service"`
}

type StripeCancelSubscriptionResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message,omitempty"`
	SubscriptionID string `json:"subscription_id,omitempty"`
}

type SubscriptionNotifyPayload struct {
	Event   string                   `json:"event"`
	Service string                   `json:"service"`
	UserID  string                   `json:"user_id"`
	OrderNo string                   `json:"order_no"`
	Plan    string                   `json:"plan"`
	Stripe  SubscriptionNotifyStripe `json:"stripe"`
	Period  SubscriptionNotifyPeriod `json:"period"`
	Status  string                   `json:"status"`
	EventID string                   `json:"event_id"`
}

type SubscriptionNotifyStripe struct {
	SubscriptionID     string `json:"subscription_id"`
	CustomerID         string `json:"customer_id"`
	CustomerEmail      string `json:"customer_email"`
	CheckoutSessionID  string `json:"checkout_session_id"`
	InvoiceID          string `json:"invoice_id"`
	CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
	SubscriptionStatus string `json:"subscription_status"`
}

type SubscriptionNotifyPeriod struct {
	CurrentPeriodStart int64 `json:"current_period_start"`
	CurrentPeriodEnd   int64 `json:"current_period_end"`
}

func (s *StripePaymentService) CreateCheckoutSession(ctx context.Context, encrypt string) (*StripeCheckoutResponse, error) {
	if s.stripeClient == nil {
		return nil, xerror.New("stripe client not initialized")
	}

	decryptData, err := aes.AESDecrypt(encrypt, []byte(s.AESKey))
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var request StripeCheckoutRequest
	if err := json.Unmarshal([]byte(decryptData), &request); err != nil {
		return nil, xerror.Wrap(err)
	}

	// check if user has active subscription
	paymentAggregate, err := s.paymentRepository.FindActiveSubscriptionByUserIDAndService(ctx, request.UserID, request.Service)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if paymentAggregate != nil {
		return nil, xerror.New("user already has an active subscription")
	}

	userAggregate, err := s.userRepository.Find(ctx, request.UserID)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if userAggregate == nil {
		return nil, xerror.New("user not found")
	}

	outTradeNo := utils.GnerateOutTradeNo("STRIPE")

	checkoutParams := &stripeclient.CreateCheckoutSessionParams{
		CustomerEmail: request.CustomerEmail,
		UserID:        request.UserID,
		Service:       request.Service,
		Interval:      request.Interval,
		OutTradeNo:    outTradeNo,
	}

	session, err := s.stripeClient.CreateCheckoutSession(ctx, checkoutParams)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	interval := enum.SubscriptionIntervalMonthly
	if request.Interval == "yearly" {
		interval = enum.SubscriptionIntervalYearly
	}

	payment := &entity.PaymentEntity{
		OutTradeNo: outTradeNo,
		UserID:     request.UserID,
		ChannelInfo: entity.PaymentChannelInfo{
			Channel:                 enum.PaymentChannelStripe,
			StripeInterval:          interval,
			StripeCustomerEmail:     request.CustomerEmail,
			StripeCheckoutSessionID: session.ID,
		},
		Service:     request.Service,
		Amount:      int(session.AmountTotal),
		Currency:    string(session.Currency),
		Description: request.Description,
		Status:      enum.PaymentStatusNotPay,
		PaymentType: enum.PaymentTypeSubscription,
	}

	_, err = s.paymentRepository.Create(ctx, &aggregate.PaymentAggregate{Payment: payment})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &StripeCheckoutResponse{
		CheckoutURL: session.URL,
		SessionID:   session.ID,
		OutTradeNo:  outTradeNo,
	}, nil
}

func (s *StripePaymentService) HandleWebhook(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	if s.stripeClient == nil {
		return nil, xerror.New("stripe client not initialized")
	}

	exists, err := s.stripeEventRepository.ExistsByEventID(ctx, event.ID)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if exists {
		s.logger.Infof(ctx, "Stripe event %s already processed, skipping", event.ID)
		return nil, nil
	}

	var paymentAggregate *aggregate.PaymentAggregate
	var handleErr error
	switch event.Type {
	case string(stripe.EventTypeCheckoutSessionCompleted):
		paymentAggregate, handleErr = s.handleCheckoutSessionCompleted(ctx, event)
	case string(stripe.EventTypeInvoicePaid):
		paymentAggregate, handleErr = s.handleInvoicePaid(ctx, event)
	case string(stripe.EventTypeInvoicePaymentFailed):
		paymentAggregate, handleErr = s.handleInvoicePaymentFailed(ctx, event)
	case string(stripe.EventTypeCustomerSubscriptionUpdated):
		paymentAggregate, handleErr = s.handleSubscriptionUpdated(ctx, event)
	case string(stripe.EventTypeCustomerSubscriptionDeleted):
		paymentAggregate, handleErr = s.handleSubscriptionDeleted(ctx, event)
	default:
		s.logger.Infof(ctx, "Unhandled Stripe event type: %s", event.Type)
		return nil, nil
	}

	if handleErr != nil {
		s.logger.Errorf(ctx, "Error handling Stripe event %s: %v", event.Type, handleErr)
		return nil, handleErr
	}

	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for event %s", event.ID)
		return nil, nil
	}

	stripeEvent := &entity.StripeEventEntity{
		EventID:        event.ID,
		EventType:      event.Type,
		SubscriptionID: paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID,
		UserID:         paymentAggregate.Payment.UserID,
		Processed:      true,
		ProcessedAt:    time.Now(),
	}
	if _, err := s.stripeEventRepository.Create(ctx, stripeEvent); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) handleCheckoutSessionCompleted(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	sessionData, err := stripeclient.ParseCheckoutSessionCompleted(event)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	outTradeNo := sessionData.Metadata["out_trade_no"]
	if outTradeNo == "" {
		s.logger.Errorf(ctx, "Missing out_trade_no in checkout session metadata")
		return nil, xerror.New("missing out_trade_no in metadata")
	}

	paymentAggregate, err := s.paymentRepository.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		return nil, xerror.New("payment not found for out_trade_no: " + outTradeNo)
	}

	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID = sessionData.SubscriptionID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerID = sessionData.CustomerID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail = sessionData.CustomerEmail
	paymentAggregate.Payment.ChannelInfo.StripeInvoiceID = sessionData.InvoiceID

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) handleInvoicePaid(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	invoiceData, err := stripeclient.ParseInvoicePaid(event)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	outTradeNo := invoiceData.Metadata["out_trade_no"]
	if outTradeNo == "" {
		s.logger.Errorf(ctx, "Missing out_trade_no in checkout session metadata")
		return nil, xerror.New("missing out_trade_no in metadata")
	}

	// find by out trade no
	paymentAggregate, err := s.paymentRepository.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", invoiceData.SubscriptionID)
		return nil, nil
	}

	paymentAggregate.Payment.PaidAt = invoiceData.PaidAt
	paymentAggregate.Payment.Status = enum.PaymentStatusSuccess
	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID = invoiceData.SubscriptionID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerID = invoiceData.CustomerID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail = invoiceData.CustomerEmail
	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus = enum.SubscriptionStatusActive
	paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodStart = time.Unix(invoiceData.CurrentPeriodStart, 0)
	paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodEnd = time.Unix(invoiceData.CurrentPeriodEnd, 0)
	paymentAggregate.Payment.ChannelInfo.StripeInvoiceID = invoiceData.InvoiceID

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) handleInvoicePaymentFailed(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	invoiceData, err := stripeclient.ParseInvoicePaymentFailed(event)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	outTradeNo := invoiceData.Metadata["out_trade_no"]
	if outTradeNo == "" {
		s.logger.Errorf(ctx, "Missing out_trade_no in checkout session metadata")
		return nil, xerror.New("missing out_trade_no in metadata")
	}

	paymentAggregate, err := s.paymentRepository.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", invoiceData.SubscriptionID)
		return nil, nil
	}

	paymentAggregate.Payment.Status = enum.PaymentStatusFailed
	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID = invoiceData.SubscriptionID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerID = invoiceData.CustomerID
	paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail = invoiceData.CustomerEmail
	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus = enum.SubscriptionStatusPastDue
	paymentAggregate.Payment.ChannelInfo.StripeInvoiceID = invoiceData.InvoiceID

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) handleSubscriptionUpdated(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	subData, err := stripeclient.ParseSubscriptionUpdated(event)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, subData.SubscriptionID)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", subData.SubscriptionID)
		return nil, nil
	}

	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus = enum.ParseSubscriptionStatus(subData.Status)
	paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodStart = subData.CurrentPeriodStart
	paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodEnd = subData.CurrentPeriodEnd
	paymentAggregate.Payment.ChannelInfo.StripeCancelAtPeriodEnd = subData.CancelAtPeriodEnd

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) handleSubscriptionDeleted(ctx context.Context, event *stripeclient.WebhookEvent) (*aggregate.PaymentAggregate, error) {
	subData, err := stripeclient.ParseSubscriptionDeleted(event)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, subData.SubscriptionID)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", subData.SubscriptionID)
		return nil, nil
	}

	paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus = enum.SubscriptionStatusCanceled
	paymentAggregate.Payment.Status = enum.PaymentStatusClosed

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) CancelSubscription(ctx context.Context, encrypt string) (*aggregate.PaymentAggregate, error) {
	if s.stripeClient == nil {
		return nil, xerror.New("stripe client not initialized")
	}

	decryptData, err := aes.AESDecrypt(encrypt, []byte(s.AESKey))
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var request StripeCancelSubscriptionRequest
	if err := json.Unmarshal([]byte(decryptData), &request); err != nil {
		return nil, xerror.Wrap(err)
	}

	if request.UserID == "" {
		return nil, xerror.New("user_id is required")
	}

	if request.Service == "" {
		return nil, xerror.New("service is required")
	}

	paymentAggregate, err := s.paymentRepository.FindActiveSubscriptionByUserIDAndService(ctx, request.UserID, request.Service)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		return nil, xerror.New("no active subscription found for user")
	}

	if paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID == "" {
		return nil, xerror.New("subscription ID not found in payment record")
	}

	// cancel at period end
	_, err = s.stripeClient.UpdateSubscription(
		ctx,
		paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID,
		&stripe.SubscriptionParams{
			CancelAtPeriodEnd: stripe.Bool(true),
		},
	)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	paymentAggregate.Payment.ChannelInfo.StripeCancelAtPeriodEnd = true

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return nil, xerror.Wrap(err)
	}

	return paymentAggregate, nil
}

func (s *StripePaymentService) SendNotification(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate, event *stripeclient.WebhookEvent) error {
	var notifyURL string
	if urlValue, ok := s.ServiceNotifyURL[paymentAggregate.Payment.Service]; ok {
		notifyURL = urlValue.(string)
	} else {
		s.logger.Warnf(ctx, "No notify URL configured for service: %s", paymentAggregate.Payment.Service)
		return nil
	}

	payloadData := SubscriptionNotifyPayload{
		Event:   event.Type,
		EventID: event.ID,
		Service: paymentAggregate.Payment.Service,
		UserID:  paymentAggregate.Payment.UserID,
		OrderNo: paymentAggregate.Payment.OutTradeNo,
		Plan:    string(paymentAggregate.Payment.ChannelInfo.StripeInterval),
		Stripe: SubscriptionNotifyStripe{
			SubscriptionID:     paymentAggregate.Payment.ChannelInfo.StripeSubscriptionID,
			SubscriptionStatus: string(paymentAggregate.Payment.ChannelInfo.StripeSubscriptionStatus),
			CustomerID:         paymentAggregate.Payment.ChannelInfo.StripeCustomerID,
			CustomerEmail:      paymentAggregate.Payment.ChannelInfo.StripeCustomerEmail,
			CheckoutSessionID:  paymentAggregate.Payment.ChannelInfo.StripeCheckoutSessionID,
			InvoiceID:          paymentAggregate.Payment.ChannelInfo.StripeInvoiceID,
			CancelAtPeriodEnd:  paymentAggregate.Payment.ChannelInfo.StripeCancelAtPeriodEnd,
		},
		Period: SubscriptionNotifyPeriod{
			CurrentPeriodStart: paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodStart.Unix(),
			CurrentPeriodEnd:   paymentAggregate.Payment.ChannelInfo.StripeCurrentPeriodEnd.Unix(),
		},
		Status: string(paymentAggregate.Payment.Status),
	}

	jsonData, err := json.Marshal(payloadData)
	if err != nil {
		return xerror.Wrap(err)
	}

	encryptedData, err := aes.AESEncrypt(string(jsonData), []byte(s.AESKey))
	if err != nil {
		return xerror.Wrap(err)
	}

	requestBody := map[string]interface{}{
		"data": encryptedData,
	}

	jsonReqBody, err := json.Marshal(requestBody)
	if err != nil {
		return xerror.Wrap(err)
	}

	resp, err := s.httpClient.Post(notifyURL, "application/json", bytes.NewReader(jsonReqBody))
	if err != nil {
		return xerror.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Errorf(ctx, "Notification to business service failed with status: %d", resp.StatusCode)
		return xerror.New("notification failed")
	}

	s.logger.Infof(ctx, "Successfully sent notification for event %s %s to %s", event.Type, event.ID, notifyURL)
	return nil
}
