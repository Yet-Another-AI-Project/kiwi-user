package config

type RedisConfig struct {
	Host     string `config:"host"`
	Password string `config:"password"`
	DbIndex  int    `config:"db_index"`
}
