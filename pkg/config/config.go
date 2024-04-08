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
	Print bool
}

func Load() (*Config, error) {
	configPathBuild := "config.yaml"

	yamlFile, err := os.ReadFile(configPathBuild)
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