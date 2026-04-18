package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/alert"
	"github.com/user/vaultwatch/internal/lease"
)

func newTestLease(id string, expiresAt time.Time) *lease.Lease {
	return lease.New(id, "secret/data/db", expiresAt)
}

func TestSlackHandler_OnAlert_Success(t *testing.T) {
	var received map[string]string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := alert.NewSlackHandler(ts.URL)
	l := newTestLease("lease-abc", time.Now().Add(5*time.Minute))

	if err := h.OnAlert(l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["text"] == "" {
		t.Error("expected non-empty text payload")
	}
}

func TestSlackHandler_OnAlert_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := alert.NewSlackHandler(ts.URL)
	l := newTestLease("lease-xyz", time.Now().Add(2*time.Minute))

	if err := h.OnAlert(l); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestSlackHandler_OnAlert_BadURL(t *testing.T) {
	h := alert.NewSlackHandler("http://127.0.0.1:0/no-server")
	l := newTestLease("lease-bad", time.Now().Add(1*time.Minute))

	if err := h.OnAlert(l); err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
