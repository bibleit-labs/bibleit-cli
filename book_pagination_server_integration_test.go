package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/bibleit-labs/bibleit-go/rest"
)

func TestServerBookPagination(t *testing.T) {
	requireDisposableCLIIntegration(t)
	for _, credential := range []struct{ name, token string }{
		{"personal", os.Getenv("BIBLEIT_INTEGRATION_TOKEN")},
		{"delegated", os.Getenv("BIBLEIT_OAUTH_TOKEN")},
		{"service", os.Getenv("BIBLEIT_SERVICE_TOKEN")},
	} {
		t.Run(credential.name, func(t *testing.T) {
			if credential.token == "" {
				t.Fatal("fixture credential missing")
			}
			directory := t.TempDir()
			t.Setenv("BIBLEIT_CONFIG", filepath.Join(directory, "config.yaml"))
			t.Setenv("BIBLEIT_TOKEN", credential.token)
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			client, err := rest.NewClient(credential.token, cliHTTPClient)
			if err != nil {
				t.Fatal("fixture client initialization failed")
			}
			a := httpClient{client: client, ctx: ctx, service: credential.name == "service"}
			key, id, err := a.workspace()
			if err != nil {
				t.Fatal("fixture workspace resolution failed")
			}
			operation := "readPersonalBook"
			if a.service {
				operation = "readBook"
			}
			params := map[string]string{key: id, "slug": "web", "book_id": "19"}
			defaultPage, _, err := a.call(operation, params, nil)
			if err != nil {
				t.Fatal("default bounded book request failed")
			}
			defaultVerses, ok := defaultPage["verses"].([]any)
			cursor, _ := defaultPage["next_cursor"].(string)
			if !ok || len(defaultVerses) != 50 || cursor == "" {
				t.Fatal("default book request did not return 50 verses and continuation")
			}
			// Independently read every fixture chapter to detect omissions in either
			// cursor traversal or the CLI's aggregation of book pages.
			var want []any
			chapterOperation := "readPersonalChapter"
			if a.service {
				chapterOperation = "readChapter"
			}
			for _, chapter := range []int{1, 23, 24} {
				chapterParams := map[string]string{key: id, "slug": "web", "book_id": "19", "chapter": strconv.Itoa(chapter)}
				data, _, err := a.call(chapterOperation, chapterParams, nil)
				if err != nil {
					t.Fatal("fixture chapter reference read failed")
				}
				verses, ok := data["verses"].([]any)
				if !ok {
					t.Fatal("fixture chapter reference has no verses")
				}
				want = append(want, verses...)
			}
			if len(want) != 104 || !reflect.DeepEqual(defaultVerses, want[:50]) {
				t.Fatal("default page differs from independently read source chapters")
			}
			pages := client.ReadPersonalBookPages(ctx, id, "web", "19", &rest.Options{Query: url.Values{"limit": {"7"}}})
			if a.service {
				pages = client.ReadBookPages(ctx, id, "web", "19", &rest.Options{Query: url.Values{"limit": {"7"}}})
			}
			var got []rest.Verse
			count := 0
			for page, err := range pages {
				if err != nil || len(page.Verses) > 7 {
					t.Fatal("bounded SDK traversal failed")
				}
				count++
				got = append(got, page.Verses...)
			}
			encoded, _ := json.Marshal(want)
			var expected []rest.Verse
			if json.Unmarshal(encoded, &expected) != nil || count < 2 || !reflect.DeepEqual(got, expected) {
				t.Fatal("paged traversal omitted, duplicated or changed a source verse")
			}
			code, out, _ := captureCLI(t, func() int { return run([]string{"--transport", "http", "--format", "raw", "read", "web", "19"}) })
			var aggregate httpData
			if code != exitOK || json.Unmarshal([]byte(out), &aggregate) != nil || !reflect.DeepEqual(aggregate["verses"], want) {
				t.Fatal("CLI automatic pagination did not preserve the complete book")
			}
			if _, exists := aggregate["next_cursor"]; exists {
				t.Fatal("complete CLI output retained a page cursor")
			}
			code, out, _ = captureCLI(t, func() int {
				return run([]string{"--transport", "http", "--format", "raw", "read", "web", "19", "--limit", "1"})
			})
			var first httpData
			if code != exitOK || json.Unmarshal([]byte(out), &first) != nil {
				t.Fatal("CLI explicit page request failed")
			}
			cursor, _ = first["next_cursor"].(string)
			if len(first["verses"].([]any)) != 1 || cursor == "" {
				t.Fatal("explicit first page lacks one verse and continuation")
			}
			code, out, _ = captureCLI(t, func() int {
				return run([]string{"--transport", "http", "--format", "raw", "read", "web", "19", "--limit", "2", "--cursor", cursor})
			})
			var second httpData
			if code != exitOK || json.Unmarshal([]byte(out), &second) != nil || !reflect.DeepEqual(second["verses"], want[1:3]) {
				t.Fatal("CLI continuation or changed page size lost source positions")
			}
			code, out, _ = captureCLI(t, func() int {
				return run([]string{"--transport", "http", "--format", "raw", "read", "web", "19", "--cursor", cursor})
			})
			var defaultContinuation httpData
			if code != exitOK || json.Unmarshal([]byte(out), &defaultContinuation) != nil || !reflect.DeepEqual(defaultContinuation["verses"], want[1:51]) {
				t.Fatal("cursor-only continuation did not use the default page size")
			}
			lastPage, _, err := a.call(operation, params, &rest.Options{Query: url.Values{"limit": {"500"}}})
			if err != nil || !reflect.DeepEqual(lastPage["verses"], want) {
				t.Fatal("maximum-sized terminal page differs from source chapters")
			}
			if next, exists := lastPage["next_cursor"]; !exists || next != nil {
				t.Fatal("terminal page must contain an explicit null cursor")
			}
			for _, malformed := range []string{"", "not-a-cursor"} {
				_, _, err := a.call(operation, params, &rest.Options{Query: url.Values{"after": {malformed}}})
				var problem *rest.ProblemError
				if !errors.As(err, &problem) || problem.Status != 422 || problem.Code != "invalid_cursor" {
					t.Fatal("malformed cursor did not return a descriptive typed error")
				}
			}
			otherBook := map[string]string{key: id, "slug": "web", "book_id": "18"}
			_, _, err = a.call(operation, otherBook, &rest.Options{Query: url.Values{"after": {cursor}}})
			var mismatch *rest.ProblemError
			if !errors.As(err, &mismatch) || mismatch.Status != 422 || mismatch.Code != "cursor_mismatch" {
				t.Fatal("a cursor for a different book was not rejected descriptively")
			}
		})
	}
}
