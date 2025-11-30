package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// Feature: CLI_INIT
// Spec: spec/commands/init.md
func TestNewInitCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewInitCommand()

	if cmd.Use != "init" {
		t.Fatalf("expected Use to be 'init', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func executeCommand(cmd *cobra.Command, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}

func TestInitCommand_DefaultConfigPath_InteractiveStub(t *testing.T) {
	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewInitCommand())

	out, err := executeCommand(root, "init")
	if err != nil {
		t.Fatalf("expected no error executing 'init' command, got: %v", err)
	}

	if !strings.Contains(out, "Initializing Stagecraft project (interactive, stub)") {
		t.Fatalf("expected interactive stub message, got: %q", out)
	}

	if !strings.Contains(out, "stagecraft.yml") {
		t.Fatalf("expected default config path 'stagecraft.yml' in output, got: %q", out)
	}
}

func TestInitCommand_NonInteractiveStub(t *testing.T) {
	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewInitCommand())

	out, err := executeCommand(root, "init", "--non-interactive", "--config", "custom.yml")
	if err != nil {
		t.Fatalf("expected no error executing 'init --non-interactive', got: %v", err)
	}

	if !strings.Contains(out, "Initializing Stagecraft project (non-interactive, stub)") {
		t.Fatalf("expected non-interactive stub message, got: %q", out)
	}

	if !strings.Contains(out, "custom.yml") {
		t.Fatalf("expected custom config path 'custom.yml' in output, got: %q", out)
	}
}
