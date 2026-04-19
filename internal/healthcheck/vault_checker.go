package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// VaultCheckerConfig holds configuration for the Vault health checker.
type VaultCheckerConfig struct {
	// Address is the Vault server address, e.g. "https://vault:8200".
	Address string
	// Timeout is the HTTP request timeout.
	Timeout time.Duration
}

// NewVaultChecker returns a Checker that queries the Vault /v1/sys/health
// endpoint. It treats standby and performance-standby as healthy.
func NewVaultChecker(cfg VaultCheckerConfig) Checker {
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	client := &http.Client{Timeout: cfg.Timeout}

	return func(ctx context.Context) error {
		url := fmt.Sprintf("%s/v1/sys/health?standbyok=true&perfstandbyok=true", cfg.Address)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return fmt.Errorf("healthcheck vault: build request: %w", err)
		}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("healthcheck vault: %w", err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("healthcheck vault: unexpected status %d", resp.StatusCode)
		}
		return nil
	}
}
