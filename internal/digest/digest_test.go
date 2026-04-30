package digest_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/digest"
	"github.com/vaultwatch/internal/lease"
)

func newLease(id, path, status string, ttl int) *lease.Lease {
	l := lease.New(id, path)
	l.Status = lease.Status(status)
	l.TTL = ttl
	return l
}

func TestCompute_StableForSameLease(t *testing.T) {
	l := newLease("id-1", "secret/db", "healthy", 3600)
	d1 := digest.Compute(l)
	d2 := digest.Compute(l)
	if d1 != d2 {
		t.Fatalf("expected stable digest, got %s and %s", d1, d2)
	}
}

func TestCompute_DiffersOnTTLChange(t *testing.T) {
	l1 := newLease("id-1", "secret/db", "healthy", 3600)
	l2 := newLease("id-1", "secret/db", "healthy", 1800)
	if digest.Compute(l1) == digest.Compute(l2) {
		t.Fatal("expected different digests for different TTLs")
	}
}

func TestChanged_FirstObservation_ReturnsTrue(t *testing.T) {
	tr := digest.New()
	l := newLease("id-2", "secret/api", "healthy", 7200)
	if !tr.Changed(l) {
		t.Fatal("expected Changed=true on first observation")
	}
}

func TestChanged_SameState_ReturnsFalse(t *testing.T) {
	tr := digest.New()
	l := newLease("id-3", "secret/api", "healthy", 7200)
	tr.Changed(l) // seed
	if tr.Changed(l) {
		t.Fatal("expected Changed=false for unchanged lease")
	}
}

func TestChanged_StatusChange_ReturnsTrue(t *testing.T) {
	tr := digest.New()
	l1 := newLease("id-4", "secret/db", "healthy", 3600)
	tr.Changed(l1)
	l2 := newLease("id-4", "secret/db", "expiring", 3600)
	if !tr.Changed(l2) {
		t.Fatal("expected Changed=true after status change")
	}
}

func TestChanged_NilLease_ReturnsFalse(t *testing.T) {
	tr := digest.New()
	if tr.Changed(nil) {
		t.Fatal("expected Changed=false for nil lease")
	}
}

func TestGet_Missing(t *testing.T) {
	tr := digest.New()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected Get to return false for unknown lease")
	}
}

func TestGet_AfterChanged(t *testing.T) {
	tr := digest.New()
	l := newLease("id-5", "secret/x", "healthy", 100)
	tr.Changed(l)
	e, ok := tr.Get("id-5")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Digest != digest.Compute(l) {
		t.Fatalf("digest mismatch: got %s", e.Digest)
	}
}

func TestPurge_RemovesStaleEntries(t *testing.T) {
	tr := digest.New()
	l := newLease("id-6", "secret/y", "healthy", 500)
	tr.Changed(l)
	tr.Purge(time.Now().Add(time.Second))
	_, ok := tr.Get("id-6")
	if ok {
		t.Fatal("expected entry to be purged")
	}
}
