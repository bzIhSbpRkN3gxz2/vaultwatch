package audit_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vaultwatch/internal/audit"
	"github.com/vaultwatch/internal/lease"
)

func TestLogger_Record_WritesJSON(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)

	if err := l.Record("lease-abc", lease.StatusExpiring, "expiring soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var e audit.Event
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if e.LeaseID != "lease-abc" {
		t.Errorf("lease_id = %q, want %q", e.LeaseID, "lease-abc")
	}
	if e.Status != lease.StatusExpiring {
		t.Errorf("status = %v, want %v", e.Status, lease.StatusExpiring)
	}
	if e.Message != "expiring soon" {
		t.Errorf("message = %q, want %q", e.Message, "expiring soon")
	}
	if e.Timestamp.IsZero() {
		t.Error("timestamp should not be zero")
	}
}

func TestLogger_DefaultWriter_UsesStdout(t *testing.T) {
	l := audit.NewLogger(nil)
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestOpenFile_CreatesAndCloses(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	fl, err := audit.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	if err := fl.Record("lease-xyz", lease.StatusHealthy, "all good"); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if err := fl.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !strings.Contains(string(data), "lease-xyz") {
		t.Errorf("expected lease-xyz in log, got: %s", data)
	}
}
