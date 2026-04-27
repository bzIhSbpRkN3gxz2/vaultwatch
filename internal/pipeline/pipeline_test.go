package pipeline_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
	"github.com/yourusername/vaultwatch/internal/pipeline"
)

func newLease(id, path string, ttl time.Duration) *lease.Lease {
	return lease.New(id, path, ttl)
}

func TestPipeline_PassesLeaseThrough(t *testing.T) {
	var received *lease.Lease

	p := pipeline.New(
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			return l, nil
		}),
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			received = l
			return l, nil
		}),
	)

	input := newLease("lease-1", "secret/db", 10*time.Minute)
	_, err := p.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received == nil || received.ID != input.ID {
		t.Errorf("expected lease %q to pass through, got %v", input.ID, received)
	}
}

func TestPipeline_StageError_HaltsExecution(t *testing.T) {
	sentinel := errors.New("stage failure")
	called := false

	p := pipeline.New(
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			return nil, sentinel
		}),
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			called = true
			return l, nil
		}),
	)

	_, err := p.Run(context.Background(), newLease("lease-2", "secret/api", 5*time.Minute))
	if !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
	if called {
		t.Error("expected subsequent stage to not be called after error")
	}
}

func TestPipeline_ContextCancellation_StopsExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ran := false
	p := pipeline.New(
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			ran = true
			return l, ctx.Err()
		}),
	)

	_, err := p.Run(ctx, newLease("lease-3", "secret/svc", 2*time.Minute))
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if !ran {
		t.Error("expected stage to run and detect cancellation")
	}
}

func TestPipeline_NoStages_ReturnsLeaseUnmodified(t *testing.T) {
	p := pipeline.New()
	input := newLease("lease-4", "secret/empty", 30*time.Second)
	out, err := p.Run(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == nil || out.ID != input.ID {
		t.Errorf("expected unmodified lease, got %v", out)
	}
}

func TestPipeline_MultipleLeases_IndependentRuns(t *testing.T) {
	var seen []string

	p := pipeline.New(
		pipeline.StageFunc(func(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
			seen = append(seen, l.ID)
			return l, nil
		}),
	)

	leases := []*lease.Lease{
		newLease("a", "secret/a", time.Minute),
		newLease("b", "secret/b", time.Minute),
		newLease("c", "secret/c", time.Minute),
	}

	for _, l := range leases {
		if _, err := p.Run(context.Background(), l); err != nil {
			t.Fatalf("unexpected error for lease %q: %v", l.ID, err)
		}
	}

	if len(seen) != 3 {
		t.Errorf("expected 3 leases processed, got %d", len(seen))
	}
}
