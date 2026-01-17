package config

type PosthogConfig struct {
	ProjectID      string `config:"project_id"`
	ProjectAPIKey  string `config:"project_api_key"`
	PersonalAPIKey string `config:"personal_api_key"`
	Endpoint       string `config:"endpoint"`
}
