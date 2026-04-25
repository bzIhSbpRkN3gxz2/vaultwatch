package staleness

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/lease"
)

func TestGuard_NoAlert_BelowThreshold(t *testing.T) {
	now := time.Now()
	tr := New(10 * time.Second)
	tr.now = func() time.Time { return now }

	alerted := false
	g := NewGuard(tr, func(_ context.Context, _ *lease.Lease, _ error) error {
		alerted = true
		return nil
	})

	l := newLease("x", lease.StatusExpiring)
	if err := g.Check(context.Background(), l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alerted {
		t.Fatal("should not have alerted below threshold")
	}
}

func TestGuard_Alert_AboveThreshold(t *testing.T) {
	now := time.Now()
	tr := New(5 * time.Second)
	tr.now = func() time.Time { return now }

	l := newLease("x", lease.StatusExpiring)
	_ = tr.Observe(l)

	tr.now = func() time.Time { return now.Add(10 * time.Second) }

	alerted := false
	g := NewGuard(tr, func(_ context.Context, _ *lease.Lease, err error) error {
		alerted = true
		return err
	})

	err := g.Check(context.Background(), l)
	if !alerted {
		t.Fatal("expected alert to be triggered")
	}
	if !errors.Is(err, ErrThresholdExceeded) {
		t.Fatalf("expected ErrThresholdExceeded, got %v", err)
	}
}

func TestGuard_PropagatesOnStaleError(t *testing.T) {
	now := time.Now()
	tr := New(1 * time.Second)
	tr.now = func() time.Time { return now }

	l := newLease("y", lease.StatusOrphaned)
	_ = tr.Observe(l)

	tr.now = func() time.Time { return now.Add(5 * time.Second) }

	sentinel := errors.New("downstream failure")
	g := NewGuard(tr, func(_ context.Context, _ *lease.Lease, _ error) error {
		return sentinel
	})

	if err := g.Check(context.Background(), l); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
