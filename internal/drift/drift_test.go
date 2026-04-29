package drift_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/drift"
	"github.com/your-org/vaultwatch/internal/lease"
)

func newLease(id string, ttl time.Duration) *lease.Lease {
	return &lease.Lease{
		LeaseID: id,
		Path:    "secret/data/test",
		TTL:     ttl,
		Issued:  time.Now(),
	}
}

func TestObserve_FirstOccurrence_SetsBaseline(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("lease-1", 60*time.Second)
	if err := d.Observe(l); err != nil {
		t.Fatalf("expected no error on first observe, got %v", err)
	}
}

func TestObserve_WithinThreshold_NoError(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("lease-2", 60*time.Second)
	_ = d.Observe(l)

	// TTL changes by 5s on a 60s base — well within 25%.
	l.TTL = 55 * time.Second
	if err := d.Observe(l); err != nil {
		t.Fatalf("expected no error within threshold, got %v", err)
	}
}

func TestObserve_ExceedsThreshold_ReturnsError(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("lease-3", 60*time.Second)
	_ = d.Observe(l)

	// TTL jumps to 120s — 100% increase, far beyond 25% threshold.
	l.TTL = 120 * time.Second
	if err := d.Observe(l); err == nil {
		t.Fatal("expected ErrDriftDetected, got nil")
	}
}

func TestObserve_NilLease_NoError(t *testing.T) {
	d := drift.New(0.25)
	if err := d.Observe(nil); err != nil {
		t.Fatalf("expected nil for nil lease, got %v", err)
	}
}

func TestObserve_EmptyLeaseID_NoError(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("", 30*time.Second)
	if err := d.Observe(l); err != nil {
		t.Fatalf("expected nil for empty lease ID, got %v", err)
	}
}

func TestReset_AllowsRebaseline(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("lease-4", 60*time.Second)
	_ = d.Observe(l)

	d.Reset("lease-4")

	// After reset, any TTL should be accepted as a new baseline.
	l.TTL = 10 * time.Second
	if err := d.Observe(l); err != nil {
		t.Fatalf("expected no error after reset, got %v", err)
	}
}

func TestPurge_RemovesOldEntries(t *testing.T) {
	d := drift.New(0.25)
	l := newLease("lease-5", 60*time.Second)
	_ = d.Observe(l)

	// Purge entries older than 0 — effectively all.
	d.Purge(0)

	// After purge, observation should succeed as a fresh baseline.
	l.TTL = 5 * time.Second
	if err := d.Observe(l); err != nil {
		t.Fatalf("expected no error after purge, got %v", err)
	}
}

func TestNew_DefaultThreshold_OnInvalidInput(t *testing.T) {
	// threshold <= 0 should fall back to 0.25
	d := drift.New(-1)
	l := newLease("lease-6", 100*time.Second)
	_ = d.Observe(l)

	// 30% drift should trigger with default 0.25 threshold.
	l.TTL = 130 * time.Second
	if err := d.Observe(l); err == nil {
		t.Fatal("expected ErrDriftDetected with default threshold")
	}
}
