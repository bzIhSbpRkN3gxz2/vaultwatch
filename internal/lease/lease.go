package lease

import (
	"fmt"
	"time"
)

// Status represents the health state of a lease.
type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusExpiring Status = "expiring"
	StatusExpired  Status = "expired"
	StatusOrphaned Status = "orphaned"
)

func (s Status) String() string { return string(s) }

// Lease holds metadata about a single Vault secret lease.
type Lease struct {
	ID        string
	Path      string
	ExpiresAt time.Time
	Orphaned  bool
}

// New constructs a Lease.
func New(id, path string, expiresAt time.Time, orphaned bool) *Lease {
	return &Lease{
		ID:        id,
		Path:      path,
		ExpiresAt: expiresAt,
		Orphaned:  orphaned,
	}
}

// Status evaluates the current state of the lease.
// warnThreshold is how far in the future expiry must be to be considered healthy.
func (l *Lease) Status(warnThreshold time.Duration) Status {
	if l.Orphaned {
		return StatusOrphaned
	}
	ttl := time.Until(l.ExpiresAt)
	switch {
	case ttl <= 0:
		return StatusExpired
	case ttl < warnThreshold:
		return StatusExpiring
	default:
		return StatusHealthy
	}
}

// TTL returns a human-readable remaining duration string.
func (l *Lease) TTL() string {
	ttl := time.Until(l.ExpiresAt)
	if ttl <= 0 {
		return "expired"
	}
	return fmt.Sprintf("%v", ttl.Truncate(time.Second))
}

// IsExpired reports whether the lease has passed its expiry time.
func (l *Lease) IsExpired() bool {
	return !l.Orphaned && !l.ExpiresAt.IsZero() && time.Now().After(l.ExpiresAt)
}
