package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestNew_InvalidRate(t *testing.T) {
	_, err := New(0)
	if err == nil {
		t.Fatal("expected error for zero rate")
	}
	_, err = New(-5)
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestNew_ValidRate(t *testing.T) {
	l, err := New(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil limiter")
	}
}

func TestWait_ConsumesToken(t *testing.T) {
	l, _ := New(10)
	ctx := context.Background()
	before := l.Available()
	if err := l.Wait(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := l.Available()
	if after >= before {
		t.Errorf("expected tokens to decrease: before=%.2f after=%.2f", before, after)
	}
}

func TestWait_CancelledContext(t *testing.T) {
	l, _ := New(0.001) // very slow refill
	// drain all tokens
	l.mu.Lock()
	l.tokens = 0
	l.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := l.Wait(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestAvailable_RefillsOverTime(t *testing.T) {
	l, _ := New(100)
	l.mu.Lock()
	l.tokens = 0
	l.lastTick = time.Now().Add(-500 * time.Millisecond)
	l.mu.Unlock()

	avail := l.Available()
	if avail < 40 {
		t.Errorf("expected ~50 tokens after 500ms at rate 100/s, got %.2f", avail)
	}
}

func TestAvailable_CapsAtMax(t *testing.T) {
	l, _ := New(10)
	l.mu.Lock()
	l.tokens = 0
	l.lastTick = time.Now().Add(-60 * time.Second)
	l.mu.Unlock()

	if avail := l.Available(); avail > 10 {
		t.Errorf("tokens should not exceed max: got %.2f", avail)
	}
}
