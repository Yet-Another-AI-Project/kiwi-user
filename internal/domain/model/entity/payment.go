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
	PaidAt      time.Time

	PaymentType        enum.PaymentType
	SubscriptionID     string
	SubscriptionStatus enum.SubscriptionStatus
	Interval           enum.SubscriptionInterval
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CustomerID         string
	CustomerEmail      string
	CheckoutSessionID  string
}

type PaymentChannelInfo struct {
	Channel       enum.PaymentChannel
	Platform      enum.WechatOpenIDPlatform
	TransactionID string
	OpenID        string
}
