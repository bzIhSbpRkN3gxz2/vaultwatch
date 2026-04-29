// Package drift detects when a lease's TTL has drifted significantly
// from its expected renewal baseline, indicating potential clock skew
// or missed renewal cycles.
package drift

import (
	"errors"
	"sync"
	"time"

	"github.com/your-org/vaultwatch/internal/lease"
)

// ErrDriftDetected is returned when a lease TTL deviates beyond the threshold.
var ErrDriftDetected = errors.New("drift: TTL drift exceeds threshold")

// Entry records the expected TTL at a point in time.
type Entry struct {
	ExpectedTTL time.Duration
	RecordedAt  time.Time
}

// Detector tracks expected TTL values and flags anomalous deviations.
type Detector struct {
	mu        sync.Mutex
	entries   map[string]Entry
	threshold float64 // fraction, e.g. 0.25 means 25% drift triggers alert
}

// New creates a Detector with the given drift threshold fraction (0 < threshold <= 1).
func New(threshold float64) *Detector {
	if threshold <= 0 || threshold > 1 {
		threshold = 0.25
	}
	return &Detector{
		entries:   make(map[string]Entry),
		threshold: threshold,
	}
}

// Observe records or evaluates the TTL for the given lease.
// On first observation the baseline is set. On subsequent calls,
// if the observed TTL deviates from the expected value by more than
// the threshold fraction, ErrDriftDetected is returned.
func (d *Detector) Observe(l *lease.Lease) error {
	if l == nil || l.LeaseID == "" {
		return nil
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	entry, exists := d.entries[l.LeaseID]

	if !exists {
		d.entries[l.LeaseID] = Entry{
			ExpectedTTL: l.TTL,
			RecordedAt:  now,
		}
		return nil
	}

	// Adjust expected TTL by elapsed time since last observation.
	elapsed := now.Sub(entry.RecordedAt)
	adjusted := entry.ExpectedTTL - elapsed
	if adjusted < 0 {
		adjusted = 0
	}

	if adjusted == 0 {
		d.entries[l.LeaseID] = Entry{ExpectedTTL: l.TTL, RecordedAt: now}
		return nil
	}

	diff := l.TTL - adjusted
	if diff < 0 {
		diff = -diff
	}

	if float64(diff)/float64(adjusted) > d.threshold {
		// Update baseline after detection so subsequent calls re-evaluate.
		d.entries[l.LeaseID] = Entry{ExpectedTTL: l.TTL, RecordedAt: now}
		return ErrDriftDetected
	}

	d.entries[l.LeaseID] = Entry{ExpectedTTL: l.TTL, RecordedAt: now}
	return nil
}

// Reset clears the recorded baseline for a lease.
func (d *Detector) Reset(leaseID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, leaseID)
}

// Purge removes all entries older than the given age.
func (d *Detector) Purge(maxAge time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	cutoff := time.Now().Add(-maxAge)
	for id, e := range d.entries {
		if e.RecordedAt.Before(cutoff) {
			delete(d.entries, id)
		}
	}
}
