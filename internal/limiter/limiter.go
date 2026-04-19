// Package limiter provides rate limiting for Vault API calls.
package limiter

import (
	"sync"
	"time"
)

// Limiter enforces a maximum number of operations per interval.
type Limiter struct {
	mu       sync.Mutex
	max      int
	interval time.Duration
	tokens   int
	last     time.Time
}

// New creates a Limiter allowing at most max calls per interval.
func New(max int, interval time.Duration) *Limiter {
	return &Limiter{
		max:      max,
		interval: interval,
		tokens:   max,
		last:     time.Now(),
	}
}

// Allow returns true if the operation is permitted under the rate limit.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens <= 0 {
		return false
	}
	l.tokens--
	return true
}

// Remaining returns the number of tokens currently available.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

// Reset restores the token count to the configured maximum.
func (l *Limiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.tokens = l.max
	l.last = time.Now()
}

// refill replenishes tokens if the interval has elapsed. Must be called with mu held.
func (l *Limiter) refill() {
	if time.Since(l.last) >= l.interval {
		l.tokens = l.max
		l.last = time.Now()
	}
}
