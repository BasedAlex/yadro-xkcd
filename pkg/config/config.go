package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Path  string `yaml:"source_url"` 
	Limit int    `yaml:"limit"`
	DbFile string `yaml:"db_file"`
	DbPath string `yaml:"db_path"`
	IndexFile string `yaml:"index_file"`
	Parallel int `yaml:"parallel"`
	SrvPort	string `yaml:"srv_port"`
	Until int `yaml:"until"`
	DSN string `yaml:"pg_dsn"`
	ConcurrencyLimit int `yaml:"concurrency_limit"`
	RateLimit int `yaml:"rate_limit"`
	TokenMaxTime int `yaml:"token_max_time"`
	JWTSecret string `yaml:"jwt_secret"`
}

func Load(configPath string) (*Config, error) {
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	c := &Config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return nil, err
	}

	return c, nil
}