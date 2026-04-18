// Package filter provides lease filtering utilities for vaultwatch.
package filter

import (
	"strings"

	"github.com/vaultwatch/internal/lease"
)

// Options holds filtering criteria for leases.
type Options struct {
	PathPrefix  string
	Statuses    []lease.Status
	MinTTL      int // seconds
	MaxTTL      int // seconds; 0 means no upper bound
}

// Filter returns only the leases from the input slice that match all
// criteria specified in opts. An empty/zero field means "no constraint".
func Filter(leases []*lease.Lease, opts Options) []*lease.Lease {
	var out []*lease.Lease
	for _, l := range leases {
		if opts.PathPrefix != "" && !strings.HasPrefix(l.Path, opts.PathPrefix) {
			continue
		}
		if len(opts.Statuses) > 0 && !statusIn(l.Status(), opts.Statuses) {
			continue
		}
		if opts.MinTTL > 0 && l.TTL < opts.MinTTL {
			continue
		}
		if opts.MaxTTL > 0 && l.TTL > opts.MaxTTL {
			continue
		}
		out = append(out, l)
	}
	return out
}

func statusIn(s lease.Status, list []lease.Status) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
