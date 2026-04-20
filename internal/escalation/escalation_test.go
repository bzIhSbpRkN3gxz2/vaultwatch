package escalation_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/escalation"
)

func testPolicy() escalation.Policy {
	return escalation.Policy{
		WarnAfter:     1,
		CriticalAfter: 3,
		PageAfter:     5,
		Window:        1 * time.Hour,
	}
}

func TestEvaluate_FirstOccurrence_Warn(t *testing.T) {
	e := escalation.New(testPolicy())
	level := e.Evaluate("lease-1")
	if level != escalation.LevelWarn {
		t.Fatalf("expected LevelWarn, got %v", level)
	}
}

func TestEvaluate_RepeatedOccurrence_Critical(t *testing.T) {
	e := escalation.New(testPolicy())
	for i := 0; i < 3; i++ {
		e.Evaluate("lease-2")
	}
	level := e.Evaluate("lease-2")
	if level != escalation.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", level)
	}
}

func TestEvaluate_ExceedsPageThreshold(t *testing.T) {
	e := escalation.New(testPolicy())
	for i := 0; i < 5; i++ {
		e.Evaluate("lease-3")
	}
	level := e.Evaluate("lease-3")
	if level != escalation.LevelPage {
		t.Fatalf("expected LevelPage, got %v", level)
	}
}

func TestReset_ClearsState(t *testing.T) {
	e := escalation.New(testPolicy())
	for i := 0; i < 5; i++ {
		e.Evaluate("lease-4")
	}
	e.Reset("lease-4")
	level := e.Evaluate("lease-4")
	if level != escalation.LevelWarn {
		t.Fatalf("expected LevelWarn after reset, got %v", level)
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	p := testPolicy()
	p.Window = 1 * time.Millisecond
	e := escalation.New(p)
	e.Evaluate("lease-5")
	time.Sleep(5 * time.Millisecond)
	e.Purge()
	// After purge, next evaluate should restart at warn
	level := e.Evaluate("lease-5")
	if level != escalation.LevelWarn {
		t.Fatalf("expected LevelWarn after purge, got %v", level)
	}
}

func TestDefaultPolicy_Values(t *testing.T) {
	p := escalation.DefaultPolicy()
	if p.WarnAfter != 1 || p.CriticalAfter != 3 || p.PageAfter != 6 {
		t.Fatalf("unexpected default policy: %+v", p)
	}
}
