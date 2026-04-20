package rotation_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/lease"
	"github.com/vaultwatch/internal/rotation"
)

func newLease(id string) *lease.Lease {
	return lease.New(id, "secret/data/test", 300)
}

func TestBegin_SetsActiveStatus(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-1")

	if err := tr.Begin(l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	r, ok := tr.Get(l.LeaseID)
	if !ok {
		t.Fatal("expected record to exist")
	}
	if r.Status != rotation.StatusActive {
		t.Errorf("expected StatusActive, got %v", r.Status)
	}
}

func TestBegin_ReturnsErrAlreadyRotating(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-2")

	_ = tr.Begin(l)
	err := tr.Begin(l)

	if err != rotation.ErrAlreadyRotating {
		t.Errorf("expected ErrAlreadyRotating, got %v", err)
	}
}

func TestComplete_SetsCompleteStatus(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-3")
	_ = tr.Begin(l)

	tr.Complete(l.LeaseID)

	r, _ := tr.Get(l.LeaseID)
	if r.Status != rotation.StatusComplete {
		t.Errorf("expected StatusComplete, got %v", r.Status)
	}
	if r.EndedAt.IsZero() {
		t.Error("expected EndedAt to be set")
	}
}

func TestFail_SetsFailedStatus(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-4")
	_ = tr.Begin(l)

	tr.Fail(l.LeaseID, "vault unreachable")

	r, _ := tr.Get(l.LeaseID)
	if r.Status != rotation.StatusFailed {
		t.Errorf("expected StatusFailed, got %v", r.Status)
	}
	if r.Error != "vault unreachable" {
		t.Errorf("unexpected error message: %q", r.Error)
	}
}

func TestGet_Missing(t *testing.T) {
	tr := rotation.New()
	_, ok := tr.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for missing lease")
	}
}

func TestPurge_RemovesOldRecords(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-5")
	_ = tr.Begin(l)
	tr.Complete(l.LeaseID)

	tr.Purge(time.Now().Add(time.Second))

	_, ok := tr.Get(l.LeaseID)
	if ok {
		t.Error("expected record to be purged")
	}
}

func TestPurge_KeepsActiveRecords(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-6")
	_ = tr.Begin(l)

	tr.Purge(time.Now().Add(time.Second))

	_, ok := tr.Get(l.LeaseID)
	if !ok {
		t.Error("expected active record to survive purge")
	}
}

func TestBegin_AllowsAfterComplete(t *testing.T) {
	tr := rotation.New()
	l := newLease("lease-7")
	_ = tr.Begin(l)
	tr.Complete(l.LeaseID)

	if err := tr.Begin(l); err != nil {
		t.Errorf("expected no error after completion, got %v", err)
	}
}
