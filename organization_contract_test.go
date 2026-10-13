package main

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

// Protocol-v1 decoding/output must accept the additive server read contract
// alongside the additive read commands without changing schema version 1.
func TestOrganizationFoundationResponseCompatibility(t *testing.T) {
	body, err := os.ReadFile("contract/fixtures/responses-v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []struct {
		Name      string `json:"name"`
		Wire      string `json:"wire"`
		Invalid   bool   `json:"invalid"`
		ErrorCode string `json:"error_code"`
		Records   []struct {
			Type   string            `json:"type"`
			Fields map[string]string `json:"fields"`
		} `json:"records"`
	}
	if err := json.Unmarshal(body, &fixtures); err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, fixture := range fixtures {
		if !strings.HasPrefix(fixture.Name, "organization-") {
			continue
		}
		checked++
		t.Run(fixture.Name, func(t *testing.T) {
			response, err := bibleit.DecodeResponse(strings.NewReader(fixture.Wire))
			if fixture.Invalid {
				if err == nil {
					t.Fatal("accepted truncated organization directory")
				}
				return
			}
			if fixture.ErrorCode != "" {
				var serverErr *bibleit.ServerError
				if !errors.As(err, &serverErr) || serverErr.Code != fixture.ErrorCode {
					t.Fatalf("lost server error: %v", err)
				}
				code, stdout, stderr := captureCLI(t, func() int {
					previousFormat := activeFormat
					defer func() { activeFormat = previousFormat }()
					activeFormat = "json"
					return protocolFailure(err)
				})
				expected := map[string]int{"unauthorized": exitUnauthenticated, "forbidden": exitForbidden, "not_found": exitNotFound, "organization_pending_verification": exitForbidden, "organization_suspended": exitForbidden, "invalid_cursor": exitFailure}[fixture.ErrorCode]
				if code != expected || stdout != "" || !strings.Contains(stderr, fixture.ErrorCode) {
					t.Fatalf("error output: %d %q %q", code, stdout, stderr)
				}
				var failure struct {
					SchemaVersion int  `json:"schema_version"`
					OK            bool `json:"ok"`
					ExitCode      int  `json:"exit_code"`
					Error         struct {
						Code string `json:"code"`
					} `json:"error"`
				}
				if err := json.Unmarshal([]byte(stderr), &failure); err != nil {
					t.Fatal(err)
				}
				if failure.SchemaVersion != 1 || failure.OK || failure.ExitCode != expected || failure.Error.Code != fixture.ErrorCode {
					t.Fatalf("changed failure envelope: %s", stderr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			code, stdout, stderr := captureCLI(t, func() int { return printResponse(response.Lines, "json") })
			if code != exitOK || stderr != "" {
				t.Fatalf("output: %d %q", code, stderr)
			}
			var envelope struct {
				SchemaVersion int            `json:"schema_version"`
				OK            bool           `json:"ok"`
				Records       []stableRecord `json:"records"`
			}
			if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
				t.Fatal(err)
			}
			if envelope.SchemaVersion != 1 || !envelope.OK || len(envelope.Records) != len(fixture.Records) {
				t.Fatalf("changed schema: %s", stdout)
			}
			for i, expected := range fixture.Records {
				actual := envelope.Records[i]
				if actual.Type != expected.Type {
					t.Fatalf("record type: %s", actual.Type)
				}
				for key, value := range expected.Fields {
					if key == "count" {
						if actual.Fields[key] != float64(len(fixture.Records)-1) {
							t.Fatalf("directory count lost numeric type: %#v", actual.Fields[key])
						}
					} else if actual.Fields[key] != value {
						t.Fatalf("field %s changed: %#v, want %q", key, actual.Fields[key], value)
					}
				}
				if state, exists := expected.Fields["state"]; exists && state != "approved" {
					if _, exists := actual.Fields["organization_id"]; exists {
						t.Fatal("pending application acquired an operational organization ID")
					}
				}
			}
		})
	}
	if checked < 13 {
		t.Fatalf("missing foundation compatibility cases: %d", checked)
	}
}
