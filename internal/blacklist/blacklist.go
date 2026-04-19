// Package blacklist tracks lease IDs that should never be renewed or alerted on.
package blacklist

import (
	"sync"
	"time"
)

// Entry holds metadata about a blacklisted lease.
type Entry struct {
	LeaseID   string
	Reason    string
	AddedAt   time.Time
	ExpiresAt time.Time // zero means permanent
}

// Blacklist stores lease IDs that should be suppressed.
type Blacklist struct {
	mu      sync.RWMutex
	entries map[string]Entry
}

// New returns an empty Blacklist.
func New() *Blacklist {
	return &Blacklist{entries: make(map[string]Entry)}
}

// Add adds a lease ID with a reason and optional TTL (0 = permanent).
func (b *Blacklist) Add(leaseID, reason string, ttl time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	b.entries[leaseID] = Entry{
		LeaseID:   leaseID,
		Reason:    reason,
		AddedAt:   time.Now(),
		ExpiresAt: exp,
	}
}

// Contains returns true if the lease ID is blacklisted and not expired.
func (b *Blacklist) Contains(leaseID string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	e, ok := b.entries[leaseID]
	if !ok {
		return false
	}
	if !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt) {
		return false
	}
	return true
}

// Remove deletes a lease ID from the blacklist.
func (b *Blacklist) Remove(leaseID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, leaseID)
}

// Purge removes all expired entries.
func (b *Blacklist) Purge() {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	for id, e := range b.entries {
		if !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt) {
			delete(b.entries, id)
		}
	}
}

// All returns a copy of all active entries.
func (b *Blacklist) All() []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]Entry, 0, len(b.entries))
	now := time.Now()
	for _, e := range b.entries {
		if e.ExpiresAt.IsZero() || now.Before(e.ExpiresAt) {
			out = append(out, e)
		}
	}
	return out
}
