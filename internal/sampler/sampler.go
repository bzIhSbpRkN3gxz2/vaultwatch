// Package sampler provides probabilistic sampling for lease alerts,
// allowing high-volume environments to reduce notification noise.
package sampler

import (
	"math/rand"
	"sync"
	"time"
)

// Sampler decides whether a given event should be sampled (passed through).
type Sampler struct {
	mu   sync.Mutex
	rate float64 // 0.0 – 1.0
	rng  *rand.Rand
}

// New returns a Sampler with the given sample rate.
// A rate of 1.0 passes every event; 0.0 drops all events.
func New(rate float64) *Sampler {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	return &Sampler{
		rate: rate,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// SetRate updates the sample rate at runtime.
func (s *Sampler) SetRate(rate float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	s.rate = rate
}

// Rate returns the current sample rate.
func (s *Sampler) Rate() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rate
}

// Sample returns true if the event should be processed.
func (s *Sampler) Sample() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.rate >= 1.0 {
		return true
	}
	if s.rate <= 0.0 {
		return false
	}
	return s.rng.Float64() < s.rate
}
