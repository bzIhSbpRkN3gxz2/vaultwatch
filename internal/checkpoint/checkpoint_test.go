package checkpoint_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/vaultwatch/internal/checkpoint"
)

func tempFile(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "checkpoint.json")
}

func TestSet_And_Get(t *testing.T) {
	cp, err := checkpoint.New(tempFile(t))
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	cp.Set(checkpoint.Entry{LeaseID: "abc", Path: "secret/db", Status: "healthy", TTL: 3600})
	e, ok := cp.Get("abc")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Path != "secret/db" {
		t.Errorf("path: got %q, want %q", e.Path, "secret/db")
	}
}

func TestGet_Missing(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	_, ok := cp.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry")
	}
}

func TestDelete(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	cp.Set(checkpoint.Entry{LeaseID: "del", Path: "x", Status: "expiring", TTL: 60})
	cp.Delete("del")
	_, ok := cp.Get("del")
	if ok {
		t.Fatal("expected entry to be deleted")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	cp.Set(checkpoint.Entry{LeaseID: "a", Status: "healthy"})
	cp.Set(checkpoint.Entry{LeaseID: "b", Status: "expiring"})
	all := cp.All()
	if len(all) != 2 {
		t.Errorf("All: got %d entries, want 2", len(all))
	}
}

func TestSave_And_Reload(t *testing.T) {
	p := tempFile(t)
	cp, _ := checkpoint.New(p)
	cp.Set(checkpoint.Entry{LeaseID: "persist", Path: "secret/key", Status: "expiring", TTL: 120})
	if err := cp.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cp2, err := checkpoint.New(p)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	e, ok := cp2.Get("persist")
	if !ok {
		t.Fatal("expected persisted entry after reload")
	}
	if e.Status != "expiring" {
		t.Errorf("status: got %q, want %q", e.Status, "expiring")
	}
}

func TestSet_EmptyLeaseID_Ignored(t *testing.T) {
	cp, _ := checkpoint.New(tempFile(t))
	cp.Set(checkpoint.Entry{LeaseID: "", Path: "ignored"})
	if len(cp.All()) != 0 {
		t.Error("expected empty entry to be ignored")
	}
}

func TestSave_WritesValidJSON(t *testing.T) {
	p := tempFile(t)
	cp, _ := checkpoint.New(p)
	cp.Set(checkpoint.Entry{LeaseID: "json-check", Status: "healthy", TTL: 9000})
	_ = cp.Save()

	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Errorf("invalid JSON: %v", err)
	}
}
