package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRegistryEveryLeafHasOfflineHelpAndValidatedExample(t *testing.T) {
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "broken.json"))
	if err := os.WriteFile(configPath(), []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, spec := range commandRegistry {
		t.Run(spec.Syntax, func(t *testing.T) {
			if seen[spec.Syntax] {
				t.Fatal("duplicate syntax")
			}
			seen[spec.Syntax] = true
			args := strings.Fields(spec.Example)
			got, err := normalizeCommand(args)
			if err != nil {
				t.Fatal(err)
			}
			if !spec.Local && args[0] != "auth" {
				if _, err = commandFor(got); err != nil {
					t.Fatal("registry/SSH grammar drift", err)
				}
			}

			code, out, stderr := captureCLI(t, func() int { return run(append(args, "--help", "--format", "json")) })
			if code != exitOK || stderr != "" || !strings.Contains(out, "schema_version") {
				t.Fatalf("offline leaf help: %d %s %s", code, out, stderr)
			}
		})
	}
}
func TestUsageValidationPrecedesCredentialsAndNetwork(t *testing.T) {
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "broken.json"))
	os.WriteFile(configPath(), []byte("broken"), 0600)
	for _, a := range [][]string{
		{"live", "stack", "info"}, {"live", "L1", "stack", "pop", "-1"}, {"live", "L1", "stack", "pop", "0"}, {"live", "L1", "stack", "pop", "2147483648"}, {"live", "L1", "stack", "pop", "99999999999999999999999"},
		{"live", "create", "Sunday", "--wat"}, {"account", "tokens", "extra"}, {"read", "web", "John", "3:16:2"}, {"auth", "login", "postgres"},
		{"live", "L1", "secret", "PRIVATE_AUDIENCE_CANARY"},
	} {
		code, out, stderr := captureCLI(t, func() int { return run(a) })
		if code != exitUsage || out != "" || strings.Contains(stderr, "PRIVATE_AUDIENCE_CANARY") || strings.Contains(stderr, "read configuration") {
			t.Fatalf("%v: %d %s %s", a, code, out, stderr)
		}
	}
}
func TestReferenceAndAliasesNormalizeAcrossTransports(t *testing.T) {
	for _, a := range [][]string{{"read", "web", "John", "3:16"}, {"read", "web", "John", "3", "16"}} {
		n, e := normalizeCommand(a)
		if e != nil || strings.Join(n, " ") != "read web John 3 16" {
			t.Fatal(n, e)
		}
		for _, service := range []bool{false, true} {
			p, e := planHTTP(n, service)
			if e != nil || !strings.Contains(p.operation, "Verse") {
				t.Fatal(p, e)
			}
		}
	}
	for _, a := range [][]string{{"account", "quota", "list"}, {"account", "token", "list"}} {
		n, e := normalizeCommand(a)
		if e != nil || len(n) != 2 {
			t.Fatal(n, e)
		}
	}
	for _, count := range []string{"", "2"} {
		a := []string{"live", "L1", "stack", "pop"}
		if count != "" {
			a = append(a, count)
		}
		if destructiveAction(a) != "" {
			t.Fatal("pop must never prompt")
		}
	}
}
func TestGeneratedCompletionMatchesRegistryLiterals(t *testing.T) {
	for _, spec := range commandRegistry {
		p := strings.Fields(spec.Syntax)
		example := strings.Fields(spec.Example)
		for i, token := range p {
			if placeholder(token) || i >= len(example) {
				continue
			}
			previous := append([]string(nil), example[:i]...)
			got := completionCandidates(append(previous, token))
			found := false
			for _, c := range got {
				if c == token {
					found = true
				}
			}
			if !found {
				t.Errorf("%s at %s: %v", spec.Syntax, token, got)
			}
		}
	}
}

func TestTranslatedNumericStackReferenceCompactMatchesExpandedAcrossTransports(t *testing.T) {
	compact := []string{"live", "L1", "stack", "push", "bpml", "31", "1:1"}
	expanded := []string{"live", "L1", "stack", "push", "bpml", "31", "1", "1"}
	want, err := commandFor(expanded)
	if err != nil {
		t.Fatal(err)
	}
	got, err := commandFor(compact)
	if err != nil || got.String() != want.String() {
		t.Fatalf("compact SSH reference differs from expanded form: %v", err)
	}
	for _, service := range []bool{false, true} {
		first, err := planHTTP(compact, service)
		if err != nil {
			t.Fatal(err)
		}
		second, err := planHTTP(expanded, service)
		if err != nil {
			t.Fatal(err)
		}
		firstJSON, _ := json.Marshal(first.options.JSON)
		secondJSON, _ := json.Marshal(second.options.JSON)
		if first.operation != second.operation || first.book != second.book || string(firstJSON) != string(secondJSON) {
			t.Fatal("compact REST reference differs from expanded form")
		}
	}
}
