package main

import (
	"fmt"
	"strings"
)

type helpRow struct{ syntax, description string }
type helpSection struct {
	title string
	rows  []helpRow
}
type helpNamespace struct{ syntax, description string }

// These are presentation labels. The registry remains the source of command
// grammar, descriptions, aliases and completion candidates.
var helpNamespaces = map[string]helpNamespace{
	"server":                   {"server", "Inspect the service."},
	"auth":                     {"auth", "Manage authentication."},
	"account":                  {"account", "Inspect your account."},
	"organization":             {"organization", "Inspect organizations and applications."},
	"organization application": {"organization application <id>", "Inspect an organization application."},
	"translation":              {"translation", "Manage your translation library."},
	"live":                     {"live", "Create and manage Live sessions."},
	"live set":                 {"live <id> set", "Update Live settings."},
	"live stack":               {"live <id> stack", "Manage retained Scripture."},
	"live secret":              {"live <id> secret", "Manage audience protection."},
	"config":                   {"config", "Manage saved CLI settings."},
}

func helpRootSection(command string) string {
	switch command {
	case "auth", "account", "organization":
		return "ACCOUNT"
	case "translation", "read", "search":
		return "SCRIPTURE"
	case "live":
		return "LIVE"
	default:
		return "COMMANDS"
	}
}

func helpHasPrefix(path, prefix []string) bool {
	return len(path) >= len(prefix) && strings.Join(path[:len(prefix)], " ") == strings.Join(prefix, " ")
}

func canonicalHelpTopic(topic []string) ([]string, *commandSpec) {
	topic = append([]string(nil), topic...)
	selected := commandSpecFor(topic)
	if selected != nil {
		topic = helpPath(strings.Fields(selected.Syntax))
	}
	// Help accepts either `live stack` or `live <actual-id> stack`.
	if len(topic) > 1 && topic[0] == "live" && !oneOf(topic[1], "list", "create", "delete", "stack", "set", "secret", "info", "stats", "start", "stop", "pause", "resume", "subscribe", "clear") {
		topic = append([]string{"live"}, topic[2:]...)
	}
	if len(topic) > 2 && topic[0] == "organization" && topic[1] == "application" && topic[2] != "status" {
		topic = append([]string{"organization", "application"}, topic[3:]...)
	}
	return topic, selected
}

func commandHelp(topic []string, width int) (string, error) {
	topic, selected := canonicalHelpTopic(topic)
	matches := []commandSpec{}
	for _, spec := range commandRegistry {
		if helpHasPrefix(helpPath(strings.Fields(spec.Syntax)), topic) {
			matches = append(matches, spec)
		}
	}
	if len(matches) == 0 {
		return "", fmt.Errorf("unknown help topic; run `bibleit help` to list commands")
	}
	if selected == nil && len(matches) == 1 && len(helpPath(strings.Fields(matches[0].Syntax))) == len(topic) {
		selected = &matches[0]
	}
	if selected != nil {
		lines := append(wrapHelpText("bibleit "+selected.Syntax, width), "")
		lines = append(lines, wrapHelpText(selected.Summary, width)...)
		return strings.Join(lines, "\n"), nil
	}
	sections := []helpSection{}
	if len(topic) == 0 {
		for _, title := range []string{"COMMANDS", "ACCOUNT", "SCRIPTURE", "LIVE"} {
			sections = append(sections, helpSection{title: title})
		}
	} else {
		sections = append(sections, helpSection{title: "COMMANDS"})
		if strings.Join(topic, " ") == "live" {
			sections = append(sections, helpSection{title: "SESSION"})
		}
	}
	seen := map[string]bool{}
	for _, spec := range matches {
		if spec.Alias != "" {
			continue
		}
		path := helpPath(strings.Fields(spec.Syntax))
		n := min(len(path), len(topic)+1)
		liveSession := strings.Join(topic, " ") == "live" && strings.HasPrefix(spec.Syntax, "live <id> ")
		if strings.Join(topic, " ") == "live" && !liveSession {
			n = len(path)
		}
		key := strings.Join(path[:n], " ")
		seenKey := key
		if liveSession {
			seenKey = "session " + key
		}
		if seen[seenKey] {
			continue
		}
		seen[seenKey] = true
		row := helpRow{syntax: spec.Syntax, description: spec.Summary}
		if len(path) > n {
			if namespace, ok := helpNamespaces[key]; ok {
				row = helpRow{namespace.syntax, namespace.description}
			} else {
				row = helpRow{strings.Join(path[:n], " "), "Show available commands."}
			}
		}
		title := "COMMANDS"
		if len(topic) == 0 {
			title = helpRootSection(path[0])
		} else if strings.Join(topic, " ") == "live" && strings.HasPrefix(spec.Syntax, "live <id> ") {
			title = "SESSION"
		}
		for i := range sections {
			if sections[i].title == title {
				sections[i].rows = append(sections[i].rows, row)
				break
			}
		}
	}
	lines := wrapHelpText("Bibleit CLI: read Scripture and manage Live sessions.", width)
	if len(topic) > 0 {
		namespace := helpNamespaces[strings.Join(topic, " ")]
		syntax := namespace.syntax
		if syntax == "" {
			syntax = strings.Join(topic, " ")
		}
		lines = wrapHelpText("bibleit "+syntax+" <command>", width)
		if namespace.description != "" {
			lines = append(lines, "")
			lines = append(lines, wrapHelpText(namespace.description, width)...)
		}
	}
	maxSyntax := 0
	for _, section := range sections {
		for _, row := range section.rows {
			maxSyntax = max(maxSyntax, cellWidth("bibleit "+row.syntax))
		}
	}
	for _, section := range sections {
		if len(section.rows) == 0 {
			continue
		}
		lines = append(lines, "", section.title)
		lines = append(lines, renderHelpRows(section.rows, width, maxSyntax, "bibleit ")...)
	}
	if len(topic) == 0 {
		lines = append(lines, "", "FLAGS")
		flags := []helpRow{{"-h, --help", "Show help."}, {"-v, --version", "Show the CLI version."}}
		lines = append(lines, renderHelpRows(flags, width, cellWidth("-v, --version"), "")...)
	}
	lines = append(lines, "")
	lines = append(lines, wrapHelpText("Run `bibleit help <command>` for command usage.", width)...)
	return strings.Join(lines, "\n"), nil
}

func renderHelpRows(rows []helpRow, width, syntaxWidth int, prefix string) []string {
	lines := []string{}
	for _, row := range rows {
		syntax := prefix + row.syntax
		if width-2-syntaxWidth-2 < 24 {
			for _, line := range wrapHelpText(syntax, width-2) {
				lines = append(lines, "  "+line)
			}
			for _, line := range wrapHelpText(row.description, width-4) {
				lines = append(lines, "    "+line)
			}
			continue
		}
		description := wrapHelpText(row.description, width-2-syntaxWidth-2)
		for i, line := range description {
			left := ""
			if i == 0 {
				left = syntax
			}
			lines = append(lines, "  "+left+strings.Repeat(" ", syntaxWidth-cellWidth(left)+2)+line)
		}
	}
	return lines
}

func wrapHelpText(text string, width int) []string {
	width = max(1, width)
	lines := []string{}
	line := ""
	for _, word := range strings.Fields(text) {
		if line != "" && cellWidth(line)+1+cellWidth(word) > width {
			lines = append(lines, line)
			line = ""
		}
		if line != "" {
			line += " "
		}
		// Long words are unusual in descriptions; split them without overflowing a
		// narrow terminal or cutting a multi-column display cell.
		parts := strings.Split(displayCells.Wrap(word, width), "\n")
		for i, part := range parts {
			if i > 0 {
				lines = append(lines, line)
				line = ""
			}
			line += part
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}
