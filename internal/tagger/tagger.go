// Package tagger provides label-based tagging for Vault leases,
// allowing leases to be annotated with arbitrary key-value metadata.
package tagger

import (
	"fmt"
	"strings"
	"sync"
)

// Tags is a map of key-value string pairs attached to a lease.
type Tags map[string]string

// Store holds tags indexed by lease ID.
type Store struct {
	mu   sync.RWMutex
	data map[string]Tags
}

// New returns an initialised tag Store.
func New() *Store {
	return &Store{data: make(map[string]Tags)}
}

// Set adds or replaces a tag on the given lease.
func (s *Store) Set(leaseID, key, value string) error {
	if leaseID == "" {
		return fmt.Errorf("tagger: leaseID must not be empty")
	}
	if key == "" {
		return fmt.Errorf("tagger: key must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[leaseID] == nil {
		s.data[leaseID] = make(Tags)
	}
	s.data[leaseID][key] = value
	return nil
}

// Get returns the tags for a lease. The returned map is a copy.
func (s *Store) Get(leaseID string) Tags {
	s.mu.RLock()
	defer s.mu.RUnlock()
	src := s.data[leaseID]
	out := make(Tags, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// Delete removes all tags for a lease.
func (s *Store) Delete(leaseID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, leaseID)
}

// Match returns all lease IDs whose tags contain all of the supplied
// key=value pairs.
func (s *Store) Match(query Tags) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var ids []string
outer:
	for id, tags := range s.data {
		for k, v := range query {
			if tags[k] != v {
				continue outer
			}
		}
		ids = append(ids, id)
	}
	return ids
}

// String returns a human-readable representation of a Tags map.
func (t Tags) String() string {
	parts := make([]string, 0, len(t))
	for k, v := range t {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}
