//go:build !windows

package main

import (
	"os"
	"testing"
)

func assertPrivateConfig(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("config mode is %o", info.Mode().Perm())
	}
}

func TestCredentialReadsRejectPermissiveFilesAndSymlinks(t *testing.T) {
	isolatedPreferences(t)
	if e := saveConfig(config{AccessToken: "bt_CANARY"}); e != nil {
		t.Fatal(e)
	}
	if e := os.Chmod(configPath(), 0644); e != nil {
		t.Fatal(e)
	}
	if _, e := loadConfig(); e == nil {
		t.Fatal("public credential file accepted")
	}
	target := configPath()
	if e := os.Chmod(target, 0600); e != nil {
		t.Fatal(e)
	}
	link := target + ".link"
	if e := os.Symlink(target, link); e != nil {
		t.Fatal(e)
	}
	t.Setenv("BIBLEIT_CONFIG", link)
	if _, e := loadConfig(); e == nil {
		t.Fatal("symlink accepted")
	}
}
