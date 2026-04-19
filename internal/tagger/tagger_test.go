package tagger_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/tagger"
)

func TestSet_And_Get(t *testing.T) {
	s := tagger.New()
	if err := s.Set("lease-1", "env", "prod"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tags := s.Get("lease-1")
	if tags["env"] != "prod" {
		t.Errorf("expected prod, got %q", tags["env"])
	}
}

func TestGet_ReturnsCopy(t *testing.T) {
	s := tagger.New()
	_ = s.Set("lease-2", "team", "infra")
	tags := s.Get("lease-2")
	tags["team"] = "mutated"
	if s.Get("lease-2")["team"] != "infra" {
		t.Error("Get should return a copy, not a reference")
	}
}

func TestGet_Missing(t *testing.T) {
	s := tagger.New()
	tags := s.Get("nonexistent")
	if len(tags) != 0 {
		t.Errorf("expected empty tags, got %v", tags)
	}
}

func TestSet_EmptyLeaseID(t *testing.T) {
	s := tagger.New()
	if err := s.Set("", "k", "v"); err == nil {
		t.Error("expected error for empty leaseID")
	}
}

func TestSet_EmptyKey(t *testing.T) {
	s := tagger.New()
	if err := s.Set("lease-3", "", "v"); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestDelete(t *testing.T) {
	s := tagger.New()
	_ = s.Set("lease-4", "env", "dev")
	s.Delete("lease-4")
	if len(s.Get("lease-4")) != 0 {
		t.Error("expected tags to be deleted")
	}
}

func TestMatch(t *testing.T) {
	s := tagger.New()
	_ = s.Set("lease-a", "env", "prod")
	_ = s.Set("lease-a", "team", "infra")
	_ = s.Set("lease-b", "env", "prod")
	_ = s.Set("lease-c", "env", "dev")

	matches := s.Match(tagger.Tags{"env": "prod"})
	if len(matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(matches))
	}

	matches = s.Match(tagger.Tags{"env": "prod", "team": "infra"})
	if len(matches) != 1 || matches[0] != "lease-a" {
		t.Errorf("expected only lease-a, got %v", matches)
	}
}

func TestTags_String(t *testing.T) {
	tags := tagger.Tags{"env": "prod"}
	if tags.String() == "" {
		t.Error("expected non-empty string representation")
	}
}
