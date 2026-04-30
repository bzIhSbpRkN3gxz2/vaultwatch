// Package checkpoint provides persistent lease-state tracking for vaultwatch.
// It records the last-known status, TTL, and observation time for each lease
// so that monitoring can resume gracefully after a process restart without
// generating duplicate alerts for previously-seen conditions.
package checkpoint
