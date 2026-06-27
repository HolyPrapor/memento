package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type DownrankRule struct {
	Paths  []string `yaml:"paths"`
	Factor float64  `yaml:"factor"`
}

type Config struct {
	Downrank []DownrankRule `yaml:"downrank"`
}

func LoadConfig(dbPath string) (*Config, error) {
	dir := filepath.Dir(dbPath)
	configPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *Config) ApplyDownrank(path string, relevance float64) float64 {
	for _, rule := range c.Downrank {
		for _, pattern := range rule.Paths {
			if ok, _ := filepath.Match(pattern, path); ok {
				return relevance * rule.Factor
			}
		}
	}
	return relevance
}
