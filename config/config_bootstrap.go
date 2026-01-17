package config

type BootstrapConfig struct {
	FirstUserName string `config:"first_user_name"`
	FirstUserPass string `config:"first_user_pass"`
}
