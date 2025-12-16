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
	"os/exec"
	"path/filepath"

	"stagecraft/ai.agent/cortex/projectroot"

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
	cmd.AddCommand(NewContextDocsCommand())

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
		Long:  "Runs XRAY scan (always against repository root) to analyze repository structure and dependencies",
		RunE:  runContextXray,
		Args:  cobra.NoArgs,
	}
}

// NewContextDocsCommand returns the `stagecraft context docs` command.
func NewContextDocsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "docs",
		Short: "Generate AI-Agent documentation",
		Long:  "Generates human-readable documentation from AI-Agent outputs (chunks.ndjson, manifest.json, XRAY index.json)",
		RunE:  runContextDocs,
		Args:  cobra.NoArgs,
	}
}

// runContextBuild executes the context:build npm script.
func runContextBuild(cmd *cobra.Command, _ []string) error {
	repoRoot, err := projectroot.Find(".")
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
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	if len(args) != 0 {
		return fmt.Errorf("xray does not accept a scan target; it always scans the repository root to avoid overwriting .ai-context/xray outputs")
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] running XRAY scan...\n")

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Always scan the repository root.
	// XRAY writes to a stable location under .ai-context/, so scanning subdirectories
	// would overwrite the primary scan output.
	scanTarget := repoRoot

	// Run XRAY from tools/context-compiler, but scan the chosen target.
	// `npm run <script> -- <args>` forwards args to the underlying command.
	// #nosec G204 - scanTarget is repoRoot derived from FindRepoRoot and not user-controlled
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

// runContextDocs executes the context:docs npm script.
func runContextDocs(cmd *cobra.Command, _ []string) error {
	repoRoot, err := projectroot.Find(".")
	if err != nil {
		return fmt.Errorf("finding repo root: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] generating AI-Agent docs...\n")

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	npmCmd := exec.CommandContext(ctx, "npm", "run", "context:docs")
	npmCmd.Dir = filepath.Join(repoRoot, "tools", "context-compiler")
	npmCmd.Stdout = cmd.OutOrStdout()
	npmCmd.Stderr = cmd.ErrOrStderr()

	if err := npmCmd.Run(); err != nil {
		return fmt.Errorf("context docs generation failed: %w", err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "[stagecraft] AI-Agent docs ready → docs/generated/ai-agent/\n")

	return nil
}
