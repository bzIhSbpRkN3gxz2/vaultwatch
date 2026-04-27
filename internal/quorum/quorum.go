// Package quorum implements a voting mechanism that requires a minimum number
// of agreeing checks before an alert or action is triggered for a lease.
package quorum

import (
	"errors"
	"sync"
	"time"
)

// ErrQuorumNotMet is returned when votes have not reached the required threshold.
var ErrQuorumNotMet = errors.New("quorum: threshold not met")

// vote records a single cast vote and when it was cast.
type vote struct {
	agree     bool
	castAt   time.Time
}

// entry holds all votes for a single lease key.
type entry struct {
	votes []vote
}

// Quorum tracks votes per lease key and reports whether a threshold is met.
type Quorum struct {
	mu        sync.Mutex
	entries   map[string]*entry
	threshold int
	window    time.Duration
	now       func() time.Time
}

// New creates a Quorum that requires at least threshold agreeing votes
// within the given window duration.
func New(threshold int, window time.Duration) *Quorum {
	if threshold < 1 {
		threshold = 1
	}
	return &Quorum{
		entries:   make(map[string]*entry),
		threshold: threshold,
		window:    window,
		now:       time.Now,
	}
}

// Vote records a vote for the given key. agree=true counts toward quorum.
func (q *Quorum) Vote(key string, agree bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	e, ok := q.entries[key]
	if !ok {
		e = &entry{}
		q.entries[key] = e
	}
	e.votes = append(e.votes, vote{agree: agree, castAt: q.now()})
}

// Reached reports whether the key has met quorum within the active window.
// Returns ErrQuorumNotMet if the threshold has not been reached.
func (q *Quorum) Reached(key string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	e, ok := q.entries[key]
	if !ok {
		return ErrQuorumNotMet
	}
	cutoff := q.now().Add(-q.window)
	count := 0
	for _, v := range e.votes {
		if v.agree && v.castAt.After(cutoff) {
			count++
		}
	}
	if count >= q.threshold {
		return nil
	}
	return ErrQuorumNotMet
}

// Reset clears all votes for the given key.
func (q *Quorum) Reset(key string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	delete(q.entries, key)
}

// Purge removes all entries whose most recent vote is older than the window.
func (q *Quorum) Purge() {
	q.mu.Lock()
	defer q.mu.Unlock()
	cutoff := q.now().Add(-q.window)
	for key, e := range q.entries {
		if len(e.votes) == 0 {
			delete(q.entries, key)
			continue
		}
		last := e.votes[len(e.votes)-1]
		if last.castAt.Before(cutoff) {
			delete(q.entries, key)
		}
	}
}
