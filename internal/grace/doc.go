// Package grace provides a Tracker that records Vault secret leases entering
// their grace period — the configurable window of time immediately before a
// lease expires. Downstream components (alerting, renewal, rotation) can
// query the tracker to decide whether urgent action is required.
package grace
