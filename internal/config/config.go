package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Provider represents a supported secret manager backend.
type Provider string

const (
	ProviderAWS   Provider = "aws"
	ProviderVault Provider = "vault"
	ProviderGCP   Provider = "gcp"
)

// ProviderConfig holds connection details for a single secret manager.
type ProviderConfig struct {
	Type    Provider          `yaml:"type"`
	Alias   string            `yaml:"alias"`
	Options map[string]string `yaml:"options"`
}

// Config is the top-level vaultswap configuration.
type Config struct {
	Version   string           `yaml:"version"`
	Providers []ProviderConfig `yaml:"providers"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// Validate checks that required fields are present and providers are valid.
func (c *Config) Validate() error {
	if c.Version == "" {
		return fmt.Errorf("version is required")
	}

	seen := make(map[string]bool)
	for i, p := range c.Providers {
		if p.Type != ProviderAWS && p.Type != ProviderVault && p.Type != ProviderGCP {
			return fmt.Errorf("provider[%d]: unknown type %q", i, p.Type)
		}
		if p.Alias == "" {
			return fmt.Errorf("provider[%d]: alias is required", i)
		}
		if seen[p.Alias] {
			return fmt.Errorf("provider[%d]: duplicate alias %q", i, p.Alias)
		}
		seen[p.Alias] = true
	}

	return nil
}
