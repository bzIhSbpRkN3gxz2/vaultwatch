package limiter_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/limiter"
)

func TestAllow_ConsumesTokens(t *testing.T) {
	l := limiter.New(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !l.Allow() {
			t.Fatalf("expected Allow()=true on call %d", i+1)
		}
	}
	if l.Allow() {
		t.Fatal("expected Allow()=false after exhausting tokens")
	}
}

func TestRemaining_DecreasesOnAllow(t *testing.T) {
	l := limiter.New(5, time.Minute)
	if got := l.Remaining(); got != 5 {
		t.Fatalf("expected 5 remaining, got %d", got)
	}
	l.Allow()
	l.Allow()
	if got := l.Remaining(); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
}

func TestReset_RestoresTokens(t *testing.T) {
	l := limiter.New(2, time.Minute)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected no tokens before reset")
	}
	l.Reset()
	if !l.Allow() {
		t.Fatal("expected Allow()=true after reset")
	}
}

func TestRefill_AfterInterval(t *testing.T) {
	l := limiter.New(2, 50*time.Millisecond)
	l.Allow()
	l.Allow()
	if l.Allow() {
		t.Fatal("expected tokens exhausted")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow() {
		t.Fatal("expected tokens refilled after interval")
	}
}

func TestAllow_ZeroMax(t *testing.T) {
	l := limiter.New(0, time.Minute)
	if l.Allow() {
		t.Fatal("expected Allow()=false with max=0")
	}
}
