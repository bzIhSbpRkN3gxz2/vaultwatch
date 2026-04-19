// Package resolver maps lease paths to owner metadata using configurable rules.
package resolver

import (
	"strings"
	"sync"
)

// Owner holds metadata about the entity that owns a lease.
type Owner struct {
	Team  string
	Email string
	App   string
}

// Rule maps a path prefix to an Owner.
type Rule struct {
	Prefix string
	Owner  Owner
}

// Resolver resolves lease paths to owners.
type Resolver struct {
	mu    sync.RWMutex
	rules []Rule
}

// New returns a Resolver loaded with the given rules.
func New(rules []Rule) *Resolver {
	r := make([]Rule, len(rules))
	copy(r, rules)
	return &Resolver{rules: r}
}

// Resolve returns the Owner for the given lease path.
// It matches the longest prefix rule. If no rule matches, the zero Owner is returned.
func (r *Resolver) Resolve(path string) (Owner, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var best Rule
	found := false
	for _, rule := range r.rules {
		if strings.HasPrefix(path, rule.Prefix) {
			if !found || len(rule.Prefix) > len(best.Prefix) {
				best = rule
				found = true
			}
		}
	}
	return best.Owner, found
}

// Add appends a new rule at runtime.
func (r *Resolver) Add(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rules = append(r.rules, rule)
}

// Rules returns a snapshot of current rules.
func (r *Resolver) Rules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Rule, len(r.rules))
	copy(out, r.rules)
	return out
}
