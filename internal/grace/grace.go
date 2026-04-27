// Package grace tracks leases that are within a configurable grace period
// before expiry, allowing downstream components to take action before a hard
// expiration occurs.
package grace

import (
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Entry records when a lease entered the grace window.
type Entry struct {
	LeaseID   string
	EnteredAt time.Time
	ExpiresAt time.Time
}

// Tracker monitors leases approaching expiry within a grace window.
type Tracker struct {
	mu      sync.Mutex
	entries map[string]Entry
	window  time.Duration
}

// New creates a Tracker that considers leases within window of expiry to be
// in their grace period.
func New(window time.Duration) *Tracker {
	return &Tracker{
		entries: make(map[string]Entry),
		window:  window,
	}
}

// Observe checks whether l is within the grace window. If so, the lease is
// recorded and true is returned. If the lease has already been recorded the
// existing entry is preserved and true is returned.
func (t *Tracker) Observe(l *lease.Lease, now time.Time) bool {
	if l == nil || l.LeaseID == "" {
		return false
	}
	expiresAt := now.Add(time.Duration(l.TTL) * time.Second)
	if time.Duration(l.TTL)*time.Second > t.window {
		return false
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.entries[l.LeaseID]; !ok {
		t.entries[l.LeaseID] = Entry{
			LeaseID:   l.LeaseID,
			EnteredAt: now,
			ExpiresAt: expiresAt,
		}
	}
	return true
}

// Get returns the grace Entry for the given leaseID and whether it exists.
func (t *Tracker) Get(leaseID string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[leaseID]
	return e, ok
}

// Remove deletes the entry for leaseID.
func (t *Tracker) Remove(leaseID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, leaseID)
}

// Purge removes all entries whose ExpiresAt is before now.
func (t *Tracker) Purge(now time.Time) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	removed := 0
	for id, e := range t.entries {
		if e.ExpiresAt.Before(now) {
			delete(t.entries, id)
			removed++
		}
	}
	return removed
}
