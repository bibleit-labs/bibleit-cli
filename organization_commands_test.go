package main

import (
	"encoding/json"
	"io"
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

func TestOrganizationCapabilityGuardAndDeletedStatus(t *testing.T) {
	for _, supported := range []bool{false, true} {
		t.Run(map[bool]string{false: "old-server", true: "foundation-server"}[supported], func(t *testing.T) {
			t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "config.json"))
			t.Setenv("BIBLEIT_TOKEN", "bt_fixture")
			previous := cliHTTPClient
			defer func() { cliHTTPClient = previous }()
			commands := []string{}
			cliHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				var payload map[string]string
				if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
					t.Fatal(err)
				}
				commands = append(commands, payload["command"])
				wire := "OK protocol_version=1 version=fixture capabilities=help,live\n"
				if len(commands) == 1 && supported {
					wire = "OK protocol_version=1 version=fixture capabilities=help,live,organization_status\n"
				}
				if len(commands) > 1 {
					wire = "OK application_id=fixture name=Example domain=example.org state=organization_deleted next_action=none verification_url=/organizations/applications/fixture\n"
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(wire))}, nil
			})}
			code, out, stderr := captureCLI(t, func() int {
				return run([]string{"--format", "json", "organization", "application", "fixture", "status"})
			})
			if !supported {
				if code != exitUsage || out != "" || !strings.Contains(stderr, "does not support organization_status") || !reflect.DeepEqual(commands, []string{"SERVER INFO"}) {
					t.Fatalf("old server: %d %q %q %v", code, out, stderr, commands)
				}
			} else {
				if code != exitOK || stderr != "" || !strings.Contains(out, "organization_deleted") || strings.Contains(out, "organization_id") || !reflect.DeepEqual(commands, []string{"SERVER INFO", "ORGANIZATION APPLICATION fixture STATUS"}) {
					t.Fatalf("foundation: %d %q %q %v", code, out, stderr, commands)
				}
			}
		})
	}
}
