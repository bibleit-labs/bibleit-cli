package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestServerScriptureUX(t *testing.T) {
	requireDisposableCLIIntegration(t)
	directory := t.TempDir()
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(directory, "credentials.json"))
	t.Setenv("BIBLEIT_PREFERENCES", filepath.Join(directory, "config.yaml"))
	t.Setenv("BIBLEIT_TOKEN", "")
	token := os.Getenv("BIBLEIT_INTEGRATION_TOKEN")
	if token == "" {
		t.Fatal("fixture credential missing")
	}
	if err := saveConfig(config{AccessToken: token}); err != nil {
		t.Fatal("fixture credential persistence failed")
	}
	operation := func(args ...string) httpResult {
		t.Helper()
		plan, err := planHTTP(args, false)
		if err != nil {
			t.Fatal("fixture planning failed")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		result, err := executeHTTP(ctx, token, plan)
		if err != nil {
			t.Fatal("fixture resource operation failed")
		}
		return result
	}
	created := operation("live", "create", "Scripture UX fixture")
	id, ok := created.data["id"].(string)
	if !ok || id == "" {
		t.Fatal("fixture Live ID missing")
	}
	defer operation("live", id, "delete")
	operation("live", id, "stack", "push", "web", "19", "23", "1")
	for _, args := range [][]string{{"read", "web", "Salmos", "23", "1"}, {"live", id, "stack", "info"}} {
		argv := append([]string{"--transport", "http", "--format", "table"}, args...)
		code, out, stderr := captureCLI(t, func() int { return run(argv) })
		if code != exitOK || stderr != "" {
			t.Fatal("human Scripture command failed")
		}
		for _, value := range []string{"TRANSLATION", "REFERENCE", "TEXT", "Psalms 23:1", "The shepherd guides me."} {
			if !strings.Contains(out, value) {
				t.Fatal("human output lost a shared Scripture field")
			}
		}
		if strings.Contains(out, "api.bibleit.app") || strings.Contains(strings.ToLower(out), "http endpoint") {
			t.Fatal("routine Scripture output contains endpoint noise")
		}
	}
	// No --yes, no input or terminal. Pop must execute directly and report success.
	code, _, stderr := captureCLI(t, func() int { return run([]string{"--transport", "http", "live", id, "stack", "pop"}) })
	if code != exitOK || stderr != "" {
		t.Fatal("stack pop unexpectedly required confirmation")
	}
	stack := operation("live", id, "stack", "info")
	rows, ok := stack.data["data"].([]any)
	if !ok || len(rows) != 0 {
		t.Fatal("unconfirmed pop did not remove the retained entry")
	}
}
