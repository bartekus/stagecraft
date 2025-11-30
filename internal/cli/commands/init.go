// internal/cli/commands/init.go
package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
)

// Feature: CLI_INIT
// Spec: spec/commands/init.md

// NewInitCommand returns the `stagecraft init` command.
func NewInitCommand() *cobra.Command {
	var nonInteractive bool
	var configPath string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Bootstrap Stagecraft into the current project",
		Long: `Initialize Stagecraft configuration in the current directory.

This command will create a minimal Stagecraft config file and guide you
through initial setup. In future iterations it will support more advanced
provider-specific bootstrapping.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			if configPath == "" {
				configPath = config.DefaultConfigPath()
			}

			exists, err := config.Exists(configPath)
			if err != nil {
				return fmt.Errorf("checking existing config at %s: %w", configPath, err)
			}

			if exists {
				fmt.Fprintf(out, "A Stagecraft config file already exists at %s.\n", configPath)
				return nil
			}

			if nonInteractive {
				fmt.Fprintf(out, "Initializing Stagecraft project (non-interactive, stub) at %s\n", configPath)
			} else {
				fmt.Fprintf(out, "Initializing Stagecraft project (interactive, stub) at %s\n", configPath)
				fmt.Fprintln(out, "NOTE: Interactive questions are not yet implemented.")
			}

			// TODO: create a minimal config.Config value and write it to disk via a helper.

			return nil
		},
	}

	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "run without interactive prompts and use defaults")
	cmd.Flags().StringVar(&configPath, "config", "", "path to Stagecraft config file (default: stagecraft.yml)")

	return cmd
}
