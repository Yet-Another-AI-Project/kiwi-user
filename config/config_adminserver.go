package config

type AdminServerConfig struct {
	Port string `config:"port" default:"8088"`
}
