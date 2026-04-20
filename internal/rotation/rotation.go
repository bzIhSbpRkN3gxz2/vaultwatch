// Package rotation provides secret rotation tracking and coordination
// for Vault leases managed by vaultwatch.
package rotation

import (
	"errors"
	"sync"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// ErrAlreadyRotating is returned when a rotation is already in progress for a lease.
var ErrAlreadyRotating = errors.New("rotation already in progress")

// Status represents the current rotation state of a lease.
type Status int

const (
	StatusIdle     Status = iota // No rotation in progress.
	StatusPending               // Rotation requested, not yet started.
	StatusActive                // Rotation is actively running.
	StatusComplete              // Rotation finished successfully.
	StatusFailed                // Rotation failed.
)

// Record holds rotation state for a single lease.
type Record struct {
	LeaseID   string
	Status    Status
	StartedAt time.Time
	EndedAt   time.Time
	Error     string
}

// Tracker maintains rotation state for leases.
type Tracker struct {
	mu      sync.Mutex
	records map[string]*Record
}

// New returns a new Tracker.
func New() *Tracker {
	return &Tracker{records: make(map[string]*Record)}
}

// Begin marks a lease rotation as active. Returns ErrAlreadyRotating if one is
// already in progress.
func (t *Tracker) Begin(l *lease.Lease) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if r, ok := t.records[l.LeaseID]; ok && r.Status == StatusActive {
		return ErrAlreadyRotating
	}
	t.records[l.LeaseID] = &Record{
		LeaseID:   l.LeaseID,
		Status:    StatusActive,
		StartedAt: time.Now(),
	}
	return nil
}

// Complete marks the rotation for leaseID as successfully completed.
func (t *Tracker) Complete(leaseID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if r, ok := t.records[leaseID]; ok {
		r.Status = StatusComplete
		r.EndedAt = time.Now()
	}
}

// Fail marks the rotation for leaseID as failed with the given error message.
func (t *Tracker) Fail(leaseID, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if r, ok := t.records[leaseID]; ok {
		r.Status = StatusFailed
		r.EndedAt = time.Now()
		r.Error = errMsg
	}
}

// Get returns the rotation Record for leaseID and whether it exists.
func (t *Tracker) Get(leaseID string) (Record, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	r, ok := t.records[leaseID]
	if !ok {
		return Record{}, false
	}
	return *r, true
}

// Purge removes all records whose rotation ended before the given cutoff.
func (t *Tracker) Purge(before time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id, r := range t.records {
		if (r.Status == StatusComplete || r.Status == StatusFailed) && r.EndedAt.Before(before) {
			delete(t.records, id)
		}
	}
}
