// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/internal/git"
	"stagecraft/internal/reports"
	"stagecraft/internal/reports/commithealth"
)

// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md

// NewCommitReportCommand returns the `stagecraft commit report` command.
func NewCommitReportCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit report",
		Short: "Generate commit health report",
		Long:  "Generates a commit health report analyzing commit message discipline",
		RunE:  runCommitReport,
	}

	// Flags in alphabetical order for deterministic help output
	cmd.Flags().String("from", "origin/main", "Start of commit range (default: origin/main)")
	cmd.Flags().String("to", "HEAD", "End of commit range (default: HEAD)")

	return cmd
}

// runCommitReport executes the commit report command.
func runCommitReport(cmd *cobra.Command, args []string) error {
	// 1. Get repository path (current working directory)
	repoPath, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	// 2. Get commit range flags
	fromFlag, _ := cmd.Flags().GetString("from")
	toFlag, _ := cmd.Flags().GetString("to")

	// 3. Load feature registry from spec/features.yaml
	knownFeatures, err := loadFeatureRegistry(repoPath)
	if err != nil {
		return fmt.Errorf("loading feature registry: %w", err)
	}

	// 4. Get commit history via git adapter
	historySource := git.NewHistorySource(repoPath)
	commits, err := historySource.Commits()
	if err != nil {
		return fmt.Errorf("retrieving commit history: %w", err)
	}

	// 5. Determine repo info
	repoInfo := commithealth.RepoInfo{
		Name:          determineRepoName(repoPath),
		DefaultBranch: "main", // TODO: Detect from git config
	}

	// 6. Build commit range info
	rangeInfo := commithealth.CommitRange{
		From:        fromFlag,
		To:          toFlag,
		Description: fmt.Sprintf("%s..%s", fromFlag, toFlag),
	}

	// 7. Generate report using Phase 3.B generator
	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		return fmt.Errorf("generating commit health report: %w", err)
	}

	// 8. Write report atomically
	reportPath := filepath.Join(repoPath, ".stagecraft", "reports", "commit-health.json")
	if err := reports.WriteJSONAtomic(reportPath, report); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}

	return nil
}

// loadFeatureRegistry loads the feature registry from spec/features.yaml.
func loadFeatureRegistry(repoPath string) (map[string]bool, error) {
	// TODO: Implement feature registry loading
	// - Read spec/features.yaml
	// - Parse YAML
	// - Extract feature IDs
	// - Return map[string]bool where key is feature ID
	return nil, fmt.Errorf("not implemented")
}

// determineRepoName determines the repository name from the path.
func determineRepoName(repoPath string) string {
	// TODO: Implement repo name detection
	// - Use git config or directory name
	// - Return repository name
	return "stagecraft"
}
