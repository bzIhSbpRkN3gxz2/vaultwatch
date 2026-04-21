package tokenstore

import (
	"errors"
	"sync"
	"time"
)

// ErrNotFound is returned when a token is not present in the store.
var ErrNotFound = errors.New("tokenstore: token not found")

// ErrExpired is returned when a token exists but has passed its TTL.
var ErrExpired = errors.New("tokenstore: token expired")

// Entry holds a Vault token and its expiry metadata.
type Entry struct {
	LeaseID   string
	Token     string
	ExpiresAt time.Time
	Meta      map[string]string
}

// Expired reports whether the entry has passed its expiry time.
func (e *Entry) Expired(now time.Time) bool {
	return now.After(e.ExpiresAt)
}

// Store is a thread-safe in-memory store for Vault tokens keyed by lease ID.
type Store struct {
	mu      sync.RWMutex
	entries map[string]*Entry
	now     func() time.Time
}

// New returns an initialised Store.
func New() *Store {
	return &Store{
		entries: make(map[string]*Entry),
		now:     time.Now,
	}
}

// Set stores or replaces a token entry.
func (s *Store) Set(e *Entry) error {
	if e == nil || e.LeaseID == "" {
		return errors.New("tokenstore: entry must have a non-empty LeaseID")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[e.LeaseID] = e
	return nil
}

// Get retrieves a token entry by lease ID. Returns ErrNotFound or ErrExpired
// when appropriate.
func (s *Store) Get(leaseID string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[leaseID]
	if !ok {
		return nil, ErrNotFound
	}
	if e.Expired(s.now()) {
		return nil, ErrExpired
	}
	return e, nil
}

// Delete removes a token entry from the store.
func (s *Store) Delete(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, leaseID)
}

// Purge removes all entries whose TTL has elapsed.
func (s *Store) Purge() int {
	now := s.now()
	s.mu.Lock()
	defer s.mu.Unlock()
	removed := 0
	for id, e := range s.entries {
		if e.Expired(now) {
			delete(s.entries, id)
			removed++
		}
	}
	return removed
}

// All returns a snapshot of all non-expired entries.
func (s *Store) All() []*Entry {
	now := s.now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		if !e.Expired(now) {
			out = append(out, e)
		}
	}
	return out
}
