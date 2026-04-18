package alert_test

import (
	"errors"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
	"github.com/user/vaultwatch/internal/lease"
)

type stubHandler struct {
	called bool
	err    error
}

func (s *stubHandler) OnAlert(_ *lease.Lease) error {
	s.called = true
	return s.err
}

func TestMultiHandler_AllCalled(t *testing.T) {
	a, b := &stubHandler{}, &stubHandler{}
	m := alert.NewMultiHandler(a, b)
	l := newTestLease("lease-1", time.Now().Add(10*time.Minute))

	if err := m.OnAlert(l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called || !b.called {
		t.Error("expected both handlers to be called")
	}
}

func TestMultiHandler_ReturnsLastError(t *testing.T) {
	errA := errors.New("handler A failed")
	errB := errors.New("handler B failed")
	a := &stubHandler{err: errA}
	b := &stubHandler{err: errB}
	m := alert.NewMultiHandler(a, b)
	l := newTestLease("lease-2", time.Now().Add(10*time.Minute))

	err := m.OnAlert(l)
	if !errors.Is(err, errB) {
		t.Errorf("expected errB, got %v", err)
	}
}

func TestMultiHandler_ContinuesOnError(t *testing.T) {
	a := &stubHandler{err: errors.New("fail")}
	b := &stubHandler{}
	m := alert.NewMultiHandler(a, b)
	l := newTestLease("lease-3", time.Now().Add(10*time.Minute))

	m.OnAlert(l) //nolint:errcheck
	if !b.called {
		t.Error("expected second handler to be called despite first error")
	}
}
