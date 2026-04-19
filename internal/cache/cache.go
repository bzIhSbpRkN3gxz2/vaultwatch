// Package cache provides a simple TTL-based in-memory cache for lease metadata.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached value and its expiry time.
type Entry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache is a thread-safe TTL cache.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	default TTL time.Duration
}

// New creates a Cache with the given default TTL.
func New(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]Entry),
		defaultTTL: defaultTTL,
	}
}

// Set stores a value under key with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetTTL(key, value, c.defaultTTL)
}

// SetTTL stores a value under key with a custom TTL.
func (c *Cache) SetTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = Entry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get retrieves a value by key. Returns (value, true) if found and not expired.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.items[key]
	if !ok || time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	return entry.Value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Purge removes all expired entries from the cache.
func (c *Cache) Purge() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	removed := 0
	for k, e := range c.items {
		if now.After(e.ExpiresAt) {
			delete(c.items, k)
			removed++
		}
	}
	return removed
}

// Len returns the number of entries (including expired but not yet purged).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
