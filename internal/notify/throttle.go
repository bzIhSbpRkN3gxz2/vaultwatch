// Package notify provides alert throttling to prevent duplicate notifications.
package notify

import (
	"sync"
	"time"
)

// Throttle suppresses repeated alerts for the same lease within a cooldown window.
type Throttle struct {
	mu       sync.Mutex
	seen     map[string]time.Time
	cooldown time.Duration
}

// NewThrottle creates a Throttle with the given cooldown duration.
func NewThrottle(cooldown time.Duration) *Throttle {
	return &Throttle{
		seen:     make(map[string]time.Time),
		cooldown: cooldown,
	}
}

// Allow returns true if an alert for the given key should be sent.
// It records the current time for the key when allowed.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if last, ok := t.seen[key]; ok {
		if time.Since(last) < t.cooldown {
			return false
		}
	}
	t.seen[key] = time.Now()
	return true
}

// Reset clears the throttle state for a specific key.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.seen, key)
}

// Purge removes all keys whose cooldown has elapsed.
func (t *Throttle) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, last := range t.seen {
		if time.Since(last) >= t.cooldown {
			delete(t.seen, k)
		}
	}
}
