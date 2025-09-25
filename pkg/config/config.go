package config

import (
	"io/ioutil"

	"github.com/sunbankio/tronevents/pkg/queue"
	"github.com/sunbankio/tronevents/pkg/redis"
	"gopkg.in/yaml.v2"
)

// Config holds the configuration for the entire daemon.
type Config struct {
	Redis       redis.Config `yaml:"redis"`
	Queue       queue.Config `yaml:"queue"`
	TronNodeURL string       `yaml:"tron_node_url"`
	LogLevel    string       `yaml:"log_level"`
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
