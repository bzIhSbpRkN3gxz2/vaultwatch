package retrier

import (
	"context"
	"errors"
	"time"
)

// Policy defines retry behaviour.
type Policy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultPolicy returns a sensible retry policy.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts: 4,
		BaseDelay:   250 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}
}

// Retrier executes a function with exponential back-off.
type Retrier struct {
	policy Policy
	sleep  func(time.Duration)
}

// New returns a Retrier using the given policy.
func New(p Policy) *Retrier {
	return &Retrier{policy: p, sleep: time.Sleep}
}

// ErrMaxAttempts is returned when all attempts are exhausted.
var ErrMaxAttempts = errors.New("retrier: max attempts reached")

// Do calls fn until it returns nil, the context is cancelled, or attempts are
// exhausted. The last non-nil error is wrapped with ErrMaxAttempts.
func (r *Retrier) Do(ctx context.Context, fn func() error) error {
	delay := r.policy.BaseDelay
	var last error
	for i := 0; i < r.policy.MaxAttempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		if last = fn(); last == nil {
			return nil
		}
		if i < r.policy.MaxAttempts-1 {
			r.sleep(delay)
			delay = time.Duration(float64(delay) * r.policy.Multiplier)
			if delay > r.policy.MaxDelay {
				delay = r.policy.MaxDelay
			}
		}
	}
	return errors.Join(ErrMaxAttempts, last)
}
