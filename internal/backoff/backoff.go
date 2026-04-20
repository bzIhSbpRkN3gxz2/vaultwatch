// Package backoff provides configurable exponential back-off strategies
// used when retrying Vault API calls or alert deliveries.
package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy defines how successive delays are computed.
type Strategy struct {
	Initial    time.Duration
	Multiplier float64
	MaxDelay   time.Duration
	Jitter     bool
}

// DefaultStrategy returns a sensible exponential back-off configuration.
func DefaultStrategy() Strategy {
	return Strategy{
		Initial:    200 * time.Millisecond,
		Multiplier: 2.0,
		MaxDelay:   30 * time.Second,
		Jitter:     true,
	}
}

// Delay returns the back-off duration for the given attempt (0-indexed).
// When Jitter is enabled a random fraction up to 25 % of the computed delay
// is added so that concurrent callers do not all retry simultaneously.
func (s Strategy) Delay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := float64(s.Initial) * math.Pow(s.Multiplier, float64(attempt))
	if base > float64(s.MaxDelay) {
		base = float64(s.MaxDelay)
	}
	if s.Jitter {
		// add up to 25 % random jitter
		base += base * 0.25 * rand.Float64() //nolint:gosec
	}
	d := time.Duration(base)
	if d > s.MaxDelay {
		d = s.MaxDelay
	}
	return d
}

// Reset returns the initial delay, equivalent to Delay(0).
func (s Strategy) Reset() time.Duration {
	return s.Delay(0)
}
