//go:build e2e

// test/e2e/init_smoke_test.go
package e2e

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

// Feature: CLI_INIT
// Spec: spec/commands/init.md

// This is a stub E2E test that expects the binary `stagecraft` to be
// in PATH or built beforehand. It is gated behind the `e2e` build tag
// so it won't run in normal `go test ./...` runs.
func TestStagecraftInit_Smoke(t *testing.T) {
	cmd := exec.Command("stagecraft", "init", "--non-interactive")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		t.Fatalf("expected 'stagecraft init' to succeed, got error: %v, output: %s", err, out.String())
	}

	if !strings.Contains(out.String(), "Initializing Stagecraft project") {
		t.Fatalf("expected output to contain init stub message, got: %q", out.String())
	}
}

// go test ./test/e2e -tags=e2e
