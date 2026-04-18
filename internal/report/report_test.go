package report_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/vaultwatch/internal/lease"
	"github.com/vaultwatch/internal/report"
)

func newLease(path string, ttl time.Duration, orphaned bool) *lease.Lease {
	return lease.New(lease.Options{
		LeaseID:  "lease/" + path,
		Path:     path,
		TTL:      ttl,
		Orphaned: orphaned,
	})
}

func TestBuild_Counts(t *testing.T) {
	leases := []*lease.Lease{
		newLease("secret/a", 2*time.Hour, false),  // healthy
		newLease("secret/b", 10*time.Minute, false), // expiring
		newLease("secret/c", 0, false),              // expired
		newLease("secret/d", time.Hour, true),       // orphaned
	}

	g := report.New(nil, false)
	s := g.Build(leases)

	if s.Total != 4 {
		t.Errorf("expected Total=4, got %d", s.Total)
	}
	if s.Healthy != 1 {
		t.Errorf("expected Healthy=1, got %d", s.Healthy)
	}
	if s.Expiring != 1 {
		t.Errorf("expected Expiring=1, got %d", s.Expiring)
	}
	if s.Expired != 1 {
		t.Errorf("expected Expired=1, got %d", s.Expired)
	}
	if s.Orphaned != 1 {
		t.Errorf("expected Orphaned=1, got %d", s.Orphaned)
	}
}

func TestBuild_IncludeAll(t *testing.T) {
	leases := []*lease.Lease{newLease("secret/x", time.Hour, false)}
	g := report.New(nil, true)
	s := g.Build(leases)
	if len(s.Leases) != 1 {
		t.Errorf("expected 1 lease in report, got %d", len(s.Leases))
	}
}

func TestWrite_ValidJSON(t *testing.T) {
	var buf bytes.Buffer
	g := report.New(&buf, false)
	s := g.Build([]*lease.Lease{newLease("secret/y", time.Hour, false)})
	if err := g.Write(s); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	var out report.Summary
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if out.Total != 1 {
		t.Errorf("expected Total=1, got %d", out.Total)
	}
}
