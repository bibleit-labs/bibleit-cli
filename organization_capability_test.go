package main

import (
	"strings"
	"testing"
)

func TestOrganizationCapabilityCompatibility(t *testing.T) {
	for _, tc := range []struct {
		name, capabilities string
		supported          bool
	}{
		{"canonical", "organization", true},
		{"removed legacy name", "organization_status", false},
		{"both", "organization,organization_status", true},
		{"among others", "translation,organization,live", true},
		{"removed name among others", "translation,organization_status,live", false},
		{"missing", "", false},
		{"unrelated", "translation,live", false},
		{"canonical substring", "organization_extra,my_organization", false},
		{"legacy substring", "organization_status_extra,my_organization_status", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := requireOrganizationCapability([]string{"OK capabilities=" + tc.capabilities, "END"})
			got := err == nil
			if got != tc.supported {
				t.Fatalf("supported = %v, want %v", got, tc.supported)
			}
			if err != nil && (!strings.Contains(err.Error(), "does not support organization;") || strings.Contains(err.Error(), "organization_status")) {
				t.Fatal("missing capability guidance does not use canonical name")
			}
		})
	}
	if requireOrganizationCapability([]string{"OK description=organization", "END"}) == nil {
		t.Fatal("accepted organization outside the capabilities field")
	}
}
