package config

type MailClientConfig struct {
	ResendAPIKey        string `config:"resend_api_key"`
	ResendFromEmail     string `config:"resend_from_email"`
	VertifyCodeTemplate string `config:"vertify_code_template"`
	VertifyCodeSubject  string `config:"vertify_code_subject"`
	FeishuWebhook       string `config:"feishu_webhook"`
}
