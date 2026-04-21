package jitter

import (
	"context"
	"time"
)

// TickFunc is the signature of a polling callback used by the scheduler.
type TickFunc func(ctx context.Context) error

// Wrap returns a new TickFunc that sleeps for a jittered duration derived
// from interval before delegating to fn.  This staggers concurrent pollers
// so they do not all hit Vault simultaneously.
func Wrap(s *Strategy, interval time.Duration, fn TickFunc) TickFunc {
	return func(ctx context.Context) error {
		delay := s.Apply(interval)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		return fn(ctx)
	}
}

// WrapSymmetric is like Wrap but uses ApplyFull so the delay may be
// shorter or longer than interval by up to Strategy.Factor.
func WrapSymmetric(s *Strategy, interval time.Duration, fn TickFunc) TickFunc {
	return func(ctx context.Context) error {
		delay := s.ApplyFull(interval)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
		return fn(ctx)
	}
}
