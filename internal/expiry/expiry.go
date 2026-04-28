// Package expiry provides utilities for computing and comparing lease
// expiration times relative to a configurable warning horizon.
package expiry

import (
	"errors"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// ErrExpired is returned when a lease has already passed its expiry time.
var ErrExpired = errors.New("expiry: lease has already expired")

// ErrNoExpiry is returned when a lease carries no expiration information.
var ErrNoExpiry = errors.New("expiry: lease has no expiry time")

// Checker evaluates lease expiration state against a warning horizon.
type Checker struct {
	warnBefore time.Duration
	now        func() time.Time
}

// New returns a Checker that emits warnings when a lease expires within
// warnBefore of the current time.
func New(warnBefore time.Duration) *Checker {
	if warnBefore <= 0 {
		warnBefore = 5 * time.Minute
	}
	return &Checker{
		warnBefore: warnBefore,
		now:        time.Now,
	}
}

// Result holds the outcome of an expiry evaluation.
type Result struct {
	// ExpiresAt is the absolute expiry time derived from the lease TTL.
	ExpiresAt time.Time
	// Remaining is the duration until expiry (negative if already expired).
	Remaining time.Duration
	// Warn is true when the lease will expire within the configured horizon.
	Warn bool
	// Expired is true when the lease has already expired.
	Expired bool
}

// Evaluate computes the expiry Result for the given lease.
// It returns ErrNoExpiry when the lease carries a zero TTL.
func (c *Checker) Evaluate(l *lease.Lease) (Result, error) {
	if l == nil {
		return Result{}, ErrNoExpiry
	}
	if l.TTL <= 0 {
		return Result{}, ErrNoExpiry
	}

	now := c.now()
	expiresAt := l.CreatedAt.Add(l.TTL)
	remaining := expiresAt.Sub(now)

	r := Result{
		ExpiresAt: expiresAt,
		Remaining: remaining,
		Expired:   remaining <= 0,
		Warn:      remaining > 0 && remaining <= c.warnBefore,
	}

	if r.Expired {
		return r, ErrExpired
	}
	return r, nil
}

// TimeUntilWarn returns how long until the warning horizon is reached.
// A non-positive value means the warning window is already active.
func (c *Checker) TimeUntilWarn(l *lease.Lease) (time.Duration, error) {
	r, err := c.Evaluate(l)
	if errors.Is(err, ErrNoExpiry) {
		return 0, err
	}
	return r.Remaining - c.warnBefore, nil
}
