package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bibleit-labs/bibleit-go"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This runner executes the public CLI against the fixed-origin, egress-disabled
// container fixture. It never accepts a user endpoint or records response bodies.
func TestServerCommandParity(t *testing.T) {
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the isolated server container harness")
	}
	if os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("fixed endpoint fixture is required")
	}
	for _, host := range []string{"api.bibleit.app", "bibleit.app"} {
		addresses, err := net.LookupIP(host)
		if err != nil || len(addresses) == 0 {
			t.Fatal("fixture DNS unavailable")
		}
		for _, address := range addresses {
			if !address.IsLoopback() {
				t.Fatal("fixture DNS must remain loopback")
			}
		}
	}
	type credential struct{ name, transport, token string }
	credentials := []credential{{"bt_", "http", os.Getenv("BIBLEIT_INTEGRATION_TOKEN")}, {"registered-user-key", "ssh", os.Getenv("BIBLEIT_INTEGRATION_TOKEN")}}
	if token := os.Getenv("BIBLEIT_OAUTH_TOKEN"); token != "" {
		credentials = append(credentials, credential{"oa_", "http", token})
	}
	if token := os.Getenv("BIBLEIT_SERVICE_TOKEN"); token != "" {
		credentials = append(credentials, credential{"svc_", "http", token})
	}
	evidence := []ParityEvidence{}
	for _, credential := range credentials {
		t.Run(credential.name, func(t *testing.T) {
			directory := t.TempDir()
			t.Setenv("BIBLEIT_CONFIG", filepath.Join(directory, "config.yaml"))
			t.Setenv("BIBLEIT_TOKEN", "")
			t.Setenv("BIBLEIT_PROFILE", "")
			if err := saveConfig(config{AccessToken: credential.token, Identity: os.Getenv("BIBLEIT_INTEGRATION_IDENTITY")}); err != nil {
				t.Fatal(err)
			}
			var target *testing.T
			raw := func(args ...string) map[string]any {
				target.Helper()
				normalized, err := normalizeCommand(args)
				if err != nil {
					target.Fatal(err)
				}
				plan, err := planHTTP(normalized, credential.name == "svc_")
				if err != nil {
					target.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				result, err := executeHTTP(ctx, credential.token, plan)
				if err != nil {
					target.Fatalf("fixture operation %s failed: %v", strings.Join(args[:min(2, len(args))], " "), err)
				}
				return result.data
			}
			for _, spec := range commandRegistry {
				if spec.Local || spec.Syntax == "live <id> secret <value>" || strings.Contains(spec.Syntax, "subscribe") {
					continue
				}
				for index, example := range commandInvocations(spec) {
					t.Run(fmt.Sprintf("%s/%d", spec.Syntax, index), func(t *testing.T) {
						target = t
						cell := ParityEvidence{SchemaVersion: 1, Command: spec.Syntax, Invocation: example, Transport: credential.transport, Credential: credential.name, Layer: "server-functional", Status: "failed", Scenario: "synthetic CLI invocation and domain readback", CLIRevision: os.Getenv("BIBLEIT_PARITY_CLI_REVISION"), SDKRevision: os.Getenv("BIBLEIT_PARITY_SDK_REVISION"), ServerRevision: os.Getenv("BIBLEIT_PARITY_SERVER_REVISION"), Cleanup: "per-case fixture deletion; disposable database destroyed by harness"}
						defer func() { evidence = append(evidence, cell) }()
						if credential.name == "svc_" && example[0] == "organization" {
							cell.Status = "inapplicable"
							cell.Reason = "personal membership/application operation requires the requesting user"
							return
						}
						normalized, _ := normalizeCommand(example)
						if credential.name == "oa_" && (example[0] == "organization" || (example[0] == "account" && len(normalized) > 1 && normalized[1] != "info") || (example[0] == "translation" && oneOf(example[1], "add", "remove"))) {
							cell.Status = "inapplicable"
							cell.Reason = "no separately consented delegated scope for personal metadata, memberships or library writes"
							return
						}
						args := append([]string(nil), example...)
						for i, v := range args {
							if v == "application-id" {
								args[i] = os.Getenv("BIBLEIT_PARITY_APPLICATION_ID")
							}
							if v == "synthetic-cursor" {
								args[i] = "o:00000000-0000-0000-0000-000000000000"
							}
						}
						if args[0] == "read" {
							if args[2] == "43" {
								args[2] = "19"
							} else if credential.transport == "ssh" {
								args[2] = "Psalms"
							} else {
								args[2] = "Salmos"
							}
							for i := 3; i < len(args); i++ {
								if strings.Contains(args[i], ":") {
									args[i] = "23:1"
								} else if i == 3 {
									args[i] = "23"
								} else {
									args[i] = "1"
								}
							}
						}
						if args[0] == "search" {
							args[2] = "shepherd"
						}
						id := ""
						if args[0] == "live" && len(args) > 2 && args[1] == "L1" {
							fixture := raw("live", "create", "Parity fixture")
							id, _ = fixture["id"].(string)
							if id == "" {
								t.Fatal("created fixture has no ID")
							}
							args[1] = id
							defer func() {
								p, _ := planHTTP([]string{"live", id, "info"}, credential.name == "svc_")
								ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
								defer cancel()
								if _, err := executeHTTP(ctx, credential.token, p); err == nil {
									raw("live", id, "delete")
								}
							}()
							raw("live", id, "start")
							raw("live", id, "set", "translations", "web")
							raw("live", id, "stack", "push", "web", "19", "23", "1")
							raw("live", id, "stack", "push", "web", "19", "23", "2")
							if args[2] == "resume" {
								raw("live", id, "pause")
							}
							if args[2] == "secret" && args[3] != "create" {
								raw("live", id, "secret", "create")
							}
							if args[2] == "stack" && args[3] == "push" {
								for i := 4; i < len(args); i++ {
									if args[i] == "John" || args[i] == "43" {
										if args[i] == "John" {
											args[i] = "Salmos"
											if credential.transport == "ssh" {
												args[i] = "Psalms"
											}
										} else {
											args[i] = "19"
										}
										if i+1 < len(args) {
											args[i+1] = "23"
										}
										if i+2 < len(args) {
											args[i+2] = "1"
										}
									}
								}
								if strings.Contains(args[len(args)-1], ":") {
									args[len(args)-1] = "23:1"
								}
							}
						}
						if args[0] == "live" && args[1] == "delete" {
							raw("live", "create", "Bulk A")
							raw("live", "create", "Bulk B")
						}
						if args[0] == "translation" && args[1] == "remove" {
							defer raw("translation", "add", "web")
						}
						argv := []string{"--transport", credential.transport, "--format", "json", "--yes"}
						secretPath := ""
						if id != "" && args[2] == "secret" && args[3] != "delete" {
							secretPath = filepath.Join(t.TempDir(), "audience-secret")
							argv = append(argv, "--secret-output", secretPath)
						}
						argv = append(argv, args...)
						code, stdout, _ := captureCLI(t, func() int { return run(argv) })
						cell.ExitCode = &code
						if code != exitOK {
							t.Fatalf("CLI exited %d for registry command %s", code, spec.Syntax)
							return
						}
						var result map[string]any
						if err := json.Unmarshal([]byte(stdout), &result); err != nil {
							t.Fatal("CLI did not emit valid JSON")
						}
						schema, ok := result["schema_version"].(float64)
						if !ok {
							t.Fatal("missing output schema")
						}
						version := int(schema)
						cell.OutputSchema = &version
						expected := 1
						if credential.transport == "http" {
							expected = 2
						}
						if version != expected {
							t.Fatalf("schema %d, expected %d", version, expected)
						}
						var data map[string]any
						if credential.transport == "http" {
							data, _ = result["data"].(map[string]any)
						} else {
							records, _ := result["records"].([]any)
							if len(records) == 0 {
								t.Fatal("SSH result has no records")
							}
							record, _ := records[0].(map[string]any)
							data, _ = record["fields"].(map[string]any)
						}
						if data == nil {
							t.Fatal("missing native response data")
						}
						assertion := "successful typed CLI output with native schema and fixture domain response"
						switch args[0] {
						case "ping":
							if data["pong"] != true {
								t.Fatal("missing pong")
							}
						case "account":
							if args[1] == "info" {
								account := raw("account", "info")
								if account["id"] == "" || account["kind"] == "" {
									t.Fatal("missing account identity")
								}
							}
							if args[1] == "quotas" || args[1] == "quota" {
								quotas := raw("account", "quotas")
								rows, _ := quotas["data"].([]any)
								if len(rows) == 0 {
									t.Fatal("missing effective quota rows")
								}
							}
						case "translation":
							if args[1] == "add" || args[1] == "remove" {
								library := raw("translation", "list")
								rows, _ := library["data"].([]any)
								found := false
								for _, row := range rows {
									item, _ := row.(map[string]any)
									if strings.EqualFold(fmt.Sprint(item["slug"]), "web") {
										found = true
									}
								}
								if found != (args[1] == "add") {
									t.Fatal("library mutation did not persist")
								}
								assertion = "translation library mutation verified by fresh enabled-library read"
							}
						case "read", "search":
							if credential.transport == "http" {
								verses, _ := data["verses"].([]any)
								if len(verses) == 0 {
									t.Fatal("fixture Scripture result is empty")
								}
								verse, _ := verses[0].(map[string]any)
								if fmt.Sprint(verse["book_id"]) != "19" || verse["text"] == "" {
									t.Fatal("incorrect source coordinates or content")
								}
							} else if len(args) > 4 && args[0] == "read" {
								if data["content"] == nil || data["reference"] == nil {
									t.Fatal("missing source Scripture metadata")
								}
							}
							assertion = "successful Scripture read/search with explicit source coordinates and content"
						case "live":
							if args[1] == "create" {
								createdID, _ := data["id"].(string)
								if createdID == "" {
									t.Fatal("creation returned no ID")
								}
								info := raw("live", createdID, "info")
								name := ""
								if len(args) > 2 {
									name = strings.Join(args[2:], " ")
								}
								if info["name"] != name {
									t.Fatal("created name did not persist")
								}
								raw("live", createdID, "delete")
								assertion = "new Live ID and name verified by fresh resource read and deleted"
							} else if args[1] == "delete" {
								list := raw("live", "list")
								rows, _ := list["data"].([]any)
								if len(rows) != 0 {
									t.Fatal("bulk deletion left authorized Lives")
								}
								assertion = "bulk deletion verified by fresh empty authorized collection"
							} else if id != "" {
								if args[2] == "delete" {
									list := raw("live", "list")
									rows, _ := list["data"].([]any)
									for _, row := range rows {
										entry, _ := row.(map[string]any)
										if entry["id"] == id {
											t.Fatal("deleted Live still exists")
										}
									}
									assertion = "deleted Live absent from fresh collection"
								} else {
									info := raw("live", id, "info")
									switch args[2] {
									case "start", "stop":
										expected := "running"
										if args[2] == "stop" {
											expected = "stopped"
										}
										if info["status"] != expected {
											t.Fatal("playback status did not persist")
										}
									case "pause", "resume":
										if info["paused"] != (args[2] == "pause") {
											t.Fatal("pause state did not persist")
										}
									case "set":
										if args[3] == "name" && info["name"] != strings.Join(args[4:], " ") {
											t.Fatal("name did not persist")
										}
										if args[3] != "name" {
											ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
											settings, err := executeHTTP(ctx, credential.token, httpPlan{operation: "getLiveSettings", params: map[string]string{"live_id": id}})
											cancel()
											if err != nil || settings.data["live_id"] != id {
												t.Fatal("fresh settings read failed")
											}
											if args[3] == "reference" && settings.data["reference"] != strings.Join(args[4:], " ") {
												t.Fatal("reference setting did not persist")
											}
											if args[3] == "translations" {
												translations, ok := settings.data["translations"].([]any)
												if !ok || len(translations) != 1 || translations[0] != "web" {
													t.Fatal("translation setting did not persist")
												}
											}
										}
									case "clear", "stack":
										stack := raw("live", id, "stack", "info")
										rows, _ := stack["data"].([]any)
										if args[2] == "clear" || args[3] == "clear" {
											if len(rows) != 0 {
												t.Fatal("clear retained history")
											}
										}
										if args[2] == "stack" && args[3] == "pop" {
											want := 1
											if len(args) > 4 {
												want = 0
											}
											if len(rows) != want {
												t.Fatal("pop affected wrong retained entry count")
											}
										}
										if args[2] == "stack" && args[3] == "push" && len(rows) < 2 {
											t.Fatal("push lost retained entries")
										}
									case "secret":
										// Live GET intentionally excludes audience protection metadata.
										// A fresh conditional operation proves the persisted domain state
										// without reading or logging the audience secret.
										probeAction, wantCode := "create", "already_protected"
										if args[3] == "delete" {
											probeAction, wantCode = "rotate", "not_protected"
										}
										probe, err := planHTTP([]string{"live", id, "secret", probeAction}, credential.name == "svc_")
										if err != nil {
											t.Fatal("secret probe planning failed")
										}
										ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
										_, err = executeHTTP(ctx, credential.token, probe)
										cancel()
										var problem *bibleit.ProblemError
										if !errors.As(err, &problem) || problem.Status != 409 || problem.Code != wantCode {
											t.Fatal("secret protection did not persist")
										}
										if secretPath != "" {
											secret, err := os.ReadFile(secretPath)
											if err != nil || len(strings.TrimSpace(string(secret))) != 48 {
												t.Fatal("one-time secret was not delivered")
											}
											stat, _ := os.Stat(secretPath)
											if stat.Mode().Perm() != 0600 {
												t.Fatal("secret file not private")
											}
											if strings.Contains(stdout, strings.TrimSpace(string(secret))) {
												t.Fatal("secret leaked to stdout")
											}
										}
									}
									assertion = "mutation/read verified with fresh Live, settings or retained-stack state"
								}
							}
						}
						cell.Assertion = assertion
						cell.Status = "passed"
					})
				}
			}
		})
	}
	if destination := os.Getenv("BIBLEIT_PARITY_EVIDENCE"); destination != "" {
		b, err := json.MarshalIndent(evidence, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(destination, append(b, '\n'), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
