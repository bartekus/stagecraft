// SPDX-License-Identifier: AGPL-3.0-or-later

// Package commands contains the CLI command constructors.
package commands

import (
	"context"
	"fmt"

	dev "stagecraft/internal/dev"
	devcompose "stagecraft/internal/dev/compose"
	devprocess "stagecraft/internal/dev/process"

	"github.com/spf13/cobra"

	"stagecraft/pkg/config"
)

// Feature: CLI_DEV
// Spec: spec/commands/dev.md

const (
	devFlagEnv       = "env"
	devFlagConfig    = "config"
	devFlagNoHTTPS   = "no-https"
	devFlagNoHosts   = "no-hosts"
	devFlagNoTraefik = "no-traefik"
	devFlagDetach    = "detach"
	devFlagVerbose   = "verbose"
)

// NewDevCommand returns the `stagecraft dev` command.
//
// v1 skeleton responsibilities:
//   - Define flags and help text
//   - Validate basic flag semantics
//   - Delegate real work to a topology runner in a later slice
func NewDevCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Run a complete local dev stack (backend, frontend, infra)",
		Long: `Run a complete local development environment for Stagecraft projects.

This command orchestrates backend and frontend services plus development
infrastructure such as Traefik, mkcert, and hosts file management. For v1,
the implementation is added in incremental slices; this command currently
validates flags and prepares the execution context.`,
		RunE: runDevCommand,
	}

	// Flags must stay lexicographically sorted by flag name.
	cmd.Flags().String(devFlagEnv, "dev", "Environment name to use")
	cmd.Flags().String(devFlagConfig, "", "Path to the Stagecraft config file (optional)")
	cmd.Flags().Bool(devFlagNoHTTPS, false, "Disable mkcert and HTTPS integration")
	cmd.Flags().Bool(devFlagNoHosts, false, "Do not modify /etc/hosts (manual DNS management)")
	cmd.Flags().Bool(devFlagNoTraefik, false, "Skip Traefik setup (providers must expose ports directly)")
	cmd.Flags().Bool(devFlagDetach, false, "Run dev stack in the background and return immediately")
	cmd.Flags().Bool(devFlagVerbose, false, "Enable verbose output for debugging")

	return cmd
}

// devOptions holds parsed flag values for stagecraft dev.
//
// This type is intentionally small and focused on CLI flags only.
// Topology and provider details live in internal/dev.
type devOptions struct {
	Env       string
	Config    string
	NoHTTPS   bool
	NoHosts   bool
	NoTraefik bool
	Detach    bool
	Verbose   bool
}

// runDevCommand is the Cobra entry point. It parses flags and delegates
// to runDevWithOptions, which contains the implementation logic.
func runDevCommand(cmd *cobra.Command, _ []string) error {
	env, err := cmd.Flags().GetString(devFlagEnv)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagEnv, err)
	}

	configPath, err := cmd.Flags().GetString(devFlagConfig)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagConfig, err)
	}

	noHTTPS, err := cmd.Flags().GetBool(devFlagNoHTTPS)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagNoHTTPS, err)
	}

	noHosts, err := cmd.Flags().GetBool(devFlagNoHosts)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagNoHosts, err)
	}

	noTraefik, err := cmd.Flags().GetBool(devFlagNoTraefik)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagNoTraefik, err)
	}

	detach, err := cmd.Flags().GetBool(devFlagDetach)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagDetach, err)
	}

	verbose, err := cmd.Flags().GetBool(devFlagVerbose)
	if err != nil {
		return fmt.Errorf("dev: get %s flag: %w", devFlagVerbose, err)
	}

	opts := devOptions{
		Env:       env,
		Config:    configPath,
		NoHTTPS:   noHTTPS,
		NoHosts:   noHosts,
		NoTraefik: noTraefik,
		Detach:    detach,
		Verbose:   verbose,
	}

	return runDevWithOptions(cmd.Context(), opts)
}

// runDevWithOptions is the core CLI_DEV implementation for v1 slices.
func runDevWithOptions(ctx context.Context, opts devOptions) error {
	if opts.Env == "" {
		return fmt.Errorf("dev: --%s must not be empty", devFlagEnv)
	}

	// 1. Load config
	cfg, err := loadConfigForEnv(opts.Config, opts.Env)
	if err != nil {
		return fmt.Errorf("dev: load config: %w", err)
	}

	// 2. Compute dev domains (placeholder defaults for now)
	domains := dev.Domains{
		Frontend: "app.localdev.test",
		Backend:  "api.localdev.test",
	}

	// 3. Construct minimal service definitions
	backendSvc := &devcompose.ServiceDefinition{
		Name: "backend",
	}
	frontendSvc := &devcompose.ServiceDefinition{
		Name: "frontend",
	}
	traefikSvc := &devcompose.ServiceDefinition{
		Name: "traefik",
	}

	builder := dev.NewBuilder(nil, nil)

	topology, err := builder.Build(
		cfg,
		domains,
		backendSvc,
		frontendSvc,
		traefikSvc,
		!opts.NoHTTPS,
		"", // certDir - to be wired from DEV_MKCERT in a later slice
	)
	if err != nil {
		return fmt.Errorf("dev: build topology: %w", err)
	}

	// 4. Persist dev config files.
	devDir := ".stagecraft/dev" // relative to project root.

	if _, err := dev.WriteFiles(devDir, topology); err != nil {
		return fmt.Errorf("dev: write dev files: %w", err)
	}

	// 5. Start processes via DEV_PROCESS_MGMT.
	procOpts := devprocess.Options{
		DevDir:    devDir,
		NoTraefik: opts.NoTraefik,
		Detach:    opts.Detach,
		Verbose:   opts.Verbose,
	}

	runner := devprocess.NewRunner()

	if err := runner.Run(ctx, procOpts); err != nil {
		return fmt.Errorf("dev: start processes: %w", err)
	}

	return nil
}

// loadConfigForEnv loads the Stagecraft config for the given env.
//
// This is intentionally thin and will be refined as CORE_CONFIG dictates.
func loadConfigForEnv(path, env string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return nil, fmt.Errorf("stagecraft config not found at %s", path)
		}
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Validate that the environment exists
	if cfg.Environments == nil {
		return nil, fmt.Errorf("no environments defined in config")
	}

	if _, ok := cfg.Environments[env]; !ok {
		return nil, fmt.Errorf("environment %q not found in config", env)
	}

	return cfg, nil
}
