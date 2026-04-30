// Package digest computes and tracks rolling digests of lease state,
// enabling change detection across poll cycles.
package digest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Entry holds the last recorded digest for a lease.
type Entry struct {
	Digest    string
	RecordedAt time.Time
}

// Tracker computes and stores digests for leases, reporting
// whether the lease state has changed since the last observation.
type Tracker struct {
	mu      sync.Mutex
	entries map[string]Entry
}

// New returns an initialised Tracker.
func New() *Tracker {
	return &Tracker{
		entries: make(map[string]Entry),
	}
}

// Compute returns a deterministic hex digest for the given lease,
// incorporating its ID, path, status and TTL.
func Compute(l *lease.Lease) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%s|%d", l.ID, l.Path, l.Status, l.TTL)
	return hex.EncodeToString(h.Sum(nil))
}

// Changed returns true if the lease digest differs from the previously
// stored value. The new digest is always persisted after the call.
func (t *Tracker) Changed(l *lease.Lease) bool {
	if l == nil {
		return false
	}
	current := Compute(l)

	t.mu.Lock()
	defer t.mu.Unlock()

	prev, ok := t.entries[l.ID]
	t.entries[l.ID] = Entry{Digest: current, RecordedAt: time.Now()}

	if !ok {
		return true
	}
	return prev.Digest != current
}

// Get returns the stored Entry for a lease ID, and whether it exists.
func (t *Tracker) Get(leaseID string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[leaseID]
	return e, ok
}

// Purge removes entries that have not been updated since the cutoff.
func (t *Tracker) Purge(cutoff time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for id, e := range t.entries {
		if e.RecordedAt.Before(cutoff) {
			delete(t.entries, id)
		}
	}
}
