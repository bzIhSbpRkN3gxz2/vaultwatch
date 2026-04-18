package alert

import (
	"errors"

	"github.com/user/vaultwatch/internal/lease"
)

// Handler is implemented by any type that can receive a lease alert.
type Handler interface {
	OnAlert(l *lease.Lease) error
}

// MultiHandler dispatches an alert to multiple handlers in order.
// It continues on error and returns a joined error combining all non-nil errors.
type MultiHandler struct {
	handlers []Handler
}

// NewMultiHandler creates a MultiHandler wrapping the provided handlers.
func NewMultiHandler(handlers ...Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

// OnAlert calls OnAlert on each underlying handler.
// All handlers are invoked regardless of errors; all errors are joined and returned.
func (m *MultiHandler) OnAlert(l *lease.Lease) error {
	var errs []error
	for _, h := range m.handlers {
		if err := h.OnAlert(l); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
