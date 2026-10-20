package main

import (
	"bytes"
	"os"
	"strings"
	"syscall"
	"testing"
)

func TestWindowsConsoleHandles(t *testing.T) {
	// Allocate a real console if the test runner supplied only redirected handles.
	// Keep an existing console attached; never detach the runner's console.
	kernel := syscall.NewLazyDLL("kernel32.dll")
	created, _, allocationErr := kernel.NewProc("AllocConsole").Call()
	if created != 0 {
		defer kernel.NewProc("FreeConsole").Call()
	}
	input, err := os.OpenFile("CONIN$", os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open console input: %v (allocation: %v)", err, allocationErr)
	}
	defer input.Close()
	output, err := os.OpenFile("CONOUT$", os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	if !terminalInput(input) || !terminalInput(output) {
		t.Fatal("real console handles were not recognized")
	}
	if terminalColumns(output) <= 0 {
		t.Fatal("console width was not detected")
	}
	var prompt bytes.Buffer
	err = confirmAction(config{Transport: "http", Endpoint: "https://example.invalid"}, "Delete Live L1", "table", strings.NewReader("yes\r\n"), &prompt, terminalInput(input) && terminalInput(output))
	if err != nil || !strings.Contains(prompt.String(), "Type") {
		t.Fatalf("Windows console confirmation: %v %q", err, prompt.String())
	}
}
