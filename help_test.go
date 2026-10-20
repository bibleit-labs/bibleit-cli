package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"unicode"
)

func TestGroupedHelpNavigation(t *testing.T) {
	root, err := commandHelp(nil, 120)
	if err != nil {
		t.Fatal(err)
	}
	for _, heading := range []string{"COMMANDS", "ACCOUNT", "SCRIPTURE", "LIVE", "FLAGS"} {
		if !strings.Contains(root, "\n"+heading+"\n") {
			t.Fatalf("missing section %s", heading)
		}
	}
	for _, row := range []string{"bibleit auth", "bibleit account", "bibleit organization", "bibleit translation", "bibleit live", "bibleit config"} {
		if !strings.Contains(root, row) {
			t.Fatalf("missing namespace %s", row)
		}
	}
	for _, noise := range []string{"live <id> stack push", "auth whoami", "account quota list", "secret <value>", "HTTP (", "SSH (", "Local;"} {
		if strings.Contains(root, noise) {
			t.Fatalf("root help contains %q", noise)
		}
	}
	if strings.ContainsRune(root, '\x1b') || strings.IndexFunc(root, func(r rune) bool { return unicode.Is(unicode.So, r) }) >= 0 {
		t.Fatal("help must be plain text without emoji")
	}
	live, err := commandHelp([]string{"live"}, 120)
	if err != nil {
		t.Fatal(err)
	}
	for _, syntax := range []string{"bibleit live delete all", "bibleit live <id> delete", "bibleit live <id> set", "bibleit live <id> stack", "bibleit live <id> secret"} {
		if !strings.Contains(live, syntax) {
			t.Fatalf("Live help lost %s", syntax)
		}
	}
	stack, err := commandHelp([]string{"live", "L1", "stack"}, 120)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range []string{"info", "push", "pop", "clear"} {
		if !strings.Contains(stack, "bibleit live <id> stack "+action) {
			t.Fatalf("stack help lost %s", action)
		}
	}
	if strings.Contains(stack, "secret") {
		t.Fatal("subcommand help includes unrelated commands")
	}
}

func TestHelpFitsTerminalWidthsAndPreservesCommands(t *testing.T) {
	topics := [][]string{nil, {"live"}, {"live", "stack"}, {"organization"}, {"organization", "application", "PRIVATE_APP_ID"}, {"config"}}
	for _, spec := range commandRegistry {
		topics = append(topics, strings.Fields(spec.Example))
	}
	for _, width := range []int{32, 40, 60, 80, 100, 120} {
		for _, topic := range topics {
			text, err := commandHelp(topic, width)
			if err != nil {
				t.Fatalf("%v: %v", topic, err)
			}
			for _, line := range strings.Split(text, "\n") {
				if cellWidth(line) > width {
					t.Fatalf("width %d, topic %v: overflow %q", width, topic, line)
				}
			}
		}
	}
	// Description wrapping must respect display cells for accented and wide text.
	for _, line := range wrapHelpText("Descrição 界界界界 é texto comprido", 12) {
		if cellWidth(line) > 12 {
			t.Fatalf("Unicode overflow: %q", line)
		}
	}
}

func TestHelpJSONKeepsEnvelopeAndCanonicalTopic(t *testing.T) {
	code, out, stderr := captureCLI(t, func() int {
		return run([]string{"live", "PRIVATE_LIVE_ID", "stack", "pop", "--help", "--format", "json"})
	})
	var doc struct {
		Schema  int  `json:"schema_version"`
		OK      bool `json:"ok"`
		Records []struct {
			Fields map[string]string `json:"fields"`
		} `json:"records"`
	}
	if code != exitOK || stderr != "" || json.Unmarshal([]byte(out), &doc) != nil || doc.Schema != 1 || !doc.OK {
		t.Fatal("invalid JSON help envelope")
	}
	fields := doc.Records[0].Fields
	if fields["topic"] != "live stack pop" || strings.Contains(out, "PRIVATE_LIVE_ID") || !strings.Contains(fields["text"], "stack pop [count]") {
		t.Fatal("help leaked the target or lost canonical usage")
	}
}

func TestRemovedSecretSyntaxIsNotAdvertised(t *testing.T) {
	for _, topic := range [][]string{nil, {"live"}, {"live", "secret"}} {
		text, err := commandHelp(topic, 120)
		if err != nil || strings.Contains(text, "secret <value>") {
			t.Fatal("removed syntax is advertised")
		}
	}
	if _, err := normalizeCommand([]string{"live", "L1", "secret", "PRIVATE_SECRET_CANARY"}); err == nil || !strings.Contains(err.Error(), "does not persist") || strings.Contains(err.Error(), "PRIVATE_SECRET_CANARY") {
		t.Fatal("removed syntax lost descriptive rejection")
	}
}

func TestInvalidConfigValuesAreUsageErrorsAndDoNotChangeFile(t *testing.T) {
	isolatedPreferences(t)
	if err := savePreferences(preferences{Version: 2, Transport: "ssh"}); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(preferencesPath())
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"transport", "invalid"}, {"format", "invalid"}, {"format", "ndjson"}} {
		for _, format := range []string{"table", "raw", "json"} {
			argv := append([]string{"config", "set"}, args...)
			argv = append(argv, "--format", format)
			code, out, stderr := captureCLI(t, func() int { return run(argv) })
			if code != exitUsage || out != "" || stderr == "" {
				t.Fatalf("%v/%s: %d %s", args, format, code, stderr)
			}
			if format == "json" {
				var doc struct {
					Exit  int `json:"exit_code"`
					Error struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if json.Unmarshal([]byte(stderr), &doc) != nil || doc.Exit != exitUsage || doc.Error.Code != "usage" {
					t.Fatal("invalid value was not classified as usage")
				}
			}
			after, err := os.ReadFile(preferencesPath())
			if err != nil || string(after) != string(before) {
				t.Fatal("invalid edit changed configuration")
			}
		}
	}
}
