package expiry_test

import (
	"errors"
	"testing"
	"time"

	"github.com/vaultwatch/internal/expiry"
	"github.com/vaultwatch/internal/lease"
)

func newLease(ttl time.Duration, createdAt time.Time) *lease.Lease {
	l := lease.New("lease-1", "secret/data/db")
	l.TTL = ttl
	l.CreatedAt = createdAt
	return l
}

func TestEvaluate_Healthy(t *testing.T) {
	now := time.Now()
	c := expiry.New(5 * time.Minute)
	l := newLease(30*time.Minute, now)

	r, err := c.Evaluate(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Expired {
		t.Error("expected not expired")
	}
	if r.Warn {
		t.Error("expected no warning for healthy lease")
	}
	if r.Remaining <= 0 {
		t.Error("expected positive remaining duration")
	}
}

func TestEvaluate_WarnWindow(t *testing.T) {
	now := time.Now()
	c := expiry.New(5 * time.Minute)
	// Lease created 26 minutes ago with a 30-minute TTL → 4 min remaining
	l := newLease(30*time.Minute, now.Add(-26*time.Minute))

	r, err := c.Evaluate(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Warn {
		t.Error("expected warn=true within warning horizon")
	}
	if r.Expired {
		t.Error("expected not expired")
	}
}

func TestEvaluate_Expired(t *testing.T) {
	now := time.Now()
	c := expiry.New(5 * time.Minute)
	// Lease created 35 minutes ago with a 30-minute TTL → already expired
	l := newLease(30*time.Minute, now.Add(-35*time.Minute))

	r, err := c.Evaluate(l)
	if !errors.Is(err, expiry.ErrExpired) {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
	if !r.Expired {
		t.Error("expected expired=true")
	}
}

func TestEvaluate_NoTTL(t *testing.T) {
	c := expiry.New(5 * time.Minute)
	l := newLease(0, time.Now())

	_, err := c.Evaluate(l)
	if !errors.Is(err, expiry.ErrNoExpiry) {
		t.Fatalf("expected ErrNoExpiry, got %v", err)
	}
}

func TestEvaluate_NilLease(t *testing.T) {
	c := expiry.New(5 * time.Minute)
	_, err := c.Evaluate(nil)
	if !errors.Is(err, expiry.ErrNoExpiry) {
		t.Fatalf("expected ErrNoExpiry for nil lease, got %v", err)
	}
}

func TestTimeUntilWarn(t *testing.T) {
	now := time.Now()
	c := expiry.New(5 * time.Minute)
	// 20 min remaining → warn triggers at 5 min → 15 min until warn
	l := newLease(30*time.Minute, now.Add(-10*time.Minute))

	d, err := c.TimeUntilWarn(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const tolerance = 2 * time.Second
	expected := 15 * time.Minute
	if d < expected-tolerance || d > expected+tolerance {
		t.Errorf("TimeUntilWarn = %v, want ~%v", d, expected)
	}
}

func TestNew_DefaultWarnBefore(t *testing.T) {
	// Zero or negative warnBefore should fall back to 5 minutes
	c := expiry.New(0)
	now := time.Now()
	// 4 min remaining → should warn with default 5-minute horizon
	l := newLease(30*time.Minute, now.Add(-26*time.Minute))

	r, err := c.Evaluate(l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Warn {
		t.Error("expected warn=true with default horizon")
	}
}
