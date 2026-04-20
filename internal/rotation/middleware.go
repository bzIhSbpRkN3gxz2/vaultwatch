package rotation

import (
	"context"
	"fmt"

	"github.com/vaultwatch/internal/lease"
)

// RotateFunc is a function that performs the actual secret rotation for a lease.
type RotateFunc func(ctx context.Context, l *lease.Lease) error

// Guard wraps a RotateFunc with rotation tracking. It prevents concurrent
// rotations for the same lease and records success or failure.
func Guard(tracker *Tracker, fn RotateFunc) RotateFunc {
	return func(ctx context.Context, l *lease.Lease) error {
		if err := tracker.Begin(l); err != nil {
			return fmt.Errorf("rotation guard: %w", err)
		}

		if err := fn(ctx, l); err != nil {
			tracker.Fail(l.LeaseID, err.Error())
			return err
		}

		tracker.Complete(l.LeaseID)
		return nil
	}
}
