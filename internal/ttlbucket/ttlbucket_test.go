package ttlbucket_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/lease"
	"github.com/yourusername/vaultwatch/internal/ttlbucket"
)

func newLease(id string, ttlSeconds int) *lease.Lease {
	return lease.New(id, "secret/data/"+id, ttlSeconds)
}

func TestClassify_Critical(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{newLease("a", 1800)}) // 30 min → critical

	if got := b.Get(ttlbucket.BucketCritical); len(got) != 1 {
		t.Fatalf("expected 1 critical lease, got %d", len(got))
	}
	if got := b.Get(ttlbucket.BucketWarning); len(got) != 0 {
		t.Fatalf("expected 0 warning leases, got %d", len(got))
	}
}

func TestClassify_Warning(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{newLease("b", 10800)}) // 3 h → warning

	if got := b.Get(ttlbucket.BucketWarning); len(got) != 1 {
		t.Fatalf("expected 1 warning lease, got %d", len(got))
	}
}

func TestClassify_Healthy(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{newLease("c", 86400)}) // 24 h → healthy

	if got := b.Get(ttlbucket.BucketHealthy); len(got) != 1 {
		t.Fatalf("expected 1 healthy lease, got %d", len(got))
	}
}

func TestClassify_MixedLeases(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{
		newLease("crit", 600),   // 10 min
		newLease("warn", 7200),  // 2 h
		newLease("ok", 43200),   // 12 h
	})

	counts := b.Counts()
	if counts[ttlbucket.BucketCritical] != 1 {
		t.Errorf("critical: want 1, got %d", counts[ttlbucket.BucketCritical])
	}
	if counts[ttlbucket.BucketWarning] != 1 {
		t.Errorf("warning: want 1, got %d", counts[ttlbucket.BucketWarning])
	}
	if counts[ttlbucket.BucketHealthy] != 1 {
		t.Errorf("healthy: want 1, got %d", counts[ttlbucket.BucketHealthy])
	}
}

func TestClassify_ReplacesOldData(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{newLease("x", 300)})
	b.Classify([]*lease.Lease{}) // reclassify with empty set

	if got := b.Get(ttlbucket.BucketCritical); len(got) != 0 {
		t.Fatalf("expected buckets cleared, got %d critical", len(got))
	}
}

func TestGet_UnknownLabel_ReturnsEmpty(t *testing.T) {
	b := ttlbucket.New()
	if got := b.Get("nonexistent"); len(got) != 0 {
		t.Fatalf("expected empty slice for unknown label, got %d", len(got))
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	b := ttlbucket.New()
	b.Classify([]*lease.Lease{newLease("y", 300)})

	copy1 := b.Get(ttlbucket.BucketCritical)
	copy1[0] = nil // mutate returned slice

	copy2 := b.Get(ttlbucket.BucketCritical)
	if copy2[0] == nil {
		t.Fatal("Get should return an independent copy")
	}
}
