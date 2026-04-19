package circuitbreaker_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/circuitbreaker"
)

func TestAllow_ClosedByDefault(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestFailure_OpensAfterThreshold(t *testing.T) {
	b := circuitbreaker.New(3, time.Second)
	for i := 0; i < 3; i++ {
		b.Failure()
	}
	if b.CurrentState() != circuitbreaker.StateOpen {
		t.Fatal("expected circuit to be open")
	}
	if err := b.Allow(); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestSuccess_ResetsClosed(t *testing.T) {
	b := circuitbreaker.New(2, time.Second)
	b.Failure()
	b.Failure()
	b.Success()
	if b.CurrentState() != circuitbreaker.StateClosed {
		t.Fatal("expected circuit to be closed after success")
	}
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil after reset, got %v", err)
	}
}

func TestHalfOpen_AfterTimeout(t *testing.T) {
	b := circuitbreaker.New(1, 10*time.Millisecond)
	b.Failure()
	if b.CurrentState() != circuitbreaker.StateOpen {
		t.Fatal("expected open")
	}
	time.Sleep(20 * time.Millisecond)
	if err := b.Allow(); err != nil {
		t.Fatalf("expected nil in half-open, got %v", err)
	}
	if b.CurrentState() != circuitbreaker.StateHalfOpen {
		t.Fatal("expected half-open state")
	}
}

func TestFailure_BelowThreshold_StaysClosed(t *testing.T) {
	b := circuitbreaker.New(5, time.Second)
	b.Failure()
	b.Failure()
	if b.CurrentState() != circuitbreaker.StateClosed {
		t.Fatal("expected closed below threshold")
	}
}
