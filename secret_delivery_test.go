package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSecretOutputIsPrivateOneTimeAndNeverOverwrites(t *testing.T) {
	calls := 0
	httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
		calls++
		if r.Method == "GET" {
			reply := httpReply(200, `{"id":"L1"}`)
			reply.Header.Set("ETag", `"live-1"`)
			return reply, nil
		}
		if r.Method != "POST" || r.URL.Path != "/v1/lives/L1/audience-secret" {
			t.Fatal(r.Method, r.URL)
		}
		return httpReply(200, `{"live_id":"L1","secret":"SECRET_CANARY","sequence":2}`), nil
	})
	path := filepath.Join(t.TempDir(), "audience-secret")
	args := []string{"live", "L1", "secret", "create", "--secret-output", path, "--format", "json"}
	code, out, stderr := captureCLI(t, func() int { return run(args) })
	if code != exitOK || calls != 2 || strings.Contains(out+stderr, "SECRET_CANARY") || !strings.Contains(out, "written_to_private_file") {
		t.Fatal(code, calls, out, stderr)
	}
	data, e := os.ReadFile(path)
	if e != nil || string(data) != "SECRET_CANARY\n" {
		t.Fatal("missing private delivery", e)
	}
	assertPrivateConfig(t, path)
	code, out, stderr = captureCLI(t, func() int { return run(args) })
	if code != exitUsage || calls != 2 || out != "" || strings.Contains(stderr, "SECRET_CANARY") {
		t.Fatal(code, calls, out, stderr)
	}
}
func TestFailedSecretMutationCleansReservedDestination(t *testing.T) {
	httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
		return httpReply(403, `{"code":"forbidden","status":403}`), nil
	})
	path := filepath.Join(t.TempDir(), "secret")
	code, _, _ := captureCLI(t, func() int { return run([]string{"live", "L1", "secret", "create", "--secret-output", path}) })
	if code != exitForbidden {
		t.Fatal(code)
	}
	if _, e := os.Stat(path); !os.IsNotExist(e) {
		t.Fatal("empty destination left after failure", e)
	}
}
