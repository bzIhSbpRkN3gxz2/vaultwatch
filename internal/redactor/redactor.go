// Package redactor masks sensitive values in lease metadata before logging or alerting.
package redactor

import (
	"strings"
	"sync"
)

// Redactor holds a set of sensitive key patterns and masks matching values.
type Redactor struct {
	mu       sync.RWMutex
	patterns []string
	mask     string
}

// New returns a Redactor with the given mask string and key patterns.
func New(mask string, patterns ...string) *Redactor {
	if mask == "" {
		mask = "***"
	}
	return &Redactor{
		patterns: patterns,
		mask:     mask,
	}
}

// AddPattern registers an additional sensitive key pattern.
func (r *Redactor) AddPattern(pattern string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patterns = append(r.patterns, pattern)
}

// RedactMap returns a copy of m with sensitive values replaced by the mask.
func (r *Redactor) RedactMap(m map[string]string) map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make(map[string]string, len(m))
	for k, v := range m {
		if r.isSensitive(k) {
			out[k] = r.mask
		} else {
			out[k] = v
		}
	}
	return out
}

// RedactString replaces any occurrence of sensitive values within s.
func (r *Redactor) RedactString(keys map[string]string, s string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, v := range keys {
		if r.isSensitive(k) && v != "" {
			s = strings.ReplaceAll(s, v, r.mask)
		}
	}
	return s
}

func (r *Redactor) isSensitive(key string) bool {
	lower := strings.ToLower(key)
	for _, p := range r.patterns {
		if strings.Contains(lower, strings.ToLower(p)) {
			return true
		}
	}
	return false
}
