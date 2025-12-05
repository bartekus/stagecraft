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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/config"
	"stagecraft/pkg/executil"
	"stagecraft/pkg/logging"
	backendproviders "stagecraft/pkg/providers/backend"
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
		// Generate plan to show what would be deployed
		planner := core.NewPlanner(cfg)
		plan, err := planner.PlanDeploy(flags.Env)
		if err != nil {
			return fmt.Errorf("generating deployment plan: %w", err)
		}

		logger.Info("Dry-run mode: would deploy application",
			logging.NewField("env", flags.Env),
			logging.NewField("version", version),
			logging.NewField("commit_sha", commitSHA),
			logging.NewField("config", absPath),
			logging.NewField("operations", len(plan.Operations)),
		)
		// Dry-run does not create or modify state file
		// It only shows what would happen
		return nil
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

	// Store deployment context in plan metadata for phase functions
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["release_id"] = release.ID
	plan.Metadata["version"] = version
	plan.Metadata["config_path"] = absPath
	plan.Metadata["workdir"], _ = os.Getwd()

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

// Phase execution functions

// TODO: Future improvement - consider using a typed DeployContext struct instead of map[string]interface{}
// This would reduce type assertions and make refactors safer. Example:
//   type DeployContext struct {
//       ReleaseID  string
//       Version    string
//       ConfigPath string
//       WorkDir    string
//   }
// Store under plan.Metadata["deploy_ctx"] as a single key.

// getDeployContext extracts deployment context from plan metadata.
func getDeployContext(plan *core.Plan) (configPath, version, workdir string, err error) {
	if plan.Metadata == nil {
		return "", "", "", fmt.Errorf("plan metadata is missing")
	}

	configPath, _ = plan.Metadata["config_path"].(string)
	version, _ = plan.Metadata["version"].(string)
	workdir, _ = plan.Metadata["workdir"].(string)

	if configPath == "" {
		return "", "", "", fmt.Errorf("config_path not found in plan metadata")
	}
	if version == "" {
		return "", "", "", fmt.Errorf("version not found in plan metadata")
	}
	if workdir == "" {
		workdir, _ = os.Getwd()
	}

	return configPath, version, workdir, nil
}

// executeBuildPhase builds Docker images using the configured backend provider.
func executeBuildPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	configPath, version, workdir, err := getDeployContext(plan)
	if err != nil {
		return fmt.Errorf("getting deployment context: %w", err)
	}

	// TODO: Future optimization - pass config through plan metadata to avoid reloading
	// Config is already loaded in runDeployWithPhases, but reloading here keeps phases independent.
	// Consider storing *config.Config in plan.Metadata["config"] when adding more context.
	// Load config
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if cfg.Backend == nil {
		return fmt.Errorf("no backend configuration found")
	}

	// Get backend provider
	providerID := cfg.Backend.Provider
	provider, err := backendproviders.Get(providerID)
	if err != nil {
		return fmt.Errorf("getting backend provider %q: %w", providerID, err)
	}

	// Get provider config
	providerCfg, err := cfg.Backend.GetProviderConfig()
	if err != nil {
		return fmt.Errorf("getting provider config: %w", err)
	}

	// Construct image tag
	// For v1, use simple format: <project-name>:<version>
	// If registry is configured, prepend it
	imageTag := fmt.Sprintf("%s:%s", cfg.Project.Name, version)
	// TODO: Add registry support when project.registry is added to config

	logger.Info("Building Docker image",
		logging.NewField("provider", providerID),
		logging.NewField("image", imageTag),
		logging.NewField("workdir", workdir),
	)

	// Build image
	opts := backendproviders.BuildDockerOptions{
		Config:   providerCfg,
		ImageTag: imageTag,
		WorkDir:  workdir,
	}

	builtImage, err := provider.BuildDocker(ctx, opts)
	if err != nil {
		return fmt.Errorf("building Docker image: %w", err)
	}

	logger.Info("Docker image built successfully",
		logging.NewField("image", builtImage),
	)

	// Store built image tag in plan metadata for push phase
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["built_image"] = builtImage

	return nil
}

// executePushPhase pushes the built Docker image to the registry.
func executePushPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	_, _, _, err := getDeployContext(plan)
	if err != nil {
		return fmt.Errorf("getting deployment context: %w", err)
	}

	// Get built image from plan metadata
	if plan.Metadata == nil {
		return fmt.Errorf("plan metadata is missing")
	}

	builtImage, ok := plan.Metadata["built_image"].(string)
	if !ok || builtImage == "" {
		return fmt.Errorf("built image not found in plan metadata (build phase may have failed)")
	}

	logger.Info("Pushing Docker image",
		logging.NewField("image", builtImage),
	)

	// Push image using docker CLI
	runner := executil.NewRunner()
	cmd := executil.NewCommand("docker", "push", builtImage)
	result, err := runner.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("pushing image %q: %w", builtImage, err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("docker push failed with exit code %d: %s", result.ExitCode, string(result.Stderr))
	}

	logger.Info("Docker image pushed successfully",
		logging.NewField("image", builtImage),
	)

	return nil
}

// executeMigratePrePhase is a placeholder for pre-deployment migrations.
// In v1, this is a no-op. Future implementation will integrate with migration engines.
func executeMigratePrePhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("MigratePre phase: placeholder (no-op in v1)")
	// TODO: Integrate with MIGRATION_ENGINE_RAW when MIGRATION_PRE_DEPLOY is implemented
	return nil
}

// executeRolloutPhase deploys the application using Docker Compose.
func executeRolloutPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	_, _, workdir, err := getDeployContext(plan)
	if err != nil {
		return fmt.Errorf("getting deployment context: %w", err)
	}

	// Get built image from plan metadata
	if plan.Metadata == nil {
		return fmt.Errorf("plan metadata is missing")
	}

	builtImage, ok := plan.Metadata["built_image"].(string)
	if !ok || builtImage == "" {
		return fmt.Errorf("built image not found in plan metadata (build phase may have failed)")
	}

	logger.Info("Rolling out deployment",
		logging.NewField("environment", plan.Environment),
		logging.NewField("image", builtImage),
	)

	// For v1 MVP: use docker compose up -d
	// Check if docker-compose.yml exists
	composePath := filepath.Join(workdir, "docker-compose.yml")
	if _, err := os.Stat(composePath); err != nil {
		return fmt.Errorf("docker-compose.yml not found at %s: %w", composePath, err)
	}

	// Update image tag in compose file or use override
	// For v1, we'll use docker compose up with image override via environment variable
	// or generate a minimal override file
	// For now, simplest approach: docker compose up -d (assumes image is already set in compose)
	// Future: DEPLOY_COMPOSE_GEN will handle proper image tag injection

	runner := executil.NewRunner()
	cmd := executil.NewCommand("docker", "compose", "-f", composePath, "up", "-d")
	cmd.Env = map[string]string{
		"IMAGE_TAG": builtImage,
	}
	// TODO: Future test - add test that verifies IMAGE_TAG env var propagation
	// This ensures the environment variable is correctly passed to docker compose.
	// Can wait until DEPLOY_COMPOSE_GEN lands for full integration test.
	result, err := runner.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("running docker compose up: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("docker compose up failed with exit code %d: %s", result.ExitCode, string(result.Stderr))
	}

	logger.Info("Deployment rolled out successfully",
		logging.NewField("environment", plan.Environment),
	)

	// Store compose path in metadata for potential cleanup/rollback
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["compose_path"] = composePath

	return nil
}

// executeMigratePostPhase is a placeholder for post-deployment migrations.
// In v1, this is a no-op. Future implementation will integrate with migration engines.
func executeMigratePostPhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Debug("MigratePost phase: placeholder (no-op in v1)")
	// TODO: Integrate with MIGRATION_ENGINE_RAW when MIGRATION_POST_DEPLOY is implemented
	return nil
}

// executeFinalizePhase performs final bookkeeping for the deployment.
// The phase status is already updated by executePhasesCommon, so this is mainly for logging.
func executeFinalizePhase(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
	logger.Info("Finalizing deployment",
		logging.NewField("environment", plan.Environment),
	)
	// Phase status update is handled by executePhasesCommon
	// Future: could mark release as current for environment here
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
