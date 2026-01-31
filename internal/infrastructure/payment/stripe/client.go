package stripe

import (
	"context"
	"io"
	"net/http"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/webhook"
)

type StripeClient struct {
	APIKey         string
	WebhookSecret  string
	SuccessURL     string
	CancelURL      string
	MonthlyPriceID string
	YearlyPriceID  string
}

type WebhookEvent struct {
	ID      string
	Type    string
	RawData []byte
}

func NewStripeClient(
	apiKey string,
	webhookSecret string,
	successURL string,
	cancelURL string,
	monthlyPriceID string,
	yearlyPriceID string,
) *StripeClient {
	stripe.Key = apiKey
	return &StripeClient{
		APIKey:         apiKey,
		WebhookSecret:  webhookSecret,
		SuccessURL:     successURL,
		CancelURL:      cancelURL,
		MonthlyPriceID: monthlyPriceID,
		YearlyPriceID:  yearlyPriceID,
	}
}

type CreateCheckoutSessionParams struct {
	CustomerEmail string
	UserID        string
	Service       string
	Interval      string
	OutTradeNo    string
}

func (c *StripeClient) CreateCheckoutSession(ctx context.Context, params *CreateCheckoutSessionParams) (*stripe.CheckoutSession, error) {
	priceID := c.MonthlyPriceID
	if params.Interval == "yearly" {
		priceID = c.YearlyPriceID
	}

	checkoutParams := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(c.SuccessURL + "?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(c.CancelURL),
	}

	if params.CustomerEmail != "" {
		checkoutParams.CustomerEmail = stripe.String(params.CustomerEmail)
	}

	checkoutParams.Metadata = map[string]string{
		"user_id":      params.UserID,
		"service":      params.Service,
		"out_trade_no": params.OutTradeNo,
		"interval":     params.Interval,
	}

	checkoutParams.SubscriptionData = &stripe.CheckoutSessionSubscriptionDataParams{
		Metadata: map[string]string{
			"user_id":      params.UserID,
			"service":      params.Service,
			"out_trade_no": params.OutTradeNo,
		},
	}

	sess, err := session.New(checkoutParams)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return sess, nil
}

func (c *StripeClient) VerifyWebhookSignature(req *http.Request) (*WebhookEvent, error) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	sig := req.Header.Get("Stripe-Signature")
	event, err := webhook.ConstructEvent(body, sig, c.WebhookSecret)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &WebhookEvent{
		ID:      event.ID,
		Type:    string(event.Type),
		RawData: event.Data.Raw,
	}, nil
}

func (c *StripeClient) GetPriceID(interval string) string {
	if interval == "yearly" {
		return c.YearlyPriceID
	}
	return c.MonthlyPriceID
}
