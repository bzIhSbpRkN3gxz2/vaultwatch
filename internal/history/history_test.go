package history_test

import (
	"testing"

	"github.com/vaultwatch/internal/history"
	"github.com/vaultwatch/internal/lease"
)

func TestRecord_And_Get(t *testing.T) {
	tr := history.New(10)
	tr.Record("lease-1", lease.StatusHealthy, lease.StatusExpiring)

	events := tr.Get("lease-1")
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].From != lease.StatusHealthy || events[0].To != lease.StatusExpiring {
		t.Errorf("unexpected transition: %v -> %v", events[0].From, events[0].To)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := history.New(10)
	if got := tr.Get("no-such-lease"); len(got) != 0 {
		t.Errorf("expected empty slice, got %v", got)
	}
}

func TestRecord_CapsAtMax(t *testing.T) {
	tr := history.New(3)
	for i := 0; i < 6; i++ {
		tr.Record("lease-x", lease.StatusHealthy, lease.StatusExpiring)
	}
	events := tr.Get("lease-x")
	if len(events) != 3 {
		t.Errorf("expected 3 events (capped), got %d", len(events))
	}
}

func TestPurge(t *testing.T) {
	tr := history.New(10)
	tr.Record("lease-2", lease.StatusExpiring, lease.StatusExpired)
	tr.Purge("lease-2")
	if got := tr.Get("lease-2"); len(got) != 0 {
		t.Errorf("expected empty after purge, got %v", got)
	}
}

func TestRecord_IsolatedPerLease(t *testing.T) {
	tr := history.New(10)
	tr.Record("a", lease.StatusHealthy, lease.StatusExpiring)
	tr.Record("b", lease.StatusExpiring, lease.StatusExpired)

	if len(tr.Get("a")) != 1 || len(tr.Get("b")) != 1 {
		t.Error("events should be isolated per lease ID")
	}
}
