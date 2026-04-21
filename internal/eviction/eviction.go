// Package eviction provides lease eviction tracking and enforcement.
// It records eviction events and prevents re-admission of recently evicted leases.
package eviction

import (
	"errors"
	"sync"
	"time"
)

// ErrEvicted is returned when a lease has been evicted and is within its quarantine window.
var ErrEvicted = errors.New("lease has been evicted")

// Entry records a single eviction event.
type Entry struct {
	LeaseID   string
	Reason    string
	EvictedAt time.Time
	ExpiresAt time.Time
}

// Tracker records and enforces lease evictions.
type Tracker struct {
	mu         sync.RWMutex
	entries    map[string]Entry
	quarantine time.Duration
	now        func() time.Time
}

// New creates a Tracker with the given quarantine duration.
// Evicted leases are blocked for the quarantine period before re-admission.
func New(quarantine time.Duration) *Tracker {
	return &Tracker{
		entries:    make(map[string]Entry),
		quarantine: quarantine,
		now:        time.Now,
	}
}

// Evict records an eviction for leaseID with the supplied reason.
func (t *Tracker) Evict(leaseID, reason string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	t.entries[leaseID] = Entry{
		LeaseID:   leaseID,
		Reason:    reason,
		EvictedAt: now,
		ExpiresAt: now.Add(t.quarantine),
	}
}

// IsEvicted returns true when leaseID is within its quarantine window.
func (t *Tracker) IsEvicted(leaseID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[leaseID]
	if !ok {
		return false
	}
	return t.now().Before(e.ExpiresAt)
}

// Get returns the eviction entry for leaseID and whether it was found.
func (t *Tracker) Get(leaseID string) (Entry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	e, ok := t.entries[leaseID]
	return e, ok
}

// Check returns ErrEvicted if leaseID is currently quarantined, nil otherwise.
func (t *Tracker) Check(leaseID string) error {
	if t.IsEvicted(leaseID) {
		return ErrEvicted
	}
	return nil
}

// Purge removes all expired eviction entries.
func (t *Tracker) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	for id, e := range t.entries {
		if now.After(e.ExpiresAt) {
			delete(t.entries, id)
		}
	}
}
