package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// The browser fixture starts with an already authenticated synthetic browser
// session. It follows the real consent page, CSRF, loopback callback, PKCE token
// exchange and CLI logout; it does not bypass these server boundaries.
func requireDisposableCLIIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the disposable server container harness")
	}
	if os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("fixed endpoint fixture required")
	}
	for _, host := range []string{"api.bibleit.app", "bibleit.app"} {
		addresses, e := net.LookupIP(host)
		if e != nil || len(addresses) == 0 {
			t.Fatal("fixture DNS unavailable")
		}
		for _, address := range addresses {
			if !address.IsLoopback() {
				t.Fatal("fixture must remain loopback")
			}
		}
	}
}

func TestServerCLILocalAndAuth(t *testing.T) {
	requireDisposableCLIIntegration(t)
	directory := t.TempDir()
	credentials, preferences := filepath.Join(directory, "credentials.json"), filepath.Join(directory, "config.yaml")
	t.Setenv("BIBLEIT_CONFIG", credentials)
	t.Setenv("BIBLEIT_PREFERENCES", preferences)
	t.Setenv("BIBLEIT_TOKEN", "")
	t.Setenv("BIBLEIT_PROFILE", "")
	for _, spec := range commandRegistry {
		if !spec.Local || strings.HasPrefix(spec.Syntax, "auth ") {
			continue
		}
		for _, invocation := range commandInvocations(spec) {
			t.Run(strings.Join(invocation, " "), func(t *testing.T) {
				code, _, _ := captureCLI(t, func() int { return run(invocation) })
				if code != exitOK {
					t.Fatalf("local invocation exited %d", code)
				}
			})
		}
	}

	executable, e := os.Executable()
	if e != nil {
		t.Fatal("test executable unavailable")
	}
	openerDirectory := filepath.Join(directory, "browser")
	if e = os.Mkdir(openerDirectory, 0700); e != nil {
		t.Fatal("browser fixture directory failed")
	}
	targetFile := filepath.Join(directory, "authorization-url")
	if e = os.WriteFile(targetFile, nil, 0600); e != nil {
		t.Fatal("private browser target failed")
	}
	// Only the OS browser launcher is substituted. The actual CLI and HTTP flow run.
	if e = os.WriteFile(filepath.Join(openerDirectory, "xdg-open"), []byte("#!/bin/sh\nprintf '%s' \"$1\" > \"$BIBLEIT_BROWSER_TARGET_FILE\"\n"), 0700); e != nil {
		t.Fatal("browser fixture launcher failed")
	}
	childEnvironment := append(os.Environ(), "BIBLEIT_RUNTIME_FIXTURE=1", "BIBLEIT_RUNTIME_ADDRESS=127.0.0.1:443", "BIBLEIT_RUNTIME_CA="+os.Getenv("SSL_CERT_FILE"), "BIBLEIT_BROWSER_TARGET_FILE="+targetFile, "PATH="+openerDirectory+string(os.PathListSeparator)+os.Getenv("PATH"))
	command := func(ctx context.Context, args ...string) *exec.Cmd {
		argv := append([]string{"-test.run=^TestRuntimeFixtureProcess$", "--"}, args...)
		cmd := exec.CommandContext(ctx, executable, argv...)
		cmd.Env = childEnvironment
		return cmd
	}
	t.Run("browser_login_and_logout", func(t *testing.T) {
		session := os.Getenv("BIBLEIT_INTEGRATION_BROWSER_SESSION")
		if session == "" {
			t.Fatal("synthetic browser session missing")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
		defer cancel()
		child := command(ctx, "--transport", "http", "--format", "json", "auth", "login")
		var stdout, stderr bytes.Buffer
		child.Stdout = &stdout
		child.Stderr = &stderr
		if e := child.Start(); e != nil {
			t.Fatal("CLI login process did not start")
		}
		done := make(chan error, 1)
		go func() { done <- child.Wait() }()
		joined := false
		defer func() {
			cancel()
			if !joined {
				<-done
			}
		}()
		var target *url.URL
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			raw, e := os.ReadFile(targetFile)
			if e == nil && len(raw) > 0 {
				target, e = url.Parse(string(raw))
				if e != nil {
					t.Fatal("invalid authorization URL")
				}
				break
			}
			time.Sleep(20 * time.Millisecond)
		}
		if target == nil || target.Scheme != "https" || target.Host != "bibleit.app" || target.Path != "/cli/auth" {
			t.Fatal("CLI did not open the fixed authorization origin")
		}
		callback, e := url.Parse(target.Query().Get("redirect_uri"))
		if e != nil || callback.Scheme != "http" || callback.Hostname() != "127.0.0.1" || callback.Path != "/callback" {
			t.Fatal("invalid CLI loopback callback")
		}
		jar, _ := cookiejar.New(nil)
		jar.SetCookies(target, []*http.Cookie{{Name: "bibleit_session", Value: session, Path: "/", Secure: true}})
		browser := &http.Client{Jar: jar, Timeout: 10 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
		page, e := browser.Get(target.String())
		if e != nil {
			t.Fatal("authorization page request failed")
		}
		body, e := io.ReadAll(io.LimitReader(page.Body, 1<<20))
		page.Body.Close()
		if e != nil || page.StatusCode != 200 {
			t.Fatal("authorization page did not render")
		}
		csrf := regexp.MustCompile(`name="csrf" value="([a-f0-9]{64})"`).FindSubmatch(body)
		if len(csrf) != 2 {
			t.Fatal("authorization page missing CSRF form")
		}
		approval, e := browser.PostForm("https://bibleit.app/cli/auth", url.Values{"csrf": {string(csrf[1])}})
		if e != nil {
			t.Fatal("consent submission failed")
		}
		approval.Body.Close()
		location, e := url.Parse(approval.Header.Get("Location"))
		if e != nil || approval.StatusCode != 303 || location.Scheme != callback.Scheme || location.Host != callback.Host || location.Path != callback.Path {
			t.Fatal("consent did not redirect to the expected CLI callback")
		}
		response, e := (&http.Client{Timeout: 5 * time.Second}).Get(location.String())
		if e != nil {
			t.Fatal("CLI callback failed")
		}
		response.Body.Close()
		if response.StatusCode != 200 {
			t.Fatal("CLI callback rejected valid authorization")
		}
		select {
		case e := <-done:
			joined = true
			if e != nil {
				t.Fatal("CLI login failed")
			}
		case <-ctx.Done():
			t.Fatal("CLI login timed out")
		}
		raw, e := os.ReadFile(credentials)
		var cfg configDocument
		if e != nil || json.Unmarshal(raw, &cfg) != nil || !strings.HasPrefix(cfg.AccessToken, "bt_") || validateCredentialFile(credentials) != nil {
			t.Fatal("CLI login did not persist protected credentials")
		}
		if strings.Contains(stdout.String(), cfg.AccessToken) || strings.Contains(stderr.String(), cfg.AccessToken) {
			t.Fatal("CLI login exposed its credential")
		}
		logout := command(ctx, "--transport", "http", "--format", "json", "auth", "logout")
		if e := logout.Run(); e != nil {
			t.Fatal("CLI logout failed")
		}
		raw, e = os.ReadFile(credentials)
		cfg = configDocument{}
		if e != nil || json.Unmarshal(raw, &cfg) != nil || cfg.AccessToken != "" {
			t.Fatal("CLI logout retained its web credential")
		}
	})
	t.Run("registered_key_login", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		identity := os.Getenv("BIBLEIT_INTEGRATION_IDENTITY")
		if identity == "" {
			t.Fatal("fixture SSH identity missing")
		}
		child := command(ctx, "--transport", "ssh", "--identity", identity, "--format", "json", "auth", "login", "ssh")
		if e := child.Run(); e != nil {
			t.Fatal("CLI SSH login failed")
		}
		raw, e := os.ReadFile(preferences)
		if e != nil || !strings.Contains(string(raw), "transport: ssh") {
			t.Fatal("CLI SSH login did not remember its selected transport")
		}
	})
}
