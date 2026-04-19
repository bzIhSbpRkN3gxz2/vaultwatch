package notify

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lease"
)

type mockHandler struct {
	called []string
	err    error
}

func (m *mockHandler) OnAlert(_ context.Context, l *lease.Lease) error {
	m.called = append(m.called, l.ID)
	return m.err
}

func newDispatchLease(id string) *lease.Lease {
	return lease.New(id, "secret/data/test", time.Now().Add(5*time.Minute), 300)
}

func TestDispatcher_Dispatch_Sends(t *testing.T) {
	h := &mockHandler{}
	th := NewThrottle(30 * time.Second)
	d := NewDispatcher(h, th, nil)

	l := newDispatchLease("lease-1")
	ok, err := d.Dispatch(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected alert to be dispatched")
	}
	if len(h.called) != 1 || h.called[0] != "lease-1" {
		t.Fatalf("expected handler called with lease-1, got %v", h.called)
	}
}

func TestDispatcher_Dispatch_Throttled(t *testing.T) {
	h := &mockHandler{}
	th := NewThrottle(30 * time.Second)
	d := NewDispatcher(h, th, nil)

	l := newDispatchLease("lease-2")
	d.Dispatch(context.Background(), l) //nolint
	ok, err := d.Dispatch(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected alert to be suppressed")
	}
	if len(h.called) != 1 {
		t.Fatalf("handler should be called once, got %d", len(h.called))
	}
}

func TestDispatcher_Dispatch_HandlerError(t *testing.T) {
	h := &mockHandler{err: errors.New("send failed")}
	th := NewThrottle(30 * time.Second)
	d := NewDispatcher(h, th, nil)

	l := newDispatchLease("lease-3")
	_, err := d.Dispatch(context.Background(), l)
	if err == nil {
		t.Fatal("expected error from handler")
	}
}

func TestDispatcher_DispatchAll_CountsDispatched(t *testing.T) {
	h := &mockHandler{}
	th := NewThrottle(30 * time.Second)
	d := NewDispatcher(h, th, nil)

	leases := []*lease.Lease{
		newDispatchLease("a"),
		newDispatchLease("b"),
		newDispatchLease("c"),
	}
	count, err := d.DispatchAll(context.Background(), leases)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 dispatched, got %d", count)
	}
}
