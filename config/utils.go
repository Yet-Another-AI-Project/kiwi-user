package config

import (
	"strings"
)

type Port struct {
	s string
}

func (p *Port) Set(s string) error {
	p.s = s
	return nil
}

func (p *Port) String() string {
	return p.s
}

type LLMOption struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
}

func ParseLLMConnectionString(connectionString string) LLMOption {
	items := strings.Split(connectionString, ";")
	return LLMOption{
		Provider: items[0],
		BaseURL:  items[1],
		APIKey:   items[2],
		Model:    items[3],
	}
}
