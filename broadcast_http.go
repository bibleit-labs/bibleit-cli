package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

// HTTP and SSH share retained write files and server authority. HTTP never opens
// a capture or falls back to a different credential or transport.
func runBroadcastHTTP(cfg config, a []string, format string, yes bool) int {
	if cfg.AccessToken == "" {
		return protocolFailure(authenticationError{message: "web authentication is not configured; run `bibleit auth login`"})
	}
	if !strings.HasPrefix(cfg.AccessToken, "bt_") {
		return fail(exitForbidden, "Broadcast Studio REST controls require an authorized personal token; no transport fallback was attempted")
	}
	p := a[2:]
	action := ""
	if p[0] == "setup" && len(p) > 2 && p[2] == "delete" {
		action = "Delete the selected saved Broadcast Studio Setup"
	}
	if p[0] == "event" && len(p) > 2 && p[2] == "end" {
		action = "End the selected YouTube broadcast"
	}
	if p[0] == "publisher" && len(p) > 2 && p[2] == "stop" {
		action = "Stop the selected publisher"
	}
	if action != "" && !yes {
		if e := confirmAction(cfg, action, format, os.Stdin, os.Stderr, terminalInput(os.Stdin) && terminalInput(os.Stderr)); e != nil {
			return outputError(exitUsage, e)
		}
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	client, e := bibleit.NewClient(cfg.AccessToken, cliHTTPClient)
	if e != nil {
		return protocolFailure(e)
	}
	control := broadcastHTTP{ctx: ctx, client: client, org: a[1]}
	result, e := control.execute(p)
	if e != nil {
		return protocolFailure(e)
	}
	return printHTTP(httpResult{data: result}, "broadcast", format)
}

type broadcastHTTP struct {
	ctx    context.Context
	client *bibleit.Client
	org    string
}

func (a broadcastHTTP) decode(r *bibleit.Response, e error) (map[string]any, error) {
	if e != nil {
		return nil, e
	}
	if r == nil {
		return nil, errors.New("missing Broadcast Studio response")
	}
	defer r.Body.Close()
	body, e := io.ReadAll(io.LimitReader(r.Body, 2*broadcastMaxDocument+1))
	if e != nil {
		return nil, e
	}
	if len(body) > 2*broadcastMaxDocument {
		return nil, errors.New("Broadcast Studio response exceeds its bounded size")
	}
	var result httpData
	if e = json.Unmarshal(body, &result); e != nil {
		return nil, errors.New("invalid Broadcast Studio JSON response")
	}
	return map[string]any(result), nil
}
func (a broadcastHTTP) execute(p []string) (map[string]any, error) {
	switch p[0] {
	case "profiles":
		return a.decode(a.client.ListBroadcastOutputProfiles(a.ctx, a.org))
	case "channel":
		return a.decode(a.client.GetBroadcastChannel(a.ctx, a.org))
	case "preflight":
		return a.decode(a.client.GetBroadcastPreflight(a.ctx, a.org))
	case "events":
		return a.decode(a.client.ListBroadcastEvents(a.ctx, a.org))
	case "publishers":
		return a.decode(a.client.ListBroadcastPublishers(a.ctx, a.org))
	case "takeover":
		return a.decode(a.client.RequestBroadcastTakeover(a.ctx, a.org, &bibleit.Options{JSON: struct{}{}}))
	case "operation":
		return a.decode(a.client.GetBroadcastOperation(a.ctx, a.org, p[1]))
	case "publisher":
		if p[2] == "get" {
			return a.decode(a.client.GetBroadcastPublisher(a.ctx, a.org, p[1]))
		}
		return a.decode(a.client.StopBroadcastPublisher(a.ctx, a.org, p[1], &bibleit.Options{IfMatch: `"publisher-` + p[3] + `"`, JSON: struct{}{}}))
	case "event":
		switch p[1] {
		case "create":
			body, e := readBroadcastJSON(p[3], broadcastChunkBytes)
			if e != nil {
				return nil, e
			}
			// Validate the same bounded metadata command before any provider operation.
			if _, e = bibleit.CreateBroadcastEventCommand(a.org, p[2], body); e != nil {
				return nil, e
			}
			return a.decode(a.client.CreateBroadcastEvent(a.ctx, a.org, &bibleit.Options{Body: bytes.NewReader(body), ContentLength: int64(len(body)), ContentType: "application/json", IdempotencyKey: p[2]}))
		case "attach":
			if _, e := bibleit.AttachBroadcastEventCommand(a.org, p[2], p[3]); e != nil {
				return nil, e
			}
			return a.decode(a.client.AttachBroadcastEvent(a.ctx, a.org, &bibleit.Options{JSON: bibleit.BroadcastAttachInput{ProviderBroadcastID: p[3]}, IdempotencyKey: p[2]}))
		}
		if p[2] == "get" {
			return a.decode(a.client.GetBroadcastEvent(a.ctx, a.org, p[1]))
		}
		state := map[string]bibleit.BroadcastTransitionInputState{"test": bibleit.BroadcastTransitionInputStateTesting, "go-live": bibleit.BroadcastTransitionInputStateLive, "end": bibleit.BroadcastTransitionInputStateComplete}[p[2]]
		return a.decode(a.client.TransitionBroadcastEvent(a.ctx, a.org, p[1], &bibleit.Options{JSON: bibleit.BroadcastTransitionInput{State: state}, IfMatch: `"event-` + p[3] + `"`, IdempotencyKey: p[4]}))
	case "write":
		r, e := loadBroadcastRequest(p[2])
		if e != nil {
			return nil, e
		}
		if r.Organization != a.org {
			return nil, errors.New("request file belongs to another workspace")
		}
		result, e := a.decode(a.client.GetBroadcastWriteRequest(a.ctx, a.org, r.RequestID))
		if e != nil {
			return nil, e
		}
		return broadcastHTTPReceipt(result)
	case "setup":
		if p[1] == "list" {
			q := url.Values{}
			if len(p) > 2 {
				q.Set("limit", p[2])
			}
			if len(p) > 3 {
				q.Set("after", p[3])
			}
			return a.decode(a.client.ListBroadcastSetups(a.ctx, a.org, &bibleit.Options{Query: q}))
		}
		if p[1] == "create" {
			return a.write(p)
		}
		if p[2] == "show" {
			return a.decode(a.client.GetBroadcastSetup(a.ctx, a.org, p[1]))
		}
		if p[2] == "export" {
			return a.export(p[1], p[3])
		}
		return a.write(p)
	}
	return nil, errors.New("unsupported Broadcast Studio operation")
}

func (a broadcastHTTP) export(id, path string) (map[string]any, error) {
	response, e := a.client.GetBroadcastSetup(a.ctx, a.org, id)
	if e != nil {
		return nil, e
	}
	defer response.Body.Close()
	body, e := io.ReadAll(io.LimitReader(response.Body, 2*broadcastMaxDocument+1))
	if e != nil || len(body) > 2*broadcastMaxDocument {
		return nil, errors.New("invalid Setup export response size")
	}
	var setup struct {
		ID           string          `json:"id"`
		Organization string          `json:"organization_id"`
		Revision     json.Number     `json:"revision"`
		Document     json.RawMessage `json:"document"`
	}
	if e = json.Unmarshal(body, &setup); e != nil || setup.ID != id || setup.Organization != a.org ||
		!positiveRevision(string(setup.Revision)) || response.Header.Get("ETag") != `"setup-`+string(setup.Revision)+`"` {
		return nil, errors.New("invalid revision-bound Setup export metadata")
	}
	document := bytes.TrimSpace(setup.Document)
	if len(document) < 2 || len(document) > broadcastMaxDocument || document[0] != '{' || !json.Valid(document) {
		return nil, errors.New("invalid Setup export document")
	}
	if e = writeBroadcastPrivate(path, document); e != nil {
		return nil, e
	}
	return map[string]any{"id": id, "organization_id": a.org, "revision": setup.Revision, "exported": true}, nil
}

func (a broadcastHTTP) write(p []string) (map[string]any, error) {
	expected, body, path, e := broadcastWriteInput(a.org, p)
	if e != nil {
		return nil, e
	}
	revision, e := strconv.ParseInt(expected.Revision, 10, 64)
	if e != nil {
		return nil, e
	}
	intent, e := bibleit.NewBroadcastWriteIntentInput(bibleit.BroadcastWriteOperation(expected.Operation), expected.Target, revision, body)
	if e != nil {
		return nil, e
	}
	retained, e := loadBroadcastRequest(path)
	if e == nil {
		expected.RequestID = retained.RequestID
		if retained != expected {
			return nil, errors.New("request file does not match these exact input bytes and revision; inspect its status before starting another operation")
		}
	} else {
		if _, statErr := os.Lstat(path); !os.IsNotExist(statErr) {
			return nil, e
		}
		result, e := a.decode(a.client.CreateBroadcastWriteRequest(a.ctx, a.org, &bibleit.Options{JSON: intent}))
		if e != nil {
			return nil, e
		}
		ticket, ok := result["request_id"].(string)
		if !ok || len(ticket) == 0 || len(ticket) > 768 {
			return nil, errors.New("invalid server request ticket")
		}
		expected.RequestID = ticket
		if _, e = bibleit.BroadcastWriteOptions(ticket, revision, body); e != nil {
			return nil, e
		}
		encoded, _ := json.MarshalIndent(expected, "", "  ")
		if e = writeBroadcastPrivate(path, append(encoded, '\n')); e != nil {
			return nil, e
		}
	}
	// A lost reply is reconciled with the original ticket and exact original bytes.
	result, e := a.decode(a.client.GetBroadcastWriteRequest(a.ctx, a.org, expected.RequestID))
	if e != nil {
		return nil, e
	}
	if ticket, ok := result["request_id"].(string); !ok || ticket != expected.RequestID {
		return nil, errors.New("write status returned a different request ticket")
	}
	state, _ := result["result_state"].(string)
	if oneOf(state, "applied", "deleted", "rejected") {
		return broadcastHTTPTerminalReceipt(result)
	}
	if state != "not_seen" {
		return nil, userMessage{"The retained write is " + safeLocalMessage(state) + ". Inspect `broadcast <workspace> write status <request-file>`; no replacement intent was created."}
	}
	options, e := bibleit.BroadcastWriteOptions(expected.RequestID, revision, body)
	if e != nil {
		return nil, e
	}
	switch expected.Operation {
	case "create":
		result, e = a.decode(a.client.CreateBroadcastSetup(a.ctx, a.org, options))
	case "save":
		result, e = a.decode(a.client.ReplaceBroadcastSetup(a.ctx, a.org, expected.Target, options))
	case "duplicate":
		result, e = a.decode(a.client.DuplicateBroadcastSetup(a.ctx, a.org, expected.Target, options))
	case "delete":
		result, e = a.decode(a.client.DeleteBroadcastSetup(a.ctx, a.org, expected.Target, options))
	default:
		return nil, errors.New("invalid Setup mutation")
	}
	if e != nil {
		return nil, e
	}
	if ticket, ok := result["request_id"].(string); !ok || ticket != expected.RequestID {
		return nil, errors.New("write result returned a different request ticket; reconcile the retained request")
	}
	return broadcastHTTPTerminalReceipt(result)
}
func broadcastHTTPTerminalReceipt(result map[string]any) (map[string]any, error) {
	state, _ := result["result_state"].(string)
	if !oneOf(state, "applied", "deleted", "rejected") {
		return nil, userMessage{"The write reply did not confirm a terminal outcome. Keep the request file and inspect its status before retrying."}
	}
	return broadcastHTTPReceipt(result)
}
func broadcastHTTPReceipt(result map[string]any) (map[string]any, error) {
	if result["result_state"] == "rejected" {
		code, _ := result["error_code"].(string)
		if code == "" {
			code = "invalid_request"
		}
		return nil, &bibleit.ServerError{Code: code, Fields: map[string]string{}}
	}
	return result, nil
}
