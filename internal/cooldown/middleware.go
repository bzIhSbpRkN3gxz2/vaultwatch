package cooldown

import (
	"fmt"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Handler is a function that processes a lease alert or action.
type Handler func(l *lease.Lease) error

// ErrCoolingDown is returned when an action is suppressed by the cooldown.
var ErrCoolingDown = fmt.Errorf("cooldown: action suppressed, key still within cooldown window")

// Guard wraps a Handler, skipping execution if the lease is within its
// cooldown window. The lease ID is used as the tracking key.
func Guard(t *Tracker, next Handler) Handler {
	return func(l *lease.Lease) error {
		if !t.Allow(l.ID) {
			return ErrCoolingDown
		}
		return next(l)
	}
}
