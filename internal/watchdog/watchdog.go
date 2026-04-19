// Package watchdog detects stale or orphaned leases that have not been
// refreshed within an expected interval and emits alerts accordingly.
package watchdog

import (
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Entry tracks the last time a lease was seen alive.
type Entry struct {
	LeaseID  string
	LastSeen time.Time
}

// Watchdog holds lease heartbeat state and detects staleness.
type Watchdog struct {
	mu      sync.Mutex
	entries map[string]Entry
	ttl     time.Duration
}

// New creates a Watchdog that considers a lease stale after ttl without a heartbeat.
func New(ttl time.Duration) *Watchdog {
	return &Watchdog{
		entries: make(map[string]Entry),
		ttl:     ttl,
	}
}

// Heartbeat records that a lease was seen at the current time.
func (w *Watchdog) Heartbeat(leaseID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.entries[leaseID] = Entry{LeaseID: leaseID, LastSeen: time.Now()}
}

// Stale returns all leases that have not received a heartbeat within the configured TTL.
func (w *Watchdog) Stale(leases []*lease.Lease) []*lease.Lease {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := time.Now()
	var stale []*lease.Lease
	for _, l := range leases {
		e, ok := w.entries[l.ID]
		if !ok || now.Sub(e.LastSeen) > w.ttl {
			stale = append(stale, l)
		}
	}
	return stale
}

// Purge removes entries for leases that are no longer tracked.
func (w *Watchdog) Purge(active map[string]struct{}) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for id := range w.entries {
		if _, ok := active[id]; !ok {
			delete(w.entries, id)
		}
	}
}
