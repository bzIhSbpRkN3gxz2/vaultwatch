package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultwatch configuration.
type Config struct {
	Vault    VaultConfig    `yaml:"vault"`
	Monitor  MonitorConfig  `yaml:"monitor"`
	Alert    AlertConfig    `yaml:"alert"`
	Audit    AuditConfig    `yaml:"audit"`
}

type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
}

type MonitorConfig struct {
	Interval      time.Duration `yaml:"interval"`
	WarnThreshold time.Duration `yaml:"warn_threshold"`
}

type AlertConfig struct {
	SlackWebhookURL string `yaml:"slack_webhook_url"`
	WebhookURL      string `yaml:"webhook_url"`
}

type AuditConfig struct {
	FilePath string `yaml:"file_path"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("config: open %q: %w", path, err)
	}
	defer f.Close()

	cfg := Default()
	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, fmt.Errorf("config: decode: %w", err)
	}
	return cfg, nil
}

// Default returns a Config populated with sensible defaults.
func Default() *Config {
	return &Config{
		Vault: VaultConfig{
			Address: "http://127.0.0.1:8200",
		},
		Monitor: MonitorConfig{
			Interval:      30 * time.Second,
			WarnThreshold: 5 * time.Minute,
		},
	}
}
