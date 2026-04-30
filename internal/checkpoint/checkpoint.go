// Package checkpoint tracks the last-seen state of a lease poll cycle,
// allowing vaultwatch to resume monitoring without re-alerting on already-
// processed leases after a restart.
package checkpoint

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Entry holds a persisted snapshot of a single lease at a point in time.
type Entry struct {
	LeaseID   string    `json:"lease_id"`
	Path      string    `json:"path"`
	Status    string    `json:"status"`
	TTL       int64     `json:"ttl"`
	RecordedAt time.Time `json:"recorded_at"`
}

// Checkpoint maintains an in-memory map of lease checkpoints and can persist
// them to / restore them from a JSON file.
type Checkpoint struct {
	mu      sync.RWMutex
	entries map[string]Entry
	path    string
}

// New creates a Checkpoint backed by the given file path.
// If the file exists its contents are loaded immediately.
func New(path string) (*Checkpoint, error) {
	cp := &Checkpoint{
		entries: make(map[string]Entry),
		path:    path,
	}
	if err := cp.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return cp, nil
}

// Set records or updates the checkpoint entry for a lease.
func (c *Checkpoint) Set(e Entry) {
	if e.LeaseID == "" {
		return
	}
	e.RecordedAt = time.Now().UTC()
	c.mu.Lock()
	c.entries[e.LeaseID] = e
	c.mu.Unlock()
}

// Get returns the checkpoint entry for the given lease ID.
func (c *Checkpoint) Get(leaseID string) (Entry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[leaseID]
	return e, ok
}

// Delete removes the checkpoint entry for the given lease ID.
func (c *Checkpoint) Delete(leaseID string) {
	c.mu.Lock()
	delete(c.entries, leaseID)
	c.mu.Unlock()
}

// All returns a copy of all current entries.
func (c *Checkpoint) All() []Entry {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Entry, 0, len(c.entries))
	for _, e := range c.entries {
		out = append(out, e)
	}
	return out
}

// Save persists the current state to the configured file path.
func (c *Checkpoint) Save() error {
	c.mu.RLock()
	data, err := json.Marshal(c.entries)
	c.mu.RUnlock()
	if err != nil {
		return err
	}
	return os.WriteFile(c.path, data, 0o600)
}

func (c *Checkpoint) load() error {
	data, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return json.Unmarshal(data, &c.entries)
}
