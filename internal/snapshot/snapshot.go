// Package snapshot captures and compares lease state across polls.
package snapshot

import (
	"sync"
	"time"
)

// Entry holds a point-in-time record of a lease.
type Entry struct {
	LeaseID   string
	TTL       time.Duration
	RecordedAt time.Time
}

// Store holds the most recent snapshot of all known leases.
type Store struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an initialised Store.
func New() *Store {
	return &Store{entries: make(map[string]Entry)}
}

// Record upserts an entry for the given lease ID.
func (s *Store) Record(leaseID string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[leaseID] = Entry{
		LeaseID:    leaseID,
		TTL:        ttl,
		RecordedAt: time.Now(),
	}
}

// Get returns the stored entry for a lease ID and whether it existed.
func (s *Store) Get(leaseID string) (Entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[leaseID]
	return e, ok
}

// Delete removes a lease entry from the store.
func (s *Store) Delete(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, leaseID)
}

// All returns a copy of all current entries.
func (s *Store) All() []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}

// TTLDecreased reports whether the TTL for a lease has decreased since last
// recorded, which can indicate an unrenewed credential drifting toward expiry.
func (s *Store) TTLDecreased(leaseID string, current time.Duration) bool {
	e, ok := s.Get(leaseID)
	if !ok {
		return false
	}
	return current < e.TTL
}

// Len returns the number of entries currently held in the store.
func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}

// Expiring returns all entries whose recorded TTL is below the given threshold.
// This is useful for surfacing leases that are approaching expiry without
// having been renewed.
func (s *Store) Expiring(threshold time.Duration) []Entry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Entry
	for _, e := range s.entries {
		if e.TTL < threshold {
			out = append(out, e)
		}
	}
	return out
}
