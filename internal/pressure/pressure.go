// Package pressure tracks lease renewal pressure across Vault paths,
// providing a score that reflects how many leases are near expiry relative
// to the total observed. High pressure indicates the system is under strain.
package pressure

import (
	"sync"
	"time"
)

// Score represents a pressure reading for a given path prefix.
type Score struct {
	Path      string
	Total     int
	Critical  int
	Warning   int
	Value     float64 // 0.0–1.0
	RecordedAt time.Time
}

// Tracker computes and stores pressure scores per path.
type Tracker struct {
	mu     sync.RWMutex
	scores map[string]*Score
	clock  func() time.Time
}

// New returns a new Tracker.
func New() *Tracker {
	return &Tracker{
		scores: make(map[string]*Score),
		clock:  time.Now,
	}
}

// Record updates the pressure score for the given path.
// total is the number of leases observed; critical and warning are subsets.
func (t *Tracker) Record(path string, total, critical, warning int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	var value float64
	if total > 0 {
		value = float64(critical*2+warning) / float64(total*2)
		if value > 1.0 {
			value = 1.0
		}
	}

	t.scores[path] = &Score{
		Path:       path,
		Total:      total,
		Critical:   critical,
		Warning:    warning,
		Value:      value,
		RecordedAt: t.clock(),
	}
}

// Get returns the current pressure score for a path, or false if not found.
func (t *Tracker) Get(path string) (Score, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.scores[path]
	if !ok {
		return Score{}, false
	}
	return *s, true
}

// All returns a snapshot of all current pressure scores.
func (t *Tracker) All() []Score {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Score, 0, len(t.scores))
	for _, s := range t.scores {
		out = append(out, *s)
	}
	return out
}

// Purge removes scores older than the given age.
func (t *Tracker) Purge(maxAge time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := t.clock()
	for k, s := range t.scores {
		if now.Sub(s.RecordedAt) > maxAge {
			delete(t.scores, k)
		}
	}
}
