package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bibleit-labs/bibleit-go"
	"os"
	"os/signal"
	"time"
)

func runHTTPWatch(cfg config, id, format string) int {
	if format == "json" {
		return fail(exitUsage, "use --format ndjson for Live subscriptions")
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := watchHTTP(ctx, cfg.AccessToken, id, format); err != nil {
		return protocolFailure(err)
	}
	return exitOK
}
func watchHTTP(ctx context.Context, token, id, format string) error {
	// Finite commands keep their existing timeout. A bounded 25-second long poll
	// needs a larger deadline, including connection and response overhead.
	h := *cliHTTPClient
	if h.Timeout < 40*time.Second {
		h.Timeout = 40 * time.Second
	}
	client, err := bibleit.NewClient(token, &h)
	if err != nil {
		return err
	}
	return consumeHTTPWatch(ctx, client, id, format)
}
func consumeHTTPWatch(ctx context.Context, client *bibleit.Client, id, format string) error {
	cursor := ""
	first := true
	encoder := json.NewEncoder(os.Stdout)
	for {
		started := time.Now()
		event, err := client.PollLiveEvents(ctx, id, cursor, 25)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			return err
		} // No retries, cursor reset, replay or credential fallback.
		cursor = event.Cursor
		switch format {
		case "raw":
			err = encoder.Encode(event)
		case "ndjson":
			err = encoder.Encode(map[string]any{"schema_version": 2, "event": event})
		default:
			if event.Type != "stream.heartbeat" && event.Data != nil {
				var live httpData
				if err = decodeWatchSnapshot(event.Data.Live, &live); err != nil {
					return fmt.Errorf("invalid Live snapshot")
				}
				if first {
					fmt.Println("Watching Live snapshots. Press Ctrl-C to stop.")
					first = false
				}
				fmt.Printf("Live %s: %s · paused=%v\n", humanValue(id), humanValue(fmt.Sprint(live["status"])), live["paused"])
				var display httpData
				if err = decodeWatchSnapshot(event.Data.Display, &display); err != nil {
					return fmt.Errorf("invalid Live display snapshot")
				}
				presentHTTPHuman(map[string]any(display), "verse")
			}
		}
		if err != nil {
			return err
		}
		// A server/proxy returning a heartbeat immediately must not create a hot loop.
		if event.Type == "stream.heartbeat" {
			wait := time.Second - time.Since(started)
			if wait > 0 {
				timer := time.NewTimer(wait)
				select {
				case <-ctx.Done():
					timer.Stop()
					return nil
				case <-timer.C:
				}
			}
		}
	}
}

func decodeWatchSnapshot(value any, target *httpData) error {
	raw, e := json.Marshal(value)
	if e != nil {
		return e
	}
	return json.Unmarshal(raw, target)
}
