package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func isolatedPreferences(t *testing.T) {
	t.Helper()
	t.Setenv("BIBLEIT_CONFIG", filepath.Join(t.TempDir(), "config.json"))
	for _, v := range []string{"BIBLEIT_PREFERENCES", "BIBLEIT_CREDENTIALS", "BIBLEIT_IDENTITY", "BIBLEIT_FORMAT", "BIBLEIT_TRANSPORT"} {
		t.Setenv(v, "")
	}
}
func TestPreferencesPrecedenceAndCredentialSeparation(t *testing.T) {
	isolatedPreferences(t)
	if e := saveConfig(config{AccessToken: "bt_CREDENTIAL_CANARY", Identity: "/legacy/key"}); e != nil {
		t.Fatal(e)
	}
	before, _ := os.ReadFile(configPath())
	for _, a := range [][]string{{"config", "set", "transport", "ssh"}, {"config", "set", "identity", "~/private-key"}, {"config", "set", "format", "json"}} {
		code, _, err := captureCLI(t, func() int { return run(a) })
		if code != exitOK {
			t.Fatal(code, err)
		}
	}
	data, _ := os.ReadFile(preferencesPath())
	after, _ := os.ReadFile(configPath())
	if string(before) != string(after) || strings.Contains(string(data), "CREDENTIAL_CANARY") {
		t.Fatal("preferences duplicated or modified credentials")
	}
	assertPrivateConfig(t, preferencesPath())
	c, e := loadConfig()
	if e != nil {
		t.Fatal(e)
	}
	c, f, e := applyPreferences(c, map[string]string{})
	if e != nil || c.Transport != "ssh" || c.Identity != "~/private-key" || f != "json" {
		t.Fatal(c.Transport, c.Identity, f, e)
	}
	t.Setenv("BIBLEIT_TRANSPORT", "http")
	t.Setenv("BIBLEIT_FORMAT", "raw")
	c, f, e = applyPreferences(c, map[string]string{"--format": "table"})
	if e != nil || c.Transport != "http" || f != "table" {
		t.Fatal(c.Transport, f, e)
	}
	_, out, stderr := captureCLI(t, func() int { return run([]string{"config", "show", "--format", "json"}) })
	if strings.Contains(out+stderr, "CREDENTIAL_CANARY") || !strings.Contains(out, "BIBLEIT_TRANSPORT") {
		t.Fatal(out, stderr)
	}
}
func TestInvalidYAMLDoesNotLeakValuesAndLocalHelpSurvives(t *testing.T) {
	isolatedPreferences(t)
	for _, s := range []string{"version: 1\ntoken: PRIVATE_CANARY\n", "version: 1\ntransport: PRIVATE_CANARY\n", "version: 1\nformat: [PRIVATE_CANARY]\n", "version: 1\ntransport: http\ntransport: ssh\n", "version: 1\n---\nversion: 1\n", "version: 1\n" + strings.Repeat(" ", 17000)} {
		os.WriteFile(preferencesPath(), []byte(s), 0600)
		if _, e := loadPreferences(); e == nil || strings.Contains(e.Error(), "PRIVATE_CANARY") {
			t.Fatal("invalid or leaking YAML error", e)
		}
		for _, a := range [][]string{{"version"}, {"help", "read"}, {"completion", "bash"}} {
			code, _, stderr := captureCLI(t, func() int { return run(a) })
			if code != exitOK || stderr != "" {
				t.Fatal(a, code, stderr)
			}
		}
	}
}
func TestConfigPathCompatibilityAndUnset(t *testing.T) {
	isolatedPreferences(t)
	legacy := configPath()
	yaml := preferencesPath()
	t.Setenv("BIBLEIT_CONFIG", yaml)
	if configPath() != legacy || preferencesPath() != yaml {
		t.Fatal("YAML selection reinterpreted credential path")
	}
	for _, a := range [][]string{{"config", "set", "transport", "ssh"}, {"config", "unset", "transport"}} {
		code, _, e := captureCLI(t, func() int { return run(a) })
		if code != exitOK {
			t.Fatal(e)
		}
	}
	p, e := loadPreferences()
	if e != nil || p.Transport != "" {
		t.Fatal(p, e)
	}
	for _, key := range []string{"token", "access_token", "endpoint", "yes"} {
		code, out, err := captureCLI(t, func() int { return run([]string{"config", "get", key}) })
		if code != exitUsage || out != "" || err == "" {
			t.Fatal(key, code, out, err)
		}
	}
}

func TestUnsetIdentityDoesNotResurrectLegacyKey(t *testing.T) {
	isolatedPreferences(t)
	if e := saveConfig(config{Identity: "/legacy/key", AccessToken: "bt_fixture"}); e != nil {
		t.Fatal(e)
	}
	code, _, stderr := captureCLI(t, func() int { return run([]string{"config", "unset", "identity"}) })
	if code != exitOK {
		t.Fatal(stderr)
	}
	c, _ := loadConfig()
	c, _, e := applyPreferences(c, map[string]string{})
	if e != nil || c.Identity != "" {
		t.Fatal("old key resurrected", c.Identity, e)
	}
}
