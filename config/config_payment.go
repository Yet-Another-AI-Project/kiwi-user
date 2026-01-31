package config

type PaymentConfig struct {
	WechatAppID          string                 `config:"wechat_app_id"`
	WechatMchID          string                 `config:"wechat_mch_id"`
	WechatMchAPIv3Key    string                 `config:"wechat_mch_api_v3_key"`
	WechatPrivateKeyPath string                 `config:"wechat_private_key_path"`
	WechatPublicKeyPath  string                 `config:"wechat_public_key_path"`
	WechatPublicKeyID    string                 `config:"wechat_public_key_id"`
	WechatCertSerialNo   string                 `config:"wechat_cert_serial_no"`
	WechatNotifyURL      string                 `config:"wechat_notify_url"`
	AESEncryptKey        string                 `config:"aes_encrypt_key"`
	ServiceNotifyURL     map[string]interface{} `config:"service_notify_url"`

	// Stripe configuration
	StripeAPIKey         string `config:"stripe_api_key"`
	StripeWebhookSecret  string `config:"stripe_webhook_secret"`
	StripeSuccessURL     string `config:"stripe_success_url"`
	StripeCancelURL      string `config:"stripe_cancel_url"`
	StripeMonthlyPriceID string `config:"stripe_monthly_price_id"`
	StripeYearlyPriceID  string `config:"stripe_yearly_price_id"`
}
