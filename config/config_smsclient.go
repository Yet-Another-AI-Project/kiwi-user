package config

type SmsConfig struct {
	AccessKey        string `config:"access_key"`         // AK
	SecretKey        string `config:"secret_key"`         // SK
	SmsAccount       string `config:"sms_account"`        // 短信账户
	SignName         string `config:"sign_name"`          // 短信签名
	VerifyTemplateID string `config:"verify_template_id"` // 验证码模板ID
	SmsTemplateId    string `config:"sms_template_id"`    // 短信模板ID
	TemplateParam    string `config:"template_param"`     // 当指定的短信模板（TemplateID）存在变量时，您需要设置变量的实际值。支持传入一个或多个参数，格式示例：{"code1":"1234", "code2":"5678"}
	DefaultScene     string `config:"default_scene"`      // 默认使用场景
	Tag              string `config:"tag"`                // 透传字段
}
