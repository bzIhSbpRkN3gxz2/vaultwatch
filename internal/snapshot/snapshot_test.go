package snapshot_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/snapshot"
)

func TestRecord_And_Get(t *testing.T) {
	s := snapshot.New()
	s.Record("lease-1", 30*time.Second)

	e, ok := s.Get("lease-1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.TTL != 30*time.Second {
		t.Fatalf("expected TTL 30s, got %v", e.TTL)
	}
	if e.LeaseID != "lease-1" {
		t.Fatalf("unexpected lease ID %q", e.LeaseID)
	}
}

func TestGet_Missing(t *testing.T) {
	s := snapshot.New()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown lease")
	}
}

func TestDelete(t *testing.T) {
	s := snapshot.New()
	s.Record("lease-2", time.Minute)
	s.Delete("lease-2")
	_, ok := s.Get("lease-2")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestAll(t *testing.T) {
	s := snapshot.New()
	s.Record("a", 10*time.Second)
	s.Record("b", 20*time.Second)

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestTTLDecreased(t *testing.T) {
	s := snapshot.New()
	s.Record("lease-3", 60*time.Second)

	if !s.TTLDecreased("lease-3", 45*time.Second) {
		t.Fatal("expected TTL decrease to be detected")
	}
	if s.TTLDecreased("lease-3", 90*time.Second) {
		t.Fatal("expected no TTL decrease when TTL grew")
	}
}

func TestTTLDecreased_UnknownLease(t *testing.T) {
	s := snapshot.New()
	if s.TTLDecreased("unknown", 10*time.Second) {
		t.Fatal("expected false for unknown lease")
	}
}
