package config

import (
	"errors"
	"gopkg.in/yaml.v3"
	"os"
)

var (
	ConfigFilePath string
)

type Check struct {
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

type Config struct {
	Checks map[string]*Check `yaml:"checks"`
}

func NewConfig() (*Config, error) {

	if ConfigFilePath == "" {
		return nil, errors.New("config file path required")
	}

	b, err := os.ReadFile(ConfigFilePath)

	if err != nil {
		return nil, err
	}

	c := &Config{}

	err = yaml.Unmarshal(b, c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
