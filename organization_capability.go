package main

import (
	"errors"
	"strings"
)

// Require the canonical domain capability advertised by supported servers.
func requireOrganizationCapability(lines []string) error {
	for _, record := range parseRecords(lines) {
		for _, capability := range strings.Split(record.Fields["capabilities"], ",") {
			switch capability {
			case "organization":
				return nil
			}
		}
	}
	return errors.New("this server does not support organization; upgrade the server or use the organization page in your browser")
}
