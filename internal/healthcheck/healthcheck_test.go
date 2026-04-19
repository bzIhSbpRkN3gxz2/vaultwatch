package healthcheck

import (
	"context"
	"errors"
	"testing"
)

func TestRegister_And_RunAll(t *testing.T) {
	r := New()
	_ = r.Register("vault", func(_ context.Context) error { return nil })

	statuses := r.RunAll(context.Background())
	if len(statuses) != 1 {
		t.Fatalf("expected 1 status, got %d", len(statuses))
	}
	if !statuses[0].Healthy {
		t.Errorf("expected healthy")
	}
	if statuses[0].Message != "ok" {
		t.Errorf("expected message 'ok', got %q", statuses[0].Message)
	}
}

func TestRegister_UnhealthyChecker(t *testing.T) {
	r := New()
	_ = r.Register("db", func(_ context.Context) error {
		return errors.New("connection refused")
	})

	statuses := r.RunAll(context.Background())
	if statuses[0].Healthy {
		t.Errorf("expected unhealthy")
	}
	if statuses[0].Message != "connection refused" {
		t.Errorf("unexpected message: %s", statuses[0].Message)
	}
}

func TestRegister_EmptyName(t *testing.T) {
	r := New()
	err := r.Register("", func(_ context.Context) error { return nil })
	if err == nil {
		t.Error("expected error for empty name")
	}
}

func TestRegister_NilChecker(t *testing.T) {
	r := New()
	err := r.Register("x", nil)
	if err == nil {
		t.Error("expected error for nil checker")
	}
}

func TestHealthy_AllPass(t *testing.T) {
	r := New()
	_ = r.Register("a", func(_ context.Context) error { return nil })
	_ = r.Register("b", func(_ context.Context) error { return nil })

	if !r.Healthy(context.Background()) {
		t.Error("expected healthy")
	}
}

func TestHealthy_OneFails(t *testing.T) {
	r := New()
	_ = r.Register("a", func(_ context.Context) error { return nil })
	_ = r.Register("b", func(_ context.Context) error { return errors.New("fail") })

	if r.Healthy(context.Background()) {
		t.Error("expected unhealthy when one checker fails")
	}
}

func TestRunAll_Empty(t *testing.T) {
	r := New()
	if got := r.RunAll(context.Background()); len(got) != 0 {
		t.Errorf("expected empty results, got %d", len(got))
	}
}
