package healthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Status represents the health of a named component.
type Status struct {
	Name    string
	Healthy bool
	Message string
	CheckedAt time.Time
}

// Checker is a function that returns an error if the component is unhealthy.
type Checker func(ctx context.Context) error

// Registry holds named health checkers and runs them on demand.
type Registry struct {
	mu       sync.RWMutex
	checkers map[string]Checker
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{checkers: make(map[string]Checker)}
}

// Register adds a named checker to the registry.
func (r *Registry) Register(name string, c Checker) error {
	if name == "" {
		return fmt.Errorf("healthcheck: name must not be empty")
	}
	if c == nil {
		return fmt.Errorf("healthcheck: checker must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checkers[name] = c
	return nil
}

// RunAll executes every registered checker and returns their statuses.
func (r *Registry) RunAll(ctx context.Context) []Status {
	r.mu.RLock()
	names := make([]string, 0, len(r.checkers))
	copy := make(map[string]Checker, len(r.checkers))
	for k, v := range r.checkers {
		names = append(names, k)
		copy[k] = v
	}
	r.mu.RUnlock()

	results := make([]Status, 0, len(names))
	for _, name := range names {
		s := Status{Name: name, CheckedAt: time.Now()}
		if err := copy[name](ctx); err != nil {
			s.Healthy = false
			s.Message = err.Error()
		} else {
			s.Healthy = true
			s.Message = "ok"
		}
		results = append(results, s)
	}
	return results
}

// Healthy returns true only when all checkers pass.
func (r *Registry) Healthy(ctx context.Context) bool {
	for _, s := range r.RunAll(ctx) {
		if !s.Healthy {
			return false
		}
	}
	return true
}
