package rotation_test

import (
	"context"
	"errors"
	"testing"

	"github.com/vaultwatch/internal/lease"
	"github.com/vaultwatch/internal/rotation"
)

func TestGuard_CompletesOnSuccess(t *testing.T) {
	tr := rotation.New()
	l := lease.New("g-lease-1", "secret/db", 120)

	called := false
	fn := rotation.Guard(tr, func(ctx context.Context, l *lease.Lease) error {
		called = true
		return nil
	})

	if err := fn(context.Background(), l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected inner function to be called")
	}

	r, ok := tr.Get(l.LeaseID)
	if !ok {
		t.Fatal("expected record")
	}
	if r.Status != rotation.StatusComplete {
		t.Errorf("expected StatusComplete, got %v", r.Status)
	}
}

func TestGuard_FailsOnError(t *testing.T) {
	tr := rotation.New()
	l := lease.New("g-lease-2", "secret/db", 120)
	rotErr := errors.New("vault error")

	fn := rotation.Guard(tr, func(ctx context.Context, l *lease.Lease) error {
		return rotErr
	})

	err := fn(context.Background(), l)
	if !errors.Is(err, rotErr) {
		t.Errorf("expected rotErr, got %v", err)
	}

	r, _ := tr.Get(l.LeaseID)
	if r.Status != rotation.StatusFailed {
		t.Errorf("expected StatusFailed, got %v", r.Status)
	}
	if r.Error != rotErr.Error() {
		t.Errorf("unexpected error message: %q", r.Error)
	}
}

func TestGuard_BlocksConcurrentRotation(t *testing.T) {
	tr := rotation.New()
	l := lease.New("g-lease-3", "secret/db", 120)

	// Manually begin so the guard sees an active rotation.
	_ = tr.Begin(l)

	fn := rotation.Guard(tr, func(ctx context.Context, l *lease.Lease) error {
		return nil
	})

	err := fn(context.Background(), l)
	if !errors.Is(err, rotation.ErrAlreadyRotating) {
		t.Errorf("expected ErrAlreadyRotating, got %v", err)
	}
}
