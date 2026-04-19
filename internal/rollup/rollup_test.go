package rollup_test

import (
	"sync"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lease"
	"github.com/vaultwatch/internal/rollup"
)

func newLease(id string) *lease.Lease {
	return lease.New(id, "secret/data/"+id, 300)
}

func TestAdd_FlushesOnMax(t *testing.T) {
	var mu sync.Mutex
	var batches []rollup.Batch

	r := rollup.New(10*time.Second, 3, func(b rollup.Batch) error {
		mu.Lock()
		batches = append(batches, b)
		mu.Unlock()
		return nil
	})

	r.Add(newLease("a"))
	r.Add(newLease("b"))
	r.Add(newLease("c")) // triggers flush

	mu.Lock()
	defer mu.Unlock()
	if len(batches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(batches))
	}
	if len(batches[0].Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(batches[0].Events))
	}
}

func TestFlush_EmptyQueue_NoHandler(t *testing.T) {
	called := false
	r := rollup.New(time.Second, 10, func(b rollup.Batch) error {
		called = true
		return nil
	})
	r.Flush()
	if called {
		t.Fatal("handler should not be called on empty queue")
	}
}

func TestFlush_ClearsQueue(t *testing.T) {
	count := 0
	r := rollup.New(time.Second, 10, func(b rollup.Batch) error {
		count += len(b.Events)
		return nil
	})
	r.Add(newLease("x"))
	r.Flush()
	r.Flush() // second flush should be a no-op
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestStart_PeriodicFlush(t *testing.T) {
	var mu sync.Mutex
	flushed := 0

	r := rollup.New(50*time.Millisecond, 100, func(b rollup.Batch) error {
		mu.Lock()
		flushed++
		mu.Unlock()
		return nil
	})
	r.Add(newLease("p"))
	r.Start()
	time.Sleep(120 * time.Millisecond)
	r.Stop()

	mu.Lock()
	defer mu.Unlock()
	if flushed == 0 {
		t.Fatal("expected at least one periodic flush")
	}
}
