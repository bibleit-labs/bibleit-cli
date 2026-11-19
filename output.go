package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	bibleit "github.com/bibleit-labs/bibleit-go"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var activeFormat string

var (
	presentationName  = regexp.MustCompile(`<<"name">>\s*=>\s*<<"((?:\\.|[^"\\])*)">>`)
	presentationIndex = regexp.MustCompile(`<<"index">>\s*=>\s*(\d+)`)
	presentationCount = regexp.MustCompile(`<<"count">>\s*=>\s*(\d+)`)
)

type stableRecord struct {
	Type    string          `json:"type"`
	Event   string          `json:"event,omitempty"`
	Raw     string          `json:"raw,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
	Fields  map[string]any  `json:"fields"`
}

func typedRecord(record outputRecord) (stableRecord, error) {
	result := stableRecord{Type: record.Type, Fields: map[string]any{}}
	for key, value := range record.Fields {
		switch {
		case oneOf(key, "paused", "owned", "protected", "retiring", "pong", "applied", "replay", "terminal", "eof") || key == "auth" && record.Type == "ok" || key == "auth" && record.Type == "auth":
			if value != "true" && value != "false" {
				return result, fmt.Errorf("invalid boolean output field %s", key)
			}
			result.Fields[key] = value == "true"
		case oneOf(key, "protocol_version", "count", "verses", "chapters", "results", "books", "commands", "connections", "actor_connections", "anonymous_connections", "running_for_seconds", "revision", "stack_entries", "position", "issued_at", "created_at", "last_used_at", "expires_at", "connected_at", "connected_for_seconds", "plan_activated_at", "lives", "tokens", "keys", "used", "book", "chapter", "verse", "document_bytes", "chunk_bytes", "offset", "bytes", "next_offset", "received_bytes", "payload_bytes", "deleted_revision", "idle_expires_at", "hard_expires_at") || key == "translations" && (record.Fields["event"] == "verse" || record.Fields["lives"] != "") || key == "limit" && value != "unlimited":
			n, err := strconv.ParseInt(value, 10, 64)
			if err != nil || n < 0 {
				return result, fmt.Errorf("invalid numeric output field %s", key)
			}
			result.Fields[key] = n
		case oneOf(key, "roles", "permissions", "capabilities", "scopes", "translations"):
			values := []string{}
			if value != "" {
				values = strings.Split(value, ",")
			}
			result.Fields[key] = values
		default:
			result.Fields[key] = value
		}
	}
	if record.Fields["document_format"] == "broadcast-lifecycle-json-v1" && record.Fields["data_base64"] != "" {
		payload, err := base64.StdEncoding.Strict().DecodeString(record.Fields["data_base64"])
		if err != nil || len(payload) > broadcastMaxDocument || !json.Valid(payload) {
			return result, errors.New("invalid Broadcast Studio JSON payload")
		}
		result.Payload = json.RawMessage(payload)
		delete(result.Fields, "data_base64")
	}
	return result, nil
}
func printResponse(lines []string, format string) int {
	if activeSecretFile != nil {
		safe := append([]string(nil), lines...)
		for i, line := range safe {
			if line == "END" {
				continue
			}
			r, e := bibleit.ParseRecord(line)
			if e != nil {
				return protocolFailure(e)
			}
			if secret := r.Fields["secret"]; secret != "" {
				replacement, e := deliverSecret(secret)
				if e != nil {
					return protocolFailure(e)
				}
				safe[i] = strings.ReplaceAll(line, secret, replacement)
			}
		}
		if !activeSecretSaved {
			return outputError(exitFailure, userMessage{"The server did not supply a one-time secret; inspect the Live session before retrying."})
		}
		lines = safe
	}
	if format == "raw" {
		fmt.Println(strings.Join(lines, "\n"))
		return exitOK
	}
	records := []outputRecord{}
	for _, line := range lines {
		if line == "END" {
			continue
		}
		record, err := bibleit.ParseRecord(line)
		if err != nil {
			return protocolFailure(err)
		}
		records = append(records, outputRecord{Type: record.Type, Fields: record.Fields, order: record.Order})
	}
	if format == "ndjson" {
		for _, line := range lines {
			if line == "END" {
				continue
			}
			if err := writeSubscriptionLine(line, format); err != nil {
				return protocolFailure(err)
			}
		}
		return exitOK
	}
	return printOutputRecords(records, format)
}
func printOutputRecords(records []outputRecord, format string) int {
	if format != "json" {
		human, err := broadcastHumanRecords(records)
		if err != nil {
			return protocolFailure(err)
		}
		printTable(human)
		return exitOK
	}
	values := []stableRecord{}
	for _, record := range records {
		value, err := typedRecord(record)
		if err != nil {
			return protocolFailure(err)
		}
		values = append(values, value)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(map[string]any{"schema_version": 1, "ok": true, "records": values}); err != nil {
		return protocolFailure(err)
	}
	return exitOK
}

// Lifecycle JSON stays native in machine output; the default table displays
// its public value instead of exposing the transport's base64 envelope.
func broadcastHumanRecords(records []outputRecord) ([]outputRecord, error) {
	out := make([]outputRecord, 0, len(records))
	for _, record := range records {
		if record.Fields["document_format"] != "broadcast-lifecycle-json-v1" {
			out = append(out, record)
			continue
		}
		typed, err := typedRecord(record)
		if err != nil || len(typed.Payload) == 0 {
			return nil, errors.New("invalid Broadcast Studio JSON payload")
		}
		var envelope map[string]json.RawMessage
		if json.Unmarshal(typed.Payload, &envelope) != nil || len(envelope) != 1 || envelope["value"] == nil {
			return nil, errors.New("invalid Broadcast Studio value envelope")
		}
		var value map[string]any
		decoder := json.NewDecoder(bytes.NewReader(envelope["value"]))
		decoder.UseNumber()
		if decoder.Decode(&value) != nil {
			return nil, errors.New("invalid Broadcast Studio value")
		}
		fields := make(map[string]string, len(value))
		if value == nil {
			fields["status"] = "No result"
			if record.Fields["record_type"] == "channel" {
				fields["status"] = "No connected channel"
			}
		}
		for key, item := range value {
			switch v := item.(type) {
			case nil:
				fields[key] = "-"
			case string:
				fields[key] = v
			case map[string]any, []any:
				body, err := json.Marshal(v)
				if err != nil {
					return nil, errors.New("invalid Broadcast Studio value")
				}
				fields[key] = string(body)
			default:
				fields[key] = fmt.Sprint(v)
			}
		}
		out = append(out, outputRecord{Type: record.Type, Fields: fields})
	}
	return out, nil
}

func printLocal(kind string, fields map[string]string, format string) int {
	order := []string{}
	for key := range fields {
		order = append(order, key)
	}
	sort.Strings(order)
	return printOutputRecords([]outputRecord{{Type: kind, Fields: fields, order: order}}, format)
}
func parseOptions(args []string, allowed map[string]bool) (map[string]string, []string, error) {
	values := map[string]string{}
	rest := []string{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			rest = append(rest, args[i+1:]...)
			break
		}
		name, value, inline := strings.Cut(args[i], "=")
		takesValue, known := allowed[name]
		if !known {
			if strings.HasPrefix(name, "-") && !oneOf(name, "-h", "--help", "-v", "--version") {
				if _, err := strconv.Atoi(name); err != nil {
					return values, rest, fmt.Errorf("unknown option; run `bibleit help` for flags, or use -- before a literal value")
				}
			}
			rest = append(rest, args[i])
			continue
		}
		if _, exists := values[name]; exists {
			return values, rest, fmt.Errorf("duplicate option %s", name)
		}
		if !takesValue {
			if inline {
				return values, rest, fmt.Errorf("option %s does not take a value", name)
			}
			values[name] = "true"
			continue
		}
		if !inline {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				return values, rest, fmt.Errorf("option %s requires a value", name)
			}
			i++
			value = args[i]
		}
		if value == "" {
			return values, rest, fmt.Errorf("option %s requires a value", name)
		}
		values[name] = value
	}
	return values, rest, nil
}
func outputError(exit int, err error) int {
	code, message, fields, status := describeError(exit, err)
	if activeFormat != "json" && activeFormat != "ndjson" {
		fmt.Fprintln(os.Stderr, "bibleit: "+message)
		return exit
	}
	detail := map[string]any{"code": code, "message": message, "fields": fields}
	if status != 0 {
		detail["http_status"] = status
	}
	_ = json.NewEncoder(os.Stderr).Encode(map[string]any{"schema_version": 1, "ok": false, "error": detail, "exit_code": exit})
	return exit
}

func humanValue(value string) string {
	var out strings.Builder
	for _, ch := range value {
		switch ch {
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		default:
			if ch < 32 || ch == 127 {
				fmt.Fprintf(&out, `\u%04x`, ch)
			} else {
				out.WriteRune(ch)
			}
		}
	}
	return out.String()
}

// The protocol exposes the active slide as an opaque Erlang term. Keep that
// complete value in raw/JSON output, but never print slide elements in a table.
func tableValue(record outputRecord, key string) string {
	value := record.Fields[key]
	if key != "presentation" {
		return humanValue(value)
	}
	if value == "" || value == "null" {
		return "—"
	}
	name := "Presentation"
	if match := presentationName.FindStringSubmatch(value); len(match) == 2 {
		if decoded, err := strconv.Unquote(`"` + match[1] + `"`); err == nil && decoded != "" {
			name = decoded
		}
	}
	if runes := []rune(name); len(runes) > 48 {
		name = string(runes[:47]) + "…"
	}
	index, count := presentationIndex.FindStringSubmatch(value), presentationCount.FindStringSubmatch(value)
	if len(index) == 2 && len(count) == 2 {
		position, positionErr := strconv.Atoi(index[1])
		total, totalErr := strconv.Atoi(count[1])
		if positionErr == nil && totalErr == nil && position >= 0 && position < total {
			return humanValue(fmt.Sprintf("%s (%d/%d)", name, position+1, total))
		}
	}
	return humanValue(name)
}
func writeSubscriptionLine(line, format string) error {
	if format != "ndjson" {
		value := line
		if format != "raw" {
			value = humanValue(line)
		}
		_, err := fmt.Fprintln(os.Stdout, value)
		return err
	}
	record, err := subscriptionRecord(line)
	if err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{"schema_version": 1, "record": record})
}
func subscriptionRecord(line string) (stableRecord, error) {
	parsed, err := bibleit.ParseRecord(line)
	if err != nil {
		return stableRecord{}, err
	}
	parts := strings.Fields(line)
	if parsed.Type == "event" && len(parts) > 1 && parts[1] == "verse" {
		if len(parts) != 3 {
			return stableRecord{}, errors.New("invalid verse event payload")
		}
		payload, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil || !json.Valid(payload) {
			return stableRecord{}, errors.New("invalid verse event payload")
		}
		return stableRecord{Type: "event", Event: "verse", Fields: map[string]any{}, Payload: json.RawMessage(payload), Raw: line}, nil
	}
	record, err := typedRecord(outputRecord{Type: parsed.Type, Fields: parsed.Fields})
	if err != nil {
		return record, err
	}
	record.Raw = line
	if parsed.Type == "event" {
		if len(parts) > 1 && !strings.Contains(parts[1], "=") {
			record.Event = parts[1]
		} else if _, ok := parsed.Fields["live"]; ok {
			record.Event = "live"
		}
	}
	return record, nil
}

type subscriptionTableDisplay struct {
	lastState string
	paused    bool
}

func (display *subscriptionTableDisplay) line(line string, initial bool) (string, error) {
	record, err := subscriptionRecord(line)
	if err != nil {
		return "", err
	}
	if initial || record.Event == "live" {
		fields := record.Fields
		field := func(key string) string {
			value, _ := fields[key].(string)
			return value
		}
		name, id, status, mode := field("name"), field("id"), field("status"), field("mode")
		if id == "" {
			id = field("live")
		}
		if name == "" {
			name = id
		}
		if status == "" {
			status = "running"
		}
		if mode == "" {
			mode = "reader"
		}
		display.paused, _ = fields["paused"].(bool)
		state := status + " · " + mode
		if display.paused {
			state += " · paused"
		}
		if mode == "presentation" {
			presentation := tableValue(outputRecord{Fields: map[string]string{"presentation": field("presentation")}}, "presentation")
			if presentation != "—" {
				state += " · " + presentation
			}
		}
		if translations, ok := fields["translations"].([]string); ok && len(translations) > 0 {
			state += " · translations: " + strings.Join(translations, ", ")
		}
		if reference := field("reference"); reference != "" {
			state += " · reference: " + reference
		}
		signature := name + "|" + id + "|" + state
		if !initial && signature == display.lastState {
			return "", nil
		}
		display.lastState = signature
		if id != "" && name != id {
			return fmt.Sprintf("Live %s (%s): %s", humanValue(name), humanValue(id), humanValue(state)), nil
		}
		return fmt.Sprintf("Live %s: %s", humanValue(name), humanValue(state)), nil
	}
	switch record.Event {
	case "verse":
		return readableVersePayload(record.Payload), nil
	case "paused":
		if display.paused {
			return "", nil
		}
		display.paused = true
		return "Live paused", nil
	case "clear":
		return "Live cleared", nil
	case "closed":
		return "Live closed", nil
	case "revoked":
		return "Live access revoked", nil
	default:
		return "Live updated", nil
	}
}

func readableVersePayload(raw json.RawMessage) string {
	var payload any
	if json.Unmarshal(raw, &payload) != nil {
		return "Verse update unavailable"
	}
	var verses []any
	switch value := payload.(type) {
	case []any:
		verses = value
	case map[string]any:
		if translations, ok := value["translations"].([]any); ok && len(translations) > 0 {
			verses = translations
		} else {
			verses = []any{value}
		}
	}
	var output []string
	for _, value := range verses {
		verse, ok := value.(map[string]any)
		if !ok {
			continue
		}
		stringField := func(key string) string {
			text, _ := verse[key].(string)
			return text
		}
		reference, translation, passage := stringField("reference"), stringField("translation"), stringField("text")
		if reference == "" {
			reference = "Verse"
		}
		header := humanValue(reference)
		if translation != "" {
			header += " · " + humanValue(translation)
		}
		if passage != "" {
			output = append(output, header+"\n"+humanValue(passage))
		} else {
			output = append(output, header)
		}
	}
	if len(output) == 0 {
		return "Verse update unavailable"
	}
	return strings.Join(output, "\n\n")
}
