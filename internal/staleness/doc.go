// Package staleness tracks how long individual leases remain in a given
// status and surfaces an error when the configured threshold is exceeded.
// It is intended to complement the watchdog package by catching leases that
// are technically alive but have not progressed (e.g. perpetually expiring).
package staleness
