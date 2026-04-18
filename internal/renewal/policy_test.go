package renewal_test

import (
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/renewal"
)

func TestDefaultPolicy(t *testing.T) {
	p := renewal.DefaultPolicy()
	if p.WarnThreshold == 0 {
		t.Error("expected non-zero WarnThreshold")
	}
	if p.RenewThreshold == 0 {
		t.Error("expected non-zero RenewThreshold")
	}
	if p.MaxRetries <= 0 {
		t.Error("expected positive MaxRetries")
	}
}

func TestPolicy_ShouldWarn(t *testing.T) {
	p := renewal.Policy{WarnThreshold: 10 * time.Minute}
	cases := []struct {
		ttl  time.Duration
		want bool
	}{
		{15 * time.Minute, false},
		{10 * time.Minute, true},
		{5 * time.Minute, true},
		{0, false},
	}
	for _, c := range cases {
		got := p.ShouldWarn(c.ttl)
		if got != c.want {
			t.Errorf("ShouldWarn(%s) = %v, want %v", c.ttl, got, c.want)
		}
	}
}

func TestPolicy_ShouldRenew(t *testing.T) {
	p := renewal.Policy{RenewThreshold: 5 * time.Minute}
	cases := []struct {
		ttl  time.Duration
		want bool
	}{
		{10 * time.Minute, false},
		{5 * time.Minute, true},
		{1 * time.Minute, true},
		{0, false},
	}
	for _, c := range cases {
		got := p.ShouldRenew(c.ttl)
		if got != c.want {
			t.Errorf("ShouldRenew(%s) = %v, want %v", c.ttl, got, c.want)
		}
	}
}
