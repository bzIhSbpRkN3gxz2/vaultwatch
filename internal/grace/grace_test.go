package grace_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/grace"
	"github.com/yourusername/vaultwatch/internal/lease"
)

func newLease(id string, ttl int) *lease.Lease {
	return &lease.Lease{
		LeaseID: id,
		Path:    "secret/data/test",
		TTL:     ttl,
	}
}

func TestObserve_WithinWindow_RecordsEntry(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	now := time.Now()
	l := newLease("lease-1", 120) // 2 minutes — inside 5-minute window

	if !tr.Observe(l, now) {
		t.Fatal("expected Observe to return true for lease within grace window")
	}
	e, ok := tr.Get("lease-1")
	if !ok {
		t.Fatal("expected entry to be recorded")
	}
	if e.LeaseID != "lease-1" {
		t.Errorf("unexpected LeaseID: %s", e.LeaseID)
	}
}

func TestObserve_OutsideWindow_ReturnsFalse(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	now := time.Now()
	l := newLease("lease-2", 600) // 10 minutes — outside window

	if tr.Observe(l, now) {
		t.Fatal("expected Observe to return false for lease outside grace window")
	}
	_, ok := tr.Get("lease-2")
	if ok {
		t.Fatal("expected no entry for lease outside grace window")
	}
}

func TestObserve_PreservesExistingEntry(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	now := time.Now()
	l := newLease("lease-3", 60)

	tr.Observe(l, now)
	later := now.Add(30 * time.Second)
	tr.Observe(l, later)

	e, _ := tr.Get("lease-3")
	if !e.EnteredAt.Equal(now) {
		t.Errorf("expected EnteredAt to be preserved as %v, got %v", now, e.EnteredAt)
	}
}

func TestObserve_NilLease_ReturnsFalse(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	if tr.Observe(nil, time.Now()) {
		t.Fatal("expected Observe to return false for nil lease")
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	now := time.Now()
	l := newLease("lease-4", 60)
	tr.Observe(l, now)
	tr.Remove("lease-4")
	_, ok := tr.Get("lease-4")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	tr := grace.New(5 * time.Minute)
	now := time.Now()

	tr.Observe(newLease("lease-5", 10), now)  // expires in 10s
	tr.Observe(newLease("lease-6", 200), now) // 200s — outside window, won't be recorded
	tr.Observe(newLease("lease-7", 30), now)  // expires in 30s

	future := now.Add(20 * time.Second)
	removed := tr.Purge(future)
	if removed != 1 {
		t.Errorf("expected 1 purged entry, got %d", removed)
	}
	_, ok := tr.Get("lease-5")
	if ok {
		t.Error("expected lease-5 to be purged")
	}
	_, ok = tr.Get("lease-7")
	if !ok {
		t.Error("expected lease-7 to still be present")
	}
}
