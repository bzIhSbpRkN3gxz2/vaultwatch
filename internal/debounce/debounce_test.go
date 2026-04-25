package debounce

import (
	"testing"
	"time"
)

func TestAllow_FirstOccurrence(t *testing.T) {
	d := New(5 * time.Second)
	if !d.Allow("lease-1") {
		t.Fatal("expected first occurrence to be allowed")
	}
}

func TestAllow_SuppressesWithinWindow(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("lease-1")
	if d.Allow("lease-1") {
		t.Fatal("expected second call within window to be suppressed")
	}
}

func TestAllow_PermitsAfterWindow(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("lease-1")

	// Advance past the window.
	d.now = func() time.Time { return now.Add(6 * time.Second) }
	if !d.Allow("lease-1") {
		t.Fatal("expected call after window to be allowed")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("lease-1")
	if !d.Allow("lease-2") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("lease-1")
	d.Reset("lease-1")
	if !d.Allow("lease-1") {
		t.Fatal("expected allow after reset")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	d := New(5 * time.Second)
	d.now = func() time.Time { return now }

	d.Allow("lease-1")
	d.Allow("lease-2")

	d.now = func() time.Time { return now.Add(10 * time.Second) }
	d.Purge()

	if d.Len() != 0 {
		t.Fatalf("expected 0 entries after purge, got %d", d.Len())
	}
}

func TestLen_TracksEntries(t *testing.T) {
	d := New(5 * time.Second)
	if d.Len() != 0 {
		t.Fatal("expected empty debouncer")
	}
	d.Allow("a")
	d.Allow("b")
	if d.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", d.Len())
	}
}
