package grace

import (
	"context"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Handler is the signature for the next stage in a processing pipeline.
type Handler func(ctx context.Context, l *lease.Lease) error

// Guard wraps a Handler and records each processed lease in the Tracker when
// it falls within the grace window. The next handler is always called
// regardless of whether the lease is in grace.
func (t *Tracker) Guard(next Handler) Handler {
	return func(ctx context.Context, l *lease.Lease) error {
		t.Observe(l, time.Now())
		return next(ctx, l)
	}
}

// InGrace returns true when the supplied lease is currently tracked as being
// within its grace period.
func (t *Tracker) InGrace(leaseID string) bool {
	_, ok := t.Get(leaseID)
	return ok
}
