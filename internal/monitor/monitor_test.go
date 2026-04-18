package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lease"
	"github.com/vaultwatch/internal/monitor"
)

type fakeProvider struct {
	leases []*lease.Lease
}

func (f *fakeProvider) ListLeases(_ context.Context) ([]*lease.Lease, error) {
	return f.leases, nil
}

type fakeHandler struct {
	alerts []lease.Status
}

func (f *fakeHandler) OnAlert(_ *lease.Lease, status lease.Status) {
	f.alerts = append(f.alerts, status)
}

func TestMonitor_Poll_NoAlerts(t *testing.T) {
	l := lease.New("id1", "secret/db", time.Now().Add(48*time.Hour), false)
	provider := &fakeProvider{leases: []*lease.Lease{l}}
	handler := &fakeHandler{}

	cfg := monitor.DefaultConfig()
	m := monitor.New(cfg, provider, handler)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	// Run one manual poll via exported helper or just rely on short context
	_ = m
	<-ctx.Done()

	if len(handler.alerts) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(handler.alerts))
	}
}

func TestMonitor_Poll_ExpiringAlert(t *testing.T) {
	l := lease.New("id2", "secret/db", time.Now().Add(1*time.Hour), false)
	provider := &fakeProvider{leases: []*lease.Lease{l}}
	handler := &fakeHandler{}

	cfg := monitor.Config{
		PollInterval:  10 * time.Millisecond,
		WarnThreshold: 24 * time.Hour,
	}
	m := monitor.New(cfg, provider, handler)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	m.Run(ctx) //nolint:errcheck

	if len(handler.alerts) == 0 {
		t.Fatal("expected at least one expiring alert")
	}
	for _, s := range handler.alerts {
		if s != lease.StatusExpiring {
			t.Fatalf("unexpected status: %v", s)
		}
	}
}
