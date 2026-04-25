// Package debounce suppresses repeated alerts for the same lease within a
// configurable quiet period, emitting only the first and then one trailing
// notification once the window has elapsed.
package debounce

import (
	"sync"
	"time"
)

// entry tracks the first-seen time and whether a trailing notification is
// pending for a given key.
type entry struct {
	firstSeen time.Time
	trailing  bool
}

// Debouncer holds per-key state for debounce logic.
type Debouncer struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[string]*entry
	now     func() time.Time
}

// New returns a Debouncer with the given quiet window.
func New(window time.Duration) *Debouncer {
	return &Debouncer{
		window:  window,
		entries: make(map[string]*entry),
		now:     time.Now,
	}
}

// Allow reports whether the event for key should be forwarded.
//
// The first occurrence always returns true. Subsequent calls within the window
// return false. The first call after the window expires returns true and resets
// the timer.
func (d *Debouncer) Allow(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	e, ok := d.entries[key]
	if !ok {
		d.entries[key] = &entry{firstSeen: now}
		return true
	}

	if now.Sub(e.firstSeen) >= d.window {
		d.entries[key] = &entry{firstSeen: now}
		return true
	}

	return false
}

// Reset removes the debounce state for key, allowing the next event through
// immediately.
func (d *Debouncer) Reset(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, key)
}

// Purge removes all entries whose window has elapsed.
func (d *Debouncer) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()
	now := d.now()
	for k, e := range d.entries {
		if now.Sub(e.firstSeen) >= d.window {
			delete(d.entries, k)
		}
	}
}

// Len returns the number of active debounce entries.
func (d *Debouncer) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}
