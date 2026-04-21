package tokenstore_test

import (
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/tokenstore"
)

func newEntry(leaseID string, ttl time.Duration) *tokenstore.Entry {
	return &tokenstore.Entry{
		LeaseID:   leaseID,
		Token:     "s.example",
		ExpiresAt: time.Now().Add(ttl),
		Meta:      map[string]string{"env": "test"},
	}
}

func TestSet_And_Get(t *testing.T) {
	s := tokenstore.New()
	e := newEntry("lease-1", time.Minute)
	if err := s.Set(e); err != nil {
		t.Fatalf("Set: unexpected error: %v", err)
	}
	got, err := s.Get("lease-1")
	if err != nil {
		t.Fatalf("Get: unexpected error: %v", err)
	}
	if got.Token != e.Token {
		t.Errorf("token mismatch: got %q want %q", got.Token, e.Token)
	}
}

func TestGet_Missing(t *testing.T) {
	s := tokenstore.New()
	_, err := s.Get("no-such-lease")
	if err != tokenstore.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestGet_Expired(t *testing.T) {
	s := tokenstore.New()
	e := newEntry("lease-exp", -time.Second)
	_ = s.Set(e)
	_, err := s.Get("lease-exp")
	if err != tokenstore.ErrExpired {
		t.Errorf("expected ErrExpired, got %v", err)
	}
}

func TestDelete(t *testing.T) {
	s := tokenstore.New()
	_ = s.Set(newEntry("lease-del", time.Minute))
	s.Delete("lease-del")
	_, err := s.Get("lease-del")
	if err != tokenstore.ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	s := tokenstore.New()
	_ = s.Set(newEntry("alive", time.Minute))
	_ = s.Set(newEntry("dead-1", -time.Second))
	_ = s.Set(newEntry("dead-2", -time.Second))
	removed := s.Purge()
	if removed != 2 {
		t.Errorf("Purge removed %d entries, want 2", removed)
	}
	if all := s.All(); len(all) != 1 {
		t.Errorf("All returned %d entries after purge, want 1", len(all))
	}
}

func TestAll_ExcludesExpired(t *testing.T) {
	s := tokenstore.New()
	_ = s.Set(newEntry("ok-1", time.Minute))
	_ = s.Set(newEntry("ok-2", time.Minute))
	_ = s.Set(newEntry("stale", -time.Second))
	all := s.All()
	if len(all) != 2 {
		t.Errorf("All returned %d entries, want 2", len(all))
	}
}

func TestSet_EmptyLeaseID_ReturnsError(t *testing.T) {
	s := tokenstore.New()
	err := s.Set(&tokenstore.Entry{Token: "s.x"})
	if err == nil {
		t.Error("expected error for empty LeaseID, got nil")
	}
}
