package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBrowserAuthorizationDenied(t *testing.T) {
	for _, code := range []string{"account_blocked", "access_denied", "authorization_expired", "unknown"} {
		t.Run(code, func(t *testing.T) {
			result := make(chan browserAuthorizationResult, 1)
			w := httptest.NewRecorder()
			browserAuthorizationCallback("expected", result)(w, httptest.NewRequest("GET", "/callback?state=expected&error="+code+"&code=must-not-exchange&error_description=SECRET", nil))
			select {
			case received := <-result:
				var auth authenticationError
				if received.code != "" || !errors.As(received.err, &auth) {
					t.Fatalf("denial was not an authentication failure: %#v", received)
				}
				if strings.Contains(received.err.Error(), "SECRET") || strings.Contains(w.Body.String(), "SECRET") {
					t.Fatal("reflected untrusted error description")
				}
				if code == "account_blocked" && !strings.Contains(received.err.Error(), "account is blocked") {
					t.Fatal("blocked-account reason missing")
				}
			default:
				t.Fatal("login would remain waiting after denial")
			}
		})
	}
}

func TestBrowserAuthorizationRejectsUntrustedCallbacks(t *testing.T) {
	for _, target := range []string{"/callback?state=wrong&error=account_blocked", "/callback?error=account_blocked", "/callback?state=expected&state=wrong&error=account_blocked"} {
		result := make(chan browserAuthorizationResult, 1)
		w := httptest.NewRecorder()
		browserAuthorizationCallback("expected", result)(w, httptest.NewRequest("GET", target, nil))
		if w.Code != http.StatusBadRequest || len(result) != 0 {
			t.Fatal("untrusted callback terminated login")
		}
	}
	result := make(chan browserAuthorizationResult, 1)
	w := httptest.NewRecorder()
	browserAuthorizationCallback("expected", result)(w, httptest.NewRequest("POST", "/callback?state=expected&error=account_blocked", nil))
	if w.Code != http.StatusMethodNotAllowed || len(result) != 0 {
		t.Fatal("non-GET callback terminated login")
	}
}

func TestBrowserAuthorizationSuccess(t *testing.T) {
	result := make(chan browserAuthorizationResult, 1)
	w := httptest.NewRecorder()
	browserAuthorizationCallback("expected", result)(w, httptest.NewRequest("GET", "/callback?state=expected&code=approved", nil))
	received := <-result
	if received.code != "approved" || received.err != nil || w.Code != http.StatusOK {
		t.Fatalf("successful callback rejected: %#v", received)
	}
}
