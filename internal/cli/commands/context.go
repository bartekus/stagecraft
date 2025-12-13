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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// NewContextCommand returns the `stagecraft context` command group.
func NewContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "AI context pipeline commands",
		Long:  "Commands for building and managing AI-readable context representations of the repository.",
	}

	cmd.AddCommand(NewContextBuildCommand())
	cmd.AddCommand(NewContextXrayCommand())

	return cmd
}

// NewContextBuildCommand returns the `stagecraft context build` command.
func NewContextBuildCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build AI context representation",
		Long:  "Builds a deterministic AI-readable context representation in .ai-context/",
		RunE:  runContextBuild,
	}
}

// NewContextXrayCommand returns the `stagecraft context xray` command.
func NewContextXrayCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "xray",
		Short: "Run XRAY scan",
		Long:  "Runs XRAY scan to analyze repository structure and dependencies",
		RunE:  runContextXray,
		Args:  cobra.RangeArgs(0, 1),
	}
}

// runContextBuild executes the context:build npm script.
func runContextBuild(cmd *cobra.Command, _ []string) error {
	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] building AI context...\n")

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	npmCmd := exec.CommandContext(ctx, "npm", "run", "context:build")
	npmCmd.Dir = filepath.Join(repoRoot, "tools", "context-compiler")
	npmCmd.Stdout = cmd.OutOrStdout()
	npmCmd.Stderr = cmd.ErrOrStderr()

	if err := npmCmd.Run(); err != nil {
		return fmt.Errorf("context build failed: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] AI context ready → .ai-context/\n")

	return nil
}

// runContextXray executes the xray:scan npm script.
func runContextXray(cmd *cobra.Command, args []string) error {
	repoRoot, err := FindRepoRoot(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] running XRAY scan...\n")

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine scan target
	var scanTarget string
	if len(args) > 0 {
		scanTarget = args[0]
	} else {
		scanTarget = repoRoot
	}
	// Normalize "." and similar to repo root
	if scanTarget == "." || scanTarget == "./." || scanTarget == "./" {
		scanTarget = repoRoot
	}
	// Resolve relative paths to absolute
	scanTargetAbs, err := filepath.Abs(scanTarget)
	if err != nil {
		return fmt.Errorf("resolving scan target: %w", err)
	}

	// Validate that scanTarget is within repoRoot to prevent path traversal
	// FindRepoRoot already returns an absolute path, so we can use it directly
	// Check if scanTarget is within repoRoot using filepath.Rel
	// If the relative path starts with "..", it's outside the repo root
	rel, err := filepath.Rel(repoRoot, scanTargetAbs)
	if err != nil {
		return fmt.Errorf("validating scan target: %w", err)
	}
	if strings.HasPrefix(rel, "..") {
		return fmt.Errorf("scan target %q is outside repository root %q", scanTarget, repoRoot)
	}

	// Use the validated absolute path
	scanTarget = scanTargetAbs

	// Run XRAY from tools/context-compiler, but scan the chosen target.
	// `npm run <script> -- <args>` forwards args to the underlying command.
	// #nosec G204 - scanTarget validated to be within repoRoot above
	npmCmd := exec.CommandContext(ctx, "npm", "run", "xray:scan", "--", scanTarget)
	npmCmd.Dir = filepath.Join(repoRoot, "tools", "context-compiler")
	npmCmd.Stdout = cmd.OutOrStdout()
	npmCmd.Stderr = cmd.ErrOrStderr()

	if err := npmCmd.Run(); err != nil {
		return fmt.Errorf("xray scan failed: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] XRAY scan complete → .ai-context/xray/\n")

	return nil
}

// FindRepoRoot walks upward from the given start directory to find the repository root.
// It stops when it finds one of: .git/, spec/, or Agent.md.
// Returns an error if the filesystem root is reached without finding a marker.
func FindRepoRoot(start string) (string, error) {
	absStart, err := filepath.Abs(start)
	if err != nil {
		return "", fmt.Errorf("resolving start path: %w", err)
	}

	current := absStart
	root := filepath.VolumeName(current) + string(filepath.Separator)
	if root == string(filepath.Separator) {
		root = "/"
	}

	for {
		// Check for repository markers
		gitDir := filepath.Join(current, ".git")
		specDir := filepath.Join(current, "spec")
		agentFile := filepath.Join(current, "Agent.md")

		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return current, nil
		}
		if info, err := os.Stat(specDir); err == nil && info.IsDir() {
			return current, nil
		}
		if info, err := os.Stat(agentFile); err == nil && !info.IsDir() {
			return current, nil
		}

		// Stop if we've reached the filesystem root
		if current == root || current == filepath.Dir(current) {
			return "", fmt.Errorf("repository root not found: no .git/, spec/, or Agent.md found")
		}

		// Move up one directory
		current = filepath.Dir(current)
	}
}
