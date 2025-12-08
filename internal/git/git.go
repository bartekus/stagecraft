// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package git provides a deterministic git adapter for commit history retrieval.
//
// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package git

import (
	"context"
	"fmt"

	"stagecraft/internal/reports/commithealth"
)

// HistorySourceImpl implements HistorySource using git commands.
type HistorySourceImpl struct {
	repoPath string
}

// NewHistorySource creates a new HistorySource that reads from the given repository path.
func NewHistorySource(repoPath string) commithealth.HistorySource {
	return &HistorySourceImpl{
		repoPath: repoPath,
	}
}

// Commits retrieves commit history from git, sorted deterministically.
func (h *HistorySourceImpl) Commits() ([]commithealth.CommitMetadata, error) {
	ctx := context.Background()
	output, err := runGitLog(ctx, h.repoPath)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	commits, err := parseGitLogOutput(output)
	if err != nil {
		return nil, fmt.Errorf("parsing git log output: %w", err)
	}

	return commits, nil
}

// runGitLog executes git log and returns the raw output.
// This function shells out to git with explicit environment variables.
func runGitLog(ctx context.Context, repoPath string) (string, error) {
	// TODO: Implement git log execution
	// - Use exec.CommandContext
	// - Set explicit environment variables (no inheritance)
	// - Format: git log --format="%H|%s|%an|%ae" --reverse
	// - Return raw output string
	return "", fmt.Errorf("not implemented")
}

// parseGitLogOutput parses git log output into CommitMetadata slices.
// This is a pure function that can be tested without shelling out to git.
func parseGitLogOutput(output string) ([]commithealth.CommitMetadata, error) {
	// TODO: Implement parsing logic
	// - Split by newlines
	// - Parse format: SHA|subject|author_name|author_email
	// - Handle edge cases (empty lines, malformed lines)
	// - Sort deterministically (by SHA or commit time as per spec)
	// - Return sorted slice
	return nil, fmt.Errorf("not implemented")
}
