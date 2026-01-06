package permission

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Action string

const (
	ActionAllow Action = "allow"
	ActionDeny  Action = "deny"
	ActionAsk   Action = "ask"
)

type Rule struct {
	Pattern string `yaml:"pattern"`
	Action  Action `yaml:"action"`
	Message string `yaml:"message,omitempty"`
}

type Config struct {
	Default Action `yaml:"default"`
	Rules   []Rule `yaml:"rules"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Default == "" {
		cfg.Default = ActionAsk
	}

	return &cfg, nil
}
