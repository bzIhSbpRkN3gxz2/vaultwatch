package audit

import (
	"fmt"
	"os"
)

// FileLogger wraps Logger with an underlying *os.File so callers
// can open, use, and close a persistent audit log file.
type FileLogger struct {
	*Logger
	f *os.File
}

// OpenFile opens or creates an append-only audit log at path and
// returns a FileLogger. Call Close when done.
func OpenFile(path string) (*FileLogger, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o640)
	if err != nil {
		return nil, fmt.Errorf("audit: open file %q: %w", path, err)
	}
	return &FileLogger{Logger: NewLogger(f), f: f}, nil
}

// Close flushes and closes the underlying file.
func (fl *FileLogger) Close() error {
	if err := fl.f.Sync(); err != nil {
		return fmt.Errorf("audit: sync: %w", err)
	}
	return fl.f.Close()
}
