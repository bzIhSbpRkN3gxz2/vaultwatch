package rollup_test

import (
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/rollup"
)

func TestSummary_EmptyBatch(t *testing.T) {
	b := rollup.Batch{FlushedAt: time.Now()}
	out := rollup.Summary(b)
	if out != "no events in batch" {
		t.Fatalf("unexpected summary: %s", out)
	}
}

func TestSummary_CountsStatuses(t *testing.T) {
	b := rollup.Batch{
		FlushedAt: time.Now(),
		Events: []rollup.Event{
			{Lease: newLease("a"), QueuedAt: time.Now()},
			{Lease: newLease("b"), QueuedAt: time.Now()},
		},
	}
	out := rollup.Summary(b)
	if !strings.Contains(out, "batch(2 events)") {
		t.Fatalf("expected event count in summary, got: %s", out)
	}
	if !strings.Contains(out, "flushed_at=") {
		t.Fatalf("expected flushed_at in summary, got: %s", out)
	}
}
