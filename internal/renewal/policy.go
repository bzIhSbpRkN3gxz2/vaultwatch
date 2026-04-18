package renewal

import "time"

// Policy defines when and how aggressively to renew leases.
type Policy struct {
	// WarnThreshold is the TTL below which a lease is considered expiring.
	WarnThreshold time.Duration
	// RenewThreshold is the TTL below which renewal is attempted.
	RenewThreshold time.Duration
	// MaxRetries is the number of renewal attempts before giving up.
	MaxRetries int
}

// DefaultPolicy returns a sensible default renewal policy.
func DefaultPolicy() Policy {
	return Policy{
		WarnThreshold:  10 * time.Minute,
		RenewThreshold: 5 * time.Minute,
		MaxRetries:     3,
	}
}

// ShouldWarn reports whether the given TTL warrants a warning alert.
func (p Policy) ShouldWarn(ttl time.Duration) bool {
	return ttl > 0 && ttl <= p.WarnThreshold
}

// ShouldRenew reports whether the given TTL warrants a renewal attempt.
func (p Policy) ShouldRenew(ttl time.Duration) bool {
	return ttl > 0 && ttl <= p.RenewThreshold
}
