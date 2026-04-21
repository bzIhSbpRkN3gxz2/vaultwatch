package window

import (
	"testing"
	"time"
)

func TestRecord_And_Count(t *testing.T) {
	w := New(10 * time.Second)
	w.Record("lease-1")
	w.Record("lease-1")

	if got := w.Count("lease-1"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestCount_Missing(t *testing.T) {
	w := New(10 * time.Second)
	if got := w.Count("nonexistent"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestEviction_OutsideWindow(t *testing.T) {
	now := time.Now()
	w := New(5 * time.Second)

	// Inject a controlled clock.
	w.nowFn = func() time.Time { return now }
	w.Record("lease-1")

	// Advance time beyond the window.
	w.nowFn = func() time.Time { return now.Add(10 * time.Second) }
	w.Record("lease-1") // this one is inside the new window

	if got := w.Count("lease-1"); got != 1 {
		t.Fatalf("expected 1 after eviction, got %d", got)
	}
}

func TestReset_ClearsEntries(t *testing.T) {
	w := New(10 * time.Second)
	w.Record("lease-1")
	w.Record("lease-1")
	w.Reset("lease-1")

	if got := w.Count("lease-1"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestPurge_RemovesExpiredLeases(t *testing.T) {
	now := time.Now()
	w := New(5 * time.Second)
	w.nowFn = func() time.Time { return now }

	w.Record("lease-1")
	w.Record("lease-2")

	// Move time forward so all entries expire.
	w.nowFn = func() time.Time { return now.Add(10 * time.Second) }
	w.Purge()

	w.mu.Lock()
	remaining := len(w.entries)
	w.mu.Unlock()

	if remaining != 0 {
		t.Fatalf("expected 0 remaining entries after purge, got %d", remaining)
	}
}

func TestRecord_MultipleLeases_Isolated(t *testing.T) {
	w := New(10 * time.Second)
	w.Record("lease-a")
	w.Record("lease-a")
	w.Record("lease-b")

	if got := w.Count("lease-a"); got != 2 {
		t.Fatalf("lease-a: expected 2, got %d", got)
	}
	if got := w.Count("lease-b"); got != 1 {
		t.Fatalf("lease-b: expected 1, got %d", got)
	}
}
