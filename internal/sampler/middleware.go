package sampler

import (
	"github.com/your-org/vaultwatch/internal/lease"
)

// AlertHandler is the interface satisfied by alert handlers in the pipeline.
type AlertHandler interface {
	OnAlert(l *lease.Lease) error
}

// Middleware wraps an AlertHandler and samples events before forwarding.
type Middleware struct {
	sampler *Sampler
	next    AlertHandler
}

// NewMiddleware returns a sampling middleware wrapping next.
func NewMiddleware(s *Sampler, next AlertHandler) *Middleware {
	return &Middleware{sampler: s, next: next}
}

// OnAlert forwards the alert only if the sampler allows it.
func (m *Middleware) OnAlert(l *lease.Lease) error {
	if !m.sampler.Sample() {
		return nil
	}
	return m.next.OnAlert(l)
}
