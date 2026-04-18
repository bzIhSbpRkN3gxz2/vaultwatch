package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/lease"
)

func TestLoggerHandler_OnAlert(t *testing.T) {
	var buf bytes.Buffer
	h := alert.NewLoggerHandler(&buf)

	l := lease.New("abc-123", "secret/prod/db", time.Now().Add(2*time.Hour), false)
	h.OnAlert(l, lease.StatusExpiring)

	output := buf.String()

	for _, want := range []string{"status=expiring", "lease_id=abc-123", "path=secret/prod/db"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected output to contain %q, got: %s", want, output)
		}
	}
}

func TestLoggerHandler_DefaultWriter(t *testing.T) {
	// Should not panic when w is nil (falls back to stderr).
	h := alert.NewLoggerHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}
