package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_Defaults(t *testing.T) {
	path := writeTempConfig(t, "{}\n")
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("expected default vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", cfg.Monitor.Interval)
	}
}

func TestLoad_OverridesValues(t *testing.T) {
	content := `
vault:
  address: "https://vault.example.com"
  token: "s.abc123"
monitor:
  interval: 1m
  warn_threshold: 10m
alert:
  slack_webhook_url: "https://hooks.slack.com/test"
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "https://vault.example.com" {
		t.Errorf("unexpected address: %q", cfg.Vault.Address)
	}
	if cfg.Vault.Token != "s.abc123" {
		t.Errorf("unexpected token: %q", cfg.Vault.Token)
	}
	if cfg.Monitor.Interval != time.Minute {
		t.Errorf("unexpected interval: %v", cfg.Monitor.Interval)
	}
	if cfg.Alert.SlackWebhookURL != "https://hooks.slack.com/test" {
		t.Errorf("unexpected slack url: %q", cfg.Alert.SlackWebhookURL)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestDefault_Values(t *testing.T) {
	cfg := config.Default()
	if cfg.Monitor.WarnThreshold != 5*time.Minute {
		t.Errorf("expected 5m warn threshold, got %v", cfg.Monitor.WarnThreshold)
	}
}
