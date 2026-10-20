package main

import (
	"fmt"
	bibleit "github.com/bibleit-labs/bibleit-go"
	"sort"
	"strconv"
	"strings"
)

// commandRegistry is the public grammar. Resource execution remains in the
// typed REST planner and explicit SSH builders; this is not a protocol adapter.
type commandSpec struct {
	Syntax, Summary, Example string
	Local                    bool
	Alias                    string
}

var commandRegistry = []commandSpec{
	{"help [topic...]", "Show offline command help.", "help read", true, ""},
	{"--version", "Print the CLI version.", "--version", true, "version"},
	{"-v", "Print the CLI version.", "-v", true, "version"},
	{"ping", "Check service availability.", "ping", false, ""},
	{"version", "Print the CLI version.", "version", true, ""},
	{"server info", "Show service information and selected transport.", "server info", false, ""},
	{"server help [topic]", "Show server command help.", "server help live", false, ""},
	{"auth login [ssh]", "Sign in to your account.", "auth login", true, ""},
	{"auth info", "Show the authenticated user or service identity.", "auth info", false, ""},
	{"auth whoami", "Show the authenticated identity.", "auth whoami", false, "auth info"},
	{"auth logout", "Revoke saved web authentication and clear local credentials.", "auth logout", true, ""},
	{"account quota list", "List effective quotas (alias).", "account quota list", false, "account quotas"},
	{"account token list", "List credential metadata (alias).", "account token list", false, "account tokens"},
	{"account info", "Inspect the current account.", "account info", false, ""},
	{"account quotas", "List effective quotas and usage.", "account quotas", false, ""},
	{"account tokens", "List credential metadata, never token secrets.", "account tokens", false, ""},
	{"organization list [cursor]", "List personal memberships and applications, one page at a time.", "organization list", false, ""},
	{"organization application <id> status", "Inspect your organization application.", "organization application application-id status", false, ""},
	{"translation list", "List the current account's enabled translations.", "translation list", false, ""},
	{"translation info <translation>", "Inspect translation metadata and rights.", "translation info web", false, ""},
	{"translation catalog <translation>", "Inspect available Scripture references.", "translation catalog web", false, ""},
	{"translation add <translation>", "Enable a translation in your library.", "translation add web", false, ""},
	{"translation remove <translation>", "Remove a translation from your library.", "translation remove web", false, ""},
	{"read <translation> <book> [chapter] [verse]", "Read a verse, chapter, or book; chapter:verse is accepted.", "read web John 3 16", false, ""},
	{"search <translation> <query>", "Search an enabled translation; quote multiword queries.", "search web love", false, ""},
	{"live list", "List visible Live sessions.", "live list", false, ""},
	{"live create [name...]", "Create a Live session.", "live create Sunday", false, ""},
	{"live delete all", "Delete all Live sessions in the authorized scope; confirmation required.", "live delete all", false, ""},
	{"live <id> info", "Inspect a Live session.", "live L1 info", false, ""},
	{"live <id> stats", "Inspect connection statistics; owner authority required.", "live L1 stats", false, ""},
	{"live <id> start", "Start a Live session.", "live L1 start", false, ""},
	{"live <id> stop", "Stop a Live session.", "live L1 stop", false, ""},
	{"live <id> pause", "Pause playback.", "live L1 pause", false, ""},
	{"live <id> resume", "Resume playback.", "live L1 resume", false, ""},
	{"live <id> delete", "Delete a Live session; confirmation required.", "live L1 delete", false, ""},
	{"live <id> subscribe", "Watch Live session changes.", "live L1 subscribe", false, ""},
	{"live <id> clear", "Clear the retained stack; confirmation required.", "live L1 clear", false, "live <id> stack clear"},
	{"live <id> set name <value...>", "Change the name.", "live L1 set name Sunday", false, ""},
	{"live <id> set reference <value...>", "Change the reference.", "live L1 set reference John 3:16", false, ""},
	{"live <id> set translations <value...>", "Change enabled playback translations.", "live L1 set translations web", false, ""},
	{"live <id> stack info", "Show retained Scripture entries.", "live L1 stack info", false, ""},
	{"live <id> stack push [translation] <book> [chapter] [verse]", "Push Scripture; omit translation to use Live settings.", "live L1 stack push web John 3 16", false, ""},
	{"live <id> stack pop [count]", "Pop a positive count (default 1), without confirmation.", "live L1 stack pop", false, ""},
	{"live <id> stack clear", "Clear all retained entries; confirmation required.", "live L1 stack clear", false, ""},
	{"live <id> secret create", "Generate a protected audience secret.", "live L1 secret create", false, ""},
	{"live <id> secret rotate", "Replace the audience secret; confirmation required.", "live L1 secret rotate", false, ""},
	{"live <id> secret delete", "Remove audience protection; confirmation required.", "live L1 secret delete", false, ""},
	{"live <id> secret <value>", "Legacy connection-only authorization; cannot authorize a later invocation.", "live L1 secret synthetic-secret", false, ""},
	{"completion <shell>", "Generate Bash, Zsh, or Fish completion without network access.", "completion bash", true, ""},
	{"config show", "Show configuration and its file location, with credentials hidden.", "config show", true, ""},
	{"config get <key>", "Read a nonsecret preference.", "config get transport", true, ""},
	{"config set <key> <value>", "Save a validated nonsecret preference.", "config set transport ssh", true, ""},
	{"config unset <key>", "Remove a saved preference.", "config unset identity", true, ""},
}

func placeholder(s string) bool { return strings.HasPrefix(s, "<") || strings.HasPrefix(s, "[") }
func matchPattern(pattern []string, args []string) bool {
	if len(pattern) == 0 {
		return len(args) == 0
	}
	token := pattern[0]
	optional := strings.HasPrefix(token, "[")
	if optional && matchPattern(pattern[1:], args) {
		return true
	}
	if strings.Contains(token, "...") {
		if len(args) == 0 {
			return optional
		}
		for n := 1; n <= len(args); n++ {
			if matchPattern(pattern[1:], args[n:]) {
				return true
			}
		}
		return false
	}
	if len(args) == 0 {
		return false
	}
	if !placeholder(token) && token != args[0] {
		return false
	}
	return matchPattern(pattern[1:], args[1:])
}
func commandSpecFor(args []string) *commandSpec {
	for i := range commandRegistry {
		if matchPattern(strings.Fields(commandRegistry[i].Syntax), args) {
			return &commandRegistry[i]
		}
	}
	return nil
}
func normalizeCommand(args []string) ([]string, error) {
	a := append([]string(nil), args...)
	for _, value := range a {
		if strings.ContainsAny(value, "\r\n\x00") {
			return nil, fmt.Errorf("command arguments must not contain line breaks or NUL characters")
		}
	}
	if len(a) == 3 && a[0] == "account" && a[2] == "list" {
		switch a[1] {
		case "quota":
			a = []string{"account", "quotas"}
		case "token":
			a = []string{"account", "tokens"}
		}
	}
	if len(a) >= 3 && a[0] == "live" && oneOf(a[1], "stack", "set", "secret") {
		return nil, fmt.Errorf("missing Live ID before %q. Usage: bibleit live <id> %s; run `bibleit live list` to find an ID", a[1], a[1])
	}
	if len(a) >= 2 && a[0] == "auth" && a[1] == "ssh" {
		return nil, fmt.Errorf("`bibleit auth ssh` was renamed; use `bibleit auth login ssh`")
	}
	if len(a) == 5 && a[0] == "live" && a[2] == "secret" && a[3] == "set" {
		return nil, fmt.Errorf("custom Live secrets are unsupported; use secret create or secret rotate")
	}
	if len(a) == 4 && a[0] == "live" && a[2] == "secret" && !oneOf(a[3], "create", "rotate", "delete") {
		return nil, fmt.Errorf("standalone Live secret authentication does not persist between CLI invocations; use an audience workflow that authenticates and subscribes on the same connection")
	}
	if len(a) == 2 && a[0] == "config" && a[1] == "path" {
		return nil, fmt.Errorf("`bibleit config path` was removed; use `bibleit config show`")
	}
	spec := commandSpecFor(a)
	if spec == nil {
		return nil, fmt.Errorf("invalid command or arguments; run `bibleit help` or `bibleit help <command>` for usage")
	}
	if len(a) == 5 && a[0] == "live" && a[2] == "stack" && a[3] == "pop" {
		n, e := strconv.Atoi(a[4])
		if e != nil || n <= 0 || n > 2147483647 {
			return nil, fmt.Errorf("pop count must be an integer from 1 to 2147483647 (default 1). Usage: bibleit live <id> stack pop [count]")
		}
	}
	// Normalize the compact form once for both typed REST and SSH execution.
	if len(a) == 4 && a[0] == "read" && strings.Contains(a[3], ":") {
		parts := strings.Split(a[3], ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid chapter:verse; use positive chapter and verse numbers")
		}
		a = append(a[:3], parts...)
	}
	if a[0] == "read" {
		if _, err := cliReference(a[1], a[2:]); err != nil {
			return nil, fmt.Errorf("invalid Scripture reference; use a book and positive chapter/verse numbers")
		}
	}
	if len(a) == 3 && a[0] == "auth" && a[1] == "login" && a[2] != "ssh" {
		return nil, fmt.Errorf("usage: bibleit auth login [ssh]")
	}
	return a, nil
}
func helpPath(args []string) []string {
	out := []string{}
	for _, s := range args {
		if !placeholder(s) {
			out = append(out, s)
		}
	}
	return out
}
func registryHelp(topic []string) int {
	if spec := commandSpecFor(topic); spec != nil {
		topic = helpPath(strings.Fields(spec.Syntax))
	}
	// An explicit target is allowed in `live L1 stack pop --help`.
	if len(topic) > 1 && topic[0] == "live" && !oneOf(topic[1], "list", "create", "delete", "stack", "set", "secret", "info", "stats", "start", "stop", "pause", "resume", "subscribe", "clear") {
		topic = append([]string{"live"}, topic[2:]...)
	}
	lines := []string{"Bibleit CLI: read Scripture and manage Live sessions.", "", "Usage:"}
	found := false
	for _, spec := range commandRegistry {
		path := helpPath(strings.Fields(spec.Syntax))
		if len(topic) > len(path) || strings.Join(path[:len(topic)], " ") != strings.Join(topic, " ") {
			continue
		}
		found = true
		lines = append(lines, "  bibleit "+spec.Syntax, "    "+spec.Summary)
	}
	if !found {
		return fail(exitUsage, "unknown help topic; run `bibleit help` to list commands")
	}
	if activeFormat == "json" {
		return printLocal("help", map[string]string{"topic": strings.Join(topic, " "), "text": strings.Join(lines, "\n")}, "json")
	}
	fmt.Println(strings.Join(lines, "\n"))
	return exitOK
}
func registryCandidates(args []string) []string {
	out := map[string]bool{}
	if len(args) > 0 && args[0] == "help" {
		args = args[1:]
		for _, s := range commandRegistry {
			p := helpPath(strings.Fields(s.Syntax))
			if len(args) < len(p) && strings.Join(p[:len(args)], " ") == strings.Join(args, " ") {
				out[p[len(args)]] = true
			}
		}
	} else {
		for _, s := range commandRegistry {
			p := strings.Fields(s.Syntax)
			if len(args) > 1 && len(p) > 1 && p[0] == "live" && placeholder(p[1]) && oneOf(args[1], "list", "create", "delete") {
				continue
			}
			if len(args) >= len(p) {
				continue
			}
			ok := true
			for i, v := range args {
				if !placeholder(p[i]) && p[i] != v {
					ok = false
					break
				}
			}
			if ok && !placeholder(p[len(args)]) {
				out[p[len(args)]] = true
			}
		}
	}
	if len(args) == 0 {
		out["help"] = true
	}
	if len(args) == 1 && args[0] == "completion" {
		for _, v := range []string{"bash", "zsh", "fish"} {
			out[v] = true
		}
	}
	if len(args) == 2 && args[0] == "auth" && args[1] == "login" {
		out["ssh"] = true
	}
	if len(args) == 2 && args[0] == "server" && args[1] == "help" {
		for _, v := range []string{"server", "account", "auth", "translation", "live", "organization"} {
			out[v] = true
		}
	}
	if len(args) == 2 && args[0] == "config" && oneOf(args[1], "get", "set", "unset") {
		for _, v := range []string{"transport", "identity", "format"} {
			out[v] = true
		}
	}
	values := []string{}
	for v := range out {
		values = append(values, v)
	}
	sort.Strings(values)
	return values
}

func stackPushReference(values []string) (bibleit.Reference, error) {
	translation := ""
	if len(values) > 1 {
		_, numeric := strconv.Atoi(values[1])
		compact := false
		if parts := strings.Split(values[1], ":"); len(parts) == 2 {
			_, c := strconv.Atoi(parts[0])
			_, v := strconv.Atoi(parts[1])
			compact = c == nil && v == nil
		}
		// A final chapter:verse identifies the three-argument translated form,
		// including numeric books: bpml 31 1:1.
		translatedCompact := len(values) == 3 && strings.Contains(values[2], ":")
		if (numeric != nil && !compact) || len(values) == 4 || translatedCompact {
			translation = values[0]
			values = values[1:]
		}
	}
	ref, e := cliReference(translation, values)
	if e != nil {
		return ref, fmt.Errorf("invalid stack reference; use [translation] <book> [chapter] [verse]")
	}
	return ref, nil
}

// Public invocation variants cover optional grammar and aliases, independently
// of whether a resource binding has been implemented or functionally verified.
func commandInvocations(spec commandSpec) [][]string {
	variants := [][]string{strings.Fields(spec.Example)}
	extra := []string{}
	switch spec.Syntax {
	case "help [topic...]":
		extra = []string{"help", "help live stack pop"}
	case "server help [topic]":
		extra = []string{"server help", "server help account"}
	case "auth login [ssh]":
		extra = []string{"auth login ssh"}
	case "organization list [cursor]":
		extra = []string{"organization list synthetic-cursor"}
	case "read <translation> <book> [chapter] [verse]":
		extra = []string{"read web John", "read web John 3", "read web John 3:16", "read web 43 3 16"}
	case "live create [name...]":
		extra = []string{"live create"}
	case "live <id> stack pop [count]":
		extra = []string{"live L1 stack pop 2"}
	case "live <id> stack push [translation] <book> [chapter] [verse]":
		extra = []string{"live L1 stack push John", "live L1 stack push John 3", "live L1 stack push John 3 16", "live L1 stack push John 3:16", "live L1 stack push web 43 3 16"}
	}
	for _, s := range extra {
		variants = append(variants, strings.Fields(s))
	}
	return variants
}

func credentialExclusion(args []string, kind string) string {
	if kind == "svc_" && len(args) > 0 && args[0] == "organization" {
		return "personal membership/application operations require the requesting user, not a service identity"
	}
	if kind != "oa_" {
		return ""
	}
	if len(args) > 0 && args[0] == "organization" {
		return "delegated consent has no server.get scope for personal organization metadata"
	}
	if len(args) == 2 && args[0] == "account" && oneOf(args[1], "quotas", "tokens") {
		return "delegated consent has no server.get/token.get scope for account quota or credential metadata"
	}
	if len(args) == 3 && args[0] == "translation" && oneOf(args[1], "add", "remove") {
		return "delegated credentials have no consented library-write scope"
	}
	return ""
}
