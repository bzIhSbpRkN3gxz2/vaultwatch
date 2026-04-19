package retrier_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/retrier"
)

func TestPolicy_MaxDelayClamp(t *testing.T) {
	p := retrier.Policy{
		MaxAttempts: 10,
		BaseDelay:   time.Second,
		MaxDelay:    2 * time.Second,
		Multiplier:  4.0,
	}
	sleepCalls := []time.Duration{}
	r := retrier.NewWithSleep(p, func(d time.Duration) {
		sleepCalls = append(sleepCalls, d)
	})
	calls := 0
	_ = r.Do(nil, func() error { //nolint
		calls++
		if calls >= 4 {
			return nil
		}
		return errAlways
	})
	for _, d := range sleepCalls {
		if d > p.MaxDelay {
			t.Fatalf("delay %v exceeded MaxDelay %v", d, p.MaxDelay)
		}
	}
}

var errAlways = retrier.ErrMaxAttempts // reuse sentinel as stand-in
