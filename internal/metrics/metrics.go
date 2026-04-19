// Package metrics provides a simple in-memory counter store for vaultwatch
// runtime statistics such as leases polled, alerts fired, and renewals attempted.
package metrics

import (
	"sync"
	"time"
)

// Counter names used across vaultwatch.
const (
	LeasesPolled    = "leases_polled"
	AlertsDispatched = "alerts_dispatched"
	RenewalsAttempted = "renewals_attempted"
	RenewalsFailed  = "renewals_failed"
)

// Snapshot holds a point-in-time copy of all counters.
type Snapshot struct {
	Counters  map[string]int64
	RecordedAt time.Time
}

// Registry holds named counters and is safe for concurrent use.
type Registry struct {
	mu       sync.Mutex
	counters map[string]int64
}

// New returns an initialised Registry.
func New() *Registry {
	return &Registry{counters: make(map[string]int64)}
}

// Inc increments the named counter by 1.
func (r *Registry) Inc(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name]++
}

// Add increments the named counter by delta.
func (r *Registry) Add(name string, delta int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += delta
}

// Get returns the current value of the named counter.
func (r *Registry) Get(name string) int64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.counters[name]
}

// Reset zeroes the named counter.
func (r *Registry) Reset(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] = 0
}

// Snapshot returns a copy of all counters at this moment.
func (r *Registry) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := make(map[string]int64, len(r.counters))
	for k, v := range r.counters {
		copy[k] = v
	}
	return Snapshot{Counters: copy, RecordedAt: time.Now()}
}
