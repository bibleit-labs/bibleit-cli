package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSingleYAMLConfigPreservesPreferencesAndCredentials(t *testing.T) {
	isolatedPreferences(t)
	path := filepath.Join(t.TempDir(), "config.yaml")
	t.Setenv("BIBLEIT_CONFIG", path)
	if err := saveConfig(config{AccessToken: "bt_SECRET_CANARY", Identity: "/original/key"}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"config", "set", "transport", "ssh"}, {"config", "set", "format", "raw"}, {"config", "set", "identity", "/preferred/key"}} {
		code, _, stderr := captureCLI(t, func() int { return run(args) })
		if code != exitOK {
			t.Fatal(stderr)
		}
	}
	p, err := loadPreferences()
	if err != nil || p.AccessToken != "bt_SECRET_CANARY" || p.Identity != "/preferred/key" || p.Transport != "ssh" || p.Format != "raw" {
		t.Fatal("preference edit lost configuration")
	}
	code, out, stderr := captureCLI(t, func() int { return run([]string{"config", "show", "--format", "json"}) })
	if code != exitOK || stderr != "" || strings.Contains(out, "SECRET_CANARY") || !strings.Contains(out, "config.yaml") {
		t.Fatal("config show leaked credentials or omitted location")
	}
	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Transport, cfg.Endpoint, cfg.Identity = "ssh", defaultSSHServer, "/new/key"
	if err = saveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	if err = rememberTransport("ssh"); err != nil {
		t.Fatal(err)
	}
	p, err = loadPreferences()
	if err != nil || p.Identity != "/new/key" || p.AccessToken != "bt_SECRET_CANARY" || p.Format != "raw" {
		t.Fatal("login persistence lost settings")
	}
	cfg.AccessToken, cfg.Identity = "", ""
	if err = saveConfig(cfg); err != nil {
		t.Fatal(err)
	}
	p, err = loadPreferences()
	if err != nil || p.AccessToken != "" || p.Identity != "" || !p.IdentitySet || p.Transport != "ssh" || p.Format != "raw" {
		t.Fatal("clearing login lost defaults")
	}
	files, err := os.ReadDir(filepath.Dir(path))
	if err != nil || len(files) != 1 || files[0].Name() != "config.yaml" {
		t.Fatal("configuration created extra files")
	}
	assertPrivateConfig(t, path)
}

func TestSingleConfigIgnoresSeparateLegacyFilesAndOverrides(t *testing.T) {
	isolatedPreferences(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	t.Setenv("BIBLEIT_CONFIG", path)
	if err := saveConfig(config{Identity: "/current/key"}); err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(dir, "config.json")
	old := []byte("PRIVATE_INVALID_LEGACY_DATA")
	if err := os.WriteFile(legacy, old, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BIBLEIT_PREFERENCES", filepath.Join(dir, "unused-preferences.yaml"))
	t.Setenv("BIBLEIT_CREDENTIALS", legacy)
	cfg, err := loadConfig()
	if err != nil || cfg.Identity != "/current/key" || cfg.AccessToken != "" || configPath() != path || preferencesPath() != path {
		t.Fatal("separate legacy files influenced unified config")
	}
	after, err := os.ReadFile(legacy)
	if err != nil || string(after) != string(old) {
		t.Fatal("legacy file was migrated or modified")
	}
}

func TestCleanHelpAndRemovedConfigPath(t *testing.T) {
	for _, topic := range [][]string{nil, {"read"}, {"config"}, {"live", "stack"}} {
		code, out, stderr := captureCLI(t, func() int { return run(append([]string{"help"}, topic...)) })
		if code != exitOK || stderr != "" {
			t.Fatal("offline help failed")
		}
		for _, pollution := range []string{"HTTP (", "SSH (", "Local;", "scopes", "capabilities apply", "schema 1", "schema 2", "Alias for:", "Example:", "config path"} {
			if strings.Contains(out, pollution) {
				t.Fatalf("help contains unwanted detail %q", pollution)
			}
		}
	}
	if _, err := normalizeCommand([]string{"config", "path"}); err == nil || !strings.Contains(err.Error(), "config show") {
		t.Fatal("removed config path lacks replacement guidance")
	}
	for _, candidate := range completionCandidates([]string{"config", ""}) {
		if candidate == "path" {
			t.Fatal("completion advertises removed config path")
		}
	}
}
