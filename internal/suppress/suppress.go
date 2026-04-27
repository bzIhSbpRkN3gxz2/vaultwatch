// Package suppress provides time-windowed suppression of repeated lease alerts.
// Once a lease ID has been suppressed, further alerts are dropped until the
// suppression window expires or the entry is explicitly released.
package suppress

import (
	"sync"
	"time"
)

// Entry records when a lease was first suppressed.
type Entry struct {
	LeaseID     string
	SuppressedAt time.Time
	ExpiresAt   time.Time
}

// Suppressor tracks suppressed lease IDs within a configurable window.
type Suppressor struct {
	mu      sync.Mutex
	entries map[string]Entry
	window  time.Duration
	now     func() time.Time
}

// New returns a Suppressor with the given suppression window.
func New(window time.Duration) *Suppressor {
	return &Suppressor{
		entries: make(map[string]Entry),
		window:  window,
		now:     time.Now,
	}
}

// Suppress marks leaseID as suppressed. Returns true if the entry was newly
// added, false if it was already suppressed.
func (s *Suppressor) Suppress(leaseID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if e, ok := s.entries[leaseID]; ok && now.Before(e.ExpiresAt) {
		return false
	}
	s.entries[leaseID] = Entry{
		LeaseID:      leaseID,
		SuppressedAt: now,
		ExpiresAt:    now.Add(s.window),
	}
	return true
}

// IsSuppressed reports whether leaseID is currently suppressed.
func (s *Suppressor) IsSuppressed(leaseID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[leaseID]
	if !ok {
		return false
	}
	return s.now().Before(e.ExpiresAt)
}

// Release removes the suppression for leaseID immediately.
func (s *Suppressor) Release(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, leaseID)
}

// Purge removes all expired suppression entries.
func (s *Suppressor) Purge() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for id, e := range s.entries {
		if !now.Before(e.ExpiresAt) {
			delete(s.entries, id)
		}
	}
}

// Len returns the number of active suppression entries.
func (s *Suppressor) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.entries)
}
