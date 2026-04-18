package filter_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/filter"
	"github.com/vaultwatch/internal/lease"
)

func newLease(path, id string, ttl int, renewable bool) *lease.Lease {
	return lease.New(lease.Options{
		LeaseID:   id,
		Path:      path,
		TTL:       ttl,
		Renewable: renewable,
		IssuedAt:  time.Now(),
	})
}

func TestFilter_ByPathPrefix(t *testing.T) {
	leases := []*lease.Lease{
		newLease("secret/db", "id1", 3600, true),
		newLease("aws/creds", "id2", 3600, true),
		newLease("secret/api", "id3", 3600, true),
	}
	got := filter.Filter(leases, filter.Options{PathPrefix: "secret/"})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilter_ByStatus(t *testing.T) {
	leases := []*lease.Lease{
		newLease("secret/db", "id1", 3600, true),  // healthy
		newLease("secret/db", "", 3600, false),    // orphaned
		newLease("secret/db", "id3", 200, true),   // expiring
	}
	got := filter.Filter(leases, filter.Options{Statuses: []lease.Status{lease.StatusOrphaned}})
	if len(got) != 1 {
		t.Fatalf("expected 1 orphaned, got %d", len(got))
	}
}

func TestFilter_ByTTLRange(t *testing.T) {
	leases := []*lease.Lease{
		newLease("secret/a", "id1", 100, true),
		newLease("secret/b", "id2", 500, true),
		newLease("secret/c", "id3", 1000, true),
	}
	got := filter.Filter(leases, filter.Options{MinTTL: 200, MaxTTL: 800})
	if len(got) != 1 || got[0].TTL != 500 {
		t.Fatalf("expected 1 lease with TTL 500, got %v", got)
	}
}

func TestFilter_NoConstraints_ReturnsAll(t *testing.T) {
	leases := []*lease.Lease{
		newLease("secret/a", "id1", 3600, true),
		newLease("secret/b", "id2", 3600, true),
	}
	got := filter.Filter(leases, filter.Options{})
	if len(got) != 2 {
		t.Fatalf("expected 2, got %d", len(got))
	}
}

func TestFilter_Empty_ReturnsNil(t *testing.T) {
	got := filter.Filter(nil, filter.Options{PathPrefix: "secret/"})
	if len(got) != 0 {
		t.Fatalf("expected empty, got %d", len(got))
	}
}
