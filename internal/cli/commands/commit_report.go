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
	"strings"

	"github.com/spf13/cobra"

	"stagecraft/internal/git"
	"stagecraft/internal/reports"
	"stagecraft/internal/reports/commithealth"
)

// Feature: PROVIDER_FRONTEND_GENERIC
// Docs: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md

var newHistorySource = git.NewHistorySource

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
	historySource := newHistorySource(repoPath)
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

// loadFeatureRegistry reads spec/features.yaml under rootDir and returns a simple
// registry of known Feature IDs. It only needs to support the minimal YAML shape
// used in tests and spec/features.yaml (lines containing "id: <FEATURE_ID>").
func loadFeatureRegistry(rootDir string) (map[string]bool, error) {
	path := filepath.Join(rootDir, "spec", "features.yaml")

	data, err := os.ReadFile(path) //nolint:gosec // G304: path is constructed from repo root + spec/features.yaml
	if err != nil {
		return nil, fmt.Errorf("loading feature registry: %w", err)
	}

	registry := make(map[string]bool)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Handle `- id: CLI_DEPLOY` and `id: CLI_DEPLOY`
		if strings.HasPrefix(trimmed, "-") {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		}

		if strings.HasPrefix(trimmed, "id:") {
			id := strings.TrimSpace(strings.TrimPrefix(trimmed, "id:"))
			if id != "" {
				registry[id] = true
			}
		}
	}

	return registry, nil
}

// determineRepoName determines the repository name from the path.
func determineRepoName(repoPath string) string {
	// TODO: Implement repo name detection
	// - Use git config or directory name
	// - Return repository name
	return "stagecraft"
}
