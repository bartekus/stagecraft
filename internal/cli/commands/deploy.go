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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/config"
	"stagecraft/pkg/executil"
	"stagecraft/pkg/logging"
)

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

// NewDeployCommand returns the `stagecraft deploy` command.
func NewDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application to environment",
		Long:  "Deploys the application to the specified environment with phase tracking and release history",
		RunE:  runDeploy,
	}

	cmd.Flags().String("version", "", "Version to deploy (defaults to git SHA)")

	// Global flags (--config, --env, --verbose, --dry-run) are inherited from root

	return cmd
}

func runDeploy(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

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
		return fmt.Errorf("environment is required; use --env flag")
	}

	absPath, err := filepath.Abs(flags.Config)
	if err != nil {
		return fmt.Errorf("resolving config path: %w", err)
	}

	// Initialize logger
	logger := logging.NewLogger(flags.Verbose)

	// Resolve version
	versionFlag, _ := cmd.Flags().GetString("version")
	version, commitSHA := resolveVersion(ctx, versionFlag, logger)

	// Check for dry-run mode
	if flags.DryRun {
		logger.Info("Dry-run mode: would deploy application",
			logging.NewField("env", flags.Env),
			logging.NewField("version", version),
			logging.NewField("commit_sha", commitSHA),
			logging.NewField("config", absPath),
		)
		// In dry-run, we still create a release to show what would happen
		// but we don't execute phases
		return createReleaseOnly(ctx, flags.Env, version, commitSHA, logger)
	}

	// Initialize state manager
	stateMgr := state.NewDefaultManager()

	// Create release at deployment start
	logger.Info("Creating release",
		logging.NewField("env", flags.Env),
		logging.NewField("version", version),
		logging.NewField("commit_sha", commitSHA),
	)
	release, err := stateMgr.CreateRelease(ctx, flags.Env, version, commitSHA)
	if err != nil {
		return fmt.Errorf("creating release: %w", err)
	}

	logger.Info("Release created",
		logging.NewField("release_id", release.ID),
	)

	// Generate deployment plan
	planner := core.NewPlanner(cfg)
	plan, err := planner.PlanDeploy(flags.Env)
	if err != nil {
		// Mark all phases as failed if plan generation fails
		markAllPhasesFailed(ctx, stateMgr, release.ID, logger)
		return fmt.Errorf("generating deployment plan: %w", err)
	}

	logger.Debug("Deployment plan generated",
		logging.NewField("operations", len(plan.Operations)),
	)

	// Execute deployment phases
	err = executePhases(ctx, stateMgr, release.ID, plan, logger)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	logger.Info("Deployment completed successfully",
		logging.NewField("release_id", release.ID),
	)

	return nil
}

// resolveVersion resolves the version and commit SHA for deployment.
func resolveVersion(ctx context.Context, versionFlag string, logger logging.Logger) (version, commitSHA string) {
	// If version flag is provided, use it
	if versionFlag != "" {
		// Try to get commit SHA from git
		commitSHA = getGitCommitSHA(ctx, logger)
		return versionFlag, commitSHA
	}

	// Otherwise, try to get from git
	commitSHA = getGitCommitSHA(ctx, logger)
	if commitSHA != "" {
		// Use commit SHA as version
		return commitSHA, commitSHA
	}

	// No git available, use "unknown" as version (state manager requires non-empty version)
	logger.Debug("Git not available, using 'unknown' as version")
	return "unknown", ""
}

// getGitCommitSHA attempts to get the current git commit SHA.
func getGitCommitSHA(ctx context.Context, logger logging.Logger) string {
	runner := executil.NewRunner()
	cmd := executil.NewCommand("git", "rev-parse", "HEAD")
	result, err := runner.Run(ctx, cmd)
	if err != nil {
		logger.Debug("Failed to get git commit SHA", logging.NewField("error", err.Error()))
		return ""
	}

	sha := strings.TrimSpace(string(result.Stdout))
	if sha == "" {
		return ""
	}

	return sha
}

// createReleaseOnly creates a release without executing phases (for dry-run).
func createReleaseOnly(ctx context.Context, env, version, commitSHA string, logger logging.Logger) error {
	stateMgr := state.NewDefaultManager()
	release, err := stateMgr.CreateRelease(ctx, env, version, commitSHA)
	if err != nil {
		return fmt.Errorf("creating release: %w", err)
	}

	logger.Info("Release would be created",
		logging.NewField("release_id", release.ID),
		logging.NewField("env", env),
		logging.NewField("version", version),
	)

	return nil
}

// orderedPhases returns all deployment phases in execution order.
func orderedPhases() []state.ReleasePhase {
	return []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
}

// executePhases executes all deployment phases in order.
func executePhases(ctx context.Context, stateMgr *state.Manager, releaseID string, plan *core.Plan, logger logging.Logger) error {
	phases := []struct {
		phase     state.ReleasePhase
		name      string
		executeFn func(context.Context, *core.Plan, logging.Logger) error
	}{
		{state.PhaseBuild, "build", buildPhaseFn},
		{state.PhasePush, "push", pushPhaseFn},
		{state.PhaseMigratePre, "migrate_pre", migratePrePhaseFn},
		{state.PhaseRollout, "rollout", rolloutPhaseFn},
		{state.PhaseMigratePost, "migrate_post", migratePostPhaseFn},
		{state.PhaseFinalize, "finalize", finalizePhaseFn},
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
			markDownstreamPhasesSkipped(ctx, stateMgr, releaseID, p.phase, logger)

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

// markDownstreamPhasesSkipped marks all phases after the failed phase as skipped.
func markDownstreamPhasesSkipped(ctx context.Context, stateMgr *state.Manager, releaseID string, failedPhase state.ReleasePhase, logger logging.Logger) {
	allPhases := orderedPhases()

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

// markAllPhasesFailed marks all phases as failed (used when plan generation fails).
func markAllPhasesFailed(ctx context.Context, stateMgr *state.Manager, releaseID string, logger logging.Logger) {
	for _, phase := range orderedPhases() {
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); err != nil {
			logger.Debug("Failed to mark phase as failed",
				logging.NewField("phase", phase),
				logging.NewField("error", err.Error()),
			)
		}
	}
}

// Phase execution functions (stubs for now - will be implemented in future iterations)

func executeBuildPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("Build phase: not yet implemented")
	// TODO: Implement build phase
	// For now, this is a no-op to allow tests to pass
	return nil
}

func executePushPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("Push phase: not yet implemented")
	// TODO: Implement push phase
	return nil
}

func executeMigratePrePhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("MigratePre phase: not yet implemented")
	// TODO: Implement pre-deployment migrations
	return nil
}

func executeRolloutPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("Rollout phase: not yet implemented")
	// TODO: Implement rollout phase
	return nil
}

func executeMigratePostPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("MigratePost phase: not yet implemented")
	// TODO: Implement post-deployment migrations
	return nil
}

func executeFinalizePhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("Finalize phase: not yet implemented")
	// TODO: Implement finalize phase
	return nil
}

// Injectable phase executors for testing
var (
	buildPhaseFn       = executeBuildPhase
	pushPhaseFn        = executePushPhase
	migratePrePhaseFn  = executeMigratePrePhase
	rolloutPhaseFn     = executeRolloutPhase
	migratePostPhaseFn = executeMigratePostPhase
	finalizePhaseFn    = executeFinalizePhase
)
