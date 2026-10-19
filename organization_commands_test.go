package main

import (
	"net/http"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOrganizationReadCommands(t *testing.T) {
	for args, want := range map[string]string{"organization list": "ORGANIZATION LIST", "organization list a:fixture": "ORGANIZATION LIST a:fixture", "organization application fixture status": "ORGANIZATION APPLICATION fixture STATUS", "server help organization": "HELP ORGANIZATION"} {
		cmd, err := commandFor(strings.Fields(args))
		if err != nil || cmd.String() != want {
			t.Fatalf("%s: %s %v", args, cmd.String(), err)
		}
	}
	for _, args := range [][]string{{"organization", "request"}, {"organization", "join", "fixture"}, {"organization", "application", "id\nPING", "status"}} {
		if _, err := commandFor(args); err == nil {
			t.Fatalf("accepted unsupported/unsafe command: %v", args)
		}
	}
	if got := completionCandidates([]string{"organization", ""}); !reflect.DeepEqual(got, []string{"application", "list"}) {
		t.Fatalf("completion: %v", got)
	}
}

func TestOrganizationHelpCompletion(t *testing.T) {
	if got := completionCandidates([]string{"server", "help", "org"}); !reflect.DeepEqual(got, []string{"organization"}) {
		t.Fatalf("help completion: %v", got)
	}
	if got := completionCandidates([]string{"organization", "application", "fixture", ""}); !reflect.DeepEqual(got, []string{"status"}) {
		t.Fatalf("status completion: %v", got)
	}
}

func TestOrganizationApplicationHasNoHTTPFallback(t *testing.T) {
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv("BIBLEIT_TOKEN", "bt_fixture")
	old := cliHTTPClient
	defer func() { cliHTTPClient = old }()
	cliHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatal("unsupported command contacted server", r.URL)
		return nil, nil
	})}
	code, out, stderr := captureCLI(t, func() int {
		return run([]string{"--format", "json", "organization", "application", "fixture", "status"})
	})
	if code != exitUsage || out != "" || !strings.Contains(stderr, "no supported REST operation") {
		t.Fatal(code, out, stderr)
	}
}
