package pressure

import (
	"strings"

	"github.com/your-org/vaultwatch/internal/lease"
)

// Classifier categorises a slice of leases by status and records pressure
// scores into the Tracker for each unique path prefix (first two segments).
type Classifier struct {
	tracker *Tracker
}

// NewClassifier returns a Classifier backed by the given Tracker.
func NewClassifier(t *Tracker) *Classifier {
	return &Classifier{tracker: t}
}

// Observe ingests a batch of leases, computes per-prefix counts, and records
// updated pressure scores. It is safe to call concurrently.
func (c *Classifier) Observe(leases []*lease.Lease) {
	type counts struct{ total, critical, warning int }
	buckets := make(map[string]*counts)

	for _, l := range leases {
		prefix := pathPrefix(l.Path)
		if _, ok := buckets[prefix]; !ok {
			buckets[prefix] = &counts{}
		}
		buckets[prefix].total++
		switch l.Status() {
		case lease.StatusExpired, lease.StatusOrphaned:
			buckets[prefix].critical++
		case lease.StatusExpiring:
			buckets[prefix].warning++
		}
	}

	for prefix, cnt := range buckets {
		c.tracker.Record(prefix, cnt.total, cnt.critical, cnt.warning)
	}
}

// pathPrefix returns the first two slash-separated segments of a Vault path.
func pathPrefix(path string) string {
	parts := strings.SplitN(strings.Trim(path, "/"), "/", 3)
	if len(parts) >= 2 {
		return parts[0] + "/" + parts[1]
	}
	return path
}
