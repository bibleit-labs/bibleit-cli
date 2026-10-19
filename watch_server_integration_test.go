package main

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

type watchFixturePoll struct{ cursor, wait string }
type watchFixtureTransport struct {
	base  http.RoundTripper
	polls chan watchFixturePoll
}

func (w watchFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if strings.HasSuffix(r.URL.Path, "/events") {
		select {
		case w.polls <- watchFixturePoll{r.URL.Query().Get("after"), r.URL.Query().Get("wait")}:
		default:
		}
	}
	return w.base.RoundTrip(r)
}

// This test requires the egress-disabled disposable server harness. It records
// no credentials, request/response bodies or real user content in test output.
func TestServerRESTWatch(t *testing.T) {
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the disposable server container harness")
	}
	if os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("fixed endpoint fixture is required")
	}
	addresses, err := net.LookupIP("api.bibleit.app")
	if err != nil || len(addresses) == 0 {
		t.Fatal("fixture DNS unavailable")
	}
	for _, address := range addresses {
		if !address.IsLoopback() {
			t.Fatal("fixture DNS must remain loopback")
		}
	}
	for _, credential := range []struct{ name, token string }{{"personal", os.Getenv("BIBLEIT_INTEGRATION_TOKEN")}, {"service", os.Getenv("BIBLEIT_SERVICE_TOKEN")}, {"delegated", os.Getenv("BIBLEIT_OAUTH_TOKEN")}} {
		t.Run(credential.name, func(t *testing.T) {
			if credential.token == "" {
				t.Fatal("fixture credential missing")
			}
			for _, format := range []string{"ndjson", "human"} {
				t.Run(format, func(t *testing.T) {
					operation := func(args ...string) (httpData, error) {
						p, e := planHTTP(args, credential.name == "service")
						if e != nil {
							return nil, e
						}
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
						defer cancel()
						result, e := executeHTTP(ctx, credential.token, p)
						return result.data, e
					}
					created, e := operation("live", "create", "Disposable watch fixture")
					if e != nil {
						t.Fatal("fixture creation failed")
					}
					id, _ := created["id"].(string)
					if id == "" {
						t.Fatal("fixture ID missing")
					}
					defer func() {
						if _, e := operation("live", id, "delete"); e != nil {
							t.Error("fixture deletion failed")
						}
					}()
					if _, e := operation("live", id, "start"); e != nil {
						t.Fatal("fixture start failed")
					}

					originalClient := cliHTTPClient
					client := *originalClient
					base := client.Transport
					if base == nil {
						base = http.DefaultTransport
					}
					polls := make(chan watchFixturePoll, 32)
					client.Transport = watchFixtureTransport{base, polls}
					cliHTTPClient = &client
					defer func() { cliHTTPClient = originalClient }()
					file, e := os.CreateTemp(t.TempDir(), "watch-output-")
					if e != nil {
						t.Fatal("private output creation failed")
					}
					defer file.Close()
					originalStdout := os.Stdout
					os.Stdout = file
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					done := make(chan error, 1)
					go func() { done <- watchHTTP(ctx, credential.token, id, format) }()
					joined := false
					defer func() {
						cancel()
						if !joined {
							select {
							case <-done:
							case <-time.After(3 * time.Second):
								t.Error("watch did not release resources after cancellation")
							}
						}
						os.Stdout = originalStdout
					}()
					next := func() watchFixturePoll {
						t.Helper()
						select {
						case poll := <-polls:
							if poll.wait != "25" {
								t.Fatal("watch did not use bounded long polling")
							}
							return poll
						case <-time.After(5 * time.Second):
							t.Fatal("watch poll did not arrive")
							return watchFixturePoll{}
						}
					}
					if next().cursor != "" {
						t.Fatal("initial poll invented a cursor")
					}
					initial := next()
					if initial.cursor == "" {
						t.Fatal("initial snapshot cursor missing")
					}
					// An unchanged Live should leave one outstanding long poll, not spin.
					select {
					case <-polls:
						t.Fatal("unchanged Live caused immediate repeated polling")
					case <-time.After(200 * time.Millisecond):
					}
					if _, e := operation("live", id, "pause"); e != nil {
						t.Fatal("fixture pause failed")
					}
					changed := next()
					if changed.cursor == "" || changed.cursor == initial.cursor {
						t.Fatal("changed snapshot did not advance its opaque cursor")
					}
					cancel()
					select {
					case e := <-done:
						joined = true
						if e != nil {
							t.Fatal("watch failed during cancellation")
						}
					case <-time.After(3 * time.Second):
						t.Fatal("watch did not stop promptly")
					}
					os.Stdout = originalStdout
					if _, e := file.Seek(0, io.SeekStart); e != nil {
						t.Fatal("private output seek failed")
					}
					output, e := io.ReadAll(io.LimitReader(file, 1<<20))
					if e != nil {
						t.Fatal("private output read failed")
					}
					if format == "human" {
						text := string(output)
						if !strings.Contains(text, "Watching Live snapshots.") || !strings.Contains(text, "paused=false") || !strings.Contains(text, "paused=true") {
							t.Fatal("human watch did not render initial and changed snapshots")
						}
					} else {
						decoder := json.NewDecoder(strings.NewReader(string(output)))
						var sequences []int64
						var cursors []string
						var paused []bool
						for decoder.More() {
							var envelope struct {
								Version int `json:"schema_version"`
								Event   struct {
									Version  int    `json:"schema_version"`
									Type     string `json:"type"`
									Cursor   string `json:"cursor"`
									Sequence int64  `json:"sequence"`
									Data     struct {
										Live struct {
											Paused bool `json:"paused"`
										} `json:"live"`
									} `json:"data"`
								} `json:"event"`
							}
							if decoder.Decode(&envelope) != nil {
								t.Fatal("watch output was not valid NDJSON")
							}
							if envelope.Version != 2 || envelope.Event.Version != 1 {
								t.Fatal("watch schema versions changed")
							}
							if envelope.Event.Type == "stream.heartbeat" {
								continue
							}
							sequences = append(sequences, envelope.Event.Sequence)
							cursors = append(cursors, envelope.Event.Cursor)
							paused = append(paused, envelope.Event.Data.Live.Paused)
						}
						if len(sequences) != 2 || sequences[1] <= sequences[0] || cursors[0] != initial.cursor || cursors[1] != changed.cursor || paused[0] || !paused[1] {
							t.Fatal("watch did not emit initial and changed snapshots in sequence")
						}
					}
				})
			}
		})
	}
}
