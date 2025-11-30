// internal/cli/commands/dev.go
package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
	backendproviders "stagecraft/pkg/providers/backend"
)

// Feature: CLI_DEV_BASIC
// Spec: spec/commands/dev-basic.md

// NewDevCommand returns the `stagecraft dev` command.
func NewDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start development environment",
		Long:  "Loads stagecraft.yml, resolves backend provider, and runs dev mode",
		RunE:  runDev,
	}

	return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get config path (future: from --config flag)
	configPath := config.DefaultConfigPath()
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	// Load and validate config
	cfg, err := config.Load(configPath)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", configPath)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Backend == nil {
		return fmt.Errorf("no backend configuration found in %s", configPath)
	}

	// Resolve backend provider
	backendID := cfg.Backend.Provider
	provider, err := backendproviders.Get(backendID)
	if err != nil {
		// Error already includes available providers
		return fmt.Errorf("resolving backend provider: %w", err)
	}

	// Get provider-specific config
	providerCfg, err := cfg.Backend.GetProviderConfig()
	if err != nil {
		return fmt.Errorf("getting provider config: %w", err)
	}

	// Determine workdir (project root)
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// Log what we're doing (if verbose)
	verbose, _ := cmd.Flags().GetBool("verbose")
	if verbose {
		fmt.Fprintf(cmd.OutOrStdout(), "Using backend provider: %s\n", backendID)
		fmt.Fprintf(cmd.OutOrStdout(), "Config file: %s\n", absPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Working directory: %s\n", workDir)
	}

	// Call provider
	opts := backendproviders.DevOptions{
		Config:  providerCfg,
		WorkDir: workDir,
		Env:     make(map[string]string), // Future: load from env files
	}

	return provider.Dev(ctx, opts)
}

