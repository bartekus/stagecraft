package commands

import (
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

func TestInitCommand_DefaultConfigPath_InteractiveStub(t *testing.T) {
	root := &cobra.Command{Use: "stagecraft"}
	root.AddCommand(NewInitCommand())

	out, err := executeCommandForGolden(root, "init")
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

	out, err := executeCommandForGolden(root, "init", "--non-interactive", "--config", "custom.yml")
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

// TestInitCommand_GoldenFiles tests CLI output against golden files.
func TestInitCommand_GoldenFiles(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		golden   string
		setupCmd func() *cobra.Command
	}{
		{
			name:   "init_default",
			args:   []string{"init"},
			golden: "init_default",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "stagecraft"}
				root.AddCommand(NewInitCommand())
				return root
			},
		},
		{
			name:   "init_non_interactive",
			args:   []string{"init", "--non-interactive"},
			golden: "init_non_interactive",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "stagecraft"}
				root.AddCommand(NewInitCommand())
				return root
			},
		},
		{
			name:   "init_custom_config",
			args:   []string{"init", "--config", "custom.yml", "--non-interactive"},
			golden: "init_custom_config",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "stagecraft"}
				root.AddCommand(NewInitCommand())
				return root
			},
		},
		{
			name:   "init_help",
			args:   []string{"init", "--help"},
			golden: "init_help",
			setupCmd: func() *cobra.Command {
				root := &cobra.Command{Use: "stagecraft"}
				root.AddCommand(NewInitCommand())
				return root
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			output, err := executeCommandForGolden(cmd, tt.args...)

			// Help commands don't return errors, but other commands might
			if err != nil && !strings.Contains(strings.Join(tt.args, " "), "--help") {
				t.Fatalf("unexpected error: %v", err)
			}

			expected := readGoldenFile(t, tt.golden)

			if *updateGolden {
				writeGoldenFile(t, tt.golden, output)
				expected = output
			}

			if output != expected {
				t.Errorf("output mismatch:\nGot:\n%s\nExpected:\n%s", output, expected)
			}
		})
	}
}
