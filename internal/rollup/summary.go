package rollup

import (
	"fmt"
	"strings"

	"github.com/vaultwatch/internal/lease"
)

// Summary produces a human-readable summary string from a Batch.
func Summary(b Batch) string {
	if len(b.Events) == 0 {
		return "no events in batch"
	}
	counts := make(map[lease.Status]int)
	for _, e := range b.Events {
		counts[e.Lease.Status()]++
	}
	parts := make([]string, 0, len(counts))
	for status, n := range counts {
		parts = append(parts, fmt.Sprintf("%s:%d", status, n))
	}
	return fmt.Sprintf("batch(%d events) [%s] flushed_at=%s",
		len(b.Events),
		strings.Join(parts, " "),
		b.FlushedAt.Format("15:04:05"),
	)
}
