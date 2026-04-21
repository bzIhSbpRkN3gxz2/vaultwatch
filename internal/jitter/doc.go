// Package jitter adds randomised noise to time durations so that
// concurrent lease-renewal goroutines do not all wake up simultaneously.
package jitter
