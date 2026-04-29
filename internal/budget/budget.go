// Package budget tracks renewal and alert budget consumption per lease path,
// preventing runaway operations when many leases expire simultaneously.
package budget

import (
	"errors"
	"sync"
	"time"
)

// ErrBudgetExceeded is returned when the operation budget for a path is exhausted.
var ErrBudgetExceeded = errors.New("budget: operation budget exceeded for period")

// Entry records budget usage for a single path prefix.
type Entry struct {
	Used      int
	Max       int
	WindowEnd time.Time
}

// Budget enforces per-path operation limits within a rolling time window.
type Budget struct {
	mu       sync.Mutex
	entries  map[string]*Entry
	max      int
	window   time.Duration
	nowFn    func() time.Time
}

// New creates a Budget that allows at most max operations per path per window.
func New(max int, window time.Duration) *Budget {
	return &Budget{
		entries: make(map[string]*Entry),
		max:     max,
		window:  window,
		nowFn:   time.Now,
	}
}

// Consume attempts to consume one unit of budget for the given path.
// Returns ErrBudgetExceeded if the budget is exhausted for the current window.
func (b *Budget) Consume(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.nowFn()
	e, ok := b.entries[path]
	if !ok || now.After(e.WindowEnd) {
		b.entries[path] = &Entry{
			Used:      1,
			Max:       b.max,
			WindowEnd: now.Add(b.window),
		}
		return nil
	}
	if e.Used >= e.Max {
		return ErrBudgetExceeded
	}
	e.Used++
	return nil
}

// Remaining returns the number of operations remaining for the given path
// in the current window. Returns max if no entry exists yet.
func (b *Budget) Remaining(path string) int {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.nowFn()
	e, ok := b.entries[path]
	if !ok || now.After(e.WindowEnd) {
		return b.max
	}
	return e.Max - e.Used
}

// Reset clears the budget entry for the given path.
func (b *Budget) Reset(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.entries, path)
}

// Purge removes all entries whose windows have expired.
func (b *Budget) Purge() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := b.nowFn()
	for path, e := range b.entries {
		if now.After(e.WindowEnd) {
			delete(b.entries, path)
		}
	}
}
