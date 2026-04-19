package retrier_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/retrier"
)

func fastPolicy() retrier.Policy {
	return retrier.Policy{
		MaxAttempts: 3,
		BaseDelay:   time.Millisecond,
		MaxDelay:    5 * time.Millisecond,
		Multiplier:  2.0,
	}
}

func TestDo_SucceedsFirstAttempt(t *testing.T) {
	r := retrier.New(fastPolicy())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	r := retrier.New(fastPolicy())
	calls := 0
	err := r.Do(context.Background(), func() error {
		calls++
		if calls < 3 {
			return errors.New("transient")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	r := retrier.New(fastPolicy())
	sentinel := errors.New("permanent")
	err := r.Do(context.Background(), func() error { return sentinel })
	if !errors.Is(err, retrier.ErrMaxAttempts) {
		t.Fatalf("expected ErrMaxAttempts, got %v", err)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected wrapped sentinel, got %v", err)
	}
}

func TestDo_RespectsContextCancel(t *testing.T) {
	r := retrier.New(fastPolicy())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := r.Do(ctx, func() error { return nil })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestDefaultPolicy_Values(t *testing.T) {
	p := retrier.DefaultPolicy()
	if p.MaxAttempts != 4 {
		t.Fatalf("expected 4 attempts, got %d", p.MaxAttempts)
	}
	if p.Multiplier != 2.0 {
		t.Fatalf("expected multiplier 2.0, got %f", p.Multiplier)
	}
}
