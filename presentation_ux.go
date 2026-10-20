package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

var activeCommand []string
var activeTransport string
var displayCells = runewidth.NewCondition()

func cellWidth(s string) int { return displayCells.StringWidth(s) }
func outputWidth() int {
	if terminalInput(os.Stdout) {
		if n, e := strconv.Atoi(os.Getenv("COLUMNS")); e == nil && n >= 32 && n <= 500 {
			return n
		}
	}
	return 100
}
func verseRows(records []outputRecord) []outputRecord {
	out := []outputRecord{}
	translation := ""
	for _, r := range records {
		if r.Fields["translation"] != "" {
			translation = r.Fields["translation"]
		}
		text, has := r.Fields["text"]
		if plain, ok := r.Fields["content"]; ok {
			text = plain
			has = true
		}
		if plain, ok := r.Fields["verse_text"]; ok {
			text = plain
			has = true
		}
		if !has {
			continue
		}
		if r.Type != "verse" && r.Type != "stack" && r.Fields["translation"] == "" && len(activeCommand) == 0 {
			continue
		}
		f := map[string]string{"translation": translation, "text": text, "position": r.Fields["position"], "reference": r.Fields["reference"]}
		if f["reference"] == "" {
			book := r.Fields["book_name"]
			if book == "" {
				book = r.Fields["book"]
			}
			if _, e := strconv.Atoi(book); e == nil {
				if len(activeCommand) >= 3 && activeCommand[0] == "read" {
					book = activeCommand[2]
				} else {
					book = "Book " + book
				}
			}
			if book != "" {
				f["reference"] = book
				if chapter := r.Fields["chapter"]; chapter != "" {
					f["reference"] += " " + chapter
					if v := r.Fields["verse"]; v != "" {
						f["reference"] += ":" + v
					}
				}
			}
		}
		if f["reference"] == "" {
			f["reference"] = "—"
		}
		// Remove only an exact structured reference prefix, never infer coordinates
		// or strip arbitrary words from Scripture text.
		_, plain := r.Fields["content"]
		_, alternatePlain := r.Fields["verse_text"]
		if f["reference"] != "—" && !plain && !alternatePlain {
			f["text"] = strings.TrimPrefix(text, f["reference"]+" ")
		}
		out = append(out, outputRecord{Type: "verse", Fields: f})
	}
	return out
}
func printVerseRows(rows []outputRecord) {
	headers := []string{"translation", "reference", "text"}
	for _, r := range rows {
		if r.Fields["position"] != "" {
			headers = append([]string{"position"}, headers...)
			break
		}
	}
	widths := make([]int, len(headers))
	for i, k := range headers {
		widths[i] = cellWidth(strings.ToUpper(k))
		for _, r := range rows {
			if k != "text" {
				widths[i] = max(widths[i], min(32, cellWidth(humanValue(r.Fields[k]))))
			}
		}
	}
	occupied := 0
	for i := 0; i < len(widths)-1; i++ {
		occupied += widths[i] + 4
	}
	widths[len(widths)-1] = max(16, outputWidth()-occupied)
	printTableRow(headers, widths, true)
	for _, r := range rows {
		cells := make([][]string, len(headers))
		height := 1
		for i, k := range headers {
			value := r.Fields[k]
			lines := []string{humanValue(value)}
			if k == "text" {
				lines = strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
			}
			for _, line := range lines {
				if k == "text" {
					line = humanValue(line)
				}
				wrapped := displayCells.Wrap(line, widths[i])
				cells[i] = append(cells[i], strings.Split(wrapped, "\n")...)
			}
			height = max(height, len(cells[i]))
		}
		for n := 0; n < height; n++ {
			v := make([]string, len(headers))
			for i := range v {
				if n < len(cells[i]) {
					v[i] = cells[i][n]
				}
			}
			printTableRow(v, widths, false)
		}
	}
}
func emptyMessage() string {
	a := activeCommand
	if len(a) > 0 {
		switch a[0] {
		case "search":
			return "No verses matched."
		case "translation":
			return "No translations found. Use `bibleit translation add <translation>` to enable one."
		case "organization":
			return "No organizations or applications found."
		case "account":
			if len(a) > 1 && a[1] == "tokens" {
				return "No active token metadata found."
			}
		case "live":
			if len(a) > 2 && a[2] == "stack" {
				return "The Live stack is empty."
			}
			return "No Live sessions found. Create one with `bibleit live create [name]`."
		}
	}
	return "No results found."
}
func humanReceipt(fields map[string]string) string {
	event := fields["event"]
	switch event {
	case "deleted":
		return "Deleted the requested Live session."
	case "clear":
		return "Cleared the retained stack."
	case "logged_out":
		return "Signed out; saved web credentials and cached SSH identity cleared."
	}
	return ""
}
func presentHuman(records []outputRecord) bool {
	if receipt := commandReceipt(records); receipt != "" {
		fmt.Println(receipt)
		return true
	}
	if rows := verseRows(records); len(rows) > 0 {
		printVerseRows(rows)
		return true
	}
	if len(records) == 0 {
		fmt.Println(emptyMessage())
		return true
	}
	if len(records) == 1 {
		r := records[0]
		if msg := humanReceipt(r.Fields); msg != "" {
			fmt.Println(msg)
			return true
		}
		for _, k := range []string{"count", "results", "verses"} {
			if r.Fields[k] == "0" {
				fmt.Println(emptyMessage())
				return true
			}
		}
	}
	return false
}
func presentHTTPHuman(data map[string]any, kind string) bool {
	f := map[string]string{}
	for _, k := range []string{"count", "popped", "deleted", "affected"} {
		if v, ok := data[k]; ok {
			f[k] = fmt.Sprint(v)
		}
	}
	if receipt := commandReceipt([]outputRecord{{Fields: f}}); receipt != "" {
		fmt.Println(receipt)
		return true
	}
	records := []outputRecord{}
	var walk func(map[string]any, string)
	walk = func(m map[string]any, translation string) {
		if t, ok := m["translation"].(string); ok {
			translation = t
		}
		f := map[string]string{}
		for _, k := range []string{"translation", "reference", "book", "book_name", "chapter", "verse", "text", "position"} {
			if v, ok := m[k]; ok && v != nil {
				switch v.(type) {
				case map[string]any, []any:
				default:
					f[k] = fmt.Sprint(v)
				}
			}
		}
		if f["translation"] == "" {
			f["translation"] = translation
		}
		if text, ok := f["text"]; ok {
			f["content"] = text
			records = append(records, outputRecord{Type: "verse", Fields: f})
		}
		for _, k := range []string{"verses", "translations", "data", "entries"} {
			if arr, ok := m[k].([]any); ok {
				for _, v := range arr {
					if r, ok := v.(map[string]any); ok {
						walk(r, translation)
					}
				}
			}
		}
	}
	walk(data, "")
	if len(records) > 0 {
		printVerseRows(verseRows(records))
		printRights(data)
		return true
	}
	if arr, ok := data["data"].([]any); ok && len(arr) == 0 {
		fmt.Println(emptyMessage())
		return true
	}
	if deleted, ok := data["deleted"].(bool); ok && deleted {
		fmt.Println("Deleted the requested resource.")
		return true
	}
	return false
}
func printRights(data map[string]any) {
	if rights, ok := data["rights"].(map[string]any); ok {
		if text, ok := rights["attribution"].(string); ok && text != "" {
			fmt.Println("\n" + humanValue(text))
		}
	}
}

func commandReceipt(records []outputRecord) string {
	a := activeCommand
	if len(a) < 3 || a[0] != "live" {
		return ""
	}
	if a[1] == "delete" && a[2] == "all" {
		for _, r := range records {
			if n, e := strconv.ParseUint(r.Fields["deleted"], 10, 64); e == nil {
				return fmt.Sprintf("Deleted %d Live sessions in the authorized scope.", n)
			}
		}
		return "Deleted Live sessions in the authorized scope."
	}
	id := humanValue(a[1])
	if a[1] == "delete" {
		return ""
	}
	target := fmt.Sprintf("Live session %q", id)
	switch a[2] {
	case "start":
		return "Started " + target + "."
	case "stop":
		return "Stopped " + target + "."
	case "pause":
		return "Paused " + target + "."
	case "resume":
		return "Resumed " + target + "."
	case "delete":
		return "Deleted " + target + "."
	case "set":
		return "Updated " + target + "."
	case "clear":
		return "Cleared the retained stack for " + target + "."
	}
	if len(a) >= 4 && a[2] == "stack" {
		if a[3] == "clear" {
			return "Cleared the retained stack for " + target + "."
		}
		if a[3] == "pop" {
			for _, r := range records {
				for _, k := range []string{"affected", "popped", "removed"} {
					if n, e := strconv.ParseUint(r.Fields[k], 10, 64); e == nil {
						return fmt.Sprintf("Popped %d stack entries from %s.", n, target)
					}
				}
			}
			// Old servers do not report actual counts: never substitute the requested count.
			return "Updated the retained stack for " + target + "."
		}
	}
	return ""
}

func humanRecords(records []outputRecord) []outputRecord {
	out := make([]outputRecord, 0, len(records))
	for _, r := range records {
		f := map[string]string{}
		for k, v := range r.Fields {
			f[k] = v
		}
		if len(activeCommand) == 2 && activeCommand[0] == "server" && activeCommand[1] == "info" {
			f["transport"] = activeTransport
			f["api_endpoint"] = "https://api.bibleit.app/v1"
			f["browser_origin"] = defaultWebURL
			f["ssh_server"] = defaultSSHServer
		}
		for old, new := range map[string]string{"actor": "user", "actor_connections": "user_connections"} {
			if v, ok := f[old]; ok {
				if _, exists := f[new]; !exists {
					f[new] = v
				}
				delete(f, old)
			}
		}
		if r.Type == "actor" {
			r.Type = "user"
		}
		r.Fields = f
		r.order = orderedHeaders([]outputRecord{r})
		out = append(out, r)
	}
	return out
}
