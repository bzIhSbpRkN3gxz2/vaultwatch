// Package circuitbreaker provides a simple circuit breaker for Vault API calls.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the circuit breaker is open.
var ErrOpen = errors.New("circuit breaker is open")

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// Breaker is a simple circuit breaker.
type Breaker struct {
	mu          sync.Mutex
	failures    int
	maxFailures int
	openUntil   time.Time
	resetAfter  time.Duration
	state       State
}

// New creates a new Breaker that opens after maxFailures consecutive failures
// and attempts reset after resetAfter duration.
func New(maxFailures int, resetAfter time.Duration) *Breaker {
	return &Breaker{
		maxFailures: maxFailures,
		resetAfter:  resetAfter,
		state:       StateClosed,
	}
}

// Allow returns nil if the call is permitted, or ErrOpen if the circuit is open.
func (b *Breaker) Allow() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.state == StateOpen {
		if time.Now().After(b.openUntil) {
			b.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	}
	return nil
}

// Success records a successful call, resetting failure count.
func (b *Breaker) Success() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.state = StateClosed
}

// Failure records a failed call and may open the circuit.
func (b *Breaker) Failure() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.failures >= b.maxFailures {
		b.state = StateOpen
		b.openUntil = time.Now().Add(b.resetAfter)
	}
}

// State returns the current circuit state.
func (b *Breaker) CurrentState() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
