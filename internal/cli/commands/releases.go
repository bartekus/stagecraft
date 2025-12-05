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
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"stagecraft/internal/core/state"
)

// Feature: CLI_RELEASES
// Spec: spec/commands/releases.md

// releasePhasesOrder is the canonical ordered list of all deployment phases.
// Used consistently across displayReleaseShow and calculateOverallStatus.
var releasePhasesOrder = []state.ReleasePhase{
	state.PhaseBuild,
	state.PhasePush,
	state.PhaseMigratePre,
	state.PhaseRollout,
	state.PhaseMigratePost,
	state.PhaseFinalize,
}

// NewReleasesCommand returns the `stagecraft releases` command group.
func NewReleasesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "releases",
		Short: "List and show deployment releases",
		Long:  "View deployment release history and details",
	}

	cmd.AddCommand(NewReleasesListCommand())
	cmd.AddCommand(NewReleasesShowCommand())

	return cmd
}

// NewReleasesListCommand returns `stagecraft releases list`.
func NewReleasesListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deployment releases (optionally filtered by environment)",
		RunE:  runReleasesList,
	}
	// --env flag inherited from root
	return cmd
}

// NewReleasesShowCommand returns `stagecraft releases show`.
func NewReleasesShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <release-id>",
		Short: "Show details of a specific release",
		Args:  cobra.ExactArgs(1),
		RunE:  runReleasesShow,
	}
	return cmd
}

func runReleasesList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Resolve global flags
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// Initialize state manager
	stateMgr := state.NewDefaultManager()

	// Check if --env was explicitly provided
	envFlagSet := cmd.Flags().Changed("env")

	var releases []*state.Release
	var showEnvColumn bool

	if envFlagSet {
		// Explicitly filtered by environment
		releases, err = stateMgr.ListReleases(ctx, flags.Env)
		if err != nil {
			return fmt.Errorf("listing releases for env %q: %w", flags.Env, err)
		}
		// Single-env view; env column only added if multi-env data slips in
		showEnvColumn = false
	} else {
		// No --env provided → list all environments
		releases, err = stateMgr.ListAllReleases(ctx)
		if err != nil {
			return fmt.Errorf("listing releases for all environments: %w", err)
		}
		// Multi-env by design → force environment column on
		showEnvColumn = true
	}

	return displayReleasesList(cmd, releases, showEnvColumn)
}

func runReleasesShow(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	releaseID := args[0]

	// Initialize state manager
	stateMgr := state.NewDefaultManager()

	// Get release
	release, err := stateMgr.GetRelease(ctx, releaseID)
	if err != nil {
		if errors.Is(err, state.ErrReleaseNotFound) {
			return fmt.Errorf("release not found: %q", releaseID)
		}
		return fmt.Errorf("getting release: %w", err)
	}

	// Display release details
	return displayReleaseShow(cmd, release)
}

// displayReleasesList displays releases in table format.
func displayReleasesList(cmd *cobra.Command, releases []*state.Release, showEnv bool) error {
	out := cmd.OutOrStdout()

	if len(releases) == 0 {
		_, _ = fmt.Fprintf(out, "No releases found\n")
		return nil
	}

	// Determine if we need to show environment column
	// If showEnv is false but we have multiple environments, show it anyway
	needsEnvColumn := showEnv
	if !needsEnvColumn {
		envs := make(map[string]bool)
		for _, r := range releases {
			envs[r.Environment] = true
		}
		needsEnvColumn = len(envs) > 1
	}

	// Print header
	if needsEnvColumn {
		_, _ = fmt.Fprintf(out, "%-20s %-12s %-15s %-19s %s\n", "RELEASE ID", "ENVIRONMENT", "VERSION", "TIMESTAMP", "STATUS")
	} else {
		_, _ = fmt.Fprintf(out, "%-20s %-15s %-19s %s\n", "RELEASE ID", "VERSION", "TIMESTAMP", "STATUS")
	}

	// Print releases (sorting is handled by the state manager: ListReleases/ListAllReleases)
	for _, release := range releases {
		status := calculateOverallStatus(release)
		timestamp := formatTimestamp(release.Timestamp)

		if needsEnvColumn {
			_, _ = fmt.Fprintf(out, "%-20s %-12s %-15s %-19s %s\n",
				release.ID, release.Environment, release.Version, timestamp, status)
		} else {
			_, _ = fmt.Fprintf(out, "%-20s %-15s %-19s %s\n",
				release.ID, release.Version, timestamp, status)
		}
	}

	return nil
}

// displayReleaseShow displays detailed release information.
func displayReleaseShow(cmd *cobra.Command, release *state.Release) error {
	out := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(out, "Release ID:        %s\n", release.ID)
	_, _ = fmt.Fprintf(out, "Environment:       %s\n", release.Environment)
	_, _ = fmt.Fprintf(out, "Version:           %s\n", release.Version)

	commitSHA := release.CommitSHA
	if commitSHA == "" {
		commitSHA = "N/A"
	}
	_, _ = fmt.Fprintf(out, "Commit SHA:        %s\n", commitSHA)

	_, _ = fmt.Fprintf(out, "Timestamp:         %s\n", formatTimestamp(release.Timestamp))

	previousID := release.PreviousID
	if previousID == "" {
		previousID = "N/A"
	}
	_, _ = fmt.Fprintf(out, "Previous Release:  %s\n", previousID)

	_, _ = fmt.Fprintf(out, "\nPhases:\n")

	// Display phases in order
	for _, phase := range releasePhasesOrder {
		status := release.Phases[phase]
		if status == "" {
			status = state.StatusPending
		}
		_, _ = fmt.Fprintf(out, "  %-15s %s\n", phase+":", status)
	}

	return nil
}

// calculateOverallStatus calculates the overall status of a release based on phase statuses.
// Iterates over the canonical ordered phase list and treats missing phases as not completed.
func calculateOverallStatus(release *state.Release) string {
	hasFailed := false
	hasRunning := false
	allCompleted := true

	for _, phase := range releasePhasesOrder {
		status := release.Phases[phase]
		switch status {
		case state.StatusFailed:
			hasFailed = true
			allCompleted = false
		case state.StatusRunning:
			hasRunning = true
			allCompleted = false
		case state.StatusCompleted:
			// keep allCompleted as-is
		default:
			// missing or any other value → not completed
			allCompleted = false
		}
	}

	if hasFailed {
		return "failed"
	}
	if allCompleted {
		return "completed"
	}
	if hasRunning {
		return "running"
	}
	return "pending"
}

// formatTimestamp formats a timestamp for display.
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}
