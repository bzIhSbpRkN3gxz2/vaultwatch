package jitter

import (
	"testing"
	"time"
)

func TestDefaultStrategy_Fields(t *testing.T) {
	s := DefaultStrategy()
	if s.Factor != 0.2 {
		t.Fatalf("expected Factor 0.2, got %v", s.Factor)
	}
}

func TestNew_ClampsNegative(t *testing.T) {
	s := New(-0.5)
	if s.Factor != 0 {
		t.Fatalf("expected Factor 0, got %v", s.Factor)
	}
}

func TestNew_ClampsAboveOne(t *testing.T) {
	s := New(1.5)
	if s.Factor != 1 {
		t.Fatalf("expected Factor 1, got %v", s.Factor)
	}
}

func TestApply_ReturnsAtLeastBase(t *testing.T) {
	s := New(0.3)
	base := 10 * time.Second
	for i := 0; i < 100; i++ {
		d := s.Apply(base)
		if d < base {
			t.Fatalf("Apply returned %v, expected >= %v", d, base)
		}
		max := base + time.Duration(float64(base)*0.3)
		if d > max {
			t.Fatalf("Apply returned %v, expected <= %v", d, max)
		}
	}
}

func TestApply_ZeroFactor_ReturnsBase(t *testing.T) {
	s := New(0)
	base := 5 * time.Second
	if got := s.Apply(base); got != base {
		t.Fatalf("expected %v, got %v", base, got)
	}
}

func TestApply_ZeroBase_ReturnsZero(t *testing.T) {
	s := New(0.5)
	if got := s.Apply(0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}

func TestApplyFull_WithinSymmetricBounds(t *testing.T) {
	s := New(0.2)
	base := 10 * time.Second
	low := time.Duration(float64(base) * 0.8)
	high := time.Duration(float64(base) * 1.2)
	for i := 0; i < 200; i++ {
		d := s.ApplyFull(base)
		if d < low || d > high {
			t.Fatalf("ApplyFull(%v) = %v out of [%v, %v]", base, d, low, high)
		}
	}
}

func TestApplyFull_ZeroFactor_ReturnsBase(t *testing.T) {
	s := New(0)
	base := 8 * time.Second
	if got := s.ApplyFull(base); got != base {
		t.Fatalf("expected %v, got %v", base, got)
	}
}
