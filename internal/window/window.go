// Package window provides a sliding time-window counter for tracking
// lease events (e.g. expirations, renewals) over a rolling duration.
package window

import (
	"sync"
	"time"
)

// Entry records a single event timestamp.
type Entry struct {
	OccurredAt time.Time
}

// Window is a thread-safe sliding time-window counter keyed by lease ID.
type Window struct {
	mu       sync.Mutex
	size     time.Duration
	entries  map[string][]Entry
	nowFn    func() time.Time
}

// New creates a Window that retains events within the given duration.
func New(size time.Duration) *Window {
	return &Window{
		size:    size,
		entries: make(map[string][]Entry),
		nowFn:   time.Now,
	}
}

// Record adds an event for the given lease ID at the current time.
func (w *Window) Record(leaseID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.nowFn()
	w.entries[leaseID] = append(w.entries[leaseID], Entry{OccurredAt: now})
	w.evict(leaseID, now)
}

// Count returns the number of events for leaseID within the window.
func (w *Window) Count(leaseID string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.evict(leaseID, w.nowFn())
	return len(w.entries[leaseID])
}

// Reset removes all recorded events for the given lease ID.
func (w *Window) Reset(leaseID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.entries, leaseID)
}

// Purge removes all entries whose entire event list has expired.
func (w *Window) Purge() {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.nowFn()
	for id := range w.entries {
		w.evict(id, now)
		if len(w.entries[id]) == 0 {
			delete(w.entries, id)
		}
	}
}

// evict removes events outside the window. Caller must hold w.mu.
func (w *Window) evict(leaseID string, now time.Time) {
	cutoff := now.Add(-w.size)
	evts := w.entries[leaseID]
	i := 0
	for i < len(evts) && evts[i].OccurredAt.Before(cutoff) {
		i++
	}
	w.entries[leaseID] = evts[i:]
}
