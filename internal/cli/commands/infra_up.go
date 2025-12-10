// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commands contains Cobra subcommands for the Stagecraft CLI.
package commands

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"stagecraft/internal/infra/bootstrap"
	"stagecraft/pkg/config"
	cloud "stagecraft/pkg/providers/cloud"
	network "stagecraft/pkg/providers/network"
)

// Feature: CLI_INFRA_UP
// Spec: spec/commands/infra-up.md

// newBootstrapService is a function variable that can be overridden in tests
// to inject a fake bootstrap service.
var newBootstrapService = func(exec bootstrap.CommandExecutor, np network.NetworkProvider) bootstrap.Service {
	return bootstrap.NewService(exec, np)
}

// bootstrapPartialFailureError represents a partial bootstrap failure (exit code 10).
type bootstrapPartialFailureError struct {
	successCount int
	failureCount int
}

func (e *bootstrapPartialFailureError) Error() string {
	return fmt.Sprintf("bootstrap completed with %d success(es) and %d failure(s)", e.successCount, e.failureCount)
}

// bootstrapGlobalFailureError represents a global bootstrap failure (exit code 3).
type bootstrapGlobalFailureError struct {
	msg string
}

func (e *bootstrapGlobalFailureError) Error() string {
	return e.msg
}

// NewInfraUpCommand returns the `stagecraft infra up` command.
func NewInfraUpCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Provision infrastructure for an environment",
		Long:  "Create infrastructure hosts using the configured cloud provider and bootstrap them.",
		RunE:  runInfraUp,
	}

	// No infra-up specific flags in v1; relies on global flags (--config, --env, etc.)
	return cmd
}

// runInfraUp executes the infra up command.
func runInfraUp(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Resolve global flags
	resolvedFlags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("infra up: resolving flags: %w", err)
	}

	// Load config
	cfg, err := config.Load(resolvedFlags.Config)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("infra up: stagecraft config not found at %s", resolvedFlags.Config)
		}
		// maps to exit code 1 (config error)
		return fmt.Errorf("infra up: failed to load config: %w", err)
	}

	// Re-resolve flags with config for environment validation
	resolvedFlags, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("infra up: resolving flags: %w", err)
	}

	// Validate cloud provider configuration
	if cfg.Cloud == nil {
		return fmt.Errorf("infra up: cloud provider is not configured")
	}

	cloudProviderID := cfg.Cloud.Provider
	if cloudProviderID == "" {
		return fmt.Errorf("infra up: cloud.provider is required")
	}

	cloudProvider, err := cloud.Get(cloudProviderID)
	if err != nil {
		// exit code 2 (CloudProvider failure) via error classification in tests
		return fmt.Errorf("infra up: cloud provider %q not found: %w", cloudProviderID, err)
	}

	// Validate network provider configuration
	if cfg.Network == nil {
		return fmt.Errorf("infra up: network provider is not configured")
	}

	networkProviderID := cfg.Network.Provider
	if networkProviderID == "" {
		return fmt.Errorf("infra up: network.provider is required")
	}

	networkProvider, err := network.Get(networkProviderID)
	if err != nil {
		return fmt.Errorf("infra up: network provider %q not found: %w", networkProviderID, err)
	}

	// --- Slice 2: Plan + Apply + Hosts ---

	// Get provider-specific config for the selected cloud provider
	var cloudProviderCfg any
	if cfg.Cloud.Providers != nil {
		cloudProviderCfg = cfg.Cloud.Providers[cloudProviderID]
	}

	// Plan infrastructure
	plan, err := cloudProvider.Plan(ctx, cloud.PlanOptions{
		Config:      cloudProviderCfg,
		Environment: resolvedFlags.Env,
	})
	if err != nil {
		// maps to exit code 2 (CloudProvider failure)
		return fmt.Errorf("infra up: cloud provider plan failed: %w", err)
	}

	// Apply infrastructure changes
	if err := cloudProvider.Apply(ctx, cloud.ApplyOptions{
		Config:      cloudProviderCfg,
		Environment: resolvedFlags.Env,
		Plan:        plan,
	}); err != nil {
		return fmt.Errorf("infra up: cloud provider apply failed: %w", err)
	}

	// Fetch resulting hosts
	providerHosts, err := cloudProvider.Hosts(ctx, cloud.HostsOptions{
		Config:      cloudProviderCfg,
		Environment: resolvedFlags.Env,
	})
	if err != nil {
		return fmt.Errorf("infra up: listing hosts failed: %w", err)
	}

	// Slice 3: map cloud.Host → bootstrap.Host (deterministic order)
	infraHosts := mapCloudHostsToBootstrapHosts(providerHosts)

	// Load bootstrap config from cfg.Infra (if present)
	bootstrapCfg := bootstrap.Config{}
	sshUser := ""
	if cfg.Infra != nil {
		bootstrapCfg.SSHUser = cfg.Infra.Bootstrap.SSHUser
		sshUser = cfg.Infra.Bootstrap.SSHUser
	}

	// Select executor based on config
	// v1 Slice 8: Use SSHExecutor if ssh_user is configured, otherwise NoopExecutor
	var executor bootstrap.CommandExecutor
	if sshUser != "" {
		executor = bootstrap.NewSSHExecutor(sshUser, nil)
	} else {
		executor = &bootstrap.NoopExecutor{}
	}

	// Invoke INFRA_HOST_BOOTSTRAP engine
	// v1 Slice 7: Pass network provider for Tailscale setup
	svc := newBootstrapService(executor, networkProvider)
	bootstrapResult, err := svc.Bootstrap(ctx, infraHosts, bootstrapCfg)
	if err != nil {
		// Global/bootstrap service failure → exit code 3
		return &bootstrapGlobalFailureError{
			msg: fmt.Sprintf("infra up: bootstrap failed: %v", err),
		}
	}

	// Print deterministic per-host results
	printBootstrapResults(bootstrapResult)

	// Determine exit code based on results
	if bootstrapResult.AllSucceeded() {
		// All hosts succeeded → exit code 0
		return nil
	}

	// Some hosts failed → exit code 10
	return &bootstrapPartialFailureError{
		successCount: bootstrapResult.SuccessCount(),
		failureCount: bootstrapResult.FailureCount(),
	}
}

// printBootstrapResults prints deterministic per-host bootstrap results.
// Results are printed in the order they appear in the result (which matches
// the sorted input order from mapCloudHostsToBootstrapHosts).
func printBootstrapResults(result *bootstrap.Result) {
	if len(result.Hosts) == 0 {
		return
	}

	// Print header
	fmt.Fprintf(os.Stdout, "Bootstrap results:\n")

	// Print per-host results
	for _, hr := range result.Hosts {
		status := "✓"
		if !hr.Success {
			status = "✗"
		}

		// Use ID if available, otherwise fall back to Name
		hostID := hr.Host.ID
		if hostID == "" {
			hostID = hr.Host.Name
		}

		if hr.Success {
			fmt.Fprintf(os.Stdout, "  %s %s (%s)\n", status, hostID, hr.Host.Name)
		} else {
			fmt.Fprintf(os.Stdout, "  %s %s (%s): %s\n", status, hostID, hr.Host.Name, hr.Error)
		}
	}

	// Print summary
	successCount := result.SuccessCount()
	failureCount := result.FailureCount()
	totalCount := len(result.Hosts)

	fmt.Fprintf(os.Stdout, "\nSummary: %d/%d hosts bootstrapped successfully", successCount, totalCount)
	if failureCount > 0 {
		fmt.Fprintf(os.Stdout, ", %d failed", failureCount)
	}
	fmt.Fprintf(os.Stdout, "\n")
}

// mapCloudHostsToBootstrapHosts converts provider-specific cloud.Host values
// into the internal bootstrap.Host model and sorts them deterministically.
//
// Sorting by ID is preferred when available; when IDs are empty, fall back to
// lexicographic Name ordering.
func mapCloudHostsToBootstrapHosts(providerHosts []cloud.Host) []bootstrap.Host {
	infraHosts := make([]bootstrap.Host, len(providerHosts))
	for i, h := range providerHosts {
		// Defensive copy of Tags to avoid sharing underlying slices.
		tagsCopy := make([]string, len(h.Tags))
		copy(tagsCopy, h.Tags)

		infraHosts[i] = bootstrap.Host{
			ID:       h.ID,
			Name:     h.Name,
			Role:     h.Role,
			PublicIP: h.PublicIP,
			Tags:     tagsCopy,
		}
	}

	sort.Slice(infraHosts, func(i, j int) bool {
		hi := infraHosts[i]
		hj := infraHosts[j]

		switch {
		case hi.ID != "" && hj.ID != "":
			return hi.ID < hj.ID
		case hi.ID == "" && hj.ID == "":
			return hi.Name < hj.Name
		case hi.ID == "":
			// Treat empty IDs as "greater" than non-empty to keep behavior stable.
			return false
		default: // hj.ID == ""
			return true
		}
	})

	return infraHosts
}
