package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

func TestBroadcastHTTPReadAndLifecycleRoutes(t *testing.T) {
	prefix := "/v1/organizations/" + broadcastTestOrg + "/broadcast"
	cases := []struct {
		args                           []string
		method, path, match, key, body string
	}{
		{[]string{"profiles"}, "GET", "/profiles", "", "", ""},
		{[]string{"channel"}, "GET", "/channel", "", "", ""},
		{[]string{"preflight"}, "GET", "/preflight", "", "", ""},
		{[]string{"events"}, "GET", "/events", "", "", ""},
		{[]string{"publishers"}, "GET", "/publishers", "", "", ""},
		{[]string{"takeover"}, "POST", "/takeover", "", "", "{}"},
		{[]string{"setup", "list", "9", broadcastTestID}, "GET", "/setups", "", "", ""},
		{[]string{"setup", broadcastTestID, "show"}, "GET", "/setups/" + broadcastTestID, "", "", ""},
		{[]string{"event", broadcastTestID, "get"}, "GET", "/events/" + broadcastTestID, "", "", ""},
		{[]string{"operation", broadcastTestID, "get"}, "GET", "/operations/" + broadcastTestID, "", "", ""},
		{[]string{"publisher", broadcastTestID, "get"}, "GET", "/publishers/" + broadcastTestID, "", "", ""},
		{[]string{"publisher", broadcastTestID, "stop", "7"}, "POST", "/publishers/" + broadcastTestID + "/stop", `"publisher-7"`, "", "{}"},
		{[]string{"event", "attach", broadcastTestID, "YouTube_id-1"}, "POST", "/events/attach", "", broadcastTestID, `{"provider_broadcast_id":"YouTube_id-1"}`},
		{[]string{"event", broadcastTestID, "test", "7", broadcastTestID}, "POST", "/events/" + broadcastTestID + "/transitions", `"event-7"`, broadcastTestID, `{"state":"testing"}`},
		{[]string{"event", broadcastTestID, "go-live", "7", broadcastTestID}, "POST", "/events/" + broadcastTestID + "/transitions", `"event-7"`, broadcastTestID, `{"state":"live"}`},
		{[]string{"event", broadcastTestID, "end", "7", broadcastTestID}, "POST", "/events/" + broadcastTestID + "/transitions", `"event-7"`, broadcastTestID, `{"state":"complete"}`},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			calls := 0
			httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
				calls++
				var body []byte
				if r.Body != nil {
					body, _ = io.ReadAll(r.Body)
				}
				if r.Method != tc.method || r.URL.Path != prefix+tc.path || r.Header.Get("If-Match") != tc.match || r.Header.Get("Idempotency-Key") != tc.key || string(body) != tc.body {
					t.Fatalf("%s %s headers=%v body=%s", r.Method, r.URL, r.Header, body)
				}
				if tc.args[0] == "setup" && tc.args[1] == "list" {
					if r.URL.Query().Get("limit") != "9" || r.URL.Query().Get("after") != broadcastTestID {
						t.Fatal(r.URL)
					}
				}
				return httpReply(200, `{"accepted":true}`), nil
			})
			args := append([]string{"--format", "json", "--yes", "broadcast", broadcastTestOrg}, tc.args...)
			code, out, stderr := captureCLI(t, func() int { return run(args) })
			if code != exitOK || calls != 1 || !strings.Contains(out, `"schema_version": 2`) || stderr != "" {
				t.Fatal(code, out, stderr, calls)
			}
		})
	}
}
func TestBroadcastHTTPWriteRetainsOriginalTicketAfterLostReply(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "setup.json")
	request := filepath.Join(dir, "request.json")
	body := []byte("{ \"version\": 1, \"name\": \"Exact bytes\" }\n")
	if e := os.WriteFile(input, body, 0600); e != nil {
		t.Fatal(e)
	}
	intents, commits, statuses := 0, 0, 0
	applied := false
	prefix := "/v1/organizations/" + broadcastTestOrg + "/broadcast"
	receipt := `{"request_id":"retained-ticket","result_state":"applied","applied":true,"id":"` + broadcastTestID + `","revision":1}`
	httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
		var data []byte
		if r.Body != nil {
			data, _ = io.ReadAll(r.Body)
		}
		switch r.URL.Path {
		case prefix + "/write-requests":
			intents++
			var v map[string]any
			json.Unmarshal(data, &v)
			sum := sha256.Sum256(body)
			if v["payload_sha256"] != hex.EncodeToString(sum[:]) || v["payload_bytes"] != float64(len(body)) || v["expected_revision"] != float64(0) || v["target_id"] != "new" {
				t.Fatal(v)
			}
			return httpReply(201, `{"request_id":"retained-ticket"}`), nil
		case prefix + "/write-requests/retained-ticket":
			statuses++
			if applied {
				return httpReply(200, receipt), nil
			}
			return httpReply(200, `{"request_id":"retained-ticket","result_state":"not_seen"}`), nil
		case prefix + "/setups":
			commits++
			if string(data) != string(body) || r.Header.Get("Idempotency-Key") != "retained-ticket" || r.Header.Get("If-Match") != "" {
				t.Fatal(string(data), r.Header)
			}
			if _, e := loadBroadcastRequest(request); e != nil {
				t.Fatal("mutation before private request persisted", e)
			}
			applied = true
			return nil, errors.New("simulated lost commit reply")
		default:
			t.Fatal(r.URL)
			return nil, nil
		}
	})
	args := []string{"--format", "json", "broadcast", broadcastTestOrg, "setup", "create", input, request}
	code, _, _ := captureCLI(t, func() int { return run(args) })
	if code == exitOK {
		t.Fatal("lost reply reported success")
	}
	code, out, stderr := captureCLI(t, func() int { return run(args) })
	if code != exitOK || !strings.Contains(out, broadcastTestID) {
		t.Fatal(code, out, stderr)
	}
	if intents != 1 || commits != 1 || statuses != 2 {
		t.Fatal(intents, commits, statuses)
	}
	os.WriteFile(input, []byte(`{"changed":true}`), 0600)
	code, _, _ = captureCLI(t, func() int { return run(args) })
	if code == exitOK || intents != 1 || commits != 1 {
		t.Fatal("changed payload replaced retained intent", code, intents, commits)
	}
}
func TestBroadcastHTTPSetupMutationsUseRevisionAndExactPayload(t *testing.T) {
	for _, op := range []string{"save", "duplicate", "delete"} {
		t.Run(op, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "setup.json")
			path := filepath.Join(dir, "request.json")
			os.WriteFile(input, []byte("{ \"name\":\"Saved\" }\n"), 0600)
			args := []string{"setup", broadcastTestID, op, "8"}
			switch op {
			case "save":
				args = append(args, input, path)
			case "duplicate":
				args = append(args, "Copy", path)
			case "delete":
				args = append(args, path)
			}
			expected, body, _, e := broadcastWriteInput(broadcastTestOrg, args)
			if e != nil {
				t.Fatal(e)
			}
			expected.RequestID = "retained-ticket"
			encoded, _ := json.Marshal(expected)
			if e = writeBroadcastPrivate(path, encoded); e != nil {
				t.Fatal(e)
			}
			calls := 0
			httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
				calls++
				if calls == 1 {
					return httpReply(200, `{"request_id":"retained-ticket","result_state":"not_seen"}`), nil
				}
				got, _ := io.ReadAll(r.Body)
				method := "POST"
				if op == "save" {
					method = "PUT"
				}
				if op == "delete" {
					method = "DELETE"
				}
				if r.Method != method || r.Header.Get("If-Match") != `"setup-8"` || r.Header.Get("Idempotency-Key") != "retained-ticket" || string(got) != string(body) {
					t.Fatal(r.Method, r.Header, string(got))
				}
				return httpReply(200, `{"request_id":"retained-ticket","result_state":"applied","applied":true}`), nil
			})
			code, out, stderr := captureCLI(t, func() int {
				return run(append([]string{"--yes", "--format", "json", "broadcast", broadcastTestOrg}, args...))
			})
			if code != exitOK || calls != 2 {
				t.Fatal(code, out, stderr, calls)
			}
		})
	}
}
func TestBroadcastHTTPExportAndCredentialBoundaries(t *testing.T) {
	for _, token := range []string{"svc_fixture", "oa_fixture"} {
		httpFixture(t, token, func(*http.Request) (*http.Response, error) {
			t.Fatal("ineligible credential reached service")
			return nil, nil
		})
		code, _, _ := captureCLI(t, func() int { return run([]string{"--format", "json", "broadcast", broadcastTestOrg, "profiles"}) })
		if code != exitForbidden {
			t.Fatal(token, code)
		}
	}
	path := filepath.Join(t.TempDir(), "export.json")
	httpFixture(t, "bt_fixture", func(*http.Request) (*http.Response, error) {
		r := httpReply(200, `{"id":"`+broadcastTestID+`","organization_id":"`+broadcastTestOrg+`","revision":9,"document":{"version":1,"name":"Export"}}`)
		r.Header.Set("ETag", `"setup-9"`)
		return r, nil
	})
	code, out, stderr := captureCLI(t, func() int {
		return run([]string{"--format", "json", "broadcast", broadcastTestOrg, "setup", broadcastTestID, "export", path})
	})
	data, e := readBroadcastJSON(path, broadcastMaxDocument)
	if code != exitOK || e != nil || !strings.Contains(string(data), "Export") {
		t.Fatal(code, out, stderr, e)
	}
}

func TestBroadcastHTTPEventCreationPreservesMetadataAndRequestUUID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "event.json")
	body := []byte("{ \"title\": \"Unlisted\", \"privacy\":\"unlisted\" }\n")
	os.WriteFile(path, body, 0600)
	calls := 0
	httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
		calls++
		got, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || !strings.HasSuffix(r.URL.Path, "/events") || string(got) != string(body) || r.Header.Get("Idempotency-Key") != broadcastTestID {
			t.Fatal(r.Method, r.URL, string(got), r.Header)
		}
		return httpReply(202, `{"id":"`+broadcastTestID+`","state":"queued"}`), nil
	})
	code, out, stderr := captureCLI(t, func() int {
		return run([]string{"--format", "json", "broadcast", broadcastTestOrg, "event", "create", broadcastTestID, path})
	})
	if code != exitOK || calls != 1 {
		t.Fatal(code, out, stderr, calls)
	}
}

func TestBroadcastErrorExitCategoriesMatchAcrossTransports(t *testing.T) {
	cases := []struct {
		code         string
		status, exit int
	}{
		{"browser_required", 409, exitConflict}, {"publisher_not_ready", 409, exitConflict}, {"operation_conflict", 409, exitConflict},
		{"invalid_asset", 422, exitUsage}, {"invalid_metadata", 422, exitUsage}, {"invalid_request", 422, exitUsage},
		{"upload_capacity", 429, exitRateLimited}, {"commit_capacity", 429, exitRateLimited},
		{"forbidden", 403, exitForbidden}, {"outcome_unknown", 503, exitFailure},
	}
	for _, tc := range cases {
		ssh, _, _ := captureCLI(t, func() int { return protocolFailure(&bibleit.ServerError{Code: tc.code}) })
		rest, _, _ := captureCLI(t, func() int { return protocolFailure(&bibleit.ProblemError{Code: tc.code, Status: tc.status}) })
		if ssh != tc.exit || rest != tc.exit {
			t.Fatal(tc.code, ssh, rest, tc.exit)
		}
	}
}

func TestBroadcastHTTPUnknownSuccessReceiptRequiresReconciliation(t *testing.T) {
	result := map[string]any{"request_id": "retained-ticket", "result_state": "unknown"}
	if _, e := broadcastHTTPTerminalReceipt(result); e == nil {
		t.Fatal("unverified outcome reported as successful write")
	}
}
func TestBroadcastHTTPDestructiveConfirmationPrecedesRequests(t *testing.T) {
	httpFixture(t, "bt_fixture", func(*http.Request) (*http.Response, error) {
		t.Fatal("request before destructive confirmation")
		return nil, nil
	})
	for _, args := range [][]string{{"event", broadcastTestID, "end", "1", broadcastTestID}, {"publisher", broadcastTestID, "stop", "1"}, {"setup", broadcastTestID, "delete", "1", "request.json"}} {
		code, _, _ := captureCLI(t, func() int { return run(append([]string{"--format", "json", "broadcast", broadcastTestOrg}, args...)) })
		if code != exitUsage {
			t.Fatal(code, args)
		}
	}
}
