package sampler

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/lease"
)

type recordingHandler struct {
	calls int
	err   error
}

func (r *recordingHandler) OnAlert(_ *lease.Lease) error {
	r.calls++
	return r.err
}

func newLease() *lease.Lease {
	return lease.New("lease-abc", "secret/data/db", 300*time.Second)
}

func TestMiddleware_PassesWhenRateOne(t *testing.T) {
	h := &recordingHandler{}
	m := NewMiddleware(New(1.0), h)
	for i := 0; i < 10; i++ {
		if err := m.OnAlert(newLease()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if h.calls != 10 {
		t.Fatalf("expected 10 calls, got %d", h.calls)
	}
}

func TestMiddleware_DropsWhenRateZero(t *testing.T) {
	h := &recordingHandler{}
	m := NewMiddleware(New(0.0), h)
	for i := 0; i < 10; i++ {
		_ = m.OnAlert(newLease())
	}
	if h.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", h.calls)
	}
}

func TestMiddleware_PropagatesError(t *testing.T) {
	expected := errors.New("handler failed")
	h := &recordingHandler{err: expected}
	m := NewMiddleware(New(1.0), h)
	err := m.OnAlert(newLease())
	if !errors.Is(err, expected) {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
