package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

func newWebhookLease() *lease.Lease {
	return lease.New("webhook/test/lease-1", time.Now().Add(5*time.Minute), "database")
}

func TestWebhookHandler_OnAlert_Success(t *testing.T) {
	var received map[string]interface{}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	h := NewWebhookHandler(ts.URL)
	l := newWebhookLease()

	if err := h.OnAlert(l); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["lease_id"] != l.ID {
		t.Errorf("expected lease_id %q, got %v", l.ID, received["lease_id"])
	}
}

func TestWebhookHandler_OnAlert_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	h := NewWebhookHandler(ts.URL)
	if err := h.OnAlert(newWebhookLease()); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookHandler_OnAlert_BadURL(t *testing.T) {
	h := NewWebhookHandler("http://127.0.0.1:0/invalid")
	if err := h.OnAlert(newWebhookLease()); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
