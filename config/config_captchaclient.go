package config

type CaptchaClientConfig struct {
	AccessKeySecret string `config:"access_key_secret"`
	AccessKeyID     string `config:"access_key_id"`
}
