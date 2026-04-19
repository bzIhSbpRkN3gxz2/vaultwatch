// Package rollup aggregates alert events into batched summaries
// to reduce noise when many leases expire simultaneously.
package rollup

import (
	"sync"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Event holds a single alert event queued for rollup.
type Event struct {
	Lease     *lease.Lease
	QueuedAt  time.Time
}

// Batch is a collected set of events flushed together.
type Batch struct {
	Events    []Event
	FlushedAt time.Time
}

// Handler is called with each flushed batch.
type Handler func(Batch) error

// Rollup buffers lease alert events and flushes them periodically.
type Rollup struct {
	mu       sync.Mutex
	queue    []Event
	window   time.Duration
	max      int
	handler  Handler
	stopCh   chan struct{}
}

// New creates a Rollup that flushes every window duration or when max events accumulate.
func New(window time.Duration, max int, h Handler) *Rollup {
	return &Rollup{
		window:  window,
		max:     max,
		handler: h,
		stopCh:  make(chan struct{}),
	}
}

// Add enqueues a lease event, flushing immediately if max is reached.
func (r *Rollup) Add(l *lease.Lease) {
	r.mu.Lock()
	r.queue = append(r.queue, Event{Lease: l, QueuedAt: time.Now()})
	should := len(r.queue) >= r.max
	r.mu.Unlock()
	if should {
		r.Flush()
	}
}

// Flush drains the queue and calls the handler.
func (r *Rollup) Flush() {
	r.mu.Lock()
	if len(r.queue) == 0 {
		r.mu.Unlock()
		return
	}
	batch := Batch{Events: r.queue, FlushedAt: time.Now()}
	r.queue = nil
	r.mu.Unlock()
	_ = r.handler(batch)
}

// Start begins the periodic flush loop.
func (r *Rollup) Start() {
	go func() {
		ticker := time.NewTicker(r.window)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.Flush()
			case <-r.stopCh:
				return
			}
		}
	}()
}

// Stop halts the periodic flush loop.
func (r *Rollup) Stop() {
	close(r.stopCh)
}
