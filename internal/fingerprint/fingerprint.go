// Package fingerprint generates stable identifiers for Vault leases
// based on their observable properties.
package fingerprint

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"

	"github.com/your-org/vaultwatch/internal/lease"
)

// Fingerprint is a stable, content-derived identifier for a lease.
type Fingerprint string

// Generator builds fingerprints from lease fields.
type Generator struct {
	includeTTL bool
}

// Option configures a Generator.
type Option func(*Generator)

// WithTTL includes the lease TTL in the fingerprint (makes it time-sensitive).
func WithTTL() Option {
	return func(g *Generator) { g.includeTTL = true }
}

// New returns a Generator with the given options.
func New(opts ...Option) *Generator {
	g := &Generator{}
	for _, o := range opts {
		o(g)
	}
	return g
}

// Compute derives a Fingerprint for the given lease.
func (g *Generator) Compute(l *lease.Lease) Fingerprint {
	parts := []string{
		"id=" + l.LeaseID,
		"path=" + l.Path,
		"status=" + string(l.Status),
	}
	if g.includeTTL {
		parts = append(parts, fmt.Sprintf("ttl=%d", int(l.TTL.Seconds())))
	}
	sort.Strings(parts)
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return Fingerprint(fmt.Sprintf("%x", h[:8]))
}

// String returns the hex string representation.
func (f Fingerprint) String() string { return string(f) }
