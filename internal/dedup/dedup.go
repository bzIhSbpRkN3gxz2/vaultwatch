// Package dedup provides lease deduplication to prevent redundant alert
// processing when the same lease appears multiple times in a poll cycle.
package dedup

import (
	"sync"
	"time"
)

// Entry holds the last-seen fingerprint and timestamp for a lease.
type Entry struct {
	Fingerprint string
	SeenAt      time.Time
}

// Deduplicator tracks lease fingerprints and suppresses duplicates within a
// configurable window.
type Deduplicator struct {
	mu      sync.Mutex
	window  time.Duration
	entries map[string]Entry
}

// New returns a Deduplicator with the given deduplication window.
func New(window time.Duration) *Deduplicator {
	return &Deduplicator{
		window:  window,
		entries: make(map[string]Entry),
	}
}

// IsDuplicate reports whether leaseID with fingerprint has been seen within
// the deduplication window. If it has not, the entry is recorded and false is
// returned.
func (d *Deduplicator) IsDuplicate(leaseID, fingerprint string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if e, ok := d.entries[leaseID]; ok {
		if e.Fingerprint == fingerprint && time.Since(e.SeenAt) < d.window {
			return true
		}
	}
	d.entries[leaseID] = Entry{Fingerprint: fingerprint, SeenAt: time.Now()}
	return false
}

// Purge removes all entries older than the deduplication window.
func (d *Deduplicator) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, e := range d.entries {
		if time.Since(e.SeenAt) >= d.window {
			delete(d.entries, id)
		}
	}
}

// Reset clears all tracked entries.
func (d *Deduplicator) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = make(map[string]Entry)
}
