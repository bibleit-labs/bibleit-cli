package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

const broadcastMaxDocument = 1048576
const broadcastChunkBytes = 3072

var broadcastUUID = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

func broadcastRegistry() []commandSpec {
	const org = "11111111-1111-1111-1111-111111111111"
	const id = "22222222-2222-2222-2222-222222222222"
	specs := []struct{ syntax, summary, example string }{
		{"profiles", "Inspect the supported output profile.", "profiles"},
		{"setup list [limit] [after]", "List saved Setups with explicit workspace pagination.", "setup list"},
		{"setup <id> show", "Inspect a saved Setup revision and document hash.", "setup " + id + " show"},
		{"setup <id> export <file>", "Export a verified, revision-bound Setup to a private file.", "setup " + id + " export setup.json"},
		{"setup create <file> <request-file>", "Create a Setup; persist its request ticket before uploading.", "setup create setup.json request.json"},
		{"setup <id> save <revision> <file> <request-file>", "Save a Setup using a required current revision and durable request file.", "setup " + id + " save 1 setup.json request.json"},
		{"setup <id> duplicate <revision> <name> <request-file>", "Duplicate the given Setup revision with a new name.", "setup " + id + " duplicate 1 Copy request.json"},
		{"setup <id> delete <revision> <request-file>", "Delete a Setup revision; confirmation required.", "setup " + id + " delete 1 request.json"},
		{"write status <request-file>", "Recover a write outcome without issuing another mutation.", "write status request.json"},
		{"channel", "Inspect the connected YouTube channel without credentials.", "channel"},
		{"events", "List YouTube broadcast events.", "events"},
		{"event <id> get", "Inspect one broadcast event.", "event " + id + " get"},
		{"event create <request-id> <metadata-file>", "Create a YouTube event using a caller-retained request UUID.", "event create " + id + " event.json"},
		{"event attach <request-id> <youtube-id>", "Attach an existing YouTube event using an idempotency UUID.", "event attach " + id + " abcdefghijk"},
		{"event <id> test <revision> <request-id>", "Transition an event to testing; requires an active browser publisher.", "event " + id + " test 1 " + id},
		{"event <id> go-live <revision> <request-id>", "Go live using the current revision and retained request UUID.", "event " + id + " go-live 1 " + id},
		{"event <id> end <revision> <request-id>", "End the broadcast; confirmation required.", "event " + id + " end 1 " + id},
		{"operation <request-id> get", "Reconcile a provider operation before deciding whether to retry.", "operation " + id + " get"},
		{"publishers", "List existing browser publisher sessions.", "publishers"},
		{"publisher <id> get", "Inspect an existing publisher session.", "publisher " + id + " get"},
		{"publisher <id> stop <revision>", "Stop an existing publisher at its current revision; confirmation required.", "publisher " + id + " stop 1"},
		{"preflight", "Check workspace broadcast prerequisites.", "preflight"},
		{"takeover", "Check takeover access; a fresh browser capture is required.", "takeover"},
	}
	result := make([]commandSpec, 0, len(specs))
	for _, s := range specs {
		result = append(result, commandSpec{"broadcast <organization> " + s.syntax, s.summary, "broadcast " + org + " " + s.example, false, ""})
	}
	return result
}

func positiveRevision(s string) bool {
	n, e := strconv.ParseInt(s, 10, 64)
	return e == nil && n > 0 && strconv.FormatInt(n, 10) == s
}
func validateBroadcastArgs(a []string) error {
	bad := errors.New("use a canonical workspace UUID, canonical resource/request UUIDs and a positive revision; run `bibleit help broadcast`")
	if len(a) < 3 || !broadcastUUID.MatchString(a[1]) {
		return bad
	}
	p := a[2:]
	switch p[0] {
	case "profiles", "channel", "events", "publishers", "preflight", "takeover":
		if len(p) != 1 {
			return bad
		}
	case "setup":
		if len(p) < 2 {
			return bad
		}
		if p[1] == "list" {
			if len(p) > 4 {
				return bad
			}
			if len(p) >= 3 {
				n, e := strconv.Atoi(p[2])
				if e != nil || n < 1 || n > 99 {
					return bad
				}
			}
			if len(p) == 4 && !broadcastUUID.MatchString(p[3]) {
				return bad
			}
			return nil
		}
		if p[1] == "create" {
			if len(p) != 4 {
				return bad
			}
			return nil
		}
		if !broadcastUUID.MatchString(p[1]) || len(p) < 3 {
			return bad
		}
		switch p[2] {
		case "show":
			if len(p) != 3 {
				return bad
			}
		case "export":
			if len(p) != 4 {
				return bad
			}
		case "save", "duplicate":
			if len(p) != 6 || !positiveRevision(p[3]) {
				return bad
			}
		case "delete":
			if len(p) != 5 || !positiveRevision(p[3]) {
				return bad
			}
		default:
			return bad
		}
	case "write":
		if len(p) != 3 || p[1] != "status" {
			return bad
		}
	case "event":
		if len(p) < 3 {
			return bad
		}
		switch p[1] {
		case "create":
			if len(p) != 4 || !broadcastUUID.MatchString(p[2]) {
				return bad
			}
		case "attach":
			if len(p) != 4 {
				return bad
			}
			if _, err := bibleit.AttachBroadcastEventCommand(a[1], p[2], p[3]); err != nil {
				return bad
			}
		default:
			if !broadcastUUID.MatchString(p[1]) {
				return bad
			}
			if p[2] == "get" {
				if len(p) != 3 {
					return bad
				}
			} else if !oneOf(p[2], "test", "go-live", "end") || len(p) != 5 || !positiveRevision(p[3]) || !broadcastUUID.MatchString(p[4]) {
				return bad
			}
		}
	case "operation":
		if len(p) != 3 || !broadcastUUID.MatchString(p[1]) || p[2] != "get" {
			return bad
		}
	case "publisher":
		if len(p) < 3 || !broadcastUUID.MatchString(p[1]) {
			return bad
		}
		if p[2] == "get" {
			if len(p) != 3 {
				return bad
			}
		} else if p[2] != "stop" || len(p) != 4 || !positiveRevision(p[3]) {
			return bad
		}
	default:
		return bad
	}
	return nil
}

func requireBroadcastCapability(lines []string, capability string) error {
	for _, r := range parseRecords(lines) {
		for _, c := range strings.Split(r.Fields["capabilities"], ",") {
			if c == capability {
				return nil
			}
		}
	}
	return userMessage{"This server does not advertise " + capability + ". Upgrade/configure the server or use Broadcast Studio in your browser."}
}

func broadcastWire(parts ...string) (string, error) {
	if len(parts) < 3 || !strings.EqualFold(parts[0], "broadcast") {
		return "", errors.New("invalid broadcast command")
	}
	command, err := bibleit.BroadcastCommand(parts[1], parts[2:]...)
	if err != nil {
		return "", err
	}
	return command.String(), nil
}

func broadcastSSHCommand(ctx context.Context, cfg config, wire string, shell bool) (*exec.Cmd, string, error) {
	cmd, identity, e := sshCommand(cfg, bibleit.PingCommand())
	if e != nil {
		return nil, "", e
	}
	args := append([]string(nil), cmd.Args[1:len(cmd.Args)-1]...)
	args = append([]string{"-o", "BatchMode=yes", "-o", "ServerAliveInterval=10", "-o", "ServerAliveCountMax=2"}, args...)
	if !shell {
		args = append(args, wire)
	}
	return exec.CommandContext(ctx, cmd.Path, args...), identity, nil
}

// Bound subprocess output before buffering, including on a malformed server.
type broadcastBoundedBuffer struct {
	bytes.Buffer
	maximum int
}

func (b *broadcastBoundedBuffer) Write(p []byte) (int, error) {
	if len(p) > b.maximum-b.Len() {
		return 0, errors.New("SSH response exceeds its bounded size")
	}
	return b.Buffer.Write(p)
}

func requestBroadcastSSH(cfg config, wire string) ([]string, error) {
	if cfg.broadcastSession != nil {
		return cfg.broadcastSession.request(wire)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd, identity, e := broadcastSSHCommand(ctx, cfg, wire, false)
	if e != nil {
		return nil, e
	}
	out := broadcastBoundedBuffer{maximum: 2 * 1048576}
	stderr := broadcastBoundedBuffer{maximum: 65536}
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if e = cmd.Run(); e != nil {
		return nil, sshCommandError(identity, stderr.String(), e)
	}
	response, e := bibleit.DecodeResponse(&out)
	return response.Lines, e
}
func broadcastRequest(cfg config, org string, parts ...string) ([]string, error) {
	wire, e := broadcastWire(append([]string{"broadcast", org}, parts...)...)
	if e != nil {
		return nil, e
	}
	return requestBroadcastSSH(cfg, wire)
}
func broadcastFields(lines []string) (map[string]string, error) {
	for _, line := range lines {
		r, e := bibleit.ParseRecord(line)
		if e != nil {
			return nil, e
		}
		if r.Type == "ok" {
			return r.Fields, nil
		}
	}
	return nil, errors.New("missing broadcast response envelope")
}

func runBroadcast(cfg config, a []string, format string, yes bool) int {
	if cfg.Transport == "http" {
		return runBroadcastHTTP(cfg, a, format, yes)
	}
	if cfg.Transport != "ssh" {
		return fail(exitUsage, "Broadcast Studio requires http or ssh transport")
	}
	session, e := openBroadcastShell(cfg)
	if e != nil {
		return protocolFailure(e)
	}
	defer session.close()
	cfg.broadcastSession = session
	p := a[2:]
	capability := "broadcast.lifecycle.v1"
	if p[0] == "profiles" || p[0] == "setup" {
		capability = "broadcast.read.v1"
	}
	if p[0] == "write" || p[0] == "setup" && (p[1] == "create" || len(p) > 2 && oneOf(p[2], "save", "duplicate", "delete")) {
		capability = "broadcast.write.v1"
	}
	info, e := requestBroadcastSSH(cfg, "SERVER INFO")
	if e != nil {
		return protocolFailure(e)
	}
	if e = requireBroadcastCapability(info, capability); e != nil {
		return outputError(exitUsage, e)
	}
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
		if e = confirmAction(cfg, action, format, os.Stdin, os.Stderr, terminalInput(os.Stdin) && terminalInput(os.Stderr)); e != nil {
			return outputError(exitUsage, e)
		}
	}
	var lines []string
	if capability == "broadcast.write.v1" {
		lines, e = broadcastWrite(cfg, a[1], p)
	} else if p[0] == "setup" && len(p) > 2 && p[2] == "export" {
		lines, e = broadcastExport(cfg, a[1], p[1], p[3])
	} else {
		wireParts := append([]string(nil), p...)
		if p[0] == "setup" && len(p) > 2 && p[2] == "show" {
			wireParts[2] = "get"
		}
		if p[0] == "event" && p[1] == "create" {
			var b []byte
			b, e = readBroadcastJSON(p[3], broadcastChunkBytes)
			if e == nil {
				wireParts[3] = base64.StdEncoding.EncodeToString(b)
			}
		}
		if e == nil {
			lines, e = broadcastRequest(cfg, a[1], wireParts...)
		}
	}
	if e != nil {
		return protocolFailure(e)
	}
	if fields, e := broadcastFields(lines); e == nil && fields["result_state"] == "rejected" {
		return protocolFailure(&bibleit.ServerError{Code: fields["error_code"], Fields: map[string]string{}})
	}
	return printResponse(lines, format)
}

func readBroadcastJSON(path string, limit int) ([]byte, error) {
	info, e := os.Lstat(path)
	if e != nil || !info.Mode().IsRegular() {
		return nil, errors.New("Broadcast Studio input must be a regular file; symlinks and device files are unsupported")
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, userMessage{"Cannot open the Broadcast Studio JSON file."}
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		return nil, errors.New("Broadcast Studio input must be a regular file")
	}
	b, e := io.ReadAll(io.LimitReader(f, int64(limit)+1))
	if e != nil || len(b) > limit {
		return nil, errors.New("Broadcast Studio JSON exceeds its bounded payload size")
	}
	if len(b) == 0 || !json.Valid(b) || len(bytes.TrimSpace(b)) == 0 || bytes.TrimSpace(b)[0] != '{' {
		return nil, errors.New("Broadcast Studio payload must be a JSON object")
	}
	return b, nil
}

func broadcastExport(cfg config, org, id, path string) ([]string, error) {
	if cfg.broadcastSession == nil {
		s, e := openBroadcastShell(cfg)
		if e != nil {
			return nil, e
		}
		defer s.close()
		cfg.broadcastSession = s
	}
	meta, e := broadcastRequest(cfg, org, "setup", id, "get")
	if e != nil {
		return nil, e
	}
	fields, e := broadcastFields(meta)
	if e != nil {
		return nil, e
	}
	revision := fields["revision"]
	size, e := strconv.Atoi(fields["document_bytes"])
	if e != nil || size < 1 || size > broadcastMaxDocument || !positiveRevision(revision) {
		return nil, errors.New("invalid Setup export metadata")
	}
	var body bytes.Buffer
	for offset := 0; offset < size; {
		lines, e := broadcastRequest(cfg, org, "setup", id, "document", revision, strconv.Itoa(offset))
		if e != nil {
			return nil, e
		}
		f, e := broadcastFields(lines)
		if e != nil {
			return nil, e
		}
		chunk, e := base64.StdEncoding.Strict().DecodeString(f["data_base64"])
		if e != nil || len(chunk) == 0 || len(chunk) > broadcastChunkBytes || offset+len(chunk) > size || f["revision"] != revision || f["document_sha256"] != fields["document_sha256"] || f["offset"] != strconv.Itoa(offset) || f["next_offset"] != strconv.Itoa(offset+len(chunk)) {
			return nil, errors.New("invalid revision-bound Setup export chunk")
		}
		body.Write(chunk)
		offset += len(chunk)
	}
	sum := sha256.Sum256(body.Bytes())
	if hex.EncodeToString(sum[:]) != fields["document_sha256"] || !json.Valid(body.Bytes()) {
		return nil, errors.New("Setup export failed integrity verification")
	}
	if e = writeBroadcastPrivate(path, body.Bytes()); e != nil {
		return nil, e
	}
	return meta, nil
}

// Request files never contain Setup documents or provider credentials. The
// retained ticket binds exact input bytes; a retry never creates a fresh intent.
type broadcastWriteRequest struct {
	Version      int    `json:"version"`
	Organization string `json:"organization_id"`
	Operation    string `json:"operation"`
	Target       string `json:"target_id"`
	Revision     string `json:"expected_revision"`
	PayloadHash  string `json:"payload_sha256"`
	RequestID    string `json:"request_id"`
}

func writeBroadcastPrivate(path string, b []byte) error {
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return userMessage{"Cannot create the private output/request file. Choose a new path; existing files are never overwritten."}
	}
	success := false
	defer func() {
		f.Close()
		if !success {
			os.Remove(path)
		}
	}()
	if e = restrictConfigFile(f); e != nil {
		return e
	}
	if _, e = f.Write(b); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	success = true
	return nil
}
func loadBroadcastRequest(path string) (broadcastWriteRequest, error) {
	var r broadcastWriteRequest
	if e := validateCredentialFile(path); e != nil {
		return r, e
	}
	f, e := os.Open(path)
	if e != nil {
		return r, userMessage{"Cannot open the retained Broadcast Studio request file."}
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, 4097))
	if e != nil || len(b) > 4096 {
		return r, errors.New("invalid retained Broadcast Studio request")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(&r); e != nil || r.Version != 1 || !broadcastUUID.MatchString(r.Organization) || len(r.RequestID) == 0 || len(r.RequestID) > 768 {
		return r, errors.New("invalid retained Broadcast Studio request")
	}
	if d.Decode(new(any)) != io.EOF {
		return r, errors.New("invalid trailing retained request data")
	}
	return r, nil
}
func broadcastWriteInput(org string, p []string) (broadcastWriteRequest, []byte, string, error) {
	var op, target, revision, path string
	var body []byte
	var e error
	if p[1] == "create" {
		op, target, revision, path = "create", "new", "0", p[3]
		body, e = readBroadcastJSON(p[2], broadcastMaxDocument)
	} else {
		op, target, revision = p[2], p[1], p[3]
		switch op {
		case "save":
			path = p[5]
			body, e = readBroadcastJSON(p[4], broadcastMaxDocument)
		case "duplicate":
			path = p[5]
			body, e = json.Marshal(map[string]string{"name": p[4]})
		case "delete":
			path = p[4]
			body = []byte("{}")
		}
	}
	if e != nil {
		return broadcastWriteRequest{}, nil, "", e
	}
	sum := sha256.Sum256(body)
	return broadcastWriteRequest{1, org, op, target, revision, hex.EncodeToString(sum[:]), ""}, body, path, nil
}
func broadcastWrite(cfg config, org string, p []string) ([]string, error) {
	if cfg.broadcastSession == nil {
		s, e := openBroadcastShell(cfg)
		if e != nil {
			return nil, e
		}
		defer s.close()
		cfg.broadcastSession = s
	}
	if p[0] == "write" {
		r, e := loadBroadcastRequest(p[2])
		if e != nil {
			return nil, e
		}
		if r.Organization != org {
			return nil, errors.New("request file belongs to another workspace")
		}
		return broadcastRequest(cfg, org, "write", "status", r.RequestID)
	}
	expected, body, path, e := broadcastWriteInput(org, p)
	if e != nil {
		return nil, e
	}
	op, target, revision := expected.Operation, expected.Target, expected.Revision
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
		lines, e := broadcastRequest(cfg, org, "write", "intent", op, target, revision, strconv.Itoa(len(body)), expected.PayloadHash)
		if e != nil {
			return nil, e
		}
		f, e := broadcastFields(lines)
		if e != nil {
			return nil, e
		}
		expected.RequestID = f["request_id"]
		if expected.RequestID == "" || len(expected.RequestID) > 768 {
			return nil, errors.New("invalid server request ticket")
		}
		encoded, _ := json.MarshalIndent(expected, "", "  ")
		if e = writeBroadcastPrivate(path, append(encoded, '\n')); e != nil {
			return nil, e
		}
	}
	// Always reconcile first. A previous commit may have succeeded without a reply.
	lines, e := broadcastRequest(cfg, org, "write", "status", expected.RequestID)
	if e != nil {
		return nil, e
	}
	f, e := broadcastFields(lines)
	if e != nil {
		return nil, e
	}
	state := f["result_state"]
	if state == "" {
		state = f["state"]
	}
	if f["applied"] == "true" || oneOf(state, "applied", "rejected", "deleted") {
		return lines, nil
	}
	if state != "not_seen" {
		return nil, userMessage{"The retained write is " + safeLocalMessage(state) + ". Inspect `broadcast <workspace> write status <request-file>`; no replacement intent was created."}
	}
	session := cfg.broadcastSession
	request := func(parts ...string) ([]string, error) {
		wire, e := broadcastWire(append([]string{"broadcast", org, "write"}, parts...)...)
		if e != nil {
			return nil, e
		}
		return session.request(wire)
	}
	lines, e = request("begin", expected.RequestID)
	if e != nil {
		return nil, e
	}
	f, e = broadcastFields(lines)
	if e != nil {
		return nil, e
	}
	if f["applied"] == "true" || oneOf(f["result_state"], "applied", "deleted", "rejected") {
		return lines, nil
	}
	transfer := f["transfer_id"]
	if !broadcastUUID.MatchString(transfer) {
		return nil, errors.New("invalid upload transfer ID")
	}
	offset, e := strconv.Atoi(f["next_offset"])
	if e != nil || offset < 0 || offset > len(body) {
		return nil, errors.New("invalid upload offset")
	}
	for offset < len(body) {
		end := min(offset+broadcastChunkBytes, len(body))
		lines, e = request("chunk", transfer, strconv.Itoa(offset), base64.StdEncoding.EncodeToString(body[offset:end]))
		if e != nil {
			return nil, e
		}
		f, e = broadcastFields(lines)
		if e != nil {
			return nil, e
		}
		if f["next_offset"] != strconv.Itoa(end) {
			return nil, errors.New("upload acknowledgement offset mismatch")
		}
		offset = end
	}
	return request("commit", expected.RequestID, transfer)
}

type broadcastShellLine struct {
	line string
	err  error
}
type broadcastShell struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	input  io.WriteCloser
	lines  chan broadcastShellLine
}

func openBroadcastShell(cfg config) (*broadcastShell, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	cmd, _, e := broadcastSSHCommand(ctx, cfg, "", true)
	if e != nil {
		cancel()
		return nil, e
	}
	in, e := cmd.StdinPipe()
	if e != nil {
		cancel()
		return nil, e
	}
	out, e := cmd.StdoutPipe()
	if e != nil {
		cancel()
		return nil, e
	}
	cmd.Stderr = io.Discard
	if e = cmd.Start(); e != nil {
		cancel()
		return nil, userMessage{"Cannot open the persistent SSH upload channel."}
	}
	s := &broadcastShell{cmd, cancel, in, make(chan broadcastShellLine)}
	go func() {
		defer close(s.lines)
		scanner := bufio.NewScanner(out)
		scanner.Buffer(make([]byte, 8192), 65536)
		for scanner.Scan() {
			select {
			case s.lines <- broadcastShellLine{line: scanner.Text()}:
			case <-ctx.Done():
				return
			}
		}
		e := scanner.Err()
		if e == nil {
			e = io.ErrUnexpectedEOF
		}
		select {
		case s.lines <- broadcastShellLine{err: e}:
		case <-ctx.Done():
		}
	}()
	return s, nil
}
func (s *broadcastShell) close() { s.cancel(); s.input.Close(); s.cmd.Wait() }
func (s *broadcastShell) request(wire string) ([]string, error) {
	if _, e := io.WriteString(s.input, wire+"\n"); e != nil {
		return nil, userMessage{"SSH control connection ended. Inspect the retained request status before retrying."}
	}
	timer := time.NewTimer(45 * time.Second)
	defer timer.Stop()
	skipped, total := 0, 0
	var frame []string
	multiple := false
	for {
		select {
		case item, ok := <-s.lines:
			if !ok || item.err != nil {
				return nil, userMessage{"SSH control connection ended. Inspect the retained request status before retrying."}
			}
			line := strings.TrimSuffix(item.line, "\r")
			for strings.HasPrefix(line, "> ") {
				line = strings.TrimPrefix(line, "> ")
			}
			if len(frame) == 0 {
				if !strings.HasPrefix(line, "OK") && !strings.HasPrefix(line, "ERR ") {
					skipped += len(line)
					if skipped > 65536 {
						return nil, errors.New("unexpected persistent SSH shell output")
					}
					continue
				}
				record, e := bibleit.ParseRecord(line)
				if e != nil {
					return nil, e
				}
				for _, key := range []string{"count", "verses", "results", "books", "commands", "connections"} {
					if _, ok := record.Fields[key]; ok {
						multiple = true
					}
				}
				frame = append(frame, line)
				total = len(line) + 1
				if strings.HasPrefix(line, "ERR ") || !multiple {
					response, e := bibleit.DecodeResponse(strings.NewReader(line + "\n"))
					return response.Lines, e
				}
			} else {
				total += len(line) + 1
				if total > 2*1048576 {
					return nil, errors.New("SSH response exceeds its bounded size")
				}
				frame = append(frame, line)
				if line == "END" {
					response, e := bibleit.DecodeResponse(strings.NewReader(strings.Join(frame, "\n") + "\n"))
					return response.Lines, e
				}
			}
		case <-timer.C:
			return nil, userMessage{"SSH control reply timed out. Inspect the retained request status; a mutation may already have completed."}
		}
	}
}
