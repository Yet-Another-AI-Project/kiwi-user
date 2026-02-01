package stripe

import (
	"encoding/json"
	"time"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/stripe/stripe-go/v84"
)

type CheckoutSessionCompleted struct {
	SessionID      string
	InvoiceID      string
	CustomerID     string
	CustomerEmail  string
	SubscriptionID string
	PaymentStatus  string
	ExpiresAt      int64
	AmountTotal    int64
	Currency       string
	Metadata       map[string]string
}

type InvoicePaid struct {
	InvoiceID          string
	CustomerID         string
	SubscriptionID     string
	AmountPaid         int64
	Currency           string
	PaidAt             time.Time
	BillingReason      string
	CurrentPeriodStart int64
	CurrentPeriodEnd   int64
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
	// Use map to handle dynamic types (string vs object)
	var rawData map[string]interface{}
	if err := json.Unmarshal(event.RawData, &rawData); err != nil {
		return nil, err
	}

	result := &CheckoutSessionCompleted{}

	// Session ID
	if id, ok := rawData["id"].(string); ok {
		result.SessionID = id
	}

	// Payment Status
	if status, ok := rawData["payment_status"].(string); ok {
		result.PaymentStatus = status
	}

	// Amount Total
	if amount, ok := rawData["amount_total"].(float64); ok {
		result.AmountTotal = int64(amount)
	}

	// Currency
	if currency, ok := rawData["currency"].(string); ok {
		result.Currency = currency
	}

	// Invoice ID
	if invoiceID, ok := rawData["invoice"].(string); ok {
		result.InvoiceID = invoiceID
	}

	// Expires At
	if expiresAt, ok := rawData["expires_at"].(int64); ok {
		result.ExpiresAt = expiresAt
	}

	// Metadata
	if metadata, ok := rawData["metadata"].(map[string]interface{}); ok {
		result.Metadata = make(map[string]string)
		for k, v := range metadata {
			if strVal, ok := v.(string); ok {
				result.Metadata[k] = strVal
			}
		}
	}

	// Customer - can be string ID or object with ID field
	if customer := rawData["customer"]; customer != nil {
		switch c := customer.(type) {
		case string:
			result.CustomerID = c
		case map[string]interface{}:
			if id, ok := c["id"].(string); ok {
				result.CustomerID = id
			}
		}
	}

	// Subscription - can be string ID or object with ID field
	if subscription := rawData["subscription"]; subscription != nil {
		switch s := subscription.(type) {
		case string:
			result.SubscriptionID = s
		case map[string]interface{}:
			if id, ok := s["id"].(string); ok {
				result.SubscriptionID = id
			}
		}
	}

	// Customer Email from customer_details
	if customerDetails, ok := rawData["customer_details"].(map[string]interface{}); ok {
		if email, ok := customerDetails["email"].(string); ok {
			result.CustomerEmail = email
		}
	}

	return result, nil
}

func ParseInvoicePaid(event *WebhookEvent) (*InvoicePaid, error) {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.RawData, &invoice); err != nil {
		return nil, err
	}

	if len(invoice.Lines.Data) == 0 {
		return nil, xerror.New("no line items found in invoice")
	}

	result := &InvoicePaid{
		InvoiceID:          invoice.ID,
		AmountPaid:         invoice.AmountPaid,
		Currency:           string(invoice.Currency),
		BillingReason:      string(invoice.BillingReason),
		CurrentPeriodStart: invoice.Lines.Data[0].Period.Start,
		CurrentPeriodEnd:   invoice.Lines.Data[0].Period.End,
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
