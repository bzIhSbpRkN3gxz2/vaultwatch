// Package ttlbucket groups leases into TTL-range buckets for prioritised alerting.
package ttlbucket

import (
	"sync"
	"time"

	"github.com/yourusername/vaultwatch/internal/lease"
)

// Bucket labels represent urgency tiers.
const (
	BucketCritical = "critical" // TTL <= 1 h
	BucketWarning  = "warning"  // TTL <= 6 h
	BucketHealthy  = "healthy"  // TTL > 6 h
)

// Thresholds used to classify leases.
var (
	CriticalThreshold = time.Hour
	WarningThreshold  = 6 * time.Hour
)

// Buckets holds leases organised by urgency tier.
type Buckets struct {
	mu      sync.RWMutex
	groups  map[string][]*lease.Lease
}

// New returns an empty Buckets store.
func New() *Buckets {
	return &Buckets{
		groups: map[string][]*lease.Lease{
			BucketCritical: {},
			BucketWarning:  {},
			BucketHealthy:  {},
		},
	}
}

// Classify assigns each lease to the appropriate bucket, replacing any
// previous classification. It is safe to call concurrently.
func (b *Buckets) Classify(leases []*lease.Lease) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.groups[BucketCritical] = b.groups[BucketCritical][:0]
	b.groups[BucketWarning] = b.groups[BucketWarning][:0]
	b.groups[BucketHealthy] = b.groups[BucketHealthy][:0]

	for _, l := range leases {
		ttl := time.Duration(l.TTL) * time.Second
		switch {
		case ttl <= CriticalThreshold:
			b.groups[BucketCritical] = append(b.groups[BucketCritical], l)
		case ttl <= WarningThreshold:
			b.groups[BucketWarning] = append(b.groups[BucketWarning], l)
		default:
			b.groups[BucketHealthy] = append(b.groups[BucketHealthy], l)
		}
	}
}

// Get returns a copy of the lease slice for the given bucket label.
// An empty slice is returned for an unknown label.
func (b *Buckets) Get(label string) []*lease.Lease {
	b.mu.RLock()
	defer b.mu.RUnlock()

	src := b.groups[label]
	out := make([]*lease.Lease, len(src))
	copy(out, src)
	return out
}

// Counts returns a map of bucket label → number of leases.
func (b *Buckets) Counts() map[string]int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	counts := make(map[string]int, len(b.groups))
	for k, v := range b.groups {
		counts[k] = len(v)
	}
	return counts
}
