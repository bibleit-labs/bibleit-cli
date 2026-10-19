package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
)

func httpReply(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
func httpFixture(t *testing.T, token string, fn roundTripFunc) {
	t.Helper()
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	t.Setenv("BIBLEIT_TOKEN", token)
	old := cliHTTPClient
	cliHTTPClient = &http.Client{Transport: fn}
	t.Cleanup(func() { cliHTTPClient = old })
}
func TestHTTPReadStructuredScripture(t *testing.T) {
	calls := 0
	httpFixture(t, "bt_test", func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "api.bibleit.app" || r.Method != "GET" {
			t.Fatal(r.URL, r.Method)
		}
		switch r.URL.Path {
		case "/v1/me":
			return httpReply(200, `{"personal_account_id":"account"}`), nil
		case "/v1/books":
			return httpReply(200, `{"data":[{"id":43,"code":"JHN","names":{"en":"John"}}],"next_cursor":null}`), nil
		case "/v1/accounts/account/translations/web/books/43/chapters/3/verses/16":
			return httpReply(200, `{"translation":"web","rights":{"attribution":"Credit"},"verses":[{"book_id":43,"book":"John","chapter":3,"verse":16,"text":"A quoted \"verse\"\nwith a second line"}]}`), nil
		default:
			t.Fatal(r.URL)
			return nil, nil
		}
	})
	code, out, stderr := captureCLI(t, func() int { return run([]string{"--format", "raw", "read", "web", "John", "3", "16"}) })
	var data struct {
		Rights map[string]any `json:"rights"`
		Verses []struct {
			BookID     int `json:"book_id"`
			Book, Text string
		} `json:"verses"`
	}
	if code != exitOK || calls != 3 || json.Unmarshal([]byte(out), &data) != nil || data.Verses[0].BookID != 43 || data.Verses[0].Book != "John" || data.Rights["attribution"] != "Credit" || !strings.Contains(data.Verses[0].Text, "\n") {
		t.Fatal(code, out, stderr, calls)
	}
}
func TestHTTPConditionalMutationNoRetry(t *testing.T) {
	calls := 0
	httpFixture(t, "bt_test", func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/v1/lives/live-1/playback" && r.URL.Path != "/v1/lives/live-1" {
			t.Fatal(r.URL)
		}
		if r.Method == "GET" {
			res := httpReply(200, `{"id":"live-1"}`)
			res.Header.Set("ETag", `"live:9"`)
			return res, nil
		}
		b, _ := io.ReadAll(r.Body)
		if r.Method != "PATCH" || r.Header.Get("If-Match") != `"live:9"` || string(b) != `{"paused":true}` {
			t.Fatal(r, string(b))
		}
		return httpReply(412, `{"code":"precondition_failed","status":412}`), nil
	})
	code, _, _ := captureCLI(t, func() int { return run([]string{"live", "live-1", "pause"}) })
	if code != exitConflict || calls != 2 {
		t.Fatal(code, calls)
	}
}
func TestHTTPCollectionsFollowOpaqueCursors(t *testing.T) {
	pages := 0
	httpFixture(t, "svc_test", func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/organization" {
			return httpReply(200, `{"id":"org"}`), nil
		}
		if r.URL.Path != "/v1/organizations/org/translations" {
			t.Fatal(r.URL)
		}
		pages++
		if pages == 1 {
			return httpReply(200, `{"data":[{"slug":"web"}],"next_cursor":"opaque / +"}`), nil
		}
		if r.URL.Query().Get("after") != "opaque / +" {
			t.Fatal(r.URL)
		}
		return httpReply(200, `{"data":[{"slug":"kjv"}],"next_cursor":null}`), nil
	})
	p, err := planHTTP([]string{"translation", "list"}, true)
	if err != nil {
		t.Fatal(err)
	}
	res, err := executeHTTP(context.Background(), "svc_test", p)
	if err != nil || pages != 2 || len(res.data["data"].([]any)) != 2 {
		t.Fatal(res, err, pages)
	}
}
func TestHTTPUnsupportedCommandsNeverUseLegacyAPI(t *testing.T) {
	httpFixture(t, "bt_test", func(r *http.Request) (*http.Response, error) {
		t.Fatal("unsupported command contacted server", r.URL)
		return nil, nil
	})
	for _, args := range [][]string{{"server", "help"}, {"account", "quotas"}, {"account", "tokens"}, {"translation", "add", "web"}, {"read", "web", "John"}, {"live", "create"}, {"live", "delete", "all"}, {"live", "L1", "stats"}, {"live", "L1", "stack", "pop"}, {"live", "L1", "secret", "rotate"}} {
		code, _, _ := captureCLI(t, func() int { return run(append([]string{"--yes"}, args...)) })
		if code != exitUsage {
			t.Fatal(args, code)
		}
	}
}

func TestHTTPJSONUsesNativeResourceData(t *testing.T) {
	httpFixture(t, "bt_test", func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/v1/lives/L1" {
			t.Fatal(r.URL)
		}
		return httpReply(200, `{"id":"L1","sequence":9007199254740993,"paused":false}`), nil
	})
	code, out, stderr := captureCLI(t, func() int { return run([]string{"--format", "json", "live", "L1", "info"}) })
	var envelope struct {
		SchemaVersion int                        `json:"schema_version"`
		Data          map[string]json.RawMessage `json:"data"`
		Records       json.RawMessage            `json:"records"`
	}
	if code != exitOK || json.Unmarshal([]byte(out), &envelope) != nil || envelope.SchemaVersion != 2 || string(envelope.Data["sequence"]) != "9007199254740993" || string(envelope.Data["paused"]) != "false" || envelope.Records != nil {
		t.Fatal(code, out, stderr)
	}
}
