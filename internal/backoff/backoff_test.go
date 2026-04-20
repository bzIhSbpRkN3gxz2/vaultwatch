package backoff_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/backoff"
)

func TestDefaultStrategy_Fields(t *testing.T) {
	s := backoff.DefaultStrategy()
	if s.Initial != 200*time.Millisecond {
		t.Fatalf("expected 200ms initial, got %v", s.Initial)
	}
	if s.Multiplier != 2.0 {
		t.Fatalf("expected multiplier 2.0, got %v", s.Multiplier)
	}
	if s.MaxDelay != 30*time.Second {
		t.Fatalf("expected 30s max delay, got %v", s.MaxDelay)
	}
	if !s.Jitter {
		t.Fatal("expected jitter to be enabled by default")
	}
}

func TestDelay_IncreasesWithAttempt(t *testing.T) {
	s := backoff.Strategy{
		Initial:    100 * time.Millisecond,
		Multiplier: 2.0,
		MaxDelay:   10 * time.Second,
		Jitter:     false,
	}
	prev := s.Delay(0)
	for attempt := 1; attempt <= 5; attempt++ {
		curr := s.Delay(attempt)
		if curr <= prev {
			t.Fatalf("attempt %d: expected delay %v > %v", attempt, curr, prev)
		}
		prev = curr
	}
}

func TestDelay_ClampsAtMaxDelay(t *testing.T) {
	s := backoff.Strategy{
		Initial:    1 * time.Second,
		Multiplier: 10.0,
		MaxDelay:   5 * time.Second,
		Jitter:     false,
	}
	for _, attempt := range []int{3, 5, 10, 50} {
		d := s.Delay(attempt)
		if d > s.MaxDelay {
			t.Fatalf("attempt %d: delay %v exceeds MaxDelay %v", attempt, d, s.MaxDelay)
		}
	}
}

func TestDelay_NegativeAttemptTreatedAsZero(t *testing.T) {
	s := backoff.Strategy{
		Initial:    500 * time.Millisecond,
		Multiplier: 2.0,
		MaxDelay:   10 * time.Second,
		Jitter:     false,
	}
	if s.Delay(-1) != s.Delay(0) {
		t.Fatal("negative attempt should produce the same delay as attempt 0")
	}
}

func TestReset_EqualsDelayZero(t *testing.T) {
	s := backoff.Strategy{
		Initial:    300 * time.Millisecond,
		Multiplier: 1.5,
		MaxDelay:   10 * time.Second,
		Jitter:     false,
	}
	if s.Reset() != s.Delay(0) {
		t.Fatalf("Reset() %v != Delay(0) %v", s.Reset(), s.Delay(0))
	}
}

func TestDelay_JitterAddsVariance(t *testing.T) {
	s := backoff.Strategy{
		Initial:    1 * time.Second,
		Multiplier: 1.0, // constant base so only jitter varies
		MaxDelay:   10 * time.Second,
		Jitter:     true,
	}
	seen := make(map[time.Duration]bool)
	for i := 0; i < 20; i++ {
		seen[s.Delay(0)] = true
	}
	// With jitter we expect more than one distinct value across 20 samples.
	if len(seen) < 2 {
		t.Fatal("jitter enabled but all delays were identical")
	}
}
