package config

import (
	"io/ioutil"

	"github.com/sunbankio/tronevents/pkg/queue"
	"github.com/sunbankio/tronevents/pkg/redis"
	"gopkg.in/yaml.v2"
)

// TronConfig holds the configuration for the Tron client.
type TronConfig struct {
	NodeURL     string `yaml:"node_url"`
	Timeout     int    `yaml:"timeout"`
	PoolSize    int    `yaml:"pool_size"`
	MaxPoolSize int    `yaml:"max_pool_size"`
}

// Config holds the configuration for the entire daemon.
type Config struct {
	Redis    redis.Config `yaml:"redis"`
	Queue    queue.Config `yaml:"queue"`
	Tron     TronConfig   `yaml:"tron"`
	LogLevel string       `yaml:"log_level"`
}

// LoadFromFile loads the configuration from a YAML file.
func LoadFromFile(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
