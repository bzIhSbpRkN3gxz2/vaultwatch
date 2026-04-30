package digest

import (
	"context"

	"github.com/vaultwatch/internal/lease"
)

// Handler is a lease processing function used in pipeline stages.
type Handler func(ctx context.Context, l *lease.Lease) error

// OnlyChanged returns a Handler that invokes next only when the lease
// state has changed since the last observation. Unchanged leases are
// silently skipped, reducing downstream noise.
func OnlyChanged(tr *Tracker, next Handler) Handler {
	return func(ctx context.Context, l *lease.Lease) error {
		if !tr.Changed(l) {
			return nil
		}
		return next(ctx, l)
	}
}
