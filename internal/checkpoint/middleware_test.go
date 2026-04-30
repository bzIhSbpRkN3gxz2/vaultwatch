package checkpoint_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/checkpoint"
	"github.com/yourusername/vaultwatch/internal/lease"
)

func newLease(id, path string, ttl time.Duration) *lease.Lease {
	return lease.New(id, path, ttl)
}

func TestWithContext_And_FromContext(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	ctx := checkpoint.WithContext(context.Background(), cp)
	got := checkpoint.FromContext(ctx)
	if got != cp {
		t.Error("expected same checkpoint from context")
	}
}

func TestFromContext_Missing(t *testing.T) {
	got := checkpoint.FromContext(context.Background())
	if got != nil {
		t.Error("expected nil when no checkpoint in context")
	}
}

func TestStage_RecordsLease(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	ctx := checkpoint.WithContext(context.Background(), cp)
	l := newLease("stage-1", "secret/db", 30*time.Minute)

	out, err := checkpoint.Stage(ctx, l)
	if err != nil {
		t.Fatalf("Stage: %v", err)
	}
	if out != l {
		t.Error("Stage should return the same lease pointer")
	}
	e, ok := cp.Get("stage-1")
	if !ok {
		t.Fatal("expected lease to be recorded in checkpoint")
	}
	if e.Path != "secret/db" {
		t.Errorf("path: got %q, want %q", e.Path, "secret/db")
	}
}

func TestStage_NoCheckpoint_IsNoop(t *testing.T) {
	l := newLease("noop", "secret/x", time.Hour)
	out, err := checkpoint.Stage(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != l {
		t.Error("expected original lease returned")
	}
}

func TestStage_NilLease_ReturnsError(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	ctx := checkpoint.WithContext(context.Background(), cp)
	_, err := checkpoint.Stage(ctx, nil)
	if err == nil {
		t.Error("expected error for nil lease")
	}
}
