package blacklist_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/blacklist"
)

func TestContains_Permanent(t *testing.T) {
	b := blacklist.New()
	b.Add("lease-1", "manual ban", 0)
	if !b.Contains("lease-1") {
		t.Fatal("expected lease-1 to be blacklisted")
	}
}

func TestContains_Missing(t *testing.T) {
	b := blacklist.New()
	if b.Contains("unknown") {
		t.Fatal("expected unknown to not be blacklisted")
	}
}

func TestContains_Expired(t *testing.T) {
	b := blacklist.New()
	b.Add("lease-2", "temp", time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if b.Contains("lease-2") {
		t.Fatal("expected expired entry to not be blacklisted")
	}
}

func TestRemove(t *testing.T) {
	b := blacklist.New()
	b.Add("lease-3", "reason", 0)
	b.Remove("lease-3")
	if b.Contains("lease-3") {
		t.Fatal("expected lease-3 to be removed")
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	b := blacklist.New()
	b.Add("lease-4", "temp", time.Millisecond)
	b.Add("lease-5", "perm", 0)
	time.Sleep(5 * time.Millisecond)
	b.Purge()
	if b.Contains("lease-4") {
		t.Fatal("expected lease-4 purged")
	}
	if !b.Contains("lease-5") {
		t.Fatal("expected lease-5 to remain")
	}
}

func TestAll_ReturnsActiveOnly(t *testing.T) {
	b := blacklist.New()
	b.Add("lease-6", "perm", 0)
	b.Add("lease-7", "temp", time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	entries := b.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 active entry, got %d", len(entries))
	}
	if entries[0].LeaseID != "lease-6" {
		t.Fatalf("unexpected lease ID %s", entries[0].LeaseID)
	}
}

func TestAll_Empty(t *testing.T) {
	b := blacklist.New()
	if len(b.All()) != 0 {
		t.Fatal("expected empty blacklist")
	}
}
