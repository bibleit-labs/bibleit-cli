package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestVerseRenderingSharesReferenceAndPreservesNativeJSON(t *testing.T) {
	old := activeCommand
	activeCommand = []string{"read", "bpml", "Salmos", "23", "1"}
	defer func() { activeCommand = old }()
	fields := map[string]string{"translation": "bpml", "reference": "Salmos 23:1", "text": "Salmos 23:1 O SENHOR é meu pastor.\nLinha dois 界 é"}
	_, out, _ := captureCLI(t, func() int {
		return printOutputRecords([]outputRecord{{Type: "ok", Fields: fields, order: []string{"translation", "reference", "text"}}}, "table")
	})
	if strings.Count(out, "Salmos 23:1") != 1 || !strings.Contains(out, "Linha dois") || strings.Contains(out, `\n`) {
		t.Fatal(out)
	}
	native := httpResult{data: map[string]any{"translation": "bpml", "verses": []any{map[string]any{"book": "Salmos", "chapter": json.Number("23"), "verse": json.Number("1"), "text": "Texto\nsegunda linha"}}}}
	_, out, _ = captureCLI(t, func() int { return printHTTP(native, "verse", "json") })
	var v map[string]any
	if json.Unmarshal([]byte(out), &v) != nil || v["schema_version"] != float64(2) || strings.Contains(out, "records") {
		t.Fatal(out)
	}
	_, out, _ = captureCLI(t, func() int { return printHTTP(native, "verse", "table") })
	if !strings.Contains(out, "Salmos 23:1") || strings.Contains(out, "book_id") {
		t.Fatal(out)
	}
	if cellWidth("界") != 2 || cellWidth("é") != 1 {
		t.Fatal("display-cell widths incorrect")
	}
}
func TestReceiptsAndEmptyStatesDoNotEchoQueriesOrInventCounts(t *testing.T) {
	old := activeCommand
	defer func() { activeCommand = old }()
	activeCommand = []string{"search", "web", "RELIGIOUS_CANARY"}
	_, out, _ := captureCLI(t, func() int { return printResponse([]string{"OK results=0", "END"}, "table") })
	if !strings.Contains(out, "No verses matched") || strings.Contains(out, "CANARY") {
		t.Fatal(out)
	}
	activeCommand = []string{"live", "L1", "stack", "pop", "20"}
	_, out, _ = captureCLI(t, func() int { return printResponse([]string{"OK event=clear"}, "table") })
	if strings.Contains(out, "20") || out == "" {
		t.Fatal(out)
	}
	activeCommand = []string{"live", "L1", "delete"}
	_, out, _ = captureCLI(t, func() int { return printResponse([]string{"OK event=deleted"}, "table") })
	if !strings.Contains(out, "Deleted") || !strings.Contains(out, "L1") {
		t.Fatal(out)
	}
}
