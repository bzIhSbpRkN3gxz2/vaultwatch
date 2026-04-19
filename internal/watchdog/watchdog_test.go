package watchdog_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
	"github.com/yourusername/vaultwatch/internal/watchdog"
)

func newLease(id string) *lease.Lease {
	return lease.New(id, "secret/data/"+id, 300)
}

func TestHeartbeat_PreventsStale(t *testing.T) {
	wd := watchdog.New(5 * time.Second)
	l := newLease("lease-1")
	wd.Heartbeat(l.ID)
	stale := wd.Stale([]*lease.Lease{l})
	if len(stale) != 0 {
		t.Fatalf("expected no stale leases, got %d", len(stale))
	}
}

func TestStale_NoHeartbeat(t *testing.T) {
	wd := watchdog.New(5 * time.Second)
	l := newLease("lease-2")
	stale := wd.Stale([]*lease.Lease{l})
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale lease, got %d", len(stale))
	}
	if stale[0].ID != l.ID {
		t.Errorf("unexpected lease id: %s", stale[0].ID)
	}
}

func TestStale_AfterTTLExpiry(t *testing.T) {
	wd := watchdog.New(10 * time.Millisecond)
	l := newLease("lease-3")
	wd.Heartbeat(l.ID)
	time.Sleep(30 * time.Millisecond)
	stale := wd.Stale([]*lease.Lease{l})
	if len(stale) != 1 {
		t.Fatalf("expected 1 stale lease after TTL expiry, got %d", len(stale))
	}
}

func TestPurge_RemovesInactiveEntries(t *testing.T) {
	wd := watchdog.New(5 * time.Second)
	wd.Heartbeat("lease-a")
	wd.Heartbeat("lease-b")

	active := map[string]struct{}{"lease-a": {}}
	wd.Purge(active)

	// lease-b should now be treated as stale (no entry)
	lb := newLease("lease-b")
	stale := wd.Stale([]*lease.Lease{lb})
	if len(stale) != 1 {
		t.Fatalf("expected lease-b to be stale after purge, got %d stale", len(stale))
	}
}

func TestStale_MultipleLeases(t *testing.T) {
	wd := watchdog.New(5 * time.Second)
	leases := []*lease.Lease{
		newLease("l1"),
		newLease("l2"),
		newLease("l3"),
	}
	wd.Heartbeat("l2")
	stale := wd.Stale(leases)
	if len(stale) != 2 {
		t.Fatalf("expected 2 stale leases, got %d", len(stale))
	}
}
