package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestRESTWatchUsesBoundedPollOpaqueCursorAndNativeEnvelope(t *testing.T) {
	calls := 0
	httpFixture(t, "bt_watch", func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Path != "/v1/lives/L1/events" || r.URL.Query().Get("wait") != "25" {
			t.Fatal(r.URL)
		}
		if calls == 1 {
			if r.URL.Query().Get("after") != "" {
				t.Fatal("initial cursor invented")
			}
			return httpReply(200, `{"schema_version":1,"type":"live.snapshot","resource_type":"live","resource_id":"L1","sequence":1,"cursor":"opaque / +","data":{"live":{"id":"L1","status":"running","paused":false},"display":{"type":"empty","verses":[]},"queue":[]}}`), nil
		}
		if r.URL.Query().Get("after") != "opaque / +" {
			t.Fatal("cursor was modified")
		}
		return httpReply(401, `{"code":"unauthorized","status":401,"title":"PRIVATE_CANARY"}`), nil
	})
	old := activeFormat
	activeFormat = "ndjson"
	defer func() { activeFormat = old }()
	code, out, stderr := captureCLI(t, func() int { return protocolFailure(watchHTTP(context.Background(), "bt_watch", "L1", "ndjson")) })
	if code != exitUnauthenticated || calls != 2 || strings.Contains(stderr, "CANARY") {
		t.Fatal(code, calls, stderr)
	}
	var line struct {
		Version int `json:"schema_version"`
		Event   struct {
			Version int    `json:"schema_version"`
			Cursor  string `json:"cursor"`
		} `json:"event"`
	}
	if json.Unmarshal([]byte(out), &line) != nil || line.Version != 2 || line.Event.Version != 1 || line.Event.Cursor != "opaque / +" {
		t.Fatal(out)
	}
}
func TestRESTWatchCancellationAndJSONRejection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	httpFixture(t, "bt_watch", func(r *http.Request) (*http.Response, error) { return nil, context.Canceled })
	if e := watchHTTP(ctx, "bt_watch", "L1", "ndjson"); e != nil {
		t.Fatal(e)
	}
	code, out, _ := captureCLI(t, func() int { return runHTTPWatch(config{AccessToken: "bt_watch"}, "L1", "json") })
	if code != exitUsage || out != "" {
		t.Fatal(code, out)
	}
}
