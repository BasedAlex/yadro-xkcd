package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host  string `yaml:"host"` // localhost
	Port  string `yaml:"port"` // 8080
	Path  string `yaml:"source_url"` // https://xkcd.com/2651/info.0.json
	Start int    `yaml:"start"`
	Limit int    `yaml:"limit"`
	DbFile string `yaml:"db_file"`
	DbPath string `yaml:"db_path"`
	Print bool
}

func Load() (*Config, error) {

	configPathBuild := "config.yaml"
	// configPathNew := "../../config.yaml"

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