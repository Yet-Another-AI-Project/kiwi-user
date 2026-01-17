package dto

type PaymentRequestContent struct {
	Description string `json:"description" binding:"required"`
	Amount      Amount `json:"amount" binding:"required"`
	UserID      string `json:"user_id" binding:"required"`
	Channel     string `json:"channel" binding:"required"`  // e.g., "wechat", "alipay"
	Platform    string `json:"platform" binding:"required"` // e.g., "miniprogram"
	Service     string `json:"service" binding:"required"`
}

type PaymentRequest struct {
	Encrypt string `json:"encrypt" binding:"required"`
}

type Amount struct {
	Total    int    `json:"total" binding:"required"`
	Currency string `json:"currency,omitempty"` // Optional, e.g., "USD", "CNY", default is "CNY"
}

type PaymentResponse struct {
	OutTradeNo       string            `json:"out_trade_no"`
	Channel          string            `json:"channel"`
	WeChatPayDetails WeChatPayResponse `json:"wechat_pay_details,omitempty"`
}

type WeChatPayResponse struct {
	AppID     string `json:"appid"`
	TimeStamp string `json:"timestamp"`
	NonceStr  string `json:"nonce_str"`
	Package   string `json:"package"`
	SignType  string `json:"sign_type"`
	PaySign   string `json:"pay_sign"`
}

type QueryPaymentStatusResponse struct {
	TradeState     string `json:"trade_state"`
	TradeStateDesc string `json:"trade_state_desc"`
	SuccessTime    string `json:"success_time,omitempty"` //yyyy-MM-DDTHH:mm:ss+TIMEZONE
	TransactionID  string `json:"transaction_id,omitempty"`
	Channel        string `json:"channel"`
}

type PaymentNotifyResponse struct {
	Code string `json:"code"`
	Msg  string `json:"massage"`
}
