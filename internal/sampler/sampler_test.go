package sampler

import (
	"testing"
)

func TestNew_DefaultRate(t *testing.T) {
	s := New(0.5)
	if s.Rate() != 0.5 {
		t.Fatalf("expected rate 0.5, got %v", s.Rate())
	}
}

func TestNew_ClampsAboveOne(t *testing.T) {
	s := New(5.0)
	if s.Rate() != 1.0 {
		t.Fatalf("expected rate clamped to 1.0, got %v", s.Rate())
	}
}

func TestNew_ClampsBelowZero(t *testing.T) {
	s := New(-1.0)
	if s.Rate() != 0.0 {
		t.Fatalf("expected rate clamped to 0.0, got %v", s.Rate())
	}
}

func TestSample_AlwaysTrue_RateOne(t *testing.T) {
	s := New(1.0)
	for i := 0; i < 100; i++ {
		if !s.Sample() {
			t.Fatal("expected Sample() to return true for rate=1.0")
		}
	}
}

func TestSample_AlwaysFalse_RateZero(t *testing.T) {
	s := New(0.0)
	for i := 0; i < 100; i++ {
		if s.Sample() {
			t.Fatal("expected Sample() to return false for rate=0.0")
		}
	}
}

func TestSetRate_UpdatesRate(t *testing.T) {
	s := New(0.5)
	s.SetRate(0.9)
	if s.Rate() != 0.9 {
		t.Fatalf("expected rate 0.9, got %v", s.Rate())
	}
}

func TestSample_Probabilistic(t *testing.T) {
	s := New(0.5)
	hits := 0
	trials := 10000
	for i := 0; i < trials; i++ {
		if s.Sample() {
			hits++
		}
	}
	ratio := float64(hits) / float64(trials)
	if ratio < 0.4 || ratio > 0.6 {
		t.Fatalf("expected ~50%% sample rate, got %.2f", ratio)
	}
}
