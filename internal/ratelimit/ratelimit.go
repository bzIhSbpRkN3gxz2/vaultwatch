// Package ratelimit provides a token-bucket rate limiter for Vault API calls.
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Limiter controls the rate of outbound Vault API requests.
type Limiter struct {
	mu       sync.Mutex
	tokens   float64
	max      float64
	rate     float64 // tokens per second
	lastTick time.Time
	clock    func() time.Time
}

// New creates a Limiter that allows up to max requests per second.
func New(maxPerSecond float64) (*Limiter, error) {
	if maxPerSecond <= 0 {
		return nil, fmt.Errorf("ratelimit: maxPerSecond must be > 0, got %f", maxPerSecond)
	}
	return &Limiter{
		tokens:   maxPerSecond,
		max:      maxPerSecond,
		rate:     maxPerSecond,
		lastTick: time.Now(),
		clock:    time.Now,
	}, nil
}

// Wait blocks until a token is available or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("ratelimit: context cancelled: %w", err)
		}
		l.mu.Lock()
		l.refill()
		if l.tokens >= 1 {
			l.tokens--
			l.mu.Unlock()
			return nil
		}
		l.mu.Unlock()
		select {
		case <-ctx.Done():
			return fmt.Errorf("ratelimit: context cancelled: %w", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// Available returns the current token count.
func (l *Limiter) Available() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

// refill adds tokens based on elapsed time. Must be called with l.mu held.
func (l *Limiter) refill() {
	now := l.clock()
	elapsed := now.Sub(l.lastTick).Seconds()
	l.tokens += elapsed * l.rate
	if l.tokens > l.max {
		l.tokens = l.max
	}
	l.lastTick = now
}
