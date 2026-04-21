// Package jitter provides utilities for adding randomised delay to
// polling intervals, reducing thundering-herd effects when many leases
// expire at roughly the same time.
package jitter

import (
	"math/rand"
	"time"
)

// Strategy holds the parameters that control how jitter is applied.
type Strategy struct {
	// Factor is the maximum fraction of the base duration that may be
	// added as random noise.  A value of 0.2 means ±20 % of base.
	Factor float64
	// rng is the source of randomness; seeded on construction.
	rng *rand.Rand
}

// DefaultStrategy returns a Strategy with a 20 % jitter factor.
func DefaultStrategy() *Strategy {
	return &Strategy{
		Factor: 0.2,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// New creates a Strategy with the supplied factor clamped to [0, 1].
func New(factor float64) *Strategy {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return &Strategy{
		Factor: factor,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Apply returns base plus a random duration in [0, base*Factor).
func (s *Strategy) Apply(base time.Duration) time.Duration {
	if s.Factor == 0 || base <= 0 {
		return base
	}
	max := float64(base) * s.Factor
	noise := time.Duration(s.rng.Float64() * max)
	return base + noise
}

// ApplyFull returns a duration in [base*(1-Factor), base*(1+Factor)].
func (s *Strategy) ApplyFull(base time.Duration) time.Duration {
	if s.Factor == 0 || base <= 0 {
		return base
	}
	half := float64(base) * s.Factor
	noise := time.Duration(s.rng.Float64()*half*2 - half)
	d := base + noise
	if d < 0 {
		return 0
	}
	return d
}
