package main

import (
	"encoding/json"
	"github.com/bibleit-labs/bibleit-go"
	"os"
	"strings"
	"testing"
)

func TestEveryRegistryResourceHasRESTBindingOrGenuineExclusion(t *testing.T) {
	ops := bibleit.Operations()
	for _, spec := range commandRegistry {
		if spec.Local || strings.Contains(spec.Syntax, "secret <value>") {
			continue
		}
		for _, invocation := range commandInvocations(spec) {
			a, e := normalizeCommand(invocation)
			if e != nil {
				t.Fatal(spec.Syntax, e)
			}
			for _, service := range []bool{false, true} {
				p, e := planHTTP(a, service)
				if service && a[0] == "organization" {
					if e == nil {
						t.Fatal("service identity substituted for personal memberships/applications")
					}
					continue
				}
				if e != nil {
					t.Errorf("missing binding %s service=%v: %v", spec.Syntax, service, e)
					continue
				}
				if _, ok := ops[p.operation]; !ok {
					t.Errorf("SDK contract missing %s for %s", p.operation, spec.Syntax)
				}
			}
		}
	}
	for _, cell := range parityInventory() {
		if cell.Layer == "server-functional" && cell.Status == "passed" {
			t.Fatal("binding tests must not claim server parity")
		}
	}
}

// An external isolated runner can request a machine-readable, pending inventory.
// This creates no production data and includes no credentials or content.
func TestExportCLIParityInventory(t *testing.T) {
	path := os.Getenv("BIBLEIT_PARITY_INVENTORY")
	if path == "" {
		t.Skip("set BIBLEIT_PARITY_INVENTORY to a disposable output path")
	}
	b, e := json.MarshalIndent(parityInventory(), "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(path, append(b, '\n'), 0600); e != nil {
		t.Fatal(e)
	}
}
