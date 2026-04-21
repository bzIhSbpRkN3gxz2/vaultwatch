package circuit_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/circuit"
)

func defaultRegistry() *circuit.Registry {
	cfg := circuit.Config{
		FailureThreshold: 2,
		HalfOpenTimeout:  50 * time.Millisecond,
	}
	return circuit.New(cfg)
}

func TestAllow_ClosedByDefault(t *testing.T) {
	r := defaultRegistry()
	if err := r.Allow("lease-1"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestFailure_OpensAfterThreshold(t *testing.T) {
	r := defaultRegistry()
	r.Failure("lease-1")
	r.Failure("lease-1")
	if err := r.Allow("lease-1"); err == nil {
		t.Fatal("expected ErrCircuitOpen after threshold failures")
	}
}

func TestSuccess_ResetsClosed(t *testing.T) {
	r := defaultRegistry()
	r.Failure("lease-2")
	r.Success("lease-2")
	if err := r.Allow("lease-2"); err != nil {
		t.Fatalf("expected circuit closed after success, got %v", err)
	}
}

func TestHalfOpen_AllowsAfterTimeout(t *testing.T) {
	r := defaultRegistry()
	r.Failure("lease-3")
	r.Failure("lease-3")
	time.Sleep(60 * time.Millisecond)
	if err := r.Allow("lease-3"); err != nil {
		t.Fatalf("expected half-open allow after timeout, got %v", err)
	}
}

func TestIsolation_SeparateLeases(t *testing.T) {
	r := defaultRegistry()
	r.Failure("lease-a")
	r.Failure("lease-a")
	if err := r.Allow("lease-b"); err != nil {
		t.Fatalf("lease-b should be unaffected by lease-a failures, got %v", err)
	}
}

func TestRemove_DeletesEntry(t *testing.T) {
	r := defaultRegistry()
	r.Failure("lease-x")
	r.Failure("lease-x")
	r.Remove("lease-x")
	if n := r.Len(); n != 0 {
		t.Fatalf("expected 0 entries after remove, got %d", n)
	}
	if err := r.Allow("lease-x"); err != nil {
		t.Fatalf("removed lease should have fresh breaker, got %v", err)
	}
}

func TestDefaultConfig_Values(t *testing.T) {
	cfg := circuit.DefaultConfig()
	if cfg.FailureThreshold <= 0 {
		t.Errorf("FailureThreshold should be positive, got %d", cfg.FailureThreshold)
	}
	if cfg.HalfOpenTimeout <= 0 {
		t.Errorf("HalfOpenTimeout should be positive, got %v", cfg.HalfOpenTimeout)
	}
}
