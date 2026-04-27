// Package cooldown enforces per-lease cooldown periods between successive
// alert or renewal actions, preventing thundering-herd bursts.
package cooldown

import (
	"sync"
	"time"
)

// Entry records the last action time for a single key.
type Entry struct {
	LastAction time.Time
	Count      int
}

// Tracker maintains cooldown state keyed by lease ID.
type Tracker struct {
	mu       sync.Mutex
	entries  map[string]Entry
	cooldown time.Duration
	now      func() time.Time
}

// New returns a Tracker that enforces the given cooldown duration.
func New(d time.Duration) *Tracker {
	return &Tracker{
		entries:  make(map[string]Entry),
		cooldown: d,
		now:      time.Now,
	}
}

// Allow reports whether the key is outside its cooldown window.
// If allowed, the entry is updated immediately.
func (t *Tracker) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	e, ok := t.entries[key]
	if ok && now.Sub(e.LastAction) < t.cooldown {
		return false
	}
	t.entries[key] = Entry{LastAction: now, Count: e.Count + 1}
	return true
}

// Get returns the current entry for a key, and whether it exists.
func (t *Tracker) Get(key string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[key]
	return e, ok
}

// Reset removes the cooldown entry for a key, allowing immediate action.
func (t *Tracker) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.entries, key)
}

// Purge removes all entries whose cooldown window has fully elapsed.
func (t *Tracker) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.now()
	for k, e := range t.entries {
		if now.Sub(e.LastAction) >= t.cooldown {
			delete(t.entries, k)
		}
	}
}
