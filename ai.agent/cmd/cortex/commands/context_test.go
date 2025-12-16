// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepoRoot(t *testing.T) {
	// Test from current directory (should find repo root)
	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		t.Fatalf("FindRepoRoot failed: %v", err)
	}

	// Verify it's actually the repo root by checking for markers
	gitDir := filepath.Join(repoRoot, ".git")
	specDir := filepath.Join(repoRoot, "spec")
	agentFile := filepath.Join(repoRoot, "Agent.md")

	found := false
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		found = true
	}
	if info, err := os.Stat(specDir); err == nil && info.IsDir() {
		found = true
	}
	if info, err := os.Stat(agentFile); err == nil && !info.IsDir() {
		found = true
	}

	if !found {
		t.Errorf("FindRepoRoot returned %q but no markers found", repoRoot)
	}

	// Test from a subdirectory
	subDir := filepath.Join(repoRoot, "internal", "cli", "commands")
	repoRoot2, err := FindRepoRoot(subDir)
	if err != nil {
		t.Fatalf("FindRepoRoot from subdirectory failed: %v", err)
	}

	if repoRoot != repoRoot2 {
		t.Errorf("FindRepoRoot from subdirectory returned %q, expected %q", repoRoot2, repoRoot)
	}
}

func TestFindRepoRoot_NotFound(t *testing.T) {
	// Test from a temp directory (should fail)
	tmpDir := t.TempDir()
	_, err := FindRepoRoot(tmpDir)
	if err == nil {
		t.Error("FindRepoRoot should fail when no repo root is found")
	}
}

func TestNewContextCommand(t *testing.T) {
	cmd := NewContextCommand()
	if cmd.Use != "context" {
		t.Errorf("expected Use to be 'context', got %q", cmd.Use)
	}

	// Verify subcommands exist
	buildCmd, _, err := cmd.Find([]string{"build"})
	if err != nil {
		t.Fatalf("expected to find 'build' subcommand, got error: %v", err)
	}
	if buildCmd.Use != "build" {
		t.Errorf("expected 'build' command Use to be 'build', got %q", buildCmd.Use)
	}

	xrayCmd, _, err := cmd.Find([]string{"xray"})
	if err != nil {
		t.Fatalf("expected to find 'xray' subcommand, got error: %v", err)
	}
	if xrayCmd.Use != "xray" {
		t.Errorf("expected 'xray' command Use to be 'xray', got %q", xrayCmd.Use)
	}

	docsCmd, _, err := cmd.Find([]string{"docs"})
	if err != nil {
		t.Fatalf("expected to find 'docs' subcommand, got error: %v", err)
	}
	if docsCmd.Use != "docs" {
		t.Errorf("expected 'docs' command Use to be 'docs', got %q", docsCmd.Use)
	}
}
