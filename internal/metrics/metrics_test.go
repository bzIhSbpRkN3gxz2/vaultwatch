package metrics_test

import (
	"testing"

	"github.com/your-org/vaultwatch/internal/metrics"
)

func TestInc(t *testing.T) {
	r := metrics.New()
	r.Inc(metrics.LeasesPolled)
	r.Inc(metrics.LeasesPolled)
	if got := r.Get(metrics.LeasesPolled); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestAdd(t *testing.T) {
	r := metrics.New()
	r.Add(metrics.RenewalsAttempted, 5)
	r.Add(metrics.RenewalsAttempted, 3)
	if got := r.Get(metrics.RenewalsAttempted); got != 8 {
		t.Fatalf("expected 8, got %d", got)
	}
}

func TestGet_UnknownCounter(t *testing.T) {
	r := metrics.New()
	if got := r.Get("nonexistent"); got != 0 {
		t.Fatalf("expected 0 for unknown counter, got %d", got)
	}
}

func TestReset(t *testing.T) {
	r := metrics.New()
	r.Add(metrics.AlertsDispatched, 10)
	r.Reset(metrics.AlertsDispatched)
	if got := r.Get(metrics.AlertsDispatched); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestSnapshot_IsCopy(t *testing.T) {
	r := metrics.New()
	r.Inc(metrics.LeasesPolled)
	r.Inc(metrics.RenewalsFailed)

	snap := r.Snapshot()

	if snap.Counters[metrics.LeasesPolled] != 1 {
		t.Fatalf("expected 1 in snapshot, got %d", snap.Counters[metrics.LeasesPolled])
	}
	if snap.RecordedAt.IsZero() {
		t.Fatal("expected non-zero RecordedAt")
	}

	// Mutating the registry must not affect the snapshot.
	r.Inc(metrics.LeasesPolled)
	if snap.Counters[metrics.LeasesPolled] != 1 {
		t.Fatal("snapshot should be independent of subsequent Inc calls")
	}
}

func TestConcurrentInc(t *testing.T) {
	r := metrics.New()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			r.Inc(metrics.LeasesPolled)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}
	if got := r.Get(metrics.LeasesPolled); got != 100 {
		t.Fatalf("expected 100 after concurrent incs, got %d", got)
	}
}
