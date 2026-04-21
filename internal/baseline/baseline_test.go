package baseline

import (
	"testing"
	"time"
)

func TestRecord_BuildsBaseline(t *testing.T) {
	tr := New(2.0)
	ttl := 3600 * time.Second
	for i := 0; i < 5; i++ {
		if err := tr.Record("secret/db", ttl); err != nil {
			t.Fatalf("unexpected error on observation %d: %v", i, err)
		}
	}
	e, ok := tr.Get("secret/db")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Count != 5 {
		t.Errorf("expected count 5, got %d", e.Count)
	}
	if e.Mean != 3600 {
		t.Errorf("expected mean 3600, got %f", e.Mean)
	}
}

func TestRecord_DetectsAnomaly(t *testing.T) {
	tr := New(2.0)
	base := 3600 * time.Second
	for i := 0; i < 10; i++ {
		if err := tr.Record("secret/db", base); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	// A TTL of 10s is far below the 3600s baseline — should be anomalous.
	err := tr.Record("secret/db", 10*time.Second)
	if err == nil {
		t.Error("expected anomaly error, got nil")
	}
}

func TestRecord_NoAnomalyBelowThreshold(t *testing.T) {
	tr := New(2.0)
	base := 3600 * time.Second
	for i := 0; i < 10; i++ {
		if err := tr.Record("secret/db", base); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	// Slight variation — should not trigger anomaly when std-dev is 0.
	if err := tr.Record("secret/db", base); err != nil {
		t.Errorf("unexpected anomaly error: %v", err)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := New(2.0)
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Error("expected false for missing path")
	}
}

func TestPurge_RemovesStaleEntries(t *testing.T) {
	tr := New(2.0)
	_ = tr.Record("secret/old", 3600*time.Second)

	cutoff := time.Now().Add(time.Second)
	tr.Purge(cutoff)

	_, ok := tr.Get("secret/old")
	if ok {
		t.Error("expected entry to be purged")
	}
}

func TestPurge_KeepsRecentEntries(t *testing.T) {
	tr := New(2.0)
	_ = tr.Record("secret/new", 3600*time.Second)

	cutoff := time.Now().Add(-time.Hour)
	tr.Purge(cutoff)

	_, ok := tr.Get("secret/new")
	if !ok {
		t.Error("expected recent entry to be retained")
	}
}

func TestDefaultThreshold_UsedWhenZero(t *testing.T) {
	tr := New(0)
	if tr.threshold != 2.0 {
		t.Errorf("expected default threshold 2.0, got %f", tr.threshold)
	}
}
