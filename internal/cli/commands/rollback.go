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

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
)

// Feature: CLI_ROLLBACK
// Spec: spec/commands/rollback.md

// PhaseFns holds the phase execution functions for dependency injection.
// This allows tests to override phase behavior without mutating global state.
type PhaseFns struct {
	Build       func(context.Context, *core.Plan, logging.Logger) error
	Push        func(context.Context, *core.Plan, logging.Logger) error
	MigratePre  func(context.Context, *core.Plan, logging.Logger) error
	Rollout     func(context.Context, *core.Plan, logging.Logger) error
	MigratePost func(context.Context, *core.Plan, logging.Logger) error
	Finalize    func(context.Context, *core.Plan, logging.Logger) error
}

// defaultPhaseFns provides the default phase execution functions.
var defaultPhaseFns = PhaseFns{
	Build:       buildPhaseFn,
	Push:        pushPhaseFn,
	MigratePre:  migratePrePhaseFn,
	Rollout:     rolloutPhaseFn,
	MigratePost: migratePostPhaseFn,
	Finalize:    finalizePhaseFn,
}

// NewRollbackCommand returns the `stagecraft rollback` command.
func NewRollbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback environment to a previous release",
		Long:  "Rolls back an environment to a previous release by creating a new deployment with the target release's version",
		RunE:  runRollback,
	}

	cmd.Flags().Bool("to-previous", false, "Rollback to immediately previous release")
	cmd.Flags().String("to-release", "", "Rollback to specific release ID")
	cmd.Flags().String("to-version", "", "Rollback to most recent release with matching version")

	// Global flags (--config, --env, --verbose, --dry-run) are inherited from root

	return cmd
}

// rollbackFlags contains the parsed rollback target flags.
type rollbackFlags struct {
	ToPrevious bool
	ToRelease  string
	ToVersion  string
}

// parseRollbackFlags parses and validates rollback target flags.
func parseRollbackFlags(cmd *cobra.Command) (rollbackFlags, error) {
	toPrevious, _ := cmd.Flags().GetBool("to-previous")
	toRelease, _ := cmd.Flags().GetString("to-release")
	toVersion, _ := cmd.Flags().GetString("to-version")

	count := 0
	if toPrevious {
		count++
	}
	if toRelease != "" {
		count++
	}
	if toVersion != "" {
		count++
	}

	if count == 0 {
		return rollbackFlags{}, fmt.Errorf("rollback target required; use --to-previous, --to-release, or --to-version")
	}

	if count > 1 {
		return rollbackFlags{}, fmt.Errorf("only one rollback target flag may be specified")
	}

	return rollbackFlags{
		ToPrevious: toPrevious,
		ToRelease:  toRelease,
		ToVersion:  toVersion,
	}, nil
}

// resolveRollbackTarget resolves the rollback target based on flags.
// current may be nil for --to-release and --to-version, but must be set for --to-previous.
func resolveRollbackTarget(ctx context.Context, stateMgr *state.Manager, env string, current *state.Release, flags rollbackFlags) (*state.Release, error) {
	// Determine which flag was set
	if flags.ToPrevious {
		if current == nil {
			return nil, fmt.Errorf("no current release found for environment %q", env)
		}
		if current.PreviousID == "" {
			return nil, fmt.Errorf("no previous release to rollback to")
		}
		target, err := stateMgr.GetRelease(ctx, current.PreviousID)
		if err != nil {
			return nil, fmt.Errorf("rollback target not found: %q", current.PreviousID)
		}
		return target, nil
	}

	if flags.ToRelease != "" {
		target, err := stateMgr.GetRelease(ctx, flags.ToRelease)
		if err != nil {
			return nil, fmt.Errorf("rollback target not found: %q", flags.ToRelease)
		}
		// Validate environment match
		if target.Environment != env {
			return nil, fmt.Errorf("release %q belongs to environment %q, not %q", flags.ToRelease, target.Environment, env)
		}
		return target, nil
	}

	if flags.ToVersion != "" {
		releases, err := stateMgr.ListReleases(ctx, env)
		if err != nil {
			return nil, fmt.Errorf("listing releases: %w", err)
		}
		// Find most recent matching version (already sorted newest first)
		for _, r := range releases {
			if r.Version == flags.ToVersion {
				return r, nil
			}
		}
		return nil, fmt.Errorf("no release found with version %q in environment %q", flags.ToVersion, env)
	}

	// This should not happen if parseRollbackFlags was called correctly
	return nil, fmt.Errorf("rollback target required; use --to-previous, --to-release, or --to-version")
}

// allPhasesRollback returns all deployment phases in execution order.
// Shared by validateRollbackTarget and orderedPhasesRollback to avoid duplication.
func allPhasesRollback() []state.ReleasePhase {
	return []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
}

// validateRollbackTarget validates that the target release is eligible for rollback.
// current may be nil if no current release exists.
func validateRollbackTarget(current, target *state.Release) error {
	// Cannot rollback to current (only check if current exists)
	if current != nil && current.ID == target.ID {
		return fmt.Errorf("cannot rollback to current release %q", target.ID)
	}

	// Must be fully deployed
	requiredPhases := allPhasesRollback()

	incompletePhases := []string{}
	for _, phase := range requiredPhases {
		if target.Phases[phase] != state.StatusCompleted {
			incompletePhases = append(incompletePhases, string(phase))
		}
	}

	if len(incompletePhases) > 0 {
		return fmt.Errorf("rollback target %q is not fully deployed (phases: %v)", target.ID, incompletePhases)
	}

	return nil
}

// runRollbackWithPhases is the internal implementation that accepts PhaseFns for dependency injection.
// This allows tests to inject custom phase functions without using global state.
func runRollbackWithPhases(cmd *cobra.Command, args []string, fns PhaseFns) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Resolve flags
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Load config
	cfg, err := config.Load(flags.Config)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Re-resolve flags with config
	flags, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Validate environment
	if flags.Env == "" {
		return fmt.Errorf("environment is required; use --env flag")
	}

	// Parse rollback flags first (before getting current release)
	rollbackFlags, err := parseRollbackFlags(cmd)
	if err != nil {
		return err
	}

	// Initialize state manager
	stateMgr := state.NewDefaultManager()

	// Get current release only if needed for --to-previous or validation
	// For --to-release and --to-version, we can resolve target first
	var current *state.Release
	if rollbackFlags.ToPrevious {
		// --to-previous requires current release
		var err error
		current, err = stateMgr.GetCurrentRelease(ctx, flags.Env)
		if err != nil {
			return fmt.Errorf("no current release found for environment %q", flags.Env)
		}
	}

	// Resolve rollback target
	// For --to-previous, current is already set. For others, we pass nil and resolve without it.
	target, err := resolveRollbackTarget(ctx, stateMgr, flags.Env, current, rollbackFlags)
	if err != nil {
		return err
	}

	// Get current release for validation if we don't have it yet
	// Note: --to-release/--to-version can succeed without a current release.
	// In that case we skip the "cannot rollback to current" check.
	if current == nil {
		var err error
		current, err = stateMgr.GetCurrentRelease(ctx, flags.Env)
		if err != nil {
			// If no current release exists, skip "cannot rollback to current" validation
			// but still validate that target is fully deployed
			current = nil
		}
	}

	// Validate target (current may be nil if no current release exists)
	if err := validateRollbackTarget(current, target); err != nil {
		return err
	}

	// Initialize logger
	logger := logging.NewLogger(flags.Verbose)

	logger.Info("Rolling back environment",
		logging.NewField("env", flags.Env),
		logging.NewField("target_release", target.ID),
		logging.NewField("target_version", target.Version),
	)

	// Handle dry-run (BEFORE creating release)
	if flags.DryRun {
		logger.Info("Dry-run mode: would rollback to release",
			logging.NewField("env", flags.Env),
			logging.NewField("target_release", target.ID),
			logging.NewField("target_version", target.Version),
			logging.NewField("target_commit", target.CommitSHA),
		)
		// Optionally generate plan to show what would happen (but don't execute)
		planner := core.NewPlanner(cfg)
		plan, err := planner.PlanDeploy(flags.Env)
		if err == nil {
			logger.Debug("Would execute deployment plan",
				logging.NewField("operations", len(plan.Operations)),
			)
		}
		// Do NOT create a release or write state in dry-run
		return nil
	}

	// Create new release with target's version/commit SHA (only in non-dry-run)
	release, err := stateMgr.CreateRelease(ctx, flags.Env, target.Version, target.CommitSHA)
	if err != nil {
		return fmt.Errorf("creating rollback release: %w", err)
	}

	logger.Info("Rollback release created",
		logging.NewField("release_id", release.ID),
	)

	// Generate deployment plan
	planner := core.NewPlanner(cfg)
	plan, err := planner.PlanDeploy(flags.Env)
	if err != nil {
		markAllPhasesFailedRollback(ctx, stateMgr, release.ID, logger)
		return fmt.Errorf("generating deployment plan: %w", err)
	}

	// Execute deployment phases using injected PhaseFns
	err = executePhasesRollback(ctx, stateMgr, release.ID, plan, logger, fns)
	if err != nil {
		return fmt.Errorf("rollback deployment failed: %w", err)
	}

	logger.Info("Rollback completed successfully",
		logging.NewField("release_id", release.ID),
	)

	return nil
}

// runRollback is the public entry point that uses default phase functions.
func runRollback(cmd *cobra.Command, args []string) error {
	return runRollbackWithPhases(cmd, args, defaultPhaseFns)
}

// orderedPhasesRollback returns all deployment phases in execution order.
// Copied from deploy.go to avoid premature refactor.
// Uses allPhasesRollback() to share the phase list with validateRollbackTarget.
func orderedPhasesRollback() []state.ReleasePhase {
	return allPhasesRollback()
}

// executePhasesRollback executes all deployment phases in order.
// Copied from deploy.go to avoid premature refactor.
// Takes PhaseFns as a parameter to allow dependency injection for testing.
func executePhasesRollback(ctx context.Context, stateMgr *state.Manager, releaseID string, plan *core.Plan, logger logging.Logger, fns PhaseFns) error {
	phases := []struct {
		phase     state.ReleasePhase
		name      string
		executeFn func(context.Context, *core.Plan, logging.Logger) error
	}{
		{state.PhaseBuild, "build", fns.Build},
		{state.PhasePush, "push", fns.Push},
		{state.PhaseMigratePre, "migrate_pre", fns.MigratePre},
		{state.PhaseRollout, "rollout", fns.Rollout},
		{state.PhaseMigratePost, "migrate_post", fns.MigratePost},
		{state.PhaseFinalize, "finalize", fns.Finalize},
	}

	for _, p := range phases {
		// Update phase to running
		logger.Info("Starting phase", logging.NewField("phase", p.name))
		if err := stateMgr.UpdatePhase(ctx, releaseID, p.phase, state.StatusRunning); err != nil {
			return fmt.Errorf("updating phase %q to running: %w", p.name, err)
		}

		// Execute phase
		err := p.executeFn(ctx, plan, logger)
		if err != nil {
			// Mark current phase as failed
			if updateErr := stateMgr.UpdatePhase(ctx, releaseID, p.phase, state.StatusFailed); updateErr != nil {
				logger.Debug("Failed to update phase status", logging.NewField("error", updateErr.Error()))
			}

			// Mark all downstream phases as skipped
			markDownstreamPhasesSkippedRollback(ctx, stateMgr, releaseID, p.phase, logger)

			return fmt.Errorf("phase %q failed: %w", p.name, err)
		}

		// Mark phase as completed
		if err := stateMgr.UpdatePhase(ctx, releaseID, p.phase, state.StatusCompleted); err != nil {
			return fmt.Errorf("updating phase %q to completed: %w", p.name, err)
		}

		logger.Info("Phase completed", logging.NewField("phase", p.name))
	}

	return nil
}

// markDownstreamPhasesSkippedRollback marks all phases after the failed phase as skipped.
// Copied from deploy.go to avoid premature refactor.
func markDownstreamPhasesSkippedRollback(ctx context.Context, stateMgr *state.Manager, releaseID string, failedPhase state.ReleasePhase, logger logging.Logger) {
	allPhases := orderedPhasesRollback()

	// Find the index of the failed phase
	failedIndex := -1
	for i, phase := range allPhases {
		if phase == failedPhase {
			failedIndex = i
			break
		}
	}

	if failedIndex == -1 {
		logger.Debug("Failed phase not found in phase list", logging.NewField("phase", failedPhase))
		return
	}

	// Mark all downstream phases as skipped
	for i := failedIndex + 1; i < len(allPhases); i++ {
		phase := allPhases[i]
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusSkipped); err != nil {
			logger.Debug("Failed to mark phase as skipped",
				logging.NewField("phase", phase),
				logging.NewField("error", err.Error()),
			)
		}
	}
}

// markAllPhasesFailedRollback marks all phases as failed (used when plan generation fails).
// Copied from deploy.go to avoid premature refactor.
func markAllPhasesFailedRollback(ctx context.Context, stateMgr *state.Manager, releaseID string, logger logging.Logger) {
	for _, phase := range orderedPhasesRollback() {
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); err != nil {
			logger.Debug("Failed to mark phase as failed",
				logging.NewField("phase", phase),
				logging.NewField("error", err.Error()),
			)
		}
	}
}
