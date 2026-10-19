package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

func TestServerSparseCatalogSSH(t *testing.T) {
	if os.Getenv("BIBLEIT_CLIENT_TEST_DISPOSABLE") != "1" {
		t.Skip("run the isolated --catalog-only server harness")
	}
	if defaultSSHServer != "bibleit.app:22" || os.Getenv("BIBLEIT_FIXED_ENDPOINT_FIXTURE") != "1" {
		t.Fatal("fixed SSH address requires the isolated fixture")
	}
	addresses, err := net.LookupIP("bibleit.app")
	if err != nil || len(addresses) == 0 {
		t.Fatal("fixture resolution failed")
	}
	for _, address := range addresses {
		if !address.IsLoopback() {
			t.Fatal("fixture resolves outside loopback")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	catalog, _ := bibleit.TranslationCatalogCommand("BPML")
	german, _ := bibleit.TranslationCatalogCommand("LUT1912")
	present, _ := bibleit.ReadCommand(bibleit.Reference{Translation: "BPML", Book: "42", Chapter: 17, Verse: 37})
	omitted, _ := bibleit.ReadCommand(bibleit.Reference{Translation: "BPML", Book: "42", Chapter: 17, Verse: 36})
	missing, _ := bibleit.TranslationCatalogCommand("not-a-translation")
	cfg := config{Identity: os.Getenv("BIBLEIT_INTEGRATION_IDENTITY")}
	knownHosts := filepath.Join(t.TempDir(), "known_hosts")
	for _, command := range []bibleit.Command{catalog, german, present, omitted, missing} {
		t.Run(command.String(), func(t *testing.T) {
			base, _, err := sshCommand(cfg, command)
			if err != nil {
				t.Fatal(err)
			}
			argv := append([]string{"-o", "UserKnownHostsFile=" + knownHosts, "-o", "BatchMode=yes"}, base.Args[1:]...)
			wire, execErr := exec.CommandContext(ctx, base.Path, argv...).Output()
			sshResult, sshErr := bibleit.DecodeCommandResponse(command, bytes.NewReader(wire))
			if sshErr != nil {
				var serverErr *bibleit.ServerError
				if !errors.As(sshErr, &serverErr) || !oneOf(serverErr.Code, "verse_not_found", "translation_not_found") {
					t.Fatal(sshErr)
				}
				return
			}
			if execErr != nil {
				t.Fatal(execErr)
			}

			code, _, stderr := captureCLI(t, func() int { return printResponse(sshResult.Lines, "json") })
			if code != exitOK {
				t.Fatalf("sparse JSON output: %s", stderr)
			}
		})
	}
}
