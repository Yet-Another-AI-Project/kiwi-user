package entity

import (
	"kiwi-user/internal/domain/model/enum"
	"time"
)

type PaymentEntity struct {
	OutTradeNo  string
	UserID      string
	ChannelInfo PaymentChannelInfo
	Service     string
	Amount      int
	Currency    string
	Description string
	Status      enum.PaymentStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PaidAt      time.Time
	PaymentType enum.PaymentType
}

type PaymentChannelInfo struct {
	Channel enum.PaymentChannel
	// wechat
	WechatPlatform      enum.WechatOpenIDPlatform
	WeChatTransactionID string
	WeChatOpenID        string

	// stripe
	StripeSubscriptionID     string
	StripeSubscriptionStatus enum.SubscriptionStatus
	StripeInterval           enum.SubscriptionInterval
	StripeCurrentPeriodStart time.Time
	StripeCurrentPeriodEnd   time.Time
	StripeCustomerID         string
	StripeCustomerEmail      string
	StripeCheckoutSessionID  string
	StripeInvoiceID          string
}
