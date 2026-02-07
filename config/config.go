package config

import (
	"reflect"

	"github.com/futurxlab/golanggraph/xerror"
	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

const (
	configFileFlag = "config.file"
)

type Config struct {
	APIServer  *APIServerConfig     `config:"api_server"`
	Log        *LogConfig           `config:"log"`
	Postgresql *PostgresqlConfig    `config:"postgres"`
	JWT        *JWTConfig           `config:"jwt"`
	Bootstrap  *BootstrapConfig     `config:"bootstrap"`
	Wechat     *WechatConfig        `config:"wechat"`
	Google     *GoogleConfig        `config:"google"`
	Posthog    *PosthogConfig       `config:"posthog"`
	Metrics    *MetricsConfig       `config:"metrics"`
	Payment    *PaymentConfig       `config:"payment"`
	Sms        *SmsConfig           `config:"sms"`
	OSS        *OSSConfig           `config:"oss"`
	Mail       *MailClientConfig    `config:"mail"`
	Captcha    *CaptchaClientConfig `config:"captcha"`
}

func NewConfig() (*Config, error) {
	c := config.New("config").
		WithOptions(func(opt *config.Options) {
			opt.DecoderConfig.TagName = "config"
		}).
		WithOptions(config.ParseDefault)

	c.AddDriver(yaml.Driver)

	// load flag config
	keys := []string{configFileFlag}
	if err := c.LoadFlags(keys); err != nil {
		return nil, xerror.Wrap(err)
	}

	configFile := c.String(configFileFlag)

	if configFile != "" {
		// load from config file
		if err := c.LoadFiles(configFile); err != nil {
			return nil, xerror.Wrap(err)
		}

		// load from env
		c.LoadOSEnvs(map[string]string{
			"API_SERVER_INDEX": "api_server.index",
		})
	}

	return decodeConfig(c)
}

func decodeConfig(c *config.Config) (*Config, error) {
	cfg := Config{
		APIServer:  &APIServerConfig{},
		Log:        &LogConfig{},
		Postgresql: &PostgresqlConfig{},
		JWT:        &JWTConfig{},
		Bootstrap:  &BootstrapConfig{},
		Wechat:     &WechatConfig{},
		Google:     &GoogleConfig{},
		Posthog:    &PosthogConfig{},
		Metrics:    &MetricsConfig{},
		Payment:    &PaymentConfig{},
		Sms:        &SmsConfig{},
		OSS:        &OSSConfig{},
		Mail:       &MailClientConfig{},
		Captcha:    &CaptchaClientConfig{},
	}

	t := reflect.TypeOf(cfg)

	v := reflect.ValueOf(cfg)

	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		configName := fieldType.Tag.Get("config")

		fieldValue := v.Field(i)
		if _, ok := c.GetValue(configName); !ok {
			c.Set(configName, fieldValue.Interface())
		}
		if err := c.BindStruct(configName, fieldValue.Elem().Addr().Interface()); err != nil {
			return nil, xerror.Wrap(err)
		}
	}

	return &cfg, nil
}
