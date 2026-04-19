package dedup_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/dedup"
)

func TestIsDuplicate_FirstSeen(t *testing.T) {
	d := dedup.New(5 * time.Second)
	if d.IsDuplicate("lease-1", "fp-a") {
		t.Fatal("expected false for first-seen lease")
	}
}

func TestIsDuplicate_SameFingerprintWithinWindow(t *testing.T) {
	d := dedup.New(5 * time.Second)
	d.IsDuplicate("lease-1", "fp-a")
	if !d.IsDuplicate("lease-1", "fp-a") {
		t.Fatal("expected true for duplicate within window")
	}
}

func TestIsDuplicate_DifferentFingerprint(t *testing.T) {
	d := dedup.New(5 * time.Second)
	d.IsDuplicate("lease-1", "fp-a")
	if d.IsDuplicate("lease-1", "fp-b") {
		t.Fatal("expected false when fingerprint changes")
	}
}

func TestIsDuplicate_AfterWindowExpiry(t *testing.T) {
	d := dedup.New(10 * time.Millisecond)
	d.IsDuplicate("lease-1", "fp-a")
	time.Sleep(20 * time.Millisecond)
	if d.IsDuplicate("lease-1", "fp-a") {
		t.Fatal("expected false after window expiry")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	d := dedup.New(10 * time.Millisecond)
	d.IsDuplicate("lease-1", "fp-a")
	d.IsDuplicate("lease-2", "fp-b")
	time.Sleep(20 * time.Millisecond)
	d.Purge()
	// After purge, entries should be re-recordable without duplicate detection.
	if d.IsDuplicate("lease-1", "fp-a") {
		t.Fatal("expected false after purge")
	}
}

func TestReset_ClearsAllEntries(t *testing.T) {
	d := dedup.New(5 * time.Second)
	d.IsDuplicate("lease-1", "fp-a")
	d.Reset()
	if d.IsDuplicate("lease-1", "fp-a") {
		t.Fatal("expected false after reset")
	}
}
