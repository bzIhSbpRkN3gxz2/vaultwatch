// Package staleness tracks how long leases have remained in a given status
// and emits an alert when a lease has been stale beyond a configured threshold.
package staleness

import (
	"errors"
	"sync"
	"time"

	"github.com/youorg/vaultwatch/internal/lease"
)

// ErrThresholdExceeded is returned when a lease has been stale too long.
var ErrThresholdExceeded = errors.New("staleness threshold exceeded")

// Entry records when a lease first entered its current status.
type Entry struct {
	Status    lease.Status
	Since     time.Time
	LeaseID   string
}

// Tracker monitors how long leases stay in a particular status.
type Tracker struct {
	mu        sync.Mutex
	entries   map[string]Entry
	threshold time.Duration
	now       func() time.Time
}

// New returns a Tracker that fires ErrThresholdExceeded when a lease
// remains in the same status longer than threshold.
func New(threshold time.Duration) *Tracker {
	return &Tracker{
		entries:   make(map[string]Entry),
		threshold: threshold,
		now:       time.Now,
	}
}

// Observe records the current status of a lease. If the status has changed
// since the last observation the timer resets. Returns ErrThresholdExceeded
// when the lease has been in the same status longer than the threshold.
func (t *Tracker) Observe(l *lease.Lease) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	existing, ok := t.entries[l.LeaseID]

	if !ok || existing.Status != l.Status {
		t.entries[l.LeaseID] = Entry{
			Status:  l.Status,
			Since:   now,
			LeaseID: l.LeaseID,
		}
		return nil
	}

	if now.Sub(existing.Since) > t.threshold {
		return ErrThresholdExceeded
	}
	return nil
}

// Get returns the current staleness entry for a lease, if any.
func (t *Tracker) Get(leaseID string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[leaseID]
	return e, ok
}

// Delete removes the staleness record for a lease.
func (t *Tracker) Delete(leaseID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, leaseID)
}

// Purge removes all entries whose lease IDs are not present in active.
func (t *Tracker) Purge(active map[string]struct{}) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id := range t.entries {
		if _, ok := active[id]; !ok {
			delete(t.entries, id)
		}
	}
}
