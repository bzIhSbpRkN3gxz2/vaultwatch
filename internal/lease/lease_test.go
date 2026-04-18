package lease

import (
	"testing"
	"time"
)

func newLease(expireOffset time.Duration) *Lease {
	now := time.Now()
	return &Lease{
		ID:          "test-lease-id",
		Path:        "database/creds/my-role",
		Renewable:   true,
		TTL:         time.Hour,
		CreatedTime: now,
		ExpireTime:  now.Add(expireOffset),
	}
}

func TestLeaseStatus_Healthy(t *testing.T) {
	l := newLease(2 * time.Hour)
	if got := l.Status(30 * time.Minute); got != StatusHealthy {
		t.Errorf("expected healthy, got %s", got)
	}
}

func TestLeaseStatus_Expiring(t *testing.T) {
	l := newLease(10 * time.Minute)
	if got := l.Status(30 * time.Minute); got != StatusExpiring {
		t.Errorf("expected expiring, got %s", got)
	}
}

func TestLeaseStatus_Expired(t *testing.T) {
	l := newLease(-1 * time.Minute)
	if got := l.Status(30 * time.Minute); got != StatusExpired {
		t.Errorf("expected expired, got %s", got)
	}
}

func TestLeaseStatus_Orphaned(t *testing.T) {
	l := &Lease{ID: "orphan", Path: "secret/orphan"}
	if got := l.Status(30 * time.Minute); got != StatusOrphaned {
		t.Errorf("expected orphaned, got %s", got)
	}
}

func TestTimeRemaining(t *testing.T) {
	l := newLease(5 * time.Minute)
	remaining := l.TimeRemaining()
	if remaining <= 0 || remaining > 5*time.Minute {
		t.Errorf("unexpected time remaining: %v", remaining)
	}
}

func TestTimeRemaining_Expired(t *testing.T) {
	l := newLease(-1 * time.Minute)
	if got := l.TimeRemaining(); got != 0 {
		t.Errorf("expected 0 for expired lease, got %v", got)
	}
}
