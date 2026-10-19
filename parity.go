package main

// ParityEvidence separates binding/fixture checks from successful server-domain
// assertions. The coordinating real-server runner can emit one record per cell.
type ParityEvidence struct {
	SchemaVersion  int      `json:"schema_version"`
	Command        string   `json:"command"`
	Invocation     []string `json:"invocation"`
	Transport      string   `json:"transport"`
	Credential     string   `json:"credential"`
	Layer          string   `json:"layer"`  // registry, fixture, server-functional, native, deployed-smoke
	Status         string   `json:"status"` // pending, passed, failed, inapplicable
	Scenario       string   `json:"scenario"`
	Operation      string   `json:"operation,omitempty"`
	Reason         string   `json:"reason,omitempty"`
	CLIRevision    string   `json:"cli_revision,omitempty"`
	SDKRevision    string   `json:"sdk_revision,omitempty"`
	ServerRevision string   `json:"server_revision,omitempty"`
	ExitCode       *int     `json:"exit_code,omitempty"`
	OutputSchema   *int     `json:"output_schema,omitempty"`
	Assertion      string   `json:"assertion,omitempty"`
	Cleanup        string   `json:"cleanup,omitempty"`
	// Captures are deliberately absent: never persist bodies/argv containing
	// secrets, identities, Scripture or search text as generic diagnostic evidence.
}

func parityInventory() []ParityEvidence {
	out := []ParityEvidence{}
	for _, spec := range commandRegistry {
		for _, invocation := range commandInvocations(spec) {
			base := ParityEvidence{SchemaVersion: 1, Command: spec.Syntax, Invocation: invocation, Layer: "server-functional", Status: "pending", Scenario: "synthetic successful domain lifecycle"}
			if spec.Local {
				base.Transport = "local"
				base.Credential = "none"
				base.Layer = "native"
				out = append(out, base)
				continue
			}
			if spec.Syntax == "live <id> secret <value>" {
				base.Transport = "ssh"
				base.Credential = "registered-user-key"
				base.Layer = "registry"
				base.Status = "inapplicable"
				base.Reason = "standalone authorization cannot persist across connections"
				out = append(out, base)
				continue
			}
			args, _ := normalizeCommand(invocation)
			for _, c := range []struct {
				name    string
				service bool
			}{{"bt_", false}, {"oa_", false}, {"svc_", true}} {
				cell := base
				cell.Transport = "http"
				cell.Credential = c.name
				p, e := planHTTP(args, c.service)
				if e != nil {
					cell.Reason = "REST dispatch gap requires implementation"
					if c.service && len(args) > 0 && args[0] == "organization" {
						cell.Status = "inapplicable"
						cell.Reason = "personal membership/application operation requires the requesting user"
					}
				} else {
					cell.Operation = p.operation
				}
				if c.name == "oa_" {
					cell.Scenario = "delegated credential with explicitly allowed scopes; server applicability must be verified"
					if len(args) > 1 && args[0] == "translation" && (args[1] == "add" || args[1] == "remove") {
						cell.Status = "inapplicable"
						cell.Reason = "delegated personal library writes have no separately consented scope"
					}
				}
				if reason := credentialExclusion(args, c.name); reason != "" {
					cell.Status = "inapplicable"
					cell.Reason = reason
				}
				out = append(out, cell)
			}
			base.Transport = "ssh"
			base.Credential = "registered-user-key"
			out = append(out, base)
		}
	}
	return out
}
