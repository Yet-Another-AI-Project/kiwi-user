package config

type PostgresqlConfig struct {
	ConnectionString string `config:"connection_string"`
}
