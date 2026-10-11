package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// Test-only process entry point: exercise run with the fixed production URL,
// verified TLS, and an exclusively loopback dialer. Not in shipped binaries.
func TestRuntimeFixtureProcess(t *testing.T) {
	if os.Getenv("BIBLEIT_RUNTIME_FIXTURE") != "1" {
		t.Skip("used only by scripts/check_cli_runtime.py")
	}
	address := os.Getenv("BIBLEIT_RUNTIME_ADDRESS")
	host, _, err := net.SplitHostPort(address)
	if err != nil || host != "127.0.0.1" {
		t.Fatal("process fixture must use IPv4 loopback")
	}
	pem, err := os.ReadFile(os.Getenv("BIBLEIT_RUNTIME_CA"))
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(pem) {
		t.Fatal("invalid fixture CA")
	}
	cliHTTPClient = &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12},
		DialContext: func(ctx context.Context, network, target string) (net.Conn, error) {
			if target != "bibleit.app:443" {
				return nil, &net.AddrError{Err: "unexpected fixture destination", Addr: target}
			}
			return (&net.Dialer{}).DialContext(ctx, network, address)
		},
	}}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Exit(run(os.Args[i+1:]))
		}
	}
	t.Fatal("missing CLI argument delimiter " + strings.Join(os.Args[1:], " "))
}
