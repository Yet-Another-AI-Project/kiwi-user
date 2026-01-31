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
		OutTradeNo:        outTradeNo,
		UserID:            request.UserID,
		ChannelInfo:       entity.PaymentChannelInfo{Channel: enum.PaymentChannelStripe},
		Service:           request.Service,
		Amount:            int(session.AmountTotal),
		Currency:          string(session.Currency),
		Description:       request.Description,
		Status:            enum.PaymentStatusNotPay,
		CreatedAt:         time.Now(),
		PaymentType:       enum.PaymentTypeSubscription,
		Interval:          interval,
		CustomerEmail:     request.CustomerEmail,
		CheckoutSessionID: session.ID,
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

func (s *StripePaymentService) HandleWebhook(ctx context.Context, req *http.Request) error {
	if s.stripeClient == nil {
		return xerror.New("stripe client not initialized")
	}

	event, err := s.stripeClient.VerifyWebhookSignature(req)
	if err != nil {
		s.logger.Errorf(ctx, "Stripe webhook signature verification failed: %v", err)
		return xerror.Wrap(err)
	}

	exists, err := s.stripeEventRepository.ExistsByEventID(ctx, event.ID)
	if err != nil {
		return xerror.Wrap(err)
	}
	if exists {
		s.logger.Infof(ctx, "Stripe event %s already processed, skipping", event.ID)
		return nil
	}

	stripeEvent := &entity.StripeEventEntity{
		EventID:   event.ID,
		EventType: event.Type,
		Processed: false,
		CreatedAt: time.Now(),
	}
	if _, err := s.stripeEventRepository.Create(ctx, stripeEvent); err != nil {
		return xerror.Wrap(err)
	}

	var handleErr error
	switch event.Type {
	case "checkout.session.completed":
		handleErr = s.handleCheckoutSessionCompleted(ctx, event)
	case "invoice.paid":
		handleErr = s.handleInvoicePaid(ctx, event)
	case "invoice.payment_failed":
		handleErr = s.handleInvoicePaymentFailed(ctx, event)
	case "customer.subscription.updated":
		handleErr = s.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		handleErr = s.handleSubscriptionDeleted(ctx, event)
	default:
		s.logger.Infof(ctx, "Unhandled Stripe event type: %s", event.Type)
	}

	if handleErr != nil {
		s.logger.Errorf(ctx, "Error handling Stripe event %s: %v", event.Type, handleErr)
		return handleErr
	}

	if err := s.stripeEventRepository.MarkProcessed(ctx, event.ID); err != nil {
		s.logger.Errorf(ctx, "Failed to mark event as processed: %v", err)
	}

	return nil
}

func (s *StripePaymentService) handleCheckoutSessionCompleted(ctx context.Context, event *stripeclient.WebhookEvent) error {
	sessionData, err := stripeclient.ParseCheckoutSessionCompleted(event)
	if err != nil {
		return xerror.Wrap(err)
	}

	outTradeNo := sessionData.Metadata["out_trade_no"]
	if outTradeNo == "" {
		s.logger.Errorf(ctx, "Missing out_trade_no in checkout session metadata")
		return xerror.New("missing out_trade_no in metadata")
	}

	paymentAggregate, err := s.paymentRepository.FindByOutTradeNo(ctx, outTradeNo)
	if err != nil {
		return xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		return xerror.New("payment not found for out_trade_no: " + outTradeNo)
	}

	paymentAggregate.Payment.Status = enum.PaymentStatusSuccess
	paymentAggregate.Payment.SubscriptionID = sessionData.SubscriptionID
	paymentAggregate.Payment.CustomerID = sessionData.CustomerID
	paymentAggregate.Payment.CustomerEmail = sessionData.CustomerEmail
	paymentAggregate.Payment.SubscriptionStatus = enum.SubscriptionStatusActive
	paymentAggregate.Payment.PaidAt = time.Now()

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return xerror.Wrap(err)
	}

	return s.sendNotification(ctx, paymentAggregate, "checkout.session.completed")
}

func (s *StripePaymentService) handleInvoicePaid(ctx context.Context, event *stripeclient.WebhookEvent) error {
	invoiceData, err := stripeclient.ParseInvoicePaid(event)
	if err != nil {
		return xerror.Wrap(err)
	}

	if invoiceData.BillingReason == "subscription_create" {
		s.logger.Infof(ctx, "Skipping invoice.paid for subscription_create, handled by checkout.session.completed")
		return nil
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, invoiceData.SubscriptionID)
	if err != nil {
		return xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", invoiceData.SubscriptionID)
		return nil
	}

	paymentAggregate.Payment.SubscriptionStatus = enum.SubscriptionStatusActive
	paymentAggregate.Payment.PaidAt = invoiceData.PaidAt

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return xerror.Wrap(err)
	}

	return s.sendNotification(ctx, paymentAggregate, "invoice.paid")
}

func (s *StripePaymentService) handleInvoicePaymentFailed(ctx context.Context, event *stripeclient.WebhookEvent) error {
	invoiceData, err := stripeclient.ParseInvoicePaymentFailed(event)
	if err != nil {
		return xerror.Wrap(err)
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, invoiceData.SubscriptionID)
	if err != nil {
		return xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", invoiceData.SubscriptionID)
		return nil
	}

	paymentAggregate.Payment.SubscriptionStatus = enum.SubscriptionStatusPastDue

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return xerror.Wrap(err)
	}

	return s.sendNotification(ctx, paymentAggregate, "invoice.payment_failed")
}

func (s *StripePaymentService) handleSubscriptionUpdated(ctx context.Context, event *stripeclient.WebhookEvent) error {
	subData, err := stripeclient.ParseSubscriptionUpdated(event)
	if err != nil {
		return xerror.Wrap(err)
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, subData.SubscriptionID)
	if err != nil {
		return xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", subData.SubscriptionID)
		return nil
	}

	paymentAggregate.Payment.SubscriptionStatus = enum.ParseSubscriptionStatus(subData.Status)
	paymentAggregate.Payment.CurrentPeriodStart = subData.CurrentPeriodStart
	paymentAggregate.Payment.CurrentPeriodEnd = subData.CurrentPeriodEnd

	if subData.Status == "canceled" {
		paymentAggregate.Payment.SubscriptionStatus = enum.SubscriptionStatusCanceled
	}

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return xerror.Wrap(err)
	}

	return s.sendNotification(ctx, paymentAggregate, "customer.subscription.updated")
}

func (s *StripePaymentService) handleSubscriptionDeleted(ctx context.Context, event *stripeclient.WebhookEvent) error {
	subData, err := stripeclient.ParseSubscriptionDeleted(event)
	if err != nil {
		return xerror.Wrap(err)
	}

	paymentAggregate, err := s.paymentRepository.FindBySubscriptionID(ctx, subData.SubscriptionID)
	if err != nil {
		return xerror.Wrap(err)
	}
	if paymentAggregate == nil {
		s.logger.Warnf(ctx, "No payment found for subscription %s", subData.SubscriptionID)
		return nil
	}

	paymentAggregate.Payment.SubscriptionStatus = enum.SubscriptionStatusCanceled
	paymentAggregate.Payment.Status = enum.PaymentStatusClosed

	if _, err := s.paymentRepository.Update(ctx, paymentAggregate); err != nil {
		return xerror.Wrap(err)
	}

	return s.sendNotification(ctx, paymentAggregate, "customer.subscription.deleted")
}

func (s *StripePaymentService) sendNotification(ctx context.Context, paymentAggregate *aggregate.PaymentAggregate, eventType string) error {
	var notifyURL string
	if urlValue, ok := s.ServiceNotifyURL[paymentAggregate.Payment.Service]; ok {
		notifyURL = urlValue.(string)
	} else {
		s.logger.Warnf(ctx, "No notify URL configured for service: %s", paymentAggregate.Payment.Service)
		return nil
	}

	payloadData := map[string]interface{}{
		"order_no":             paymentAggregate.Payment.OutTradeNo,
		"pay_order_no":         paymentAggregate.Payment.ChannelInfo.TransactionID,
		"subscription_id":      paymentAggregate.Payment.SubscriptionID,
		"customer_id":          paymentAggregate.Payment.CustomerID,
		"amount":               int64(paymentAggregate.Payment.Amount),
		"currency":             paymentAggregate.Payment.Currency,
		"pay_time":             paymentAggregate.Payment.PaidAt.Format(time.RFC3339),
		"pay_state":            string(paymentAggregate.Payment.Status),
		"subscription_status":  string(paymentAggregate.Payment.SubscriptionStatus),
		"interval":             string(paymentAggregate.Payment.Interval),
		"event_type":           eventType,
		"channel":              "stripe",
		"payment_type":         string(paymentAggregate.Payment.PaymentType),
		"current_period_start": paymentAggregate.Payment.CurrentPeriodStart.Format(time.RFC3339),
		"current_period_end":   paymentAggregate.Payment.CurrentPeriodEnd.Format(time.RFC3339),
		"user_id":              paymentAggregate.Payment.UserID,
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
		"encrypt": encryptedData,
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

	s.logger.Infof(ctx, "Successfully sent notification for event %s to %s", eventType, notifyURL)
	return nil
}
