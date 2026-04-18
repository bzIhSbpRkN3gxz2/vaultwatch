package lease

import "time"

// Status represents the health state of a lease.
type Status string

const (
	StatusHealthy  Status = "healthy"
	StatusExpiring Status = "expiring"
	StatusExpired  Status = "expired"
	StatusOrphaned Status = "orphaned"
)

// Lease holds metadata about a Vault secret lease.
type Lease struct {
	ID          string
	Path        string
	Renewable   bool
	TTL         time.Duration
	ExpireTime  time.Time
	CreatedTime time.Time
}

// Status returns the current health status of the lease.
func (l *Lease) Status(warnThreshold time.Duration) Status {
	now := time.Now()

	if l.ExpireTime.IsZero() {
		return StatusOrphaned
	}
	if now.After(l.ExpireTime) {
		return StatusExpired
	}
	if l.ExpireTime.Sub(now) <= warnThreshold {
		return StatusExpiring
	}
	return StatusHealthy
}

// TimeRemaining returns the duration until the lease expires.
// Returns 0 if already expired.
func (l *Lease) TimeRemaining() time.Duration {
	if l.ExpireTime.IsZero() {
		return 0
	}
	remaining := time.Until(l.ExpireTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}
