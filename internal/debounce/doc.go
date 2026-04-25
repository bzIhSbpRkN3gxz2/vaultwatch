// Package debounce prevents alert storms by suppressing repeated notifications
// for the same lease within a configurable quiet window. Only the first event
// in each window is forwarded; subsequent events are dropped until the window
// resets.
package debounce
