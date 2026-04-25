package staleness

import (
	"context"
	"fmt"

	"github.com/youorg/vaultwatch/internal/lease"
)

// AlertFunc is called when a lease exceeds the staleness threshold.
type AlertFunc func(ctx context.Context, l *lease.Lease, err error) error

// Guard wraps an AlertFunc and skips invocation for leases that have been
// stale beyond the tracker's threshold, delegating to onStale instead.
type Guard struct {
	tracker *Tracker
	onStale AlertFunc
}

// NewGuard returns a Guard backed by the given Tracker.
func NewGuard(tracker *Tracker, onStale AlertFunc) *Guard {
	return &Guard{tracker: tracker, onStale: onStale}
}

// Check observes the lease's current status and, if the staleness threshold
// has been exceeded, invokes onStale. It returns any error from onStale or
// from the underlying observation.
func (g *Guard) Check(ctx context.Context, l *lease.Lease) error {
	err := g.tracker.Observe(l)
	if err == ErrThresholdExceeded {
		return g.onStale(ctx, l, fmt.Errorf("lease %s stale in status %s: %w", l.LeaseID, l.Status, err))
	}
	return err
}
