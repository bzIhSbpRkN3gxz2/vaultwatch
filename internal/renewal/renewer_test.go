package renewal_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	vaultapi "github.com/hashicorp/vault/api"

	"github.com/youorg/vaultwatch/internal/lease"
	"github.com/youorg/vaultwatch/internal/renewal"
)

func newVaultClient(t *testing.T, srv *httptest.Server) *vaultapi.Client {
	t.Helper()
	cfg := vaultapi.DefaultConfig()
	cfg.Address = srv.URL
	c, err := vaultapi.NewClient(cfg)
	if err != nil {
		t.Fatalf("vault client: %v", err)
	}
	return c
}

func TestMaybeRenew_SkipsHighTTL(t *testing.T) {
	r := renewal.New(renewal.Config{Threshold: 5 * time.Minute})
	l := &lease.Lease{LeaseID: "secret/data/foo/abc123", TTL: 10 * time.Minute}
	renewed, err := r.MaybeRenew(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if renewed {
		t.Error("expected no renewal for high TTL")
	}
}

func TestMaybeRenew_SkipsEmptyLeaseID(t *testing.T) {
	r := renewal.New(renewal.Config{})
	l := &lease.Lease{LeaseID: "", TTL: 1 * time.Minute}
	renewed, err := r.MaybeRenew(context.Background(), l)
	if err != nil || renewed {
		t.Errorf("expected skip for empty leaseID, got renewed=%v err=%v", renewed, err)
	}
}

func TestMaybeRenew_RenewsLowTTL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"lease_id":"secret/data/foo/abc123","lease_duration":3600,"renewable":true}`))
	}))
	defer srv.Close()

	client := newVaultClient(t, srv)
	r := renewal.New(renewal.Config{Client: client, Threshold: 5 * time.Minute})
	l := &lease.Lease{LeaseID: "secret/data/foo/abc123", TTL: 2 * time.Minute}

	renewed, err := r.MaybeRenew(context.Background(), l)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !renewed {
		t.Error("expected renewal for low TTL")
	}
	if l.TTL != 3600*time.Second {
		t.Errorf("expected TTL 3600s, got %s", l.TTL)
	}
}
