// Package pressure provides a per-path lease pressure tracker for vaultwatch.
// It computes a normalised pressure score (0.0–1.0) based on the ratio of
// critical and warning leases to total leases observed on a given path prefix.
package pressure
