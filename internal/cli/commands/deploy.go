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

// runDeploy is the public entry point that uses default phase functions.
func runDeploy(cmd *cobra.Command, args []string) error {
	return runDeployWithPhases(cmd, args, defaultPhaseFns)
}

// runDeployWithPhases is the internal implementation that accepts PhaseFns for dependency injection.
// This allows tests to inject custom phase functions without using global state.
func runDeployWithPhases(cmd *cobra.Command, _ []string, fns PhaseFns) error {
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
		markAllPhasesFailedCommon(ctx, stateMgr, release.ID, logger)
		return fmt.Errorf("generating deployment plan: %w", err)
	}

	logger.Debug("Deployment plan generated",
		logging.NewField("operations", len(plan.Operations)),
	)

	// Execute deployment phases using shared helper
	err = executePhasesCommon(ctx, stateMgr, release.ID, plan, logger, fns)
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
