package config

type WechatConfig struct {
	// 移动端
	AppID     string `config:"app_id"`
	AppSecret string `config:"app_secret"`

	// 小程序
	MiniProgramID     string `config:"mini_program_id"`
	MiniProgramSecret string `config:"mini_program_secret"`

	// 网站应用
	WebID     string `config:"web_id"`
	WebSecret string `config:"web_secret"`

	// 公众号
	OfficalAccountID     string `config:"offical_account_id"`
	OfficalAccountSecret string `config:"offical_account_secret"`

	// 企业微信
	QyWechatCorpID     string `config:"qy_wechat_corp_id"`
	QyWechatCorpSecret string `config:"qy_wechat_corp_secret"`
}
