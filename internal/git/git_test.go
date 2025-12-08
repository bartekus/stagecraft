// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package git

import (
	"testing"

	"stagecraft/internal/reports/commithealth"
)

func TestParseGitLogOutput_SingleCommit(t *testing.T) {
	t.Parallel()

	output := "abc123|feat(CLI_DEPLOY): add rollback support|John Doe|john@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}

	commit := commits[0]
	if commit.SHA != "abc123" {
		t.Errorf("expected SHA=abc123, got %s", commit.SHA)
	}
	if commit.Message != "feat(CLI_DEPLOY): add rollback support" {
		t.Errorf("expected message='feat(CLI_DEPLOY): add rollback support', got %s", commit.Message)
	}
	if commit.AuthorName != "John Doe" {
		t.Errorf("expected author=John Doe, got %s", commit.AuthorName)
	}
	if commit.AuthorEmail != "john@example.com" {
		t.Errorf("expected email=john@example.com, got %s", commit.AuthorEmail)
	}
}

func TestParseGitLogOutput_MultipleCommits(t *testing.T) {
	t.Parallel()

	output := "abc123|feat(CLI_DEPLOY): add rollback support|John Doe|john@example.com\n" +
		"def456|fix(CORE_CONFIG): fix config loading|Jane Smith|jane@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}

	// Verify deterministic ordering (should be sorted)
	// TODO: Verify sort order matches spec (likely by SHA lexicographically)
}

func TestParseGitLogOutput_WeirdButValidMessages(t *testing.T) {
	t.Parallel()

	// Test messages with colons, brackets, multiple lines
	output := "abc123|feat(CLI_DEPLOY): add support for :special: chars|John Doe|john@example.com\n" +
		"def456|fix(CORE): handle [brackets] in messages|Jane Smith|jane@example.com\n"

	commits, err := parseGitLogOutput(output)
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
}

func TestParseGitLogOutput_InvalidLines(t *testing.T) {
	t.Parallel()

	// Test truncated or malformed lines
	output := "abc123|incomplete line\n" +
		"def456|feat(CLI_DEPLOY): valid commit|John Doe|john@example.com\n"

	_, err := parseGitLogOutput(output)
	if err == nil {
		t.Error("expected error for invalid line, got nil")
	}
}

func TestParseGitLogOutput_EmptyOutput(t *testing.T) {
	t.Parallel()

	commits, err := parseGitLogOutput("")
	if err != nil {
		t.Fatalf("parseGitLogOutput failed: %v", err)
	}

	if len(commits) != 0 {
		t.Errorf("expected 0 commits for empty output, got %d", len(commits))
	}
}

func TestHistorySourceImpl_Commits(t *testing.T) {
	t.Parallel()

	// TODO: This test should use a fake/stub implementation
	// For now, we'll skip actual git execution
	// In real implementation, we'll inject a fake runGitLog function

	t.Skip("requires fake git implementation or isolated test repo")
}

// fakeHistorySource is a test helper that implements HistorySource without shelling out.
type fakeHistorySource struct {
	commits []commithealth.CommitMetadata
}

func (f *fakeHistorySource) Commits() ([]commithealth.CommitMetadata, error) {
	return f.commits, nil
}

func TestFakeHistorySource(t *testing.T) {
	t.Parallel()

	fake := &fakeHistorySource{
		commits: []commithealth.CommitMetadata{
			{
				SHA:         "abc123",
				Message:     "feat(CLI_DEPLOY): add rollback support",
				AuthorName:  "John Doe",
				AuthorEmail: "john@example.com",
			},
		},
	}

	commits, err := fake.Commits()
	if err != nil {
		t.Fatalf("fake.Commits() failed: %v", err)
	}

	if len(commits) != 1 {
		t.Fatalf("expected 1 commit, got %d", len(commits))
	}
}
