package tokenstore_test

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/tokenstore"
)

func TestWithContext_And_FromContext(t *testing.T) {
	s := tokenstore.New()
	ctx := tokenstore.WithContext(context.Background(), s)
	got, err := tokenstore.FromContext(ctx)
	if err != nil {
		t.Fatalf("FromContext: unexpected error: %v", err)
	}
	if got != s {
		t.Error("FromContext returned a different store instance")
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, err := tokenstore.FromContext(context.Background())
	if err == nil {
		t.Error("expected error for empty context, got nil")
	}
}

func TestLookup_Found(t *testing.T) {
	s := tokenstore.New()
	e := &tokenstore.Entry{
		LeaseID:   "lease-ctx",
		Token:     "s.ctx-token",
		ExpiresAt: time.Now().Add(time.Minute),
	}
	_ = s.Set(e)
	ctx := tokenstore.WithContext(context.Background(), s)
	got, err := tokenstore.Lookup(ctx, "lease-ctx")
	if err != nil {
		t.Fatalf("Lookup: unexpected error: %v", err)
	}
	if got.Token != e.Token {
		t.Errorf("token mismatch: got %q want %q", got.Token, e.Token)
	}
}

func TestLookup_NotFound(t *testing.T) {
	s := tokenstore.New()
	ctx := tokenstore.WithContext(context.Background(), s)
	_, err := tokenstore.Lookup(ctx, "missing")
	if err != tokenstore.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLookup_NoStore(t *testing.T) {
	_, err := tokenstore.Lookup(context.Background(), "any")
	if err == nil {
		t.Error("expected error when no store in context")
	}
}
