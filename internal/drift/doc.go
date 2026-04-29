// Package drift detects anomalous TTL changes in Vault leases.
// It compares each observed TTL against an elapsed-adjusted baseline
// and returns ErrDriftDetected when the deviation exceeds a configured
// fractional threshold, signalling clock skew or missed renewals.
package drift
