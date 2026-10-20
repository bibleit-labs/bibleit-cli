package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"

	"github.com/bibleit-labs/bibleit-go/rest"
)

type httpPlan struct {
	operation, kind, book              string
	params                             map[string]string
	options                            rest.Options
	workspace, collection, conditional bool
	discovery, helpTopic               string
}

// planHTTP selects resource operations from CLI arguments, independently of SSH
// command construction. Unsupported resources never use command HTTP endpoints.
func planHTTP(args []string, service bool) (httpPlan, error) {
	p := httpPlan{params: map[string]string{}, kind: "ok"}
	pick := func(personal, organization string) string {
		if service {
			return organization
		}
		return personal
	}
	switch {
	case isLiveSubscribe(args):
		p.operation = "getLiveEvents"
		p.params["live_id"] = args[1]
	case len(args) == 1 && args[0] == "ping":
		p.operation = "getOpenApi"
		p.discovery = "ping"
	case len(args) == 2 && args[0] == "server" && args[1] == "info":
		p.operation = "getOpenApi"
		p.discovery = "info"
	case len(args) >= 2 && len(args) <= 3 && args[0] == "server" && args[1] == "help":
		p.operation = "getOpenApi"
		p.discovery = "help"
		p.kind = "help"
		if len(args) == 3 {
			p.helpTopic = args[2]
		}
	case len(args) == 2 && args[0] == "auth" && oneOf(args[1], "info", "whoami"):
		p.operation = pick("getPersonalIdentity", "getServiceOrganization")
		p.kind = "identity"
	case len(args) == 2 && args[0] == "account" && args[1] == "info":
		p.operation = pick("getPersonalAccount", "getOrganization")
		p.workspace = true
		p.kind = "account"
	case len(args) == 2 && args[0] == "account" && args[1] == "quotas":
		p.operation = pick("getPersonalQuotas", "getOrganizationQuotas")
		p.workspace = true
		p.kind = "quota"
	case len(args) == 2 && args[0] == "account" && args[1] == "tokens":
		p.operation = pick("listPersonalTokens", "listServiceTokens")
		p.workspace = true
		p.collection = true
		p.kind = "token"
	case len(args) >= 2 && len(args) <= 3 && args[0] == "organization" && args[1] == "list":
		if service {
			return p, fmt.Errorf("organization membership listing requires a personal credential; service tokens identify one organization and cannot list personal memberships")
		}
		p.operation = "listPersonalOrganizations"
		p.kind = "organization"
		if len(args) == 3 {
			p.options.Query = url.Values{"after": {args[2]}}
		}
	case len(args) == 4 && args[0] == "organization" && args[1] == "application" && args[3] == "status":
		if service {
			return p, fmt.Errorf("organization application status requires the requesting user's personal credential")
		}
		p.operation = "getPersonalOrganizationApplication"
		p.params["application_id"] = args[2]
		p.kind = "application"
	case len(args) == 2 && args[0] == "translation" && args[1] == "list":
		p.operation = pick("listPersonalTranslations", "listOrganizationTranslations")
		p.workspace = true
		p.collection = true
		p.kind = "translation"
	case len(args) == 3 && args[0] == "translation" && args[1] == "info":
		p.operation = "getTranslation"
		p.params["slug"] = strings.ToLower(args[2])
		p.kind = "translation"
	case len(args) == 3 && args[0] == "translation" && args[1] == "catalog":
		p.operation = "getTranslationCatalogueDetails"
		p.params["slug"] = strings.ToLower(args[2])
		p.kind = "book"
	case len(args) == 3 && args[0] == "translation" && oneOf(args[1], "add", "remove"):
		p.workspace = true
		p.params["slug"] = strings.ToLower(args[2])
		p.kind = "translation"
		if args[1] == "add" {
			p.operation = pick("enablePersonalTranslation", "enableOrganizationTranslation")
			p.options.JSON = map[string]any{}
		} else {
			p.operation = pick("disablePersonalTranslation", "disableOrganizationTranslation")
		}
	case len(args) >= 3 && len(args) <= 5 && args[0] == "read":
		p.workspace = true
		p.params["slug"] = strings.ToLower(args[1])
		p.book = args[2]
		p.kind = "verse"
		if len(args) == 3 {
			p.operation = pick("readPersonalBook", "readBook")
			break
		}
		p.params["chapter"] = args[3]
		p.operation = pick("readPersonalChapter", "readChapter")
		if !positiveHTTPNumber(args[3], 152) {
			return p, fmt.Errorf("chapter must be between 1 and 152")
		}
		if len(args) == 5 {
			if !positiveHTTPNumber(args[4], 255) {
				return p, fmt.Errorf("verse must be between 1 and 255")
			}
			p.params["verse"] = args[4]
			p.operation = pick("readPersonalVerse", "readVerse")
		}
	case len(args) == 3 && args[0] == "search":
		p.workspace = true
		p.params["slug"] = strings.ToLower(args[1])
		p.operation = pick("searchPersonalScripture", "searchScripture")
		p.kind = "verse"
		p.options.Query = url.Values{"q": {args[2]}, "limit": {"100"}}
	case len(args) == 2 && args[0] == "live" && args[1] == "list":
		p.workspace = true
		p.operation = pick("listPersonalLives", "listLives")
		p.collection = true
		p.kind = "live"
	case len(args) >= 2 && args[0] == "live" && args[1] == "create":
		name := strings.Join(args[2:], " ")

		p.workspace = true
		p.operation = pick("createPersonalLive", "createLive")
		body := rest.CreateLive{}
		if name != "" {
			body.Name = name
		}
		p.options.JSON = body
		p.kind = "live"
	case len(args) == 3 && args[0] == "live" && args[1] == "delete" && args[2] == "all":
		p.operation = pick("deletePersonalLives", "deleteOrganizationLives")
		p.workspace = true
		p.kind = "deletion"
	case len(args) == 3 && args[0] == "live" && args[2] == "stats":
		p.operation = "getLiveStatistics"
		p.params["live_id"] = args[1]
		p.kind = "statistics"
	case len(args) >= 4 && args[0] == "live" && args[2] == "stack":
		p.params["live_id"] = args[1]
		p.kind = "stack"
		switch args[3] {
		case "info":
			p.operation = "getLiveStack"
		case "clear":
			p.operation = "clearLiveStack"
			p.conditional = true
		case "pop":
			p.operation = "popLiveStack"
			p.conditional = true
			count := int64(1)
			if len(args) == 5 {
				n, e := strconv.ParseInt(args[4], 10, 64)
				if e != nil || n <= 0 || n > 2147483647 {
					return p, fmt.Errorf("pop count must be an integer from 1 to 2147483647")
				}
				count = n
			}
			p.options.JSON = rest.LiveStackRemoval{Count: &count}
		case "push":
			ref, e := stackPushReference(args[4:])
			if e != nil {
				return p, e
			}
			p.operation = "pushLiveStack"
			p.conditional = true
			p.book = ref.Book
			body := rest.LiveStackPush{}
			if ref.Translation != "" {
				body.Reference.Translation = &ref.Translation
			}
			if ref.Chapter != 0 {
				n := int64(ref.Chapter)
				body.Reference.Chapter = &n
			}
			if ref.Verse != 0 {
				n := int64(ref.Verse)
				body.Reference.Verse = &n
			}
			p.options.JSON = body
		default:
			return p, fmt.Errorf("unknown stack action; use info, push, pop, or clear")
		}
	case len(args) == 4 && args[0] == "live" && args[2] == "secret" && oneOf(args[3], "create", "rotate", "delete"):
		p.params["live_id"] = args[1]
		p.kind = "secret"
		p.conditional = true
		switch args[3] {
		case "create":
			p.operation = "createLiveAudienceSecret"
			p.options.JSON = map[string]any{}
		case "rotate":
			p.operation = "rotateLiveAudienceSecret"
			p.options.JSON = map[string]any{}
		case "delete":
			p.operation = "deleteLiveAudienceSecret"
		}
	case len(args) >= 5 && args[0] == "live" && args[2] == "set" && oneOf(args[3], "reference", "translations"):
		p.operation = "updateLiveSettings"
		p.params["live_id"] = args[1]
		p.conditional = true
		p.kind = "settings"
		patch := rest.LiveSettingsPatch{}
		if args[3] == "reference" {
			value := strings.Join(args[4:], " ")
			patch.Reference = &value
		} else {
			patch.Translations = args[4:]
		}
		p.options.JSON = patch
	case len(args) == 3 && args[0] == "live" && args[1] != "delete" && oneOf(args[2], "info", "start", "stop", "pause", "resume", "clear", "delete"):
		p.params["live_id"] = args[1]
		p.kind = "live"
		if args[2] == "info" {
			p.operation = "getLive"
			break
		}
		p.conditional = true
		p.operation = "updateLivePlayback"
		switch args[2] {
		case "start":
			p.options.JSON = map[string]any{"status": "running"}
		case "stop":
			p.options.JSON = map[string]any{"status": "stopped"}
		case "pause":
			p.options.JSON = map[string]any{"paused": true}
		case "resume":
			p.options.JSON = map[string]any{"paused": false}
		case "clear":
			p.operation = "replaceLiveDisplay"
			p.options.JSON = map[string]any{"type": "empty"}
		case "delete":
			p.operation = "deleteLive"
		}
	case len(args) >= 5 && args[0] == "live" && args[2] == "set" && args[3] == "name":
		p.operation = "updateLiveMetadata"
		p.params["live_id"] = args[1]
		p.conditional = true
		p.options.JSON = rest.LiveMetadataPatch{Name: strings.Join(args[4:], " ")}
		p.kind = "live"
	default:
		return p, fmt.Errorf("invalid command; run `bibleit help` for supported syntax")
	}
	return p, nil
}
func positiveHTTPNumber(s string, max int) bool {
	n, err := strconv.Atoi(s)
	return err == nil && n > 0 && n <= max
}

func runHTTP(cfg config, args []string, format string, yes bool) int {
	return runHTTPPage(cfg, args, format, yes, nil)
}

// A nil page preserves the complete CLI read. Explicit pagination displays one
// page and retains its continuation cursor in machine-readable output.
func runHTTPPage(cfg config, args []string, format string, yes bool, page url.Values) int {
	if strings.HasPrefix(cfg.AccessToken, "oa_") {
		if reason := credentialExclusion(args, "oa_"); reason != "" {
			return fail(exitForbidden, "%s; explicitly select an authorized credential; no transport fallback was attempted", reason)
		}
	}
	plan, err := planHTTP(args, strings.HasPrefix(cfg.AccessToken, "svc_"))
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	if page != nil {
		plan.options.Query = page
	}
	if cfg.AccessToken == "" && plan.discovery == "" {
		return protocolFailure(authenticationError{message: "web authentication is not configured; run `bibleit auth login`"})
	}
	if isLiveSubscribe(args) {
		return runHTTPWatch(cfg, args[1], format)
	}
	if action := destructiveAction(args); action != "" && !yes {
		cfg.Endpoint = rest.Endpoint
		if err := confirmAction(cfg, action, format, os.Stdin, os.Stderr, terminalInput(os.Stdin) && terminalInput(os.Stderr)); err != nil {
			if e, ok := err.(*confirmationError); ok && e.code == "confirmation_required" {
				return outputError(exitUsage, err)
			}
			return outputError(exitFailure, err)
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	result, err := executeHTTP(ctx, cfg.AccessToken, plan)
	if err != nil {
		return protocolFailure(err)
	}
	return printHTTP(result, plan.kind, format)
}

// httpResult retains structured REST data through to output; it never encodes
// synthetic command responses or reparses command-protocol records.
type httpResult struct{ data map[string]any }

// Preserve integer values exactly when forwarding native resource JSON.
type httpData map[string]any

func (d *httpData) UnmarshalJSON(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return err
	}
	*d = object
	return nil
}

type httpClient struct {
	client  *rest.Client
	ctx     context.Context
	service bool
}

func executeHTTP(ctx context.Context, token string, p httpPlan) (httpResult, error) {
	c, err := rest.NewClient(token, cliHTTPClient)
	if err != nil {
		return httpResult{}, err
	}
	a := httpClient{client: c, ctx: ctx, service: strings.HasPrefix(token, "svc_")}
	if p.workspace {
		key, id, err := a.workspace()
		if err != nil {
			return httpResult{}, err
		}
		p.params[key] = id
	}
	if p.book != "" {
		id, err := a.book(p.book)
		if err != nil {
			return httpResult{}, err
		}
		if body, ok := p.options.JSON.(rest.LiveStackPush); ok {
			n, e := strconv.ParseInt(id, 10, 64)
			if e != nil {
				return httpResult{}, fmt.Errorf("invalid resolved book ID")
			}
			body.Reference.BookID = n
			p.options.JSON = body
		} else {
			p.params["book_id"] = id
		}
	}
	if p.conditional {
		_, etag, err := a.call("getLive", map[string]string{"live_id": p.params["live_id"]}, nil)
		if err != nil {
			return httpResult{}, err
		}
		if etag == "" {
			return httpResult{}, fmt.Errorf("Live response did not include an ETag")
		}
		p.options.IfMatch = etag
	}
	if p.collection {
		rows, err := a.rows(p.operation, p.params)
		if err != nil {
			return httpResult{}, err
		}
		return httpResult{data: map[string]any{"data": rows, "next_cursor": nil}}, nil
	}
	if oneOf(p.operation, "readPersonalBook", "readBook") && p.options.Query == nil {
		return a.completeBook(p)
	}
	data, _, err := a.call(p.operation, p.params, &p.options)
	if err != nil {
		return httpResult{}, err
	}
	if p.discovery != "" {
		if _, ok := data["openapi"].(string); !ok {
			return httpResult{}, fmt.Errorf("API discovery response is missing openapi")
		}
		if _, ok := data["info"].(map[string]any); !ok {
			return httpResult{}, fmt.Errorf("API discovery response is missing info")
		}
		if p.discovery == "ping" {
			data = map[string]any{"pong": true}
		} else if p.discovery == "help" {
			commands := []any{}
			for _, spec := range commandRegistry {
				if spec.Local {
					continue
				}
				if p.helpTopic != "" && !strings.HasPrefix(spec.Syntax, p.helpTopic+" ") && spec.Syntax != p.helpTopic {
					continue
				}
				commands = append(commands, map[string]any{"usage": "bibleit " + spec.Syntax, "summary": spec.Summary})
			}
			data = map[string]any{"source": "local CLI registry with API discovery", "permission_filtered": false, "commands": commands, "note": "API contract discovery does not establish effective credential permissions."}
		} else {
			data = map[string]any{"endpoint": rest.Endpoint, "openapi": data["openapi"], "info": data["info"]}
		}
	}
	if data == nil {
		data = map[string]any{"deleted": true}
		for k, v := range p.params {
			data[k] = v
		}
	}
	return httpResult{data: data}, nil
}
func (a httpClient) call(op string, p map[string]string, o *rest.Options) (map[string]any, string, error) {
	r, err := a.client.Call(a.ctx, op, p, o)
	if err != nil {
		return nil, "", err
	}
	defer r.Body.Close()
	if r.StatusCode == 204 {
		return nil, r.Header.Get("ETag"), nil
	}
	var data httpData
	err = r.Decode(&data)
	if err == nil && data == nil {
		err = fmt.Errorf("expected a REST JSON object")
	}
	return data, r.Header.Get("ETag"), err
}
func (a httpClient) workspace() (string, string, error) {
	op, key, field := "getPersonalIdentity", "account_id", "personal_account_id"
	if a.service {
		op, key, field = "getServiceOrganization", "organization_id", "id"
	}
	data, _, err := a.call(op, nil, nil)
	if err != nil {
		return "", "", err
	}
	id, _ := data[field].(string)
	if id == "" {
		return "", "", fmt.Errorf("REST identity response did not include %s", field)
	}
	return key, id, nil
}
func (a httpClient) rows(op string, p map[string]string) ([]any, error) {
	rows := []any{}
	after := ""
	seen := map[string]bool{}
	for {
		q := url.Values{"limit": {"100"}}
		if after != "" {
			q.Set("after", after)
		}
		d, _, err := a.call(op, p, &rest.Options{Query: q})
		if err != nil {
			return nil, err
		}
		items, ok := d["data"].([]any)
		if !ok {
			return nil, fmt.Errorf("REST collection response is missing data")
		}
		rows = append(rows, items...)
		next, _ := d["next_cursor"].(string)
		if next == "" {
			return rows, nil
		}
		if seen[next] {
			return nil, fmt.Errorf("repeated REST collection cursor")
		}
		seen[next] = true
		after = next
	}
}
func (a httpClient) book(input string) (string, error) {
	if positiveHTTPNumber(input, 255) {
		return input, nil
	}
	rows, err := a.rows("listBooks", nil)
	if err != nil {
		return "", err
	}
	for _, row := range rows {
		b, _ := row.(map[string]any)
		matches := []string{fmt.Sprint(b["code"])}
		if aliases, ok := b["aliases"].([]any); ok {
			for _, v := range aliases {
				matches = append(matches, fmt.Sprint(v))
			}
		}
		if names, ok := b["names"].(map[string]any); ok {
			for _, v := range names {
				matches = append(matches, fmt.Sprint(v))
			}
		}
		for _, v := range matches {
			if strings.EqualFold(v, input) {
				return fmt.Sprint(b["id"]), nil
			}
		}
	}
	return "", fmt.Errorf("unknown book %q", input)
}
func printHTTP(result httpResult, kind, format string) int {
	if kind == "secret" && activeSecretFile != nil {
		if secret, ok := result.data["secret"].(string); ok {
			replacement, e := deliverSecret(secret)
			if e != nil {
				return protocolFailure(e)
			}
			safe := map[string]any{}
			for k, v := range result.data {
				safe[k] = v
			}
			safe["secret"] = replacement
			result.data = safe
		} else {
			return outputError(exitFailure, userMessage{"The server did not supply a one-time secret; inspect the Live session before retrying."})
		}
	}
	if format == "raw" {
		return encodeHTTPJSON(result.data)
	}
	if format == "table" && presentHTTPHuman(result.data, kind) {
		if cursor, ok := result.data["next_cursor"].(string); kind == "verse" && ok && cursor != "" {
			fmt.Printf("\nNext cursor: %s\nContinue this book with --cursor <cursor>; use --format json to retain page metadata.\n", humanValue(cursor))
		}
		return exitOK
	}
	rows := []map[string]any{}
	hasRows := false
	for _, key := range []string{"data", "verses"} {
		if data, ok := result.data[key].([]any); ok {
			hasRows = true
			metadata := map[string]any{}
			for k, v := range result.data {
				if k != key && k != "next_cursor" {
					metadata[k] = v
				}
			}
			if len(metadata) > 0 {
				rows = append(rows, metadata)
			}
			for _, v := range data {
				if row, ok := v.(map[string]any); ok {
					rows = append(rows, row)
				}
			}
			break
		}
	}
	if !hasRows {
		rows = append(rows, result.data)
	}
	if format == "json" {
		return encodeHTTPJSON(map[string]any{"schema_version": 2, "ok": true, "data": result.data})
	}

	records := []outputRecord{}
	for _, row := range rows {
		fields := map[string]string{}
		order := []string{}
		for k, v := range row {
			if v == nil {
				continue
			}
			var s string
			switch x := v.(type) {
			case string:
				s = x
			default:
				b, _ := json.Marshal(x)
				s = string(b)
			}
			fields[k] = s
			order = append(order, k)
		}
		sort.Strings(order)
		records = append(records, outputRecord{Type: kind, Fields: fields, order: order})
	}
	return printOutputRecords(records, format)
}

func encodeHTTPJSON(value any) int {
	e := json.NewEncoder(os.Stdout)
	e.SetIndent("", "  ")
	if err := e.Encode(value); err != nil {
		return protocolFailure(err)
	}
	return exitOK
}
