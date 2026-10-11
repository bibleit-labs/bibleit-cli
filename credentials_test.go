package main

import (
	"encoding/base64"
	"encoding/json"
	bibleit "github.com/bibleit-labs/bibleit-go"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func captureCLI(t *testing.T, fn func() int) (int, string, string) {
	t.Helper()
	oldOut, oldErr := os.Stdout, os.Stderr
	outR, outW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	errR, errW, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout, os.Stderr = outW, errW
	defer func() { os.Stdout, os.Stderr = oldOut, oldErr; outR.Close(); errR.Close(); outW.Close(); errW.Close() }()
	stdoutResult, stderrResult := make(chan []byte, 1), make(chan []byte, 1)
	go func() { data, _ := io.ReadAll(outR); stdoutResult <- data }()
	go func() { data, _ := io.ReadAll(errR); stderrResult <- data }()
	code := fn()
	outW.Close()
	errW.Close()
	return code, string(<-stdoutResult), string(<-stderrResult)
}
func TestVersionedJSONAndErrors(t *testing.T) {
	code, stdout, stderr := captureCLI(t, func() int {
		return printResponse([]string{"OK actor=user count=1", "TOKEN id=001 source=cli issued_at=100 scopes=\"token.get,future.read\" retiring=false label=\"line\\nlabel\" future=007", "END"}, "json")
	})
	var envelope struct {
		Version int  `json:"schema_version"`
		OK      bool `json:"ok"`
		Records []struct {
			Fields map[string]any `json:"fields"`
		} `json:"records"`
	}
	if json.Unmarshal([]byte(stdout), &envelope) != nil || code != exitOK || stderr != "" || envelope.Version != 1 || !envelope.OK {
		t.Fatal("invalid JSON envelope")
	}
	fields := envelope.Records[1].Fields
	if fields["id"] != "001" || fields["issued_at"] != float64(100) || fields["retiring"] != false || fields["future"] != "007" || fields["label"] != "line\nlabel" {
		t.Fatal("incorrect field types")
	}
	if len(fields["scopes"].([]any)) != 2 {
		t.Fatal("lost permission list")
	}
	previous := activeFormat
	activeFormat = "json"
	defer func() { activeFormat = previous }()
	code, stdout, stderr = captureCLI(t, func() int {
		return protocolFailure(&bibleit.ServerError{Code: "rate_limited", Fields: map[string]string{"retry_after_ms": "250", "future": "kept"}})
	})
	var failure map[string]any
	if code != exitRateLimited || stdout != "" || json.Unmarshal([]byte(stderr), &failure) != nil || failure["ok"] != false || failure["schema_version"] != float64(1) {
		t.Fatal("invalid error envelope")
	}
	if failure["error"].(map[string]any)["fields"].(map[string]any)["retry_after_ms"] != "250" {
		t.Fatal("lost retry metadata")
	}
	code, stdout, _ = captureCLI(t, func() int { return printResponse([]string{"OK count=bad"}, "json") })
	if code != exitFailure || stdout != "" {
		t.Fatal("accepted malformed typed field")
	}
}
func TestNDJSONAndControlRendering(t *testing.T) {
	_, stdout, _ := captureCLI(t, func() int {
		for _, line := range []string{"OK id=live subscribed=true", "EVENT verse " + base64.StdEncoding.EncodeToString([]byte(`[{"translation":"web","text":"João\nverse"}]`)), "EVENT closed"} {
			if err := writeSubscriptionLine(line, "ndjson"); err != nil {
				t.Error(err)
			}
		}
		return exitOK
	})
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatal("not one JSON object per record")
	}
	for i, line := range lines {
		var value map[string]any
		if json.Unmarshal([]byte(line), &value) != nil || value["schema_version"] != float64(1) {
			t.Fatal("invalid stream JSON")
		}
		if i == 2 && value["record"].(map[string]any)["event"] != "closed" {
			t.Fatal("lost event name")
		}
	}
	if humanValue("line\nlabel\t\x1b[31m") != "line\\nlabel\\t\\u001b[31m" {
		t.Fatal("unescaped human controls")
	}
	values, rest, err := parseOptions([]string{"--transport=http", "search", "web", "--", "--format"}, map[string]bool{"--transport": true, "--format": true})
	if err != nil || values["--transport"] != "http" || strings.Join(rest, " ") != "search web --format" {
		t.Fatal("invalid flag terminator")
	}
}

func TestInvalidCredentialDocumentsAndLocalHelp(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("BIBLEIT_CONFIG", path)
	for _, data := range []string{
		`null`, `{"unknown":true}`, `{"profiles":{}}`,
		`{"version":2,"active_profile":"a","profiles":{}}`,
		`{"version":1,"active_profile":"missing","profiles":{"a":{"transport":"http","endpoint":"https://a.example"}}}`,
		`{"version":1,"active_profile":"a","profiles":{"a":{"transport":"http","endpoint":"https://a.example","identity":"/tmp/key"}}}`,
		`{"version":1,"active_profile":"a","profiles":{"a":{"transport":"ssh","endpoint":"a.example:22","access_token":"secret"}}}`,
	} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadConfig(); err == nil {
			t.Fatal("accepted invalid profile document")
		}
	}
	for _, args := range [][]string{{"--format", "json", "version"}, {"--format", "json", "help", "auth"}} {
		code, stdout, stderr := captureCLI(t, func() int { return run(args) })
		var value map[string]any
		if code != exitOK || stderr != "" || json.Unmarshal([]byte(stdout), &value) != nil || value["schema_version"] != float64(1) {
			t.Fatal("local metadata required config")
		}
	}
}
func TestStreamPayloadAndEmptyErrorFields(t *testing.T) {
	payload := `{"text":"João\nverse","chapter":23,"future":true}`
	record, err := subscriptionRecord("EVENT verse " + base64.StdEncoding.EncodeToString([]byte(payload)))
	if err != nil || string(record.Payload) != payload || record.Event != "verse" || len(record.Fields) != 0 {
		t.Fatal("lost native verse payload")
	}
	for _, line := range []string{"EVENT verse broken", "EVENT verse " + base64.StdEncoding.EncodeToString([]byte("not JSON"))} {
		if _, err := subscriptionRecord(line); err == nil {
			t.Fatal("accepted malformed verse payload")
		}
	}
	previous := activeFormat
	activeFormat = "json"
	defer func() { activeFormat = previous }()
	_, _, stderr := captureCLI(t, func() int { return protocolFailure(&bibleit.HTTPError{StatusCode: 503}) })
	var result map[string]any
	if json.Unmarshal([]byte(stderr), &result) != nil || result["error"].(map[string]any)["fields"] == nil {
		t.Fatal("nullable error fields")
	}
}

func TestFixedServiceRejectsEndpointOverridesAndForeignCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("BIBLEIT_CONFIG", path)
	t.Setenv("BIBLEIT_PROFILE", "")
	t.Setenv("BIBLEIT_TOKEN", "bt_test")
	t.Setenv("BIBLEIT_ENDPOINT", "https://other.example")
	beforeClient := cliHTTPClient
	defer func() { cliHTTPClient = beforeClient }()
	requests := 0
	cliHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests++
		if r.URL.String() != defaultWebURL+"/api/cli/command" {
			t.Fatalf("unexpected destination: %s", r.URL)
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("OK pong=true\n"))}, nil
	})}
	for _, args := range [][]string{
		{"--endpoint", "https://other.example", "ping"},
		{"profile", "add", "foreign", "--endpoint", "https://other.example"},
	} {
		code, _, _ := captureCLI(t, func() int { return run(args) })
		if code != exitUsage || requests != 0 {
			t.Fatal("accepted endpoint override", args, code, requests)
		}
	}
	code, _, stderr := captureCLI(t, func() int { return run([]string{"ping"}) })
	if code != exitOK || requests != 1 {
		t.Fatal("environment changed endpoint", code, stderr)
	}
	requests = 0
	for _, data := range []string{
		`{"access_token":"bt_legacy"}`,
		`{"version":1,"active_profile":"old","profiles":{"old":{"transport":"http","endpoint":"https://other.example","access_token":"bt_old"}}}`,
		`{"version":1,"active_profile":"old","profiles":{"old":{"transport":"http","endpoint":"http://127.0.0.1:8080","access_token":"bt_old"}}}`,
		`{"version":1,"active_profile":"old","profiles":{"old":{"transport":"ssh","endpoint":"other.example:2222"}}}`,
	} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		code, _, _ := captureCLI(t, func() int { return run([]string{"ping"}) })
		after, _ := os.ReadFile(path)
		if code != exitFailure || requests != 0 || string(after) != data {
			t.Fatal("retargeted or changed old credentials", code, requests)
		}
	}
	if strings.Contains(strings.Join(completionCandidates([]string{"profile", "add", "new", "--"}), " "), "--endpoint") {
		t.Fatal("completion advertises endpoint overrides")
	}
}

func TestSingleCredentialRoutingAndPreservation(t *testing.T) {
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv("BIBLEIT_PROFILE", "nonexistent")
	t.Setenv("BIBLEIT_TOKEN", "")
	c := config{AccessToken: "bt_cached", Identity: "/tmp/key"}
	if err := saveConfig(c); err != nil {
		t.Fatal(err)
	}
	c, err := loadConfig()
	if err != nil || c.Transport != "http" {
		t.Fatal(c, err)
	}
	old := cliHTTPClient
	defer func() { cliHTTPClient = old }()
	var bearer string
	cliHTTPClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		bearer = r.Header.Get("Authorization")
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("OK pong=true\n"))}, nil
	})}
	for _, args := range [][]string{{"ping"}, {"--transport", "http", "ping"}} {
		code, _, stderr := captureCLI(t, func() int { return run(args) })
		if code != exitOK || bearer != "Bearer bt_cached" {
			t.Fatal(code, stderr, bearer)
		}
	}
	t.Setenv("BIBLEIT_TOKEN", "bt_environment")
	code, _, stderr := captureCLI(t, func() int { return run([]string{"ping"}) })
	if code != exitOK || bearer != "Bearer bt_environment" {
		t.Fatal(code, stderr)
	}
	c, _ = loadConfig()
	if c.AccessToken != "bt_cached" {
		t.Fatal("persisted environment token")
	}
	c.Transport = "ssh"
	c.Endpoint = defaultSSHServer
	automated, err := automationCredential(c)
	if err != nil || automated.AccessToken != "bt_cached" {
		t.Fatal("SSH used environment token")
	}
	if err := saveConfig(c); err != nil {
		t.Fatal(err)
	}
	c, _ = loadConfig()
	if c.Transport != "http" || c.Identity != "/tmp/key" || c.AccessToken != "bt_cached" {
		t.Fatal("SSH persistence lost HTTP credentials or changed default")
	}
	for _, args := range [][]string{{"profile", "list"}, {"--profile", "old", "ping"}, {"--transport", "invalid", "ping"}, {"--identity", "/tmp/key", "ping"}, {"--transport", "http", "auth", "login", "ssh"}} {
		code, _, _ := captureCLI(t, func() int { return run(args) })
		if code != exitUsage {
			t.Fatal("accepted removed/conflicting option", args, code)
		}
	}
}
func TestCredentialBindingAndOldConfigIsNeverOverwritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	t.Setenv("BIBLEIT_CONFIG", path)
	for _, data := range []string{
		`{"version":2,"endpoint":"https://bibleit.app","ssh_server":"foreign.example:22","identity":"key"}`,
		`{"version":2,"endpoint":"https://foreign.example","ssh_server":"bibleit.app:22","access_token":"old"}`,
		`{"version":2,"endpoint":"http://localhost:8080","ssh_server":"127.0.0.1:2222","access_token":"old","identity":"key"}`,
		`{"version":2,"endpoint":"https://bibleit-server.fly.dev","ssh_server":"bibleit-server.fly.dev:22","access_token":"old","identity":"key"}`,
		`{"version":1,"active_profile":"a","profiles":{"a":{"transport":"http","endpoint":"http://localhost:8080","access_token":"old"}}}`,
	} {
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadConfig(); err == nil {
			t.Fatal("accepted old or foreign binding")
		}
		if err := saveConfig(config{AccessToken: "new"}); err == nil {
			t.Fatal("overwrote unsupported config")
		}
		after, _ := os.ReadFile(path)
		if string(after) != data {
			t.Fatal("changed old credentials")
		}
	}
}
