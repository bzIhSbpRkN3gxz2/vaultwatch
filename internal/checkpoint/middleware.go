package checkpoint

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultwatch/internal/lease"
)

type contextKey struct{}

// WithContext stores the Checkpoint in the context.
func WithContext(ctx context.Context, cp *Checkpoint) context.Context {
	return context.WithValue(ctx, contextKey{}, cp)
}

// FromContext retrieves the Checkpoint from the context.
// Returns nil if none was stored.
func FromContext(ctx context.Context) *Checkpoint {
	cp, _ := ctx.Value(contextKey{}).(*Checkpoint)
	return cp
}

// Stage returns a pipeline-compatible stage function that records each lease
// into the Checkpoint found in the context, then passes the lease through
// unchanged. If no Checkpoint is present the stage is a no-op.
func Stage(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
	cp := FromContext(ctx)
	if cp == nil {
		return l, nil
	}
	if l == nil {
		return nil, fmt.Errorf("checkpoint stage: nil lease")
	}
	cp.Set(Entry{
		LeaseID: l.ID,
		Path:    l.Path,
		Status:  string(l.Status()),
		TTL:     int64(l.TTL().Seconds()),
	})
	return l, nil
}
