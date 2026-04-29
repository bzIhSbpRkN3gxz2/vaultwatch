package budget

import (
	"testing"
	"time"
)

func TestConsume_WithinBudget(t *testing.T) {
	b := New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if err := b.Consume("secret/db"); err != nil {
			t.Fatalf("unexpected error on attempt %d: %v", i+1, err)
		}
	}
}

func TestConsume_ExceedsBudget(t *testing.T) {
	b := New(2, time.Minute)
	_ = b.Consume("secret/db")
	_ = b.Consume("secret/db")
	if err := b.Consume("secret/db"); err != ErrBudgetExceeded {
		t.Fatalf("expected ErrBudgetExceeded, got %v", err)
	}
}

func TestConsume_ResetsAfterWindow(t *testing.T) {
	now := time.Now()
	b := New(1, time.Minute)
	b.nowFn = func() time.Time { return now }

	_ = b.Consume("secret/db")
	if err := b.Consume("secret/db"); err != ErrBudgetExceeded {
		t.Fatalf("expected budget exceeded before window reset")
	}

	// Advance past the window.
	b.nowFn = func() time.Time { return now.Add(2 * time.Minute) }
	if err := b.Consume("secret/db"); err != nil {
		t.Fatalf("expected budget reset after window, got %v", err)
	}
}

func TestRemaining_DecreasesOnConsume(t *testing.T) {
	b := New(5, time.Minute)
	if r := b.Remaining("secret/aws"); r != 5 {
		t.Fatalf("expected 5 remaining, got %d", r)
	}
	_ = b.Consume("secret/aws")
	_ = b.Consume("secret/aws")
	if r := b.Remaining("secret/aws"); r != 3 {
		t.Fatalf("expected 3 remaining, got %d", r)
	}
}

func TestReset_ClearsBudget(t *testing.T) {
	b := New(1, time.Minute)
	_ = b.Consume("secret/db")
	b.Reset("secret/db")
	if err := b.Consume("secret/db"); err != nil {
		t.Fatalf("expected budget cleared after reset, got %v", err)
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	now := time.Now()
	b := New(2, time.Minute)
	b.nowFn = func() time.Time { return now }
	_ = b.Consume("secret/db")

	// Advance past window so entry is stale.
	b.nowFn = func() time.Time { return now.Add(2 * time.Minute) }
	b.Purge()

	b.mu.Lock()
	_, exists := b.entries["secret/db"]
	b.mu.Unlock()
	if exists {
		t.Fatal("expected expired entry to be purged")
	}
}

func TestConsume_IndependentPaths(t *testing.T) {
	b := New(1, time.Minute)
	_ = b.Consume("secret/db")
	if err := b.Consume("secret/cache"); err != nil {
		t.Fatalf("expected independent path to have its own budget, got %v", err)
	}
}
