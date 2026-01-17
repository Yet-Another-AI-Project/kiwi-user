package wechatpay

import (
	"time"
)

type PaymentNotify struct {
	AppID          string      `json:"appid"`
	MchID          string      `json:"mchid"`
	OutTradeNo     string      `json:"out_trade_no"`
	TransactionID  string      `json:"transaction_id"`
	TradeType      string      `json:"trade_type"`
	TradeState     string      `json:"trade_state"`
	TradeStateDesc string      `json:"trade_state_desc"`
	BankType       string      `json:"bank_type"`
	Attach         string      `json:"attach,omitempty"`
	SuccessTime    time.Time   `json:"success_time"`
	Payer          PayerInfo   `json:"payer"`
	Amount         OrderAmount `json:"amount"`
}

type PayerInfo struct {
	OpenID string `json:"openid"`
}

type OrderAmount struct {
	Total         int    `json:"total"`
	PayerTotal    int    `json:"payer_total"`
	Currency      string `json:"currency"`
	PayerCurrency string `json:"payer_currency"`
}
