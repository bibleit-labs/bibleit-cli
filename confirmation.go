package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type confirmationError struct{ code, message string }

func (e *confirmationError) Error() string { return e.message }

// Only validated command arguments reach this classification.
func destructiveAction(args []string) string {
	spec := commandSpecFor(args)
	if spec == nil {
		return ""
	}
	return spec.confirmation(args)
}
func (spec commandSpec) confirmation(args []string) string {
	switch spec.Syntax {
	case "live delete all":
		return "Delete ALL Lives managed by the authenticated user (including retained stacks and subscriber connections)"
	case "live <id> delete":
		return fmt.Sprintf("Delete Live %q, including its retained stack and subscriber connections", args[1])
	case "live <id> clear", "live <id> stack clear":
		return fmt.Sprintf("Clear every retained stack entry in Live %q", args[1])
	case "live <id> secret rotate":
		return fmt.Sprintf("Replace the secret for Live %q; the previous audience secret will stop working", args[1])
	case "live <id> secret delete":
		return fmt.Sprintf("Remove secret protection from Live %q, opening audience access", args[1])
	}
	return ""
}

func confirmAction(cfg config, action, format string, input io.Reader, output io.Writer, interactive bool) error {
	context := action
	if !interactive || format == "json" || format == "ndjson" {
		return &confirmationError{"confirmation_required", context + "; rerun with --yes to confirm"}
	}
	answer := "yes"
	if strings.HasPrefix(action, "Delete ALL Lives") {
		answer = "delete all"
	}
	if _, err := fmt.Fprintf(output, "%s.\nType %q to confirm: ", humanValue(context), answer); err != nil {
		return err
	}
	// A single bounded line: never consume piped commands or accept a prefix.
	line, err := bufio.NewReader(io.LimitReader(input, 1025)).ReadString('\n')
	if err != nil || len(line) > 1024 || strings.TrimSpace(line) != answer {
		return &confirmationError{"cancelled", "cancelled; no request sent"}
	}
	return nil
}
