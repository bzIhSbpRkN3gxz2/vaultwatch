package redactor_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/redactor"
)

func TestRedactMap_MasksSensitiveKeys(t *testing.T) {
	r := redactor.New("***", "password", "token", "secret")
	input := map[string]string{
		"username":      "alice",
		"password":      "hunter2",
		"access_token":  "abc123",
		"client_secret": "xyz",
	}
	out := r.RedactMap(input)
	if out["username"] != "alice" {
		t.Errorf("expected username unchanged, got %q", out["username"])
	}
	for _, k := range []string{"password", "access_token", "client_secret"} {
		if out[k] != "***" {
			t.Errorf("expected %q to be masked, got %q", k, out[k])
		}
	}
}

func TestRedactMap_DoesNotMutateOriginal(t *testing.T) {
	r := redactor.New("***", "secret")
	input := map[string]string{"secret": "topsecret"}
	r.RedactMap(input)
	if input["secret"] != "topsecret" {
		t.Error("original map was mutated")
	}
}

func TestRedactString_ReplacesValues(t *testing.T) {
	r := redactor.New("[REDACTED]", "token")
	keys := map[string]string{"token": "mytoken123"}
	s := r.RedactString(keys, "Authorization: Bearer mytoken123")
	if s != "Authorization: Bearer [REDACTED]" {
		t.Errorf("unexpected result: %q", s)
	}
}

func TestRedactString_SkipsNonSensitiveKeys(t *testing.T) {
	r := redactor.New("***", "secret")
	keys := map[string]string{"username": "alice"}
	s := r.RedactString(keys, "user=alice")
	if s != "user=alice" {
		t.Errorf("expected no change, got %q", s)
	}
}

func TestAddPattern_ExtendsSensitiveKeys(t *testing.T) {
	r := redactor.New("***", "password")
	r.AddPattern("apikey")
	out := r.RedactMap(map[string]string{"apikey": "key123", "other": "val"})
	if out["apikey"] != "***" {
		t.Errorf("expected apikey masked, got %q", out["apikey"])
	}
	if out["other"] != "val" {
		t.Errorf("expected other unchanged, got %q", out["other"])
	}
}

func TestNew_DefaultMask(t *testing.T) {
	r := redactor.New("", "secret")
	out := r.RedactMap(map[string]string{"secret": "val"})
	if out["secret"] != "***" {
		t.Errorf("expected default mask ***, got %q", out["secret"])
	}
}
