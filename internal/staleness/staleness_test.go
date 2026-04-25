package staleness

import (
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/lease"
)

func newLease(id string, status lease.Status) *lease.Lease {
	l := lease.New(id, "secret/data/"+id, 300)
	l.Status = status
	return l
}

func TestObserve_FirstSeen_NoError(t *testing.T) {
	tr := New(5 * time.Second)
	l := newLease("abc", lease.StatusExpiring)
	if err := tr.Observe(l); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestObserve_WithinThreshold_NoError(t *testing.T) {
	now := time.Now()
	tr := New(10 * time.Second)
	tr.now = func() time.Time { return now }

	l := newLease("abc", lease.StatusExpiring)
	_ = tr.Observe(l)

	tr.now = func() time.Time { return now.Add(5 * time.Second) }
	if err := tr.Observe(l); err != nil {
		t.Fatalf("expected nil within threshold, got %v", err)
	}
}

func TestObserve_ExceedsThreshold_ReturnsError(t *testing.T) {
	now := time.Now()
	tr := New(10 * time.Second)
	tr.now = func() time.Time { return now }

	l := newLease("abc", lease.StatusExpiring)
	_ = tr.Observe(l)

	tr.now = func() time.Time { return now.Add(15 * time.Second) }
	if err := tr.Observe(l); err != ErrThresholdExceeded {
		t.Fatalf("expected ErrThresholdExceeded, got %v", err)
	}
}

func TestObserve_StatusChange_ResetsTimer(t *testing.T) {
	now := time.Now()
	tr := New(5 * time.Second)
	tr.now = func() time.Time { return now }

	l := newLease("abc", lease.StatusExpiring)
	_ = tr.Observe(l)

	tr.now = func() time.Time { return now.Add(8 * time.Second) }
	l.Status = lease.StatusExpired
	if err := tr.Observe(l); err != nil {
		t.Fatalf("status change should reset timer, got %v", err)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := New(time.Minute)
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected false for missing entry")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	tr := New(time.Minute)
	l := newLease("abc", lease.StatusHealthy)
	_ = tr.Observe(l)
	tr.Delete("abc")
	_, ok := tr.Get("abc")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestPurge_RemovesInactive(t *testing.T) {
	tr := New(time.Minute)
	for _, id := range []string{"a", "b", "c"} {
		_ = tr.Observe(newLease(id, lease.StatusHealthy))
	}
	active := map[string]struct{}{"a": {}}
	tr.Purge(active)
	if _, ok := tr.Get("b"); ok {
		t.Error("expected b to be purged")
	}
	if _, ok := tr.Get("a"); !ok {
		t.Error("expected a to remain")
	}
}
