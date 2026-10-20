package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func cliBookPage(verse int, cursor string) string {
	return fmt.Sprintf(`{"translation":"web","rights":{"edition":"fixture"},"verses":[{"book_id":19,"book":"Psalms","chapter":1,"verse":%d,"text":"fixture"}],"next_cursor":%s}`, verse, cursor)
}

func TestHTTPBookReadFollowsPagesAndPreservesCompleteOutput(t *testing.T) {
	for _, token := range []string{"bt_fixture", "oa_fixture", "svc_fixture"} {
		t.Run(token[:3], func(t *testing.T) {
			pages := 0
			httpFixture(t, token, func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/v1/me" {
					return httpReply(200, `{"personal_account_id":"account"}`), nil
				}
				if r.URL.Path == "/v1/organization" {
					return httpReply(200, `{"id":"org"}`), nil
				}
				pages++
				if !strings.HasSuffix(r.URL.Path, "/translations/web/books/19") || r.URL.Query().Get("limit") != "50" {
					t.Fatal(r.URL)
				}
				cursor := `"opaque / +"`
				if pages == 2 {
					if r.URL.Query().Get("after") != "opaque / +" {
						t.Fatal(r.URL)
					}
					cursor = "null"
				}
				return httpReply(200, cliBookPage(pages, cursor)), nil
			})
			code, out, stderr := captureCLI(t, func() int { return run([]string{"--format", "raw", "read", "web", "19"}) })
			var result map[string]json.RawMessage
			if code != exitOK || stderr != "" || pages != 2 || json.Unmarshal([]byte(out), &result) != nil || result["next_cursor"] != nil {
				t.Fatal(code, pages, out, stderr)
			}
			var verses []map[string]any
			if json.Unmarshal(result["verses"], &verses) != nil || len(verses) != 2 || verses[1]["verse"] != float64(2) {
				t.Fatal(out)
			}
		})
	}
}

func TestHTTPExplicitBookPageKeepsCursorAndDoesNotFollow(t *testing.T) {
	for _, format := range []string{"json", "raw", "table"} {
		t.Run(format, func(t *testing.T) {
			pages := 0
			httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/v1/me" {
					return httpReply(200, `{"personal_account_id":"account"}`), nil
				}
				pages++
				if r.URL.Query().Get("limit") != "1" || r.URL.Query().Get("after") != "opaque / +" {
					t.Fatal(r.URL)
				}
				return httpReply(200, cliBookPage(2, `"next-cursor"`)), nil
			})
			code, out, stderr := captureCLI(t, func() int {
				return run([]string{"read", "web", "19", "--limit", "1", "--cursor", "opaque / +", "--format", format})
			})
			if code != exitOK || stderr != "" || pages != 1 || !strings.Contains(out, "next-cursor") {
				t.Fatal(code, pages, out, stderr)
			}
		})
	}
}

func TestHTTPBookPageFailuresNeverPrintPartialBook(t *testing.T) {
	for _, failure := range []string{"stale_cursor", "repeated_cursor"} {
		t.Run(failure, func(t *testing.T) {
			pages := 0
			httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/v1/me" {
					return httpReply(200, `{"personal_account_id":"account"}`), nil
				}
				pages++
				if pages == 1 {
					return httpReply(200, cliBookPage(1, `"next"`)), nil
				}
				if failure == "stale_cursor" {
					return httpReply(409, `{"status":409,"code":"stale_cursor","detail":"PRIVATE_CANARY"}`), nil
				}
				return httpReply(200, cliBookPage(2, `"next"`)), nil
			})
			code, out, stderr := captureCLI(t, func() int { return run([]string{"--format", "json", "read", "web", "19"}) })
			if code == exitOK || pages != 2 || out != "" || strings.Contains(stderr, "PRIVATE_CANARY") || !strings.Contains(strings.ToLower(stderr), "restart") {
				t.Fatal(code, pages, out, stderr)
			}
		})
	}
}

func TestBookPageFlagsRejectInvalidScopeAndValuesBeforeNetwork(t *testing.T) {
	httpFixture(t, "bt_fixture", func(r *http.Request) (*http.Response, error) {
		t.Fatal("invalid page flags contacted server", r.URL)
		return nil, nil
	})
	for _, args := range [][]string{
		{"read", "web", "19", "--limit", "0"}, {"read", "web", "19", "--limit", "501"},
		{"read", "web", "19", "--limit", "a"}, {"read", "web", "19", "--cursor", ""},
		{"read", "web", "19", "1", "--limit", "1"}, {"read", "web", "19", "1:1", "--cursor", "next"},
		{"search", "web", "word", "--limit", "1"}, {"organization", "list", "--limit", "1"},
		{"--transport", "ssh", "read", "web", "19", "--limit", "1"},
	} {
		code, out, _ := captureCLI(t, func() int { return run(args) })
		if code != exitUsage || out != "" {
			t.Fatal(args, code, out)
		}
	}
}
