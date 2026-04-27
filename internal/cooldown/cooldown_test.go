package cooldown

import (
	"testing"
	"time"
)

func TestAllow_FirstOccurrence(t *testing.T) {
	tr := New(5 * time.Second)
	if !tr.Allow("lease-1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SuppressedWithinCooldown(t *testing.T) {
	now := time.Now()
	tr := New(10 * time.Second)
	tr.now = func() time.Time { return now }

	tr.Allow("lease-1")
	if tr.Allow("lease-1") {
		t.Fatal("expected second call within cooldown to be suppressed")
	}
}

func TestAllow_PermitsAfterCooldown(t *testing.T) {
	base := time.Now()
	tr := New(5 * time.Second)
	tr.now = func() time.Time { return base }

	tr.Allow("lease-1")

	tr.now = func() time.Time { return base.Add(6 * time.Second) }
	if !tr.Allow("lease-1") {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	base := time.Now()
	tr := New(10 * time.Second)
	tr.now = func() time.Time { return base }

	tr.Allow("lease-1")
	if !tr.Allow("lease-2") {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestAllow_IncrementsCount(t *testing.T) {
	base := time.Now()
	tr := New(1 * time.Second)
	tr.now = func() time.Time { return base }
	tr.Allow("lease-1")

	tr.now = func() time.Time { return base.Add(2 * time.Second) }
	tr.Allow("lease-1")

	e, ok := tr.Get("lease-1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Count != 2 {
		t.Fatalf("expected count 2, got %d", e.Count)
	}
}

func TestReset_AllowsImmediately(t *testing.T) {
	base := time.Now()
	tr := New(30 * time.Second)
	tr.now = func() time.Time { return base }

	tr.Allow("lease-1")
	tr.Reset("lease-1")

	if !tr.Allow("lease-1") {
		t.Fatal("expected reset key to be allowed immediately")
	}
}

func TestPurge_RemovesElapsedEntries(t *testing.T) {
	base := time.Now()
	tr := New(5 * time.Second)
	tr.now = func() time.Time { return base }

	tr.Allow("lease-1")
	tr.Allow("lease-2")

	tr.now = func() time.Time { return base.Add(10 * time.Second) }
	tr.Purge()

	if _, ok := tr.Get("lease-1"); ok {
		t.Fatal("expected purged entry to be removed")
	}
	if _, ok := tr.Get("lease-2"); ok {
		t.Fatal("expected purged entry to be removed")
	}
}

func TestGet_Missing(t *testing.T) {
	tr := New(5 * time.Second)
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}
