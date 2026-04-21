// Package baseline tracks the expected TTL range for a lease path and
// detects anomalous TTL values that deviate significantly from the norm.
package baseline

import (
	"fmt"
	"sync"
	"time"
)

// Entry holds the running statistics for a single lease path.
type Entry struct {
	Count  int
	Mean   float64
	M2     float64 // variance accumulator (Welford's algorithm)
	Min    float64
	Max    float64
	Update time.Time
}

// StdDev returns the population standard deviation of observed TTL values.
func (e *Entry) StdDev() float64 {
	if e.Count < 2 {
		return 0
	}
	v := e.M2 / float64(e.Count)
	if v < 0 {
		v = 0
	}
	// integer sqrt approximation via Newton–Raphson
	x := v
	for i := 0; i < 64; i++ {
		if x == 0 {
			break
		}
		x = (x + v/x) / 2
	}
	return x
}

// Tracker accumulates TTL observations per lease path and flags anomalies.
type Tracker struct {
	mu        sync.Mutex
	entries   map[string]*Entry
	threshold float64 // number of std-devs considered anomalous
}

// New returns a Tracker with the given anomaly threshold (e.g. 2.0 for 2σ).
func New(threshold float64) *Tracker {
	if threshold <= 0 {
		threshold = 2.0
	}
	return &Tracker{
		entries:   make(map[string]*Entry),
		threshold: threshold,
	}
}

// Record adds a TTL observation for path and returns an error if the value is
// anomalous relative to the established baseline.
func (t *Tracker) Record(path string, ttl time.Duration) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	v := ttl.Seconds()
	e, ok := t.entries[path]
	if !ok {
		e = &Entry{Min: v, Max: v}
		t.entries[path] = e
	}

	e.Count++
	delta := v - e.Mean
	e.Mean += delta / float64(e.Count)
	e.M2 += delta * (v - e.Mean)
	if v < e.Min {
		e.Min = v
	}
	if v > e.Max {
		e.Max = v
	}
	e.Update = time.Now()

	if e.Count >= 5 {
		sd := e.StdDev()
		if sd > 0 {
			diff := v - e.Mean
			if diff < 0 {
				diff = -diff
			}
			if diff > t.threshold*sd {
				return fmt.Errorf("baseline: anomalous TTL %.0fs for path %q (mean=%.0f sd=%.0f)", v, path, e.Mean, sd)
			}
		}
	}
	return nil
}

// Get returns a copy of the baseline entry for path, and whether it exists.
func (t *Tracker) Get(path string) (Entry, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.entries[path]
	if !ok {
		return Entry{}, false
	}
	return *e, true
}

// Purge removes entries that have not been updated since cutoff.
func (t *Tracker) Purge(cutoff time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, e := range t.entries {
		if e.Update.Before(cutoff) {
			delete(t.entries, k)
		}
	}
}
