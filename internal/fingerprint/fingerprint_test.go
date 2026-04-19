package fingerprint_test

import (
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/fingerprint"
	"github.com/your-org/vaultwatch/internal/lease"
)

func newLease(id, path string, ttl time.Duration) *lease.Lease {
	l := lease.New(id, path, ttl)
	return l
}

func TestCompute_StableForSameLease(t *testing.T) {
	g := fingerprint.New()
	l := newLease("lease-1", "secret/db", 30*time.Minute)
	a := g.Compute(l)
	b := g.Compute(l)
	if a != b {
		t.Fatalf("expected stable fingerprint, got %s and %s", a, b)
	}
}

func TestCompute_DiffersForDifferentLeaseID(t *testing.T) {
	g := fingerprint.New()
	a := g.Compute(newLease("lease-1", "secret/db", 30*time.Minute))
	b := g.Compute(newLease("lease-2", "secret/db", 30*time.Minute))
	if a == b {
		t.Fatal("expected different fingerprints for different lease IDs")
	}
}

func TestCompute_DiffersForDifferentPath(t *testing.T) {
	g := fingerprint.New()
	a := g.Compute(newLease("lease-1", "secret/db", 30*time.Minute))
	b := g.Compute(newLease("lease-1", "secret/cache", 30*time.Minute))
	if a == b {
		t.Fatal("expected different fingerprints for different paths")
	}
}

func TestCompute_WithTTL_DiffersOnTTLChange(t *testing.T) {
	g := fingerprint.New(fingerprint.WithTTL())
	a := g.Compute(newLease("lease-1", "secret/db", 30*time.Minute))
	b := g.Compute(newLease("lease-1", "secret/db", 10*time.Minute))
	if a == b {
		t.Fatal("expected different fingerprints when TTL changes with WithTTL option")
	}
}

func TestCompute_WithoutTTL_StableAcrossTTLChange(t *testing.T) {
	g := fingerprint.New()
	a := g.Compute(newLease("lease-1", "secret/db", 30*time.Minute))
	b := g.Compute(newLease("lease-1", "secret/db", 10*time.Minute))
	if a != b {
		t.Fatal("expected same fingerprint when TTL changes without WithTTL option")
	}
}

func TestFingerprint_String(t *testing.T) {
	g := fingerprint.New()
	f := g.Compute(newLease("lease-x", "secret/x", time.Minute))
	if f.String() == "" {
		t.Fatal("expected non-empty fingerprint string")
	}
}
