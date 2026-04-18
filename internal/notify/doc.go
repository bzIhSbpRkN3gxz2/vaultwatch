// Package notify implements alert throttling for vaultwatch.
// It prevents alert handlers from being flooded with repeated
// notifications for the same lease within a configurable cooldown window.
package notify
