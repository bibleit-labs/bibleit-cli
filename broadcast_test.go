package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

const broadcastTestOrg = "11111111-1111-1111-1111-111111111111"
const broadcastTestID = "22222222-2222-2222-2222-222222222222"

func TestBroadcastCapabilityAndGrammar(t *testing.T) {
	for _, spec := range broadcastRegistry() {
		a, e := normalizeCommand(strings.Fields(spec.Example))
		if e != nil || validateBroadcastArgs(a) != nil {
			t.Fatalf("%s: %v", spec.Syntax, e)
		}
	}
	for _, a := range [][]string{
		{"broadcast", "Mittel", "profiles"},
		{"broadcast", broadcastTestOrg, "setup", broadcastTestID, "save", "0", "a.json", "r.json"},
		{"broadcast", broadcastTestOrg, "setup", "list", "100"},
		{"broadcast", broadcastTestOrg, "event", broadcastTestID, "end", "1", "bad"},
		{"broadcast", broadcastTestOrg, "publisher", broadcastTestID, "stop", "-1"},
	} {
		if validateBroadcastArgs(a) == nil {
			t.Fatalf("accepted invalid command %v", a)
		}
	}
	if requireBroadcastCapability([]string{`OK capabilities="organization,broadcast.read.v1"`}, "broadcast.write.v1") == nil {
		t.Fatal("read-only server accepted writes")
	}
	if requireBroadcastCapability([]string{`OK capabilities="broadcast.read.v1,broadcast.write.v1"`}, "broadcast.write.v1") != nil {
		t.Fatal("write capability missing")
	}
	if _, e := broadcastWire("broadcast", broadcastTestOrg, "write", "status", "ticket\nAUTH"); e == nil {
		t.Fatal("accepted line injection")
	}
}

func TestBroadcastPrivateFilesAndJSONBounds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "setup.json")
	if e := writeBroadcastPrivate(path, []byte("{}")); e != nil {
		t.Fatal(e)
	}
	if e := writeBroadcastPrivate(path, []byte("changed")); e == nil {
		t.Fatal("overwrote existing file")
	}
	if b, e := readBroadcastJSON(path, 2); e != nil || string(b) != "{}" {
		t.Fatal(e)
	}
	if _, e := readBroadcastJSON(path, 1); e == nil {
		t.Fatal("ignored byte bound")
	}
	for _, b := range []string{"[]", "null", "", `{"x":`, strings.Repeat("a", 100)} {
		os.WriteFile(path, []byte(b), 0600)
		if _, e := readBroadcastJSON(path, 50); e == nil {
			t.Fatalf("accepted invalid input %q", b)
		}
	}
}

// A disposable OpenSSH process fixture exercises the same CLI subprocess and
// persistent shell adapter as a real authenticated server. It models a commit
// whose reply is lost, then reconciles it without another intent or mutation.
func installBroadcastSSHFixture(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX OpenSSH process fixture")
	}
	if _, e := exec.LookPath("python3"); e != nil {
		t.Skip("Python process fixture unavailable")
	}
	dir := t.TempDir()
	script := `#!/usr/bin/env python3
import sys,os,json,base64,hashlib
root=os.environ['BIBLEIT_BROADCAST_FIXTURE']
statepath=root+'/state.json'
def state():
 try:return json.load(open(statepath))
 except FileNotFoundError:return {'intents':0,'commits':0,'applied':False}
def save(s):
 with open(statepath,'w') as f:json.dump(s,f)
def execute(w):
 s=state();p=w.split()
 if w=='SERVER INFO':return 'OK capabilities="broadcast.read.v1,broadcast.write.v1,broadcast.lifecycle.v1"'
 if p[2:4]==['setup','list']:return 'OK count=0\nEND'
 if p[2:4]==['write','intent']:
  s['intents']+=1;save(s);return 'OK request_id="signed-ticket" input_hash="safe-hash" actor_id="actor"'
 if p[2:4]==['write','status']:return 'OK result_state="applied" applied=true id="22222222-2222-2222-2222-222222222222" revision=1' if s['applied'] else 'OK result_state="not_seen" applied=false'
 if p[2:4]==['write','begin']:return 'OK transfer_id="33333333-3333-3333-3333-333333333333" next_offset=0'
 if p[2:4]==['write','chunk']:
  data=base64.b64decode(p[6]);s['chunk_bytes']=len(data);save(s);return 'OK next_offset='+str(int(p[5])+len(data))
 if p[2:4]==['write','commit']:
  s['commits']+=1;s['applied']=True;save(s)
  if os.environ.get('BIBLEIT_BROADCAST_LOST_REPLY')=='1':sys.exit(0)
  return 'OK result_state="applied" applied=true id="22222222-2222-2222-2222-222222222222" revision=1'
 return 'ERR bad_command'
s=state();s['connections']=s.get('connections',0)+1;save(s)
if sys.argv[-1].startswith('bibleit-cli@'):
 print('Bibleit authenticated shell',flush=True)
 print('> ',end='',flush=True)
 for line in sys.stdin:
  print(execute(line.strip()),flush=True)
  print('> ',end='',flush=True)
else:print(execute(sys.argv[-1]),flush=True)
`
	path := filepath.Join(dir, "ssh")
	if e := os.WriteFile(path, []byte(script), 0700); e != nil {
		t.Fatal(e)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("BIBLEIT_BROADCAST_FIXTURE", dir)
	return dir
}
func TestBroadcastPersistentWriteLostReplyReconcilesSameTicket(t *testing.T) {
	dir := installBroadcastSSHFixture(t)
	cfg := config{Transport: "ssh", Endpoint: "localhost:22"}
	input := filepath.Join(dir, "input.json")
	request := filepath.Join(dir, "request.json")
	os.WriteFile(input, []byte(`{"name":"a"}`), 0600)
	p := []string{"setup", "create", input, request}
	t.Setenv("BIBLEIT_BROADCAST_LOST_REPLY", "1")
	if _, e := broadcastWrite(cfg, broadcastTestOrg, p); e == nil {
		t.Fatal("lost reply reported success")
	}
	r, e := loadBroadcastRequest(request)
	if e != nil || r.RequestID != "signed-ticket" {
		t.Fatal("ticket not retained", e)
	}
	lines, e := broadcastWrite(cfg, broadcastTestOrg, p)
	if e != nil {
		t.Fatal(e)
	}
	f, _ := broadcastFields(lines)
	if f["applied"] != "true" {
		t.Fatal("missing reconciled receipt")
	}
	var s struct{ Intents, Commits, Connections int }
	b, _ := os.ReadFile(filepath.Join(dir, "state.json"))
	json.Unmarshal(b, &s)
	if s.Intents != 1 || s.Commits != 1 || s.Connections != 2 {
		t.Fatalf("duplicated operation %+v", s)
	}
	os.WriteFile(input, []byte(`{"name":"changed"}`), 0600)
	if _, e = broadcastWrite(cfg, broadcastTestOrg, p); e == nil {
		t.Fatal("reused ticket for changed bytes")
	}
}
func TestBroadcastOutputModesAndAuthentication(t *testing.T) {
	for _, format := range []string{"table", "raw", "json"} {
		code, out, stderr := captureCLI(t, func() int {
			return printResponse([]string{`OK result_state="applied" applied=true revision=3`}, format)
		})
		if code != 0 || out == "" || stderr != "" {
			t.Fatalf("%s: %d %s", format, code, stderr)
		}
	}
	code, _, _ := captureCLI(t, func() int {
		return runBroadcast(config{Transport: "http"}, []string{"broadcast", broadcastTestOrg, "profiles"}, "json", false)
	})
	if code != exitUnauthenticated {
		t.Fatal("HTTP broadcast accepted missing authentication")
	}
	if _, e := bibleit.DecodeResponse(strings.NewReader("ERR forbidden\n")); e == nil {
		t.Fatal("fixture decoder ignored denial")
	}
}

func TestBroadcastOutputTypingAndBounds(t *testing.T) {
	r, e := typedRecord(outputRecord{Type: "ok", Fields: map[string]string{"applied": "true", "replay": "false", "revision": "2", "next_offset": "3072"}})
	if e != nil || r.Fields["applied"] != true || r.Fields["replay"] != false || r.Fields["next_offset"] != int64(3072) {
		t.Fatal("incorrect machine-readable broadcast types", r, e)
	}
	payload := `{"value":{"state":"connected"}}`
	r, e = typedRecord(outputRecord{Type: "ok", Fields: map[string]string{"document_format": "broadcast-lifecycle-json-v1", "data_base64": "eyJ2YWx1ZSI6eyJzdGF0ZSI6ImNvbm5lY3RlZCJ9fQ=="}})
	if e != nil || string(r.Payload) != payload || r.Fields["data_base64"] != nil {
		t.Fatal("lifecycle JSON payload not decoded", e)
	}
	buffer := broadcastBoundedBuffer{maximum: 4}
	if _, e = buffer.Write([]byte("1234")); e != nil {
		t.Fatal(e)
	}
	if _, e = buffer.Write([]byte("5")); e == nil || buffer.Len() != 4 {
		t.Fatal("output buffer bound exceeded")
	}
}

func TestBroadcastPersistentReadEndFraming(t *testing.T) {
	dir := installBroadcastSSHFixture(t)
	cfg := config{Transport: "ssh", Endpoint: "localhost:22"}
	session, e := openBroadcastShell(cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer session.close()
	cfg.broadcastSession = session
	for _, parts := range [][]string{{"setup", "list"}, {"setup", "list"}} {
		lines, e := broadcastRequest(cfg, broadcastTestOrg, parts...)
		if e != nil || len(lines) == 0 {
			t.Fatal("multi-record shell framing", e)
		}
	}
	lines, e := requestBroadcastSSH(cfg, "SERVER INFO")
	if e != nil || requireBroadcastCapability(lines, "broadcast.write.v1") != nil {
		t.Fatal("following response desynchronized", e)
	}
	var s struct{ Connections int }
	b, _ := os.ReadFile(filepath.Join(dir, "state.json"))
	json.Unmarshal(b, &s)
	if s.Connections != 1 {
		t.Fatal("read control requests opened multiple connections")
	}
}
