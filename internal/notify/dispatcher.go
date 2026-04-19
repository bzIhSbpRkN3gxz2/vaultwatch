package notify

import (
	"context"
	"fmt"
	"log"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/lease"
)

// Dispatcher sends alerts through a handler, suppressing duplicates via Throttle.
type Dispatcher struct {
	handler  alert.Handler
	throttle *Throttle
	logger   *log.Logger
}

// NewDispatcher creates a Dispatcher with the given handler and throttle.
func NewDispatcher(h alert.Handler, t *Throttle, l *log.Logger) *Dispatcher {
	if l == nil {
		l = log.Default()
	}
	return &Dispatcher{handler: h, throttle: t, logger: l}
}

// Dispatch sends an alert for the given lease if not throttled.
// Returns true if the alert was dispatched, false if suppressed.
func (d *Dispatcher) Dispatch(ctx context.Context, l *lease.Lease) (bool, error) {
	if d.throttle.Allow(l.ID) {
		if err := d.handler.OnAlert(ctx, l); err != nil {
			return false, fmt.Errorf("dispatch alert for lease %s: %w", l.ID, err)
		}
		d.logger.Printf("alert dispatched for lease %s (status=%s)", l.ID, l.Status)
		return true, nil
	}
	d.logger.Printf("alert suppressed for lease %s (throttled)", l.ID)
	return false, nil
}

// DispatchAll dispatches alerts for a slice of leases.
// Returns the count of dispatched alerts and the first error encountered.
func (d *Dispatcher) DispatchAll(ctx context.Context, leases []*lease.Lease) (int, error) {
	var firstErr error
	count := 0
	for _, l := range leases {
		ok, err := d.Dispatch(ctx, l)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if ok {
			count++
		}
	}
	return count, firstErr
}
