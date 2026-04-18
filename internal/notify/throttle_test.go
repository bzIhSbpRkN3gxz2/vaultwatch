package notify_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/notify"
)

func TestThrottle_AllowsFirst(t *testing.T) {
	th := notify.NewThrottle(5 * time.Minute)
	if !th.Allow("lease-1") {
		t.Fatal("expected first call to be allowed")
	}
}

func TestThrottle_SuppressesDuplicate(t *testing.T) {
	th := notify.NewThrottle(5 * time.Minute)
	th.Allow("lease-1")
	if th.Allow("lease-1") {
		t.Fatal("expected duplicate within cooldown to be suppressed")
	}
}

func TestThrottle_AllowsAfterCooldown(t *testing.T) {
	th := notify.NewThrottle(10 * time.Millisecond)
	th.Allow("lease-1")
	time.Sleep(20 * time.Millisecond)
	if !th.Allow("lease-1") {
		t.Fatal("expected allow after cooldown elapsed")
	}
}

func TestThrottle_Reset(t *testing.T) {
	th := notify.NewThrottle(5 * time.Minute)
	th.Allow("lease-1")
	th.Reset("lease-1")
	if !th.Allow("lease-1") {
		t.Fatal("expected allow after reset")
	}
}

func TestThrottle_Purge(t *testing.T) {
	th := notify.NewThrottle(10 * time.Millisecond)
	th.Allow("lease-1")
	th.Allow("lease-2")
	time.Sleep(20 * time.Millisecond)
	th.Purge()
	// After purge, both keys should be allowed again
	if !th.Allow("lease-1") {
		t.Fatal("expected lease-1 to be allowed after purge")
	}
	if !th.Allow("lease-2") {
		t.Fatal("expected lease-2 to be allowed after purge")
	}
}

func TestThrottle_IndependentKeys(t *testing.T) {
	th := notify.NewThrottle(5 * time.Minute)
	th.Allow("lease-1")
	if !th.Allow("lease-2") {
		t.Fatal("expected different key to be allowed independently")
	}
}
