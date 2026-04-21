package eviction

import (
	"testing"
	"time"
)

func TestEvict_And_IsEvicted(t *testing.T) {
	tr := New(5 * time.Minute)
	tr.Evict("lease-1", "expired")

	if !tr.IsEvicted("lease-1") {
		t.Fatal("expected lease-1 to be evicted")
	}
	if tr.IsEvicted("lease-2") {
		t.Fatal("lease-2 should not be evicted")
	}
}

func TestEvict_Check_ReturnsError(t *testing.T) {
	tr := New(5 * time.Minute)
	tr.Evict("lease-1", "orphaned")

	if err := tr.Check("lease-1"); err != ErrEvicted {
		t.Fatalf("expected ErrEvicted, got %v", err)
	}
	if err := tr.Check("lease-2"); err != nil {
		t.Fatalf("expected nil for unknown lease, got %v", err)
	}
}

func TestEvict_Get_ReturnsEntry(t *testing.T) {
	tr := New(10 * time.Minute)
	tr.Evict("lease-3", "manual")

	e, ok := tr.Get("lease-3")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if e.LeaseID != "lease-3" {
		t.Errorf("expected LeaseID lease-3, got %s", e.LeaseID)
	}
	if e.Reason != "manual" {
		t.Errorf("expected reason manual, got %s", e.Reason)
	}
}

func TestEvict_Get_Missing(t *testing.T) {
	tr := New(5 * time.Minute)
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected missing entry to return false")
	}
}

func TestEvict_ExpiresAfterQuarantine(t *testing.T) {
	now := time.Now()
	tr := New(1 * time.Minute)
	tr.now = func() time.Time { return now }
	tr.Evict("lease-4", "expired")

	// advance past quarantine
	tr.now = func() time.Time { return now.Add(2 * time.Minute) }

	if tr.IsEvicted("lease-4") {
		t.Fatal("expected lease-4 quarantine to have expired")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	tr := New(1 * time.Minute)
	tr.now = func() time.Time { return now }
	tr.Evict("lease-5", "stale")
	tr.Evict("lease-6", "stale")

	// advance time so both entries expire
	tr.now = func() time.Time { return now.Add(2 * time.Minute) }
	tr.Purge()

	if _, ok := tr.Get("lease-5"); ok {
		t.Error("expected lease-5 to be purged")
	}
	if _, ok := tr.Get("lease-6"); ok {
		t.Error("expected lease-6 to be purged")
	}
}

func TestPurge_KeepsActiveEntries(t *testing.T) {
	now := time.Now()
	tr := New(10 * time.Minute)
	tr.now = func() time.Time { return now }
	tr.Evict("lease-7", "active")

	tr.now = func() time.Time { return now.Add(1 * time.Minute) }
	tr.Purge()

	if _, ok := tr.Get("lease-7"); !ok {
		t.Error("expected lease-7 to remain after purge")
	}
}
