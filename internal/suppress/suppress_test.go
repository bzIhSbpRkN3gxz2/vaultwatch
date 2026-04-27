package suppress

import (
	"testing"
	"time"
)

func newSuppressor(window time.Duration) *Suppressor {
	s := New(window)
	return s
}

func TestSuppress_NewEntry(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	added := s.Suppress("lease-1")
	if !added {
		t.Fatal("expected Suppress to return true for new entry")
	}
}

func TestSuppress_DuplicateWithinWindow(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	s.Suppress("lease-1")
	added := s.Suppress("lease-1")
	if added {
		t.Fatal("expected Suppress to return false for duplicate within window")
	}
}

func TestIsSuppressed_True(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	s.Suppress("lease-2")
	if !s.IsSuppressed("lease-2") {
		t.Fatal("expected lease-2 to be suppressed")
	}
}

func TestIsSuppressed_False_NotAdded(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	if s.IsSuppressed("lease-unknown") {
		t.Fatal("expected unknown lease to not be suppressed")
	}
}

func TestIsSuppressed_False_AfterExpiry(t *testing.T) {
	s := newSuppressor(10 * time.Millisecond)
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Suppress("lease-3")

	// Advance clock past window
	s.now = func() time.Time { return fixed.Add(20 * time.Millisecond) }
	if s.IsSuppressed("lease-3") {
		t.Fatal("expected lease-3 to no longer be suppressed after window expiry")
	}
}

func TestRelease_RemovesEntry(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	s.Suppress("lease-4")
	s.Release("lease-4")
	if s.IsSuppressed("lease-4") {
		t.Fatal("expected lease-4 to be released")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	s := newSuppressor(10 * time.Millisecond)
	fixed := time.Now()
	s.now = func() time.Time { return fixed }
	s.Suppress("lease-5")
	s.Suppress("lease-6")

	s.now = func() time.Time { return fixed.Add(20 * time.Millisecond) }
	s.Purge()

	if s.Len() != 0 {
		t.Fatalf("expected 0 entries after purge, got %d", s.Len())
	}
}

func TestLen_ReflectsActiveEntries(t *testing.T) {
	s := newSuppressor(5 * time.Minute)
	s.Suppress("lease-a")
	s.Suppress("lease-b")
	if s.Len() != 2 {
		t.Fatalf("expected 2 entries, got %d", s.Len())
	}
	s.Release("lease-a")
	if s.Len() != 1 {
		t.Fatalf("expected 1 entry after release, got %d", s.Len())
	}
}
