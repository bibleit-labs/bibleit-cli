package main

import (
	"bytes"
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Only a disposable loopback SSH daemon is accepted. The fixture owner creates
// the database, key and workspace; this test never uses production endpoints.
func TestBroadcastStagingSSHContract(t *testing.T) {
	if os.Getenv("BIBLEIT_BROADCAST_TEST_DISPOSABLE") != "1" {
		t.Skip("requires the isolated Broadcast Studio SSH staging fixture")
	}
	endpoint := os.Getenv("BIBLEIT_BROADCAST_TEST_ENDPOINT")
	host, _, e := net.SplitHostPort(endpoint)
	if e != nil || host != "127.0.0.1" {
		t.Fatal("staging endpoint must be IPv4 loopback")
	}
	org := os.Getenv("BIBLEIT_BROADCAST_TEST_ORG")
	if !broadcastUUID.MatchString(org) {
		t.Fatal("missing staging workspace UUID")
	}
	cfg := config{Transport: "ssh", Endpoint: endpoint, Identity: os.Getenv("BIBLEIT_BROADCAST_TEST_IDENTITY")}
	if cfg.Identity == "" {
		t.Fatal("missing disposable registered SSH identity")
	}
	session, e := openBroadcastShell(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer session.close()
	cfg.broadcastSession = session
	for _, capability := range []string{"broadcast.read.v1", "broadcast.write.v1", "broadcast.lifecycle.v1"} {
		info, e := requestBroadcastSSH(cfg, "SERVER INFO")
		if e != nil {
			t.Fatal(e)
		}
		if e = requireBroadcastCapability(info, capability); e != nil {
			t.Fatal(e)
		}
	}
	dir := t.TempDir()
	input := filepath.Join(dir, "input.json")
	doc := map[string]any{"version": 1, "name": "CLI isolated SSH staging", "outputProfile": map[string]any{"id": "1080p30", "version": 1}, "sources": []any{}, "scenes": []any{map[string]any{"id": "starting-soon", "name": "Starting soon", "layers": []any{}, "audioOverrides": []any{}, "clips": []any{}, "slate": true, "cameraOff": true, "audioMuted": true}}, "defaultSceneId": "starting-soon", "mixer": map[string]any{"channels": []any{}, "master": map[string]any{"gainDb": 0, "muted": false, "effects": []any{}}}}
	body, _ := json.Marshal(doc)
	body = append(body, bytes.Repeat([]byte(" "), broadcastMaxDocument-len(body))...)
	if e = os.WriteFile(input, body, 0600); e != nil {
		t.Fatal(e)
	}
	write := func(parts ...string) map[string]string {
		t.Helper()
		lines, e := broadcastWrite(cfg, org, parts)
		if e != nil {
			t.Fatal(e)
		}
		f, e := broadcastFields(lines)
		if e != nil {
			t.Fatal(e)
		}
		return f
	}
	get := func(parts ...string) map[string]string {
		t.Helper()
		lines, e := broadcastRequest(cfg, org, parts...)
		if e != nil {
			t.Fatal(e)
		}
		f, e := broadcastFields(lines)
		if e != nil {
			t.Fatal(e)
		}
		return f
	}
	createRequest := filepath.Join(dir, "create-request.json")
	created := write("setup", "create", input, createRequest)
	id := created["id"]
	if !broadcastUUID.MatchString(id) || created["revision"] != "1" || created["applied"] != "true" {
		t.Fatalf("create receipt: %+v", created)
	}
	deleted := false
	defer func() {
		if !deleted {
			current, e := broadcastRequest(cfg, org, "setup", id, "get")
			if e == nil {
				f, _ := broadcastFields(current)
				_, _ = broadcastWrite(cfg, org, []string{"setup", id, "delete", f["revision"], filepath.Join(dir, "cleanup-request.json")})
			}
		}
	}()
	replay := write("setup", "create", input, createRequest)
	if replay["id"] != id {
		t.Fatal("same-ticket retry duplicated setup")
	}
	status := write("write", "status", createRequest)
	if status["id"] != id || status["applied"] != "true" {
		t.Fatal("status did not reconcile create")
	}
	if get("setup", id, "get")["revision"] != "1" {
		t.Fatal("get changed revision")
	}
	if _, e = broadcastExport(cfg, org, id, filepath.Join(dir, "export.json")); e != nil {
		t.Fatal(e)
	}
	doc["name"] = "CLI isolated SSH edited"
	body, _ = json.Marshal(doc)
	os.WriteFile(input, body, 0600)
	saved := write("setup", id, "save", "1", input, filepath.Join(dir, "save-request.json"))
	if saved["revision"] != "2" {
		t.Fatal("save failed", saved)
	}
	conflictPath := filepath.Join(dir, "conflict-request.json")
	conflict := write("setup", id, "save", "1", input, conflictPath)
	if conflict["result_state"] != "rejected" || conflict["error_code"] != "conflict" {
		t.Fatal("stale save did not give durable conflict", conflict)
	}
	conflict = write("write", "status", conflictPath)
	if conflict["error_code"] != "conflict" {
		t.Fatal("conflict status not retained")
	}
	duplicate := write("setup", id, "duplicate", "2", "CLI isolated copy", filepath.Join(dir, "duplicate-request.json"))
	copyID := duplicate["id"]
	if !broadcastUUID.MatchString(copyID) || copyID == id {
		t.Fatal("duplicate failed", duplicate)
	}
	write("setup", copyID, "delete", "1", filepath.Join(dir, "copy-delete-request.json"))
	deletePath := filepath.Join(dir, "delete-request.json")
	receipt := write("setup", id, "delete", "2", deletePath)
	if receipt["applied"] != "true" {
		t.Fatal("delete failed", receipt)
	}
	deleted = true
	if write("write", "status", deletePath)["id"] != id {
		t.Fatal("deleted object receipt lost")
	}
	for _, command := range [][]string{{"profiles"}, {"setup", "list"}, {"channel"}, {"events"}, {"publishers"}, {"preflight"}} {
		lines, e := broadcastRequest(cfg, org, command...)
		if e != nil || len(lines) == 0 {
			t.Fatalf("read %s: %v", strings.Join(command, " "), e)
		}
	}
	t.Log("Passed actual SSH full-1MiB create/get/export/save/stale-conflict/duplicate/delete, durable retries/status and lifecycle reads; disposable Setups deleted")
}
