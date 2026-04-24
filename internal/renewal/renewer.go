package renewal

import (
	"context"
	"fmt"
	"log"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/youorg/vaultwatch/internal/lease"
)

// Renewer attempts to renew Vault leases before they expire.
type Renewer struct {
	client    *vaultapi.Client
	threshold time.Duration
	logger    *log.Logger
}

// Config holds configuration for the Renewer.
type Config struct {
	Client    *vaultapi.Client
	Threshold time.Duration // renew when TTL drops below this
	Logger    *log.Logger
}

// New creates a new Renewer with the given config.
func New(cfg Config) *Renewer {
	if cfg.Threshold == 0 {
		cfg.Threshold = 5 * time.Minute
	}
	if cfg.Logger == nil {
		cfg.Logger = log.Default()
	}
	return &Renewer{
		client:    cfg.Client,
		threshold: cfg.Threshold,
		logger:    cfg.Logger,
	}
}

// MaybeRenew renews the lease if its remaining TTL is below the threshold.
// Returns true if a renewal was attempted.
func (r *Renewer) MaybeRenew(ctx context.Context, l *lease.Lease) (bool, error) {
	if l.LeaseID == "" {
		return false, nil
	}
	if l.TTL > r.threshold {
		return false, nil
	}

	secret, err := r.client.Sys().RenewWithContext(ctx, l.LeaseID, int(l.TTL.Seconds()))
	if err != nil {
		return true, fmt.Errorf("renew lease %s: %w", l.LeaseID, err)
	}

	newTTL := time.Duration(secret.LeaseDuration) * time.Second
	if newTTL == 0 {
		return true, fmt.Errorf("renew lease %s: server returned zero TTL", l.LeaseID)
	}

	r.logger.Printf("renewed lease %s: new TTL %s", l.LeaseID, newTTL)
	l.TTL = newTTL
	return true, nil
}
