package config

type LogConfig struct {
	Level   string `config:"level" default:"info"`
	Format  string `config:"format" default:"console"`
	File    string `config:"file"`
	LLMFile string `config:"llm_file"`
	BiFile  string `config:"bi_file"`
}
