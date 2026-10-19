package main

import (
	"encoding/json"
	"errors"
	bibleit "github.com/bibleit-labs/bibleit-go"
	"github.com/bibleit-labs/bibleit-go/rest"
	"net/http"
	"strings"
	"testing"
)

func TestSafeErrorsAcrossFormats(t *testing.T) {
	errs := []error{
		&bibleit.ServerError{Code: "not_found", Fields: map[string]string{"owner": "PRIVATE_CANARY", "text": "RELIGIOUS_CANARY", "secret": "SECRET_CANARY"}},
		&bibleit.HTTPError{StatusCode: 403, Code: "forbidden", Fields: map[string]any{"nested": map[string]any{"token": "SECRET_CANARY"}}},
		&rest.ProblemError{Status: 429, Code: "rate_limited", Title: "PRIVATE_CANARY", Type: "https://example.invalid/SECRET_CANARY", Header: http.Header{"Retry-After": []string{"2"}}},
		errors.New("PRIVATE_CANARY SECRET_CANARY RELIGIOUS_CANARY"),
	}
	old := activeFormat
	defer func() { activeFormat = old }()
	for _, format := range []string{"table", "raw", "json", "ndjson"} {
		activeFormat = format
		for _, err := range errs {
			code, out, stderr := captureCLI(t, func() int { return protocolFailure(err) })
			if code == exitOK || out != "" || strings.Contains(stderr, "CANARY") {
				t.Fatal(format, code, out, stderr)
			}
			if format == "json" || format == "ndjson" {
				var v map[string]any
				if json.Unmarshal([]byte(stderr), &v) != nil || v["schema_version"] != float64(1) {
					t.Fatal(stderr)
				}
			}
		}
	}
	_, message, fields, status := describeError(exitRateLimited, errs[2])
	if status != 429 || fields["retry_after_ms"] != "2000" || !strings.Contains(message, "2000") {
		t.Fatal(message, fields, status)
	}
}
func TestSSHFailureNeverPrintsServerStderrOrKeyPath(t *testing.T) {
	err := sshCommandError("/private/PRIVATE_CANARY/key", "Permission denied SECRET_CANARY", errors.New("RELIGIOUS_CANARY"))
	_, message, _, _ := describeError(exitUnauthenticated, err)
	if strings.Contains(message, "CANARY") || !strings.Contains(message, "SSH") {
		t.Fatal(message)
	}
}
func TestMissingAndConcealedResourceErrorsRemainEquivalent(t *testing.T) {
	a, b := &bibleit.ServerError{Code: "not_found"}, &bibleit.ServerError{Code: "not_found", Fields: map[string]string{"owner": "secret", "exists": "true"}}
	ca, ma, fa, _ := describeError(exitNotFound, a)
	cb, mb, fb, _ := describeError(exitNotFound, b)
	if ca != cb || ma != mb || len(fa) != 0 || len(fb) != 0 {
		t.Fatal("existence disclosed")
	}
}

func TestTemporaryServerFailureGivesAvailabilityGuidance(t *testing.T) {
	err := &rest.ProblemError{Status: 503, Code: "temporarily_unavailable", Detail: "PRIVATE_CANARY"}
	_, message, _, status := describeError(exitFailure, err)
	if status != 503 || !strings.Contains(message, "temporarily unavailable") || strings.Contains(message, "PRIVATE_CANARY") || strings.Contains(message, "command help") {
		t.Fatal(message, status)
	}
}

func TestRESTFailurePreservesSafeRetryAfterGuidance(t *testing.T) {
	old := activeFormat
	activeFormat = "json"
	defer func() { activeFormat = old }()
	err := &rest.ProblemError{Status: 429, Code: "rate_limited", Detail: "PRIVATE_CANARY", Header: http.Header{"Retry-After": {"2"}}}
	code, out, stderr := captureCLI(t, func() int { return protocolFailure(err) })
	var envelope struct {
		Error struct {
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if code != exitRateLimited || out != "" || json.Unmarshal([]byte(stderr), &envelope) != nil || envelope.Error.Fields["retry_after_ms"] != "2000" || strings.Contains(stderr, "PRIVATE_CANARY") {
		t.Fatal(code, out, stderr)
	}
}
