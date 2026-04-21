package pressure

import (
	"testing"
	"time"
)

func TestRecord_And_Get(t *testing.T) {
	tr := New()
	tr.Record("secret/db", 10, 2, 3)

	s, ok := tr.Get("secret/db")
	if !ok {
		t.Fatal("expected score to exist")
	}
	if s.Total != 10 || s.Critical != 2 || s.Warning != 3 {
		t.Fatalf("unexpected counts: %+v", s)
	}
	// value = (2*2 + 3) / (10*2) = 7/20 = 0.35
	if s.Value < 0.34 || s.Value > 0.36 {
		t.Fatalf("unexpected value: %f", s.Value)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := New()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Fatal("expected no score for unknown path")
	}
}

func TestRecord_ZeroTotal_ZeroValue(t *testing.T) {
	tr := New()
	tr.Record("secret/empty", 0, 0, 0)
	s, ok := tr.Get("secret/empty")
	if !ok {
		t.Fatal("expected score to exist")
	}
	if s.Value != 0.0 {
		t.Fatalf("expected 0.0, got %f", s.Value)
	}
}

func TestRecord_ClampsAtOne(t *testing.T) {
	tr := New()
	// All critical: value should clamp at 1.0
	tr.Record("secret/overload", 5, 5, 5)
	s, _ := tr.Get("secret/overload")
	if s.Value > 1.0 {
		t.Fatalf("value exceeded 1.0: %f", s.Value)
	}
}

func TestAll_ReturnsCopy(t *testing.T) {
	tr := New()
	tr.Record("a", 4, 1, 1)
	tr.Record("b", 6, 2, 2)
	all := tr.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(all))
	}
}

func TestPurge_RemovesOldEntries(t *testing.T) {
	now := time.Now()
	tr := &Tracker{
		scores: make(map[string]*Score),
		clock:  func() time.Time { return now },
	}
	tr.Record("old/path", 3, 1, 1)

	// Advance clock beyond maxAge
	tr.clock = func() time.Time { return now.Add(10 * time.Minute) }
	tr.Purge(5 * time.Minute)

	_, ok := tr.Get("old/path")
	if ok {
		t.Fatal("expected old entry to be purged")
	}
}

func TestPurge_KeepsRecentEntries(t *testing.T) {
	now := time.Now()
	tr := &Tracker{
		scores: make(map[string]*Score),
		clock:  func() time.Time { return now },
	}
	tr.Record("new/path", 3, 1, 1)

	tr.clock = func() time.Time { return now.Add(2 * time.Minute) }
	tr.Purge(5 * time.Minute)

	_, ok := tr.Get("new/path")
	if !ok {
		t.Fatal("expected recent entry to be retained")
	}
}
