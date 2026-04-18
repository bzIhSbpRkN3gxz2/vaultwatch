// Package renewal provides lease renewal logic for vaultwatch.
//
// It exposes a Renewer that integrates with the Vault API client to
// automatically extend leases whose remaining TTL falls below a
// configurable threshold, and a Policy type that encapsulates the
// thresholds and retry behaviour used during renewal decisions.
package renewal
