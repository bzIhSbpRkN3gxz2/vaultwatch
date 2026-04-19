// Package history tracks lease status transitions over time.
package history

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Event records a single status transition for a lease.
type Event struct {
	LeaseID   string
	From      lease.Status
	To        lease.Status
	RecordedAt time.Time
}

// Tracker maintains a history of status transitions per lease.
type Tracker struct {
	mu     sync.Mutex
	events map[string][]Event
	max    int
}

// New returns a Tracker that keeps at most maxPerLease events per lease ID.
func New(maxPerLease int) *Tracker {
	if maxPerLease <= 0 {
		maxPerLease = 50
	}
	return &Tracker{
		events: make(map[string][]Event),
		max:    maxPerLease,
	}
}

// Record appends a transition event for the given lease.
func (t *Tracker) Record(leaseID string, from, to lease.Status) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ev := Event{
		LeaseID:    leaseID,
		From:       from,
		To:         to,
		RecordedAt: time.Now(),
	}
	list := append(t.events[leaseID], ev)
	if len(list) > t.max {
		list = list[len(list)-t.max:]
	}
	t.events[leaseID] = list
}

// Get returns all recorded events for a lease ID.
func (t *Tracker) Get(leaseID string) []Event {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Event, len(t.events[leaseID]))
	copy(out, t.events[leaseID])
	return out
}

// Purge removes all history for a lease ID.
func (t *Tracker) Purge(leaseID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.events, leaseID)
}
