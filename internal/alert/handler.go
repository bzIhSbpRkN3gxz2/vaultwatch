package alert

import "github.com/user/vaultwatch/internal/lease"

// Handler is implemented by any type that can receive a lease alert.
type Handler interface {
	OnAlert(l *lease.Lease) error
}

// MultiHandler dispatches an alert to multiple handlers in order.
// It continues on error and returns the last non-nil error.
type MultiHandler struct {
	handlers []Handler
}

// NewMultiHandler creates a MultiHandler wrapping the provided handlers.
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// OnAlert calls OnAlert on each underlying handler.
func (m *MultiHandler) OnAlert(l *lease.Lease) error {
	var lastErr error
	for _, h := range m.handlers {
		if err := h.OnAlert(l); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
