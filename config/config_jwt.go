package config

type JWTConfig struct {
	PublicKeyPath            string `config:"public_key_path" default:""`
	PrivateKeyPath           string `config:"private_key_path" default:""`
	AccessTokenExpireSecond  int64  `config:"access_token_expire" default:"600"`
	RefreshTokenExpireSecond int64  `config:"refresh_token_expire" default:"86400"`
}
