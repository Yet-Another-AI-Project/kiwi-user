package stripe

import (
	"encoding/json"
	"time"

	"github.com/stripe/stripe-go/v82"
)

type CheckoutSessionCompleted struct {
	SessionID      string
	CustomerID     string
	CustomerEmail  string
	SubscriptionID string
	PaymentStatus  string
	AmountTotal    int64
	Currency       string
	Metadata       map[string]string
}

type InvoicePaid struct {
	InvoiceID      string
	CustomerID     string
	SubscriptionID string
	AmountPaid     int64
	Currency       string
	PaidAt         time.Time
	BillingReason  string
}

type InvoicePaymentFailed struct {
	InvoiceID          string
	CustomerID         string
	SubscriptionID     string
	AmountDue          int64
	Currency           string
	AttemptCount       int64
	NextPaymentAttempt time.Time
}

type SubscriptionUpdated struct {
	SubscriptionID     string
	CustomerID         string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         time.Time
	PriceID            string
}

type SubscriptionDeleted struct {
	SubscriptionID string
	CustomerID     string
	CanceledAt     time.Time
}

func ParseCheckoutSessionCompleted(event *WebhookEvent) (*CheckoutSessionCompleted, error) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.RawData, &session); err != nil {
		return nil, err
	}

	result := &CheckoutSessionCompleted{
		SessionID:     session.ID,
		PaymentStatus: string(session.PaymentStatus),
		AmountTotal:   session.AmountTotal,
		Currency:      string(session.Currency),
		Metadata:      session.Metadata,
	}

	if session.Customer != nil {
		result.CustomerID = session.Customer.ID
	}

	if session.Subscription != nil {
		result.SubscriptionID = session.Subscription.ID
	}

	if session.CustomerDetails != nil {
		result.CustomerEmail = session.CustomerDetails.Email
	}

	return result, nil
}

func ParseInvoicePaid(event *WebhookEvent) (*InvoicePaid, error) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.RawData, &invoice); err != nil {
		return nil, err
	}

	result := &InvoicePaid{
		InvoiceID:     invoice.ID,
		AmountPaid:    invoice.AmountPaid,
		Currency:      string(invoice.Currency),
		BillingReason: string(invoice.BillingReason),
	}

	if invoice.Customer != nil {
		result.CustomerID = invoice.Customer.ID
	}

	if invoice.Parent != nil && invoice.Parent.SubscriptionDetails != nil && invoice.Parent.SubscriptionDetails.Subscription != nil {
		result.SubscriptionID = invoice.Parent.SubscriptionDetails.Subscription.ID
	}

	if invoice.StatusTransitions != nil && invoice.StatusTransitions.PaidAt > 0 {
		result.PaidAt = time.Unix(invoice.StatusTransitions.PaidAt, 0)
	}

	return result, nil
}

func ParseInvoicePaymentFailed(event *WebhookEvent) (*InvoicePaymentFailed, error) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.RawData, &invoice); err != nil {
		return nil, err
	}

	result := &InvoicePaymentFailed{
		InvoiceID:    invoice.ID,
		AmountDue:    invoice.AmountDue,
		Currency:     string(invoice.Currency),
		AttemptCount: invoice.AttemptCount,
	}

	if invoice.Customer != nil {
		result.CustomerID = invoice.Customer.ID
	}

	// stripe-go/v82: invoice does not expose Subscription directly; it is available under Parent.SubscriptionDetails.Subscription
	if invoice.Parent != nil && invoice.Parent.SubscriptionDetails != nil && invoice.Parent.SubscriptionDetails.Subscription != nil {
		result.SubscriptionID = invoice.Parent.SubscriptionDetails.Subscription.ID
	}

	if invoice.NextPaymentAttempt > 0 {
		result.NextPaymentAttempt = time.Unix(invoice.NextPaymentAttempt, 0)
	}

	return result, nil
}

func ParseSubscriptionUpdated(event *WebhookEvent) (*SubscriptionUpdated, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.RawData, &sub); err != nil {
		return nil, err
	}

	result := &SubscriptionUpdated{
		SubscriptionID:    sub.ID,
		Status:            string(sub.Status),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
	}

	// stripe-go/v82: current period start/end live on subscription items.
	if sub.Items != nil && len(sub.Items.Data) > 0 {
		result.CurrentPeriodStart = time.Unix(sub.Items.Data[0].CurrentPeriodStart, 0)
		result.CurrentPeriodEnd = time.Unix(sub.Items.Data[0].CurrentPeriodEnd, 0)
	}

	if sub.Customer != nil {
		result.CustomerID = sub.Customer.ID
	}

	if sub.CanceledAt > 0 {
		result.CanceledAt = time.Unix(sub.CanceledAt, 0)
	}

	if len(sub.Items.Data) > 0 && sub.Items.Data[0].Price != nil {
		result.PriceID = sub.Items.Data[0].Price.ID
	}

	return result, nil
}

func ParseSubscriptionDeleted(event *WebhookEvent) (*SubscriptionDeleted, error) {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.RawData, &sub); err != nil {
		return nil, err
	}

	result := &SubscriptionDeleted{
		SubscriptionID: sub.ID,
	}

	if sub.Customer != nil {
		result.CustomerID = sub.Customer.ID
	}

	if sub.CanceledAt > 0 {
		result.CanceledAt = time.Unix(sub.CanceledAt, 0)
	}

	return result, nil
}
