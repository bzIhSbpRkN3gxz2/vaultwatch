// Package circuit provides a per-lease circuit breaker registry that wraps
// the core circuitbreaker primitive and tracks open/closed state per lease ID.
package circuit

import (
	"errors"
	"sync"
	"time"

	"github.com/vaultwatch/internal/circuitbreaker"
)

// ErrCircuitOpen is returned when a lease's circuit is open.
var ErrCircuitOpen = errors.New("circuit open: lease operations suspended")

// Config holds tuning parameters for the registry.
type Config struct {
	FailureThreshold int
	HalfOpenTimeout  time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		FailureThreshold: 3,
		HalfOpenTimeout:  30 * time.Second,
	}
}

// Registry manages one circuit breaker per lease ID.
type Registry struct {
	mu       sync.Mutex
	cfg      Config
	breakers map[string]*circuitbreaker.Breaker
}

// New creates a Registry with the given Config.
func New(cfg Config) *Registry {
	return &Registry{
		cfg:      cfg,
		breakers: make(map[string]*circuitbreaker.Breaker),
	}
}

func (r *Registry) breaker(leaseID string) *circuitbreaker.Breaker {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.breakers[leaseID]; ok {
		return b
	}
	b := circuitbreaker.New(r.cfg.FailureThreshold, r.cfg.HalfOpenTimeout)
	r.breakers[leaseID] = b
	return b
}

// Allow returns nil if the lease's circuit is closed, ErrCircuitOpen otherwise.
func (r *Registry) Allow(leaseID string) error {
	if !r.breaker(leaseID).Allow() {
		return ErrCircuitOpen
	}
	return nil
}

// Success records a successful operation for the lease.
func (r *Registry) Success(leaseID string) {
	r.breaker(leaseID).Success()
}

// Failure records a failed operation for the lease.
func (r *Registry) Failure(leaseID string) {
	r.breaker(leaseID).Failure()
}

// Remove deletes the circuit breaker entry for a lease.
func (r *Registry) Remove(leaseID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.breakers, leaseID)
}

// Len returns the number of tracked leases.
func (r *Registry) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.breakers)
}
