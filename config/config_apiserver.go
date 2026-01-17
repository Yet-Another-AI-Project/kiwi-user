package config

type APIServerConfig struct {
	Port             string `config:"port" default:"8080"`
	Index            int64  `config:"index" default:"0"`
	ExternalAPIToken string `config:"external_api_token" default:""`
}
