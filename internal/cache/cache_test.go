package cache_test

import (
	"testing"
	"time"

	"github.com/vaultwatch/internal/cache"
)

func TestSet_And_Get(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to be found")
	}
	if v.(string) != "value1" {
		t.Fatalf("expected value1, got %v", v)
	}
}

func TestGet_Missing(t *testing.T) {
	c := cache.New(time.Minute)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected miss for unknown key")
	}
}

func TestGet_Expired(t *testing.T) {
	c := cache.New(time.Millisecond)
	c.Set("expiring", 42)
	time.Sleep(5 * time.Millisecond)
	_, ok := c.Get("expiring")
	if ok {
		t.Fatal("expected expired entry to be a miss")
	}
}

func TestSetTTL_CustomDuration(t *testing.T) {
	c := cache.New(time.Minute)
	c.SetTTL("short", "x", time.Millisecond)
	v, ok := c.Get("short")
	if !ok || v.(string) != "x" {
		t.Fatal("expected hit immediately after set")
	}
	time.Sleep(5 * time.Millisecond)
	_, ok = c.Get("short")
	if ok {
		t.Fatal("expected miss after custom TTL expired")
	}
}

func TestDelete(t *testing.T) {
	c := cache.New(time.Minute)
	c.Set("k", "v")
	c.Delete("k")
	_, ok := c.Get("k")
	if ok {
		t.Fatal("expected key to be deleted")
	}
}

func TestPurge_RemovesExpired(t *testing.T) {
	c := cache.New(time.Millisecond)
	c.Set("a", 1)
	c.Set("b", 2)
	time.Sleep(5 * time.Millisecond)
	c.SetTTL("c", 3, time.Minute)
	removed := c.Purge()
	if removed != 2 {
		t.Fatalf("expected 2 removed, got %d", removed)
	}
	if c.Len() != 1 {
		t.Fatalf("expected 1 remaining, got %d", c.Len())
	}
}

func TestLen(t *testing.T) {
	c := cache.New(time.Minute)
	if c.Len() != 0 {
		t.Fatal("expected empty cache")
	}
	c.Set("x", 1)
	c.Set("y", 2)
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
}
