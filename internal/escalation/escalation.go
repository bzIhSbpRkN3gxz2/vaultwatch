// Package escalation provides tiered alert escalation based on lease severity and repeat counts.
package escalation

import (
	"sync"
	"time"
)

// Level represents an escalation tier.
type Level int

const (
	LevelNone  Level = iota
	LevelWarn        // first occurrence
	LevelCritical    // repeated within window
	LevelPage        // exceeded critical threshold
)

// Policy defines thresholds for escalation.
type Policy struct {
	WarnAfter     int
	CriticalAfter int
	PageAfter     int
	Window        time.Duration
}

// DefaultPolicy returns sensible escalation defaults.
func DefaultPolicy() Policy {
	return Policy{
		WarnAfter:     1,
		CriticalAfter: 3,
		PageAfter:     6,
		Window:        30 * time.Minute,
	}
}

type entry struct {
	count   int
	firstAt time.Time
}

// Escalator tracks alert counts per lease and returns the appropriate level.
type Escalator struct {
	mu     sync.Mutex
	policy Policy
	state  map[string]*entry
}

// New creates a new Escalator with the given policy.
func New(p Policy) *Escalator {
	return &Escalator{
		policy: p,
		state:  make(map[string]*entry),
	}
}

// Evaluate increments the alert count for leaseID and returns the current escalation level.
func (e *Escalator) Evaluate(leaseID string) Level {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()
	ent, ok := e.state[leaseID]
	if !ok || now.Sub(ent.firstAt) > e.policy.Window {
		e.state[leaseID] = &entry{count: 1, firstAt: now}
		ent = e.state[leaseID]
	} else {
		ent.count++
	}

	switch {
	case ent.count >= e.policy.PageAfter:
		return LevelPage
	case ent.count >= e.policy.CriticalAfter:
		return LevelCritical
	case ent.count >= e.policy.WarnAfter:
		return LevelWarn
	default:
		return LevelNone
	}
}

// Reset clears the escalation state for a lease.
func (e *Escalator) Reset(leaseID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.state, leaseID)
}

// Purge removes entries whose window has expired.
func (e *Escalator) Purge() {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now()
	for id, ent := range e.state {
		if now.Sub(ent.firstAt) > e.policy.Window {
			delete(e.state, id)
		}
	}
}
