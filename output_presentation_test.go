package main

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func TestLivePresentationTableKeepsSlidePayloadOutOfView(t *testing.T) {
	presentation := `#{<<"blank">> => false,<<"count">> => 29,<<"index">> => 0,<<"name">> => <<"A vida de Jesus">>,<<"slide">> => #{<<"elements">> => [#{<<"asset">> => <<"large-private-asset">>}]}}`
	lines := []string{
		"OK count=2",
		"LIVE id=reader name=Reading mode=reader presentation=null",
		"LIVE id=deck name=Sunday mode=presentation presentation=" + presentation,
		"END",
	}
	code, table, stderr := captureCLI(t, func() int { return printResponse(lines, "table") })
	if code != exitOK || stderr != "" || !strings.Contains(table, "A vida de Jesus (1/29)") || !strings.Contains(table, "—") {
		t.Fatalf("presentation summary missing: %q (%d, %q)", table, code, stderr)
	}
	if strings.Contains(table, "large-private-asset") || strings.Contains(table, "elements") {
		t.Fatalf("slide payload leaked into table: %q", table)
	}
	code, encoded, stderr := captureCLI(t, func() int { return printResponse(lines, "json") })
	var response struct {
		Records []struct {
			Fields map[string]any `json:"fields"`
		} `json:"records"`
	}
	if code != exitOK || stderr != "" || json.Unmarshal([]byte(encoded), &response) != nil || len(response.Records) != 3 || response.Records[2].Fields["presentation"] != presentation {
		t.Fatalf("JSON presentation field changed: %q (%d, %q)", encoded, code, stderr)
	}
	code, raw, stderr := captureCLI(t, func() int { return printResponse(lines, "raw") })
	if code != exitOK || stderr != "" || !strings.Contains(raw, presentation) {
		t.Fatalf("raw presentation field changed: %q (%d, %q)", raw, code, stderr)
	}
}

func TestSubscriptionTableShowsLiveStateAndDecodedVerses(t *testing.T) {
	display := subscriptionTableDisplay{}
	initial, err := display.line(`OK id=live name="Sunday Reading" status=running mode=reader paused=false sequence=1 translations=bpml reference="Salmos 23:1" presentation=null`, true)
	if err != nil || initial != "Live Sunday Reading (live): running · reader · translations: bpml · reference: Salmos 23:1" {
		t.Fatalf("snapshot: %q %v", initial, err)
	}
	update, err := display.line(`EVENT live=live name="Sunday Reading" status=running mode=reader paused=false sequence=2 translations=bpml reference="Salmos 23:1" presentation=null`, false)
	if err != nil || update != "" {
		t.Fatalf("unchanged state should be quiet: %q %v", update, err)
	}
	translationUpdate, err := display.line(`EVENT live=live name="Sunday Reading" status=running mode=reader paused=false sequence=3 translations=bpml,web reference="Salmos 23:1" presentation=null`, false)
	if err != nil || !strings.Contains(translationUpdate, "translations: bpml, web") {
		t.Fatalf("translation change was hidden: %q %v", translationUpdate, err)
	}
	referenceUpdate, err := display.line(`EVENT live=live name="Sunday Reading" status=running mode=reader paused=false sequence=4 translations=bpml,web reference="Salmos 23:2" presentation=null`, false)
	if err != nil || !strings.Contains(referenceUpdate, "reference: Salmos 23:2") {
		t.Fatalf("reference change was hidden: %q %v", referenceUpdate, err)
	}
	payload := `{"reference":"Salmos 23:1","translation":"bpml","text":"first","translations":[{"reference":"Salmos 23:1","translation":"bpml","text":"O Senhor é meu pastor."},{"reference":"Salmos 23:1","translation":"web","text":"The Lord is my shepherd."}]}`
	verse, err := display.line("EVENT verse "+base64.StdEncoding.EncodeToString([]byte(payload)), false)
	if err != nil || !strings.Contains(verse, "Salmos 23:1 · bpml\nO Senhor é meu pastor.") || !strings.Contains(verse, "Salmos 23:1 · web\nThe Lord is my shepherd.") || strings.Contains(verse, "EVENT verse") {
		t.Fatalf("decoded verse: %q %v", verse, err)
	}
	paused, err := display.line(`EVENT live=live name="Sunday Reading" status=running mode=reader paused=true sequence=5 translations=bpml,web reference="Salmos 23:2" presentation=null`, false)
	if err != nil || !strings.Contains(paused, "paused") {
		t.Fatalf("pause state: %q %v", paused, err)
	}
	if duplicate, err := display.line("EVENT paused", false); err != nil || duplicate != "" {
		t.Fatalf("duplicate pause: %q %v", duplicate, err)
	}
	code, raw, stderr := captureCLI(t, func() int {
		return printResponse([]string{"EVENT verse " + base64.StdEncoding.EncodeToString([]byte(payload))}, "raw")
	})
	if code != exitOK || stderr != "" || raw != "EVENT verse "+base64.StdEncoding.EncodeToString([]byte(payload))+"\n" {
		t.Fatalf("raw wire changed: %q %q", raw, stderr)
	}
}

func TestLiveInfoSummarizesPresentationAndHandlesUnknownShape(t *testing.T) {
	fields := []string{"OK id=deck mode=presentation presentation=#{<<\"count\">> => 3,<<\"index\">> => 1,<<\"name\">> => <<\"Sunday\">>,<<\"slide\">> => #{<<\"elements\">> => []}}"}
	code, table, stderr := captureCLI(t, func() int { return printResponse(fields, "table") })
	if code != exitOK || stderr != "" || !strings.Contains(table, "Sunday (2/3)") || strings.Contains(table, "elements") {
		t.Fatalf("Live info was not summarized: %q (%d, %q)", table, code, stderr)
	}
	for _, value := range []string{"unexpected-value", `#{<<"slide">> => #{<<"elements">> => []}}`} {
		if result := tableValue(outputRecord{Type: "live", Fields: map[string]string{"presentation": value}}, "presentation"); result != "Presentation" {
			t.Fatalf("unknown presentation shape printed as %q", result)
		}
	}
}
