package blacklist

import (
	"context"
	"fmt"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Filter wraps a slice of leases and removes any that are blacklisted.
func Filter(bl *Blacklist, leases []*lease.Lease) []*lease.Lease {
	out := make([]*lease.Lease, 0, len(leases))
	for _, l := range leases {
		if !bl.Contains(l.ID) {
			out = append(out, l)
		}
	}
	return out
}

// RenewGuard returns an error if the lease is blacklisted, preventing renewal.
func RenewGuard(bl *Blacklist, leaseID string) error {
	if bl.Contains(leaseID) {
		return fmt.Errorf("lease %s is blacklisted: renewal suppressed", leaseID)
	}
	return nil
}

// AlertGuard returns false if the lease should be suppressed from alerting.
func AlertGuard(bl *Blacklist, leaseID string) bool {
	return !bl.Contains(leaseID)
}

// ContextKey is used to store a Blacklist in a context.
type ContextKey struct{}

// WithContext stores the blacklist in the given context.
func WithContext(ctx context.Context, bl *Blacklist) context.Context {
	return context.WithValue(ctx, ContextKey{}, bl)
}

// FromContext retrieves a Blacklist from context, or returns an empty one.
func FromContext(ctx context.Context) *Blacklist {
	if bl, ok := ctx.Value(ContextKey{}).(*Blacklist); ok && bl != nil {
		return bl
	}
	return New()
}
