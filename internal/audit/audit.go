package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Event represents a single audit log entry.
type Event struct {
	Timestamp time.Time   `json:"timestamp"`
	LeaseID   string      `json:"lease_id"`
	Status    lease.Status `json:"status"`
	Message   string      `json:"message"`
}

// Logger writes structured audit events as JSON lines.
type Logger struct {
	w io.Writer
}

// NewLogger returns an audit Logger writing to w.
// If w is nil, os.Stdout is used.
func NewLogger(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{w: w}
}

// Record encodes an audit Event to the underlying writer.
func (l *Logger) Record(leaseID string, status lease.Status, msg string) error {
	e := Event{
		Timestamp: time.Now().UTC(),
		LeaseID:   leaseID,
		Status:    status,
		Message:   msg,
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal event: %w", err)
	}
	_, err = fmt.Fprintf(l.w, "%s\n", b)
	return err
}
