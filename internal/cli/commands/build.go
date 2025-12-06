// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
)

// Feature: CLI_BUILD
// Spec: spec/commands/build.md

// NewBuildCommand returns the `stagecraft build` command.
func NewBuildCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build application images using the configured backend provider",
		Long:  "Build application images for a given environment using the configured backend provider, without deploying.",
		RunE:  runBuild,
	}

	cmd.Flags().String("version", "", "Explicit image version/tag to use")
	cmd.Flags().Bool("push", false, "Push images to registry after successful build")
	cmd.Flags().String("services", "", "Comma-separated list of services to build")

	// Global flags (--config, --env, --verbose, --dry-run) are inherited from root

	return cmd
}

// runBuild is the public entry point that uses default phase functions.
func runBuild(cmd *cobra.Command, args []string) error {
	return runBuildWithPhases(cmd, args, defaultPhaseFns)
}

// runBuildWithPhases is the internal implementation that accepts PhaseFns for dependency injection.
// This allows tests to inject custom phase functions without using global state.
func runBuildWithPhases(cmd *cobra.Command, _ []string, fns PhaseFns) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Resolve global flags
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Load config to validate environment if needed
	cfg, err := config.Load(flags.Config)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", flags.Config)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	// Re-resolve flags with config for environment validation
	flags, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Validate environment is provided
	if flags.Env == "" {
		return fmt.Errorf("build: --env is required")
	}

	// Validate environment exists
	if _, exists := cfg.Environments[flags.Env]; !exists {
		return fmt.Errorf("invalid environment: %s", flags.Env)
	}

	absPath, err := filepath.Abs(flags.Config)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	// Initialize logger
	logger := logging.NewLogger(flags.Verbose)

	// Parse build-specific flags
	versionFlag, _ := cmd.Flags().GetString("version")
	pushFlag, _ := cmd.Flags().GetBool("push")
	servicesFlag, _ := cmd.Flags().GetString("services")

	// Parse services list if provided
	var services []string
	if servicesFlag != "" {
		services = parseServicesList(servicesFlag)
	}

	// Resolve version (same logic as deploy)
	version, commitSHA := resolveVersion(ctx, versionFlag, logger)

	// Generate deployment plan (we'll filter to build phases)
	planner := core.NewPlanner(cfg)
	plan, err := planner.PlanDeploy(flags.Env)
	if err != nil {
		return fmt.Errorf("build: plan generation failed: %w", err)
	}

	// Store build context in plan metadata for phase functions
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["version"] = version
	plan.Metadata["config_path"] = absPath
	plan.Metadata["workdir"], _ = os.Getwd()
	plan.Metadata["push"] = pushFlag

	// Handle dry-run mode
	if flags.DryRun {
		return renderBuildPlan(cmd, plan, flags.Env, version, services)
	}

	// For build command, we only execute build and optionally push phases
	// We need to create a release for state tracking, but only for build phases
	stateMgr := state.NewDefaultManager()

	// Create a release for build tracking (similar to deploy)
	logger.Info("Creating build release",
		logging.NewField("env", flags.Env),
		logging.NewField("version", version),
		logging.NewField("commit_sha", commitSHA),
	)
	release, err := stateMgr.CreateRelease(ctx, flags.Env, version, commitSHA)
	if err != nil {
		return fmt.Errorf("creating build release: %w", err)
	}

	logger.Info("Build release created",
		logging.NewField("release_id", release.ID),
	)

	plan.Metadata["release_id"] = release.ID

	// Execute only build and push phases (if --push is set)
	buildPhases := []state.ReleasePhase{state.PhaseBuild}
	if pushFlag {
		buildPhases = append(buildPhases, state.PhasePush)
	}

	// Execute build phases using shared helper
	err = executeBuildPhases(ctx, stateMgr, release.ID, plan, logger, fns, buildPhases)
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	logger.Info("Build completed successfully",
		logging.NewField("release_id", release.ID),
	)

	return nil
}

// parseServicesList parses a comma-separated list of services.
func parseServicesList(servicesFlag string) []string {
	if servicesFlag == "" {
		return nil
	}
	parts := strings.Split(servicesFlag, ",")
	services := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			services = append(services, trimmed)
		}
	}
	return services
}

// renderBuildPlan renders the build plan for dry-run mode.
func renderBuildPlan(cmd *cobra.Command, plan *core.Plan, env, version string, services []string) error {
	out := cmd.OutOrStdout()

	// Extract provider from plan operations
	providerID := "unknown"
	for _, op := range plan.Operations {
		if op.Type == core.OpTypeBuild {
			if p, ok := op.Metadata["provider"].(string); ok {
				providerID = p
				break
			}
		}
	}

	// Write human-readable output to command's output stream
	_, _ = fmt.Fprintf(out, "[DRY RUN] Build plan for environment %q:\n", env)
	_, _ = fmt.Fprintf(out, "Version: %s\n", version)
	_, _ = fmt.Fprintf(out, "Provider: %s\n", providerID)

	if len(services) > 0 {
		_, _ = fmt.Fprintf(out, "Services:\n")
		for _, svc := range services {
			_, _ = fmt.Fprintf(out, " - %s\n", svc)
		}
	}

	_, _ = fmt.Fprintf(out, "No images will be built or pushed.\n")

	return nil
}

// executeBuildPhases executes only the specified build-related phases.
func executeBuildPhases(
	ctx context.Context,
	stateMgr *state.Manager,
	releaseID string,
	plan *core.Plan,
	logger logging.Logger,
	fns PhaseFns,
	phases []state.ReleasePhase,
) error {
	for _, phase := range phases {
		phaseName := string(phase)

		// Log phase start
		logger.Info("Starting phase", logging.NewField("phase", phaseName))

		// Set phase status to running
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusRunning); err != nil {
			return fmt.Errorf("updating phase %q to running: %w", phaseName, err)
		}

		// Get phase function
		phaseFn, err := phaseFnFor(phase, fns)
		if err != nil {
			// This should never happen with valid phases, but handle it gracefully
			if updateErr := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); updateErr != nil {
				logger.Debug("Failed to update phase status", logging.NewField("error", updateErr.Error()))
			}
			return fmt.Errorf("getting phase function for %q: %w", phaseName, err)
		}

		// Execute phase
		err = phaseFn(ctx, plan, logger)
		if err != nil {
			// Mark current phase as failed
			if updateErr := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); updateErr != nil {
				logger.Debug("Failed to update phase status", logging.NewField("error", updateErr.Error()))
			}

			// Mark all remaining phases as skipped
			for _, remainingPhase := range phases {
				if remainingPhase == phase {
					// Skip the failed phase itself
					continue
				}
				// Check if this phase comes after the failed one
				failedIndex := -1
				remainingIndex := -1
				for i, p := range phases {
					if p == phase {
						failedIndex = i
					}
					if p == remainingPhase {
						remainingIndex = i
					}
				}
				if remainingIndex > failedIndex {
					if skipErr := stateMgr.UpdatePhase(ctx, releaseID, remainingPhase, state.StatusSkipped); skipErr != nil {
						logger.Debug("Failed to mark phase as skipped",
							logging.NewField("phase", remainingPhase),
							logging.NewField("error", skipErr.Error()),
						)
					}
				}
			}

			return fmt.Errorf("phase %q failed: %w", phaseName, err)
		}

		// Mark phase as completed
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusCompleted); err != nil {
			return fmt.Errorf("updating phase %q to completed: %w", phaseName, err)
		}

		logger.Info("Phase completed", logging.NewField("phase", phaseName))
	}

	return nil
}
