package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	bibleit "github.com/bibleit-labs/bibleit-go"
	"github.com/bibleit-labs/bibleit-go/rest"
)

type userMessage struct{ message string }

func (e userMessage) Error() string { return e.message }

var safeErrorCode = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
var credentialValue = regexp.MustCompile(`\b(bt_|svc_|oa_)[A-Za-z0-9_.~+/-]+`)

func safeLocalMessage(s string) string {
	return humanValue(credentialValue.ReplaceAllString(s, "[redacted]"))
}
func describeError(exit int, err error) (string, string, map[string]any, int) {
	names := []string{"ok", "failure", "usage", "unauthenticated", "forbidden", "not_found", "conflict", "rate_limited"}
	code := names[exit]
	fields := map[string]any{}
	status := 0
	var server *bibleit.ServerError
	var httpErr *bibleit.HTTPError
	var problem *rest.ProblemError
	if errors.As(err, &server) {
		code = server.Code
		for k, v := range server.Fields {
			if k == "retry_after_ms" {
				if n, e := strconv.ParseUint(v, 10, 64); e == nil {
					fields[k] = strconv.FormatUint(n, 10)
				}
			}
		}
	}
	if errors.As(err, &httpErr) {
		status = httpErr.StatusCode
		if httpErr.Code != "" {
			code = httpErr.Code
		}
	}
	if errors.As(err, &problem) {
		status = problem.Status
		if problem.Code != "" {
			code = problem.Code
		}
		if d := problem.RetryAfter(); d > 0 {
			fields["retry_after_ms"] = fmt.Sprint(d.Milliseconds())
		}
	}
	if !safeErrorCode.MatchString(code) {
		code = names[exit]
	}
	message := "The operation failed. Check your connection and configuration, then retry an inspection command. A failed mutation may already have completed."
	switch code {
	case "unauthorized", "unauthenticated", "invalid_token", "revoked":
		message = "Authentication is missing, expired or revoked. Run `bibleit auth login`, or replace an explicitly supplied BIBLEIT_TOKEN."
	case "forbidden":
		message = "This identity is not allowed to perform this action. Check the selected transport, account and credential permissions."
	case "not_found", "live_not_found", "actor_not_found", "user_not_found":
		message = "The requested resource was not found or is unavailable to this identity. Check its ID and the selected account."
	case "translation_not_found":
		message = "The requested translation was not found or is unavailable. Check `bibleit translation list`."
	case "book_not_found", "chapter_not_found", "verse_not_found", "invalid_reference":
		message = "The requested Scripture reference is unavailable. Check `bibleit translation catalog <translation>` for valid references."
	case "quota_exceeded":
		message = "The account has reached its allowed quota. Check `bibleit account quotas` before retrying."
	case "conflict", "precondition_failed", "precondition_required":
		message = "The resource changed or requires a current revision. Inspect it again and review the action before retrying; no automatic retry was sent."
	case "rate_limited":
		message = "Too many requests. Wait before retrying."
		if delay, ok := fields["retry_after_ms"]; ok {
			message = fmt.Sprintf("Too many requests. Retry after %s ms.", delay)
		}
	case "live_stopped":
		message = "The Live session is stopped. Start it before updating playback."
	case "bad_request", "validation_error", "invalid_request":
		message = "The request is invalid. Check the command's help and argument values."
	case "invalid_cursor":
		message = "The pagination cursor is invalid or no longer usable. Restart the book read without --cursor."
	case "cursor_mismatch":
		message = "This cursor belongs to a different account, translation or book. Keep the original reference and account, or restart without --cursor."
	case "stale_cursor":
		message = "The translation edition changed during pagination. Restart the book read without --cursor to read the current edition."
	case "verse_too_large":
		message = "A source verse exceeds the service's response-size limit. The book could not be returned completely; report the translation and reference."
	case "database_timeout", "database_unavailable", "service_unavailable", "temporarily_unavailable":
		message = "The service is temporarily unavailable. A mutation may already have completed; inspect its state before retrying."
	}
	if code == "not_found" || code == "live_not_found" {
		if len(activeCommand) > 0 && activeCommand[0] == "live" {
			message = "The Live session was not found or is unavailable to this identity. Check `bibleit live list` and the selected account."
		}
	}
	if (server != nil || httpErr != nil || problem != nil) && strings.HasPrefix(message, "The operation failed.") {
		message = "The service rejected this operation (" + code + "). Check command help and server capabilities before retrying."
	}
	var c *confirmationError
	if errors.As(err, &c) {
		return c.code, safeLocalMessage(c.message), fields, status
	}
	var local userMessage
	if errors.As(err, &local) {
		message = safeLocalMessage(local.message)
	}
	var auth authenticationError
	if errors.As(err, &auth) {
		message = safeLocalMessage(auth.message)
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			message = "The request timed out. A mutation may already have completed; inspect its state before retrying."
		} else {
			message = "Could not connect to the service. Check your connection and retry an inspection command."
		}
	}
	if errors.Is(err, errors.ErrUnsupported) {
		message = "This operation is not supported by the selected service capability."
	}
	if errors.Is(err, rest.ErrInvalidBookPage) {
		message = "The book pages could not be combined safely. No partial book was returned. Restart the read; if it fails again, report the problem."
	}
	if errors.Is(err, context.Canceled) {
		message = "The operation was canceled."
	}
	if server != nil || httpErr != nil || problem != nil {
		if message == "" {
			message = "The service rejected this operation."
		}
	}
	return code, strings.TrimSpace(message), fields, status
}
