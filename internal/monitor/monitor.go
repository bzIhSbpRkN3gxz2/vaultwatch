package monitor

import (
	"context"
	"log"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Config holds monitor configuration.
type Config struct {
	PollInterval  time.Duration
	WarnThreshold time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		PollInterval:  30 * time.Second,
		WarnThreshold: 24 * time.Hour,
	}
}

// LeaseProvider is anything that can return a list of leases.
type LeaseProvider interface {
	ListLeases(ctx context.Context) ([]*lease.Lease, error)
}

// AlertHandler is called when a lease needs attention.
type AlertHandler interface {
	OnAlert(l *lease.Lease, status lease.Status)
}

// Monitor polls a LeaseProvider and dispatches alerts.
type Monitor struct {
	cfg      Config
	provider LeaseProvider
	handler  AlertHandler
}

// New creates a new Monitor.
func New(cfg Config, provider LeaseProvider, handler AlertHandler) *Monitor {
	return &Monitor{cfg: cfg, provider: provider, handler: handler}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.poll(ctx); err != nil {
				log.Printf("monitor: poll error: %v", err)
			}
		}
	}
}

func (m *Monitor) poll(ctx context.Context) error {
	leases, err := m.provider.ListLeases(ctx)
	if err != nil {
		return err
	}
	for _, l := range leases {
		status := l.Status(m.cfg.WarnThreshold)
		if status != lease.StatusHealthy {
			m.handler.OnAlert(l, status)
		}
	}
	return nil
}
