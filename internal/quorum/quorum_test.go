package quorum

import (
	"testing"
	"time"
)

func TestReached_NotMet_NoVotes(t *testing.T) {
	q := New(2, 10*time.Second)
	if err := q.Reached("lease-1"); err != ErrQuorumNotMet {
		t.Fatalf("expected ErrQuorumNotMet, got %v", err)
	}
}

func TestReached_MetExactly(t *testing.T) {
	q := New(2, 10*time.Second)
	q.Vote("lease-1", true)
	q.Vote("lease-1", true)
	if err := q.Reached("lease-1"); err != nil {
		t.Fatalf("expected quorum to be met, got %v", err)
	}
}

func TestReached_DisagreeVotesDontCount(t *testing.T) {
	q := New(2, 10*time.Second)
	q.Vote("lease-1", false)
	q.Vote("lease-1", false)
	q.Vote("lease-1", true)
	if err := q.Reached("lease-1"); err != ErrQuorumNotMet {
		t.Fatalf("expected ErrQuorumNotMet, got %v", err)
	}
}

func TestReached_VotesOutsideWindowIgnored(t *testing.T) {
	now := time.Now()
	q := New(2, 5*time.Second)
	q.now = func() time.Time { return now }

	// Cast votes in the past, outside the window.
	q.now = func() time.Time { return now.Add(-10 * time.Second) }
	q.Vote("lease-1", true)
	q.Vote("lease-1", true)

	// Evaluate at current time.
	q.now = func() time.Time { return now }
	if err := q.Reached("lease-1"); err != ErrQuorumNotMet {
		t.Fatalf("expected ErrQuorumNotMet for stale votes, got %v", err)
	}
}

func TestReset_ClearsVotes(t *testing.T) {
	q := New(1, 10*time.Second)
	q.Vote("lease-1", true)
	q.Reset("lease-1")
	if err := q.Reached("lease-1"); err != ErrQuorumNotMet {
		t.Fatalf("expected ErrQuorumNotMet after reset, got %v", err)
	}
}

func TestPurge_RemovesStaleEntries(t *testing.T) {
	now := time.Now()
	q := New(1, 5*time.Second)
	q.now = func() time.Time { return now.Add(-10 * time.Second) }
	q.Vote("lease-old", true)

	q.now = func() time.Time { return now }
	q.Vote("lease-new", true)

	q.Purge()

	q.mu.Lock()
	defer q.mu.Unlock()
	if _, ok := q.entries["lease-old"]; ok {
		t.Fatal("expected lease-old to be purged")
	}
	if _, ok := q.entries["lease-new"]; !ok {
		t.Fatal("expected lease-new to survive purge")
	}
}

func TestNew_ThresholdBelowOneClampedToOne(t *testing.T) {
	q := New(0, 10*time.Second)
	if q.threshold != 1 {
		t.Fatalf("expected threshold=1, got %d", q.threshold)
	}
}
