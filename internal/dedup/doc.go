// Package dedup implements lease-level deduplication for vaultwatch, ensuring
// that identical lease states are not re-alerted within a configurable window.
package dedup
