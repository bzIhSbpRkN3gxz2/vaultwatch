package report

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/vaultwatch/internal/lease"
)

// Summary holds aggregated lease statistics for a report snapshot.
type Summary struct {
	GeneratedAt time.Time      `json:"generated_at"`
	Total       int            `json:"total"`
	Healthy     int            `json:"healthy"`
	Expiring    int            `json:"expiring"`
	Expired     int            `json:"expired"`
	Orphaned    int            `json:"orphaned"`
	Leases      []*lease.Lease `json:"leases,omitempty"`
}

// Generator builds Summary reports from a slice of leases.
type Generator struct {
	w          io.Writer
	includeAll bool
}

// New returns a Generator writing to w. If includeAll is true, the full
// lease list is embedded in the report.
func New(w io.Writer, includeAll bool) *Generator {
	if w == nil {
		w = os.Stdout
	}
	return &Generator{w: w, includeAll: includeAll}
}

// Build computes a Summary from leases.
func (g *Generator) Build(leases []*lease.Lease) Summary {
	s := Summary{
		GeneratedAt: time.Now().UTC(),
		Total:       len(leases),
	}
	for _, l := range leases {
		switch l.Status() {
		case lease.StatusHealthy:
			s.Healthy++
		case lease.StatusExpiring:
			s.Expiring++
		case lease.StatusExpired:
			s.Expired++
		case lease.StatusOrphaned:
			s.Orphaned++
		}
	}
	if g.includeAll {
		s.Leases = leases
	}
	return s
}

// Write serialises the Summary as JSON to the generator's writer.
func (g *Generator) Write(s Summary) error {
	enc := json.NewEncoder(g.w)
	enc.SetIndent("", "  ")
	return enc.Encode(s)
}
