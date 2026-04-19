package resolver_test

import (
	"testing"

	"github.com/yourusername/vaultwatch/internal/resolver"
)

func TestResolve_MatchesPrefix(t *testing.T) {
	r := resolver.New([]resolver.Rule{
		{Prefix: "secret/team-a/", Owner: resolver.Owner{Team: "team-a", Email: "a@example.com"}},
		{Prefix: "secret/team-b/", Owner: resolver.Owner{Team: "team-b", Email: "b@example.com"}},
	})

	owner, ok := r.Resolve("secret/team-a/db")
	if !ok {
		t.Fatal("expected match")
	}
	if owner.Team != "team-a" {
		t.Errorf("got team %q, want team-a", owner.Team)
	}
}

func TestResolve_LongestPrefixWins(t *testing.T) {
	r := resolver.New([]resolver.Rule{
		{Prefix: "secret/", Owner: resolver.Owner{Team: "default"}},
		{Prefix: "secret/team-a/", Owner: resolver.Owner{Team: "team-a"}},
	})

	owner, ok := r.Resolve("secret/team-a/db")
	if !ok {
		t.Fatal("expected match")
	}
	if owner.Team != "team-a" {
		t.Errorf("longest prefix should win, got %q", owner.Team)
	}
}

func TestResolve_NoMatch(t *testing.T) {
	r := resolver.New([]resolver.Rule{
		{Prefix: "secret/team-a/", Owner: resolver.Owner{Team: "team-a"}},
	})

	_, ok := r.Resolve("aws/creds/my-role")
	if ok {
		t.Error("expected no match")
	}
}

func TestResolve_EmptyRules(t *testing.T) {
	r := resolver.New(nil)
	_, ok := r.Resolve("secret/anything")
	if ok {
		t.Error("expected no match with empty rules")
	}
}

func TestAdd_NewRuleResolvable(t *testing.T) {
	r := resolver.New(nil)
	r.Add(resolver.Rule{Prefix: "database/creds/", Owner: resolver.Owner{App: "payments"}})

	owner, ok := r.Resolve("database/creds/readonly")
	if !ok {
		t.Fatal("expected match after Add")
	}
	if owner.App != "payments" {
		t.Errorf("got app %q, want payments", owner.App)
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	rules := []resolver.Rule{{Prefix: "secret/", Owner: resolver.Owner{Team: "x"}}}
	r := resolver.New(rules)
	got := r.Rules()
	if len(got) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(got))
	}
	got[0].Prefix = "mutated/"
	if r.Rules()[0].Prefix != "secret/" {
		t.Error("Rules should return a copy")
	}
}
