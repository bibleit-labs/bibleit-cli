package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func broadcastOutputRecord(kind, value string) outputRecord {
	return outputRecord{Type: "ok", Fields: map[string]string{"document_format": "broadcast-lifecycle-json-v1", "record_type": kind, "data_base64": base64.StdEncoding.EncodeToString([]byte(`{"value":` + value + `}`))}}
}
func TestBroadcastHumanLifecycleShowsPublicValue(t *testing.T) {
	tests := []struct {
		name   string
		record outputRecord
		want   []string
	}{
		{"preflight", broadcastOutputRecord("preflight", `{"active_publishers":0,"browser_capture_required":true,"automatic_restart":false}`), []string{"ACTIVE_PUBLISHERS", "0", "BROWSER_CAPTURE_REQUIRED", "true", "AUTOMATIC_RESTART", "false"}},
		{"channel", broadcastOutputRecord("channel", `{"title":"Mittel","state":"connected","revision":9223372036854775807,"last_error":null}`), []string{"Mittel", "connected", "9223372036854775807"}},
		{"disconnected", broadcastOutputRecord("channel", `null`), []string{"No connected channel"}},
		{"profile", outputRecord{Type: "output_profile", Fields: map[string]string{"id": "1080p30", "width": "1920", "height": "1080", "fps": "30"}}, []string{"1080p30", "1920", "1080", "30"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, out, stderr := captureCLI(t, func() int { return printOutputRecords([]outputRecord{tt.record}, "table") })
			if code != 0 || stderr != "" {
				t.Fatalf("exit=%d stderr=%s", code, stderr)
			}
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("missing %s in %s", want, out)
				}
			}
			for _, noise := range []string{"DATA_BASE64", "DOCUMENT_FORMAT", "RECORD_TYPE", tt.record.Fields["data_base64"]} {
				if noise != "" && strings.Contains(out, noise) {
					t.Fatalf("transport envelope shown: %s", out)
				}
			}
		})
	}
}
func TestBroadcastHumanProjectionPreservesMachineOutput(t *testing.T) {
	record := broadcastOutputRecord("preflight", `{"active_publishers":0}`)
	lines := []string{`OK document_format="broadcast-lifecycle-json-v1" record_type="preflight" data_base64="` + record.Fields["data_base64"] + `"`}
	_, raw, _ := captureCLI(t, func() int { return printResponse(lines, "raw") })
	if raw != strings.Join(lines, "\n")+"\n" {
		t.Fatal("raw transport changed")
	}
	_, out, _ := captureCLI(t, func() int { return printResponse(lines, "json") })
	var decoded struct {
		Schema  int `json:"schema_version"`
		Records []struct {
			Payload map[string]any `json:"payload"`
			Fields  map[string]any `json:"fields"`
		} `json:"records"`
	}
	if json.Unmarshal([]byte(out), &decoded) != nil || decoded.Schema != 1 || len(decoded.Records) != 1 {
		t.Fatal(out)
	}
	if _, ok := decoded.Records[0].Payload["value"]; !ok {
		t.Fatal("JSON value envelope changed")
	}
	if decoded.Records[0].Fields["document_format"] != "broadcast-lifecycle-json-v1" {
		t.Fatal("JSON metadata changed")
	}
	if record.Fields["data_base64"] == "" {
		t.Fatal("input mutated")
	}
}
func TestBroadcastHumanProjectionRejectsMalformedPayload(t *testing.T) {
	for _, body := range []string{"not base64", base64.StdEncoding.EncodeToString([]byte(`{"other":1}`)), base64.StdEncoding.EncodeToString([]byte(`{"value":0}`)), base64.StdEncoding.EncodeToString([]byte(`{"value":{},"extra":true}`)), base64.StdEncoding.EncodeToString([]byte(strings.Repeat(" ", broadcastMaxDocument+1)))} {
		record := outputRecord{Type: "ok", Fields: map[string]string{"document_format": "broadcast-lifecycle-json-v1", "data_base64": body}}
		code, out, stderr := captureCLI(t, func() int { return printOutputRecords([]outputRecord{record}, "table") })
		if code == 0 || out != "" || strings.Contains(stderr, body) {
			t.Fatalf("invalid payload exposed or accepted: exit=%d", code)
		}
	}
}

func TestBroadcastHumanProjectionEscapesTerminalControls(t *testing.T) {
	record := broadcastOutputRecord("channel", `{"title":"Mittel\u001b[2J\nname","state":"connected"}`)
	code, out, stderr := captureCLI(t, func() int { return printOutputRecords([]outputRecord{record}, "table") })
	if code != 0 || stderr != "" || strings.Contains(out, "\x1b") || !strings.Contains(out, `Mittel\u001b[2J\nname`) {
		t.Fatalf("unsafe table output: %q", out)
	}
}
