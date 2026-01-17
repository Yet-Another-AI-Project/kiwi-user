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
}

type PaymentChannelInfo struct {
	Channel       enum.PaymentChannel
	Platform      enum.WechatOpenIDPlatform
	TransactionID string
	OpenID        string
}
