// Package circuit provides a per-lease circuit breaker registry for vaultwatch.
// It wraps the low-level circuitbreaker primitive and associates breaker state
// with individual lease IDs so that repeated failures on one lease do not
// affect operations on others.
package circuit
