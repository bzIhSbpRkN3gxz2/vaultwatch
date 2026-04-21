package window

import (
	"fmt"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// FrequencyGuard wraps an alert handler and suppresses alerts when the
// number of events for a lease exceeds maxCount within the window period.
type FrequencyGuard struct {
	window   *Window
	maxCount int
	next     func(l lease.Lease) error
}

// NewFrequencyGuard creates a FrequencyGuard using the provided Window.
// Alerts are forwarded to next only when the per-lease count is <= maxCount.
func NewFrequencyGuard(w *Window, maxCount int, next func(l lease.Lease) error) *FrequencyGuard {
	return &FrequencyGuard{
		window:   w,
		maxCount: maxCount,
		next:     next,
	}
}

// OnAlert records the event and forwards the alert if within the allowed
// frequency. Returns a sentinel error when the event is suppressed.
func (g *FrequencyGuard) OnAlert(l lease.Lease) error {
	g.window.Record(l.ID)
	count := g.window.Count(l.ID)

	if count > g.maxCount {
		return fmt.Errorf("window: alert suppressed for lease %s (count=%d, max=%d)",
			l.ID, count, g.maxCount)
	}

	return g.next(l)
}

// NewDefaultGuard creates a FrequencyGuard with a 1-minute window and a
// default max of 3 alerts per lease.
func NewDefaultGuard(next func(l lease.Lease) error) *FrequencyGuard {
	return NewFrequencyGuard(New(time.Minute), 3, next)
}
