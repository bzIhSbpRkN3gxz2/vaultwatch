package alert

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// LoggerHandler writes alerts as structured log lines to a writer.
type LoggerHandler struct {
	out io.Writer
}

// NewLoggerHandler creates a LoggerHandler writing to w.
// If w is nil, os.Stderr is used.
func NewLoggerHandler(w io.Writer) *LoggerHandler {
	if w == nil {
		w = os.Stderr
	}
	return &LoggerHandler{out: w}
}

// OnAlert implements monitor.AlertHandler.
func (h *LoggerHandler) OnAlert(l *lease.Lease, status lease.Status) {
	fmt.Fprintf(
		h.out,
		"time=%s level=WARN status=%s lease_id=%s path=%s expires_at=%s\n",
		time.Now().UTC().Format(time.RFC3339),
		status,
		l.ID,
		l.Path,
		l.ExpiresAt.UTC().Format(time.RFC3339),
	)
}
