// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package core

import (
	"context"
)

// Feature: CLI_BUILD
// Spec: spec/commands/build.md

// BuildOptions defines options for executing build phases.
//
// Feature: CLI_BUILD
// Spec: spec/commands/build.md
type BuildOptions struct {
	Env      string
	Version  string
	Push     bool
	DryRun   bool
	Services []string
}

// ExecuteBuild executes build-related phases for a given environment.
//
// This function is intended to be shared between CLI_DEPLOY (build step) and CLI_BUILD.
//
// Behaviour is defined in spec/commands/build.md.
//
// NOTE: Currently, the build execution logic is implemented directly in the CLI command
// (internal/cli/commands/build.go). This function is a placeholder for future refactoring
// to extract shared build semantics. For now, CLI_BUILD uses executeBuildPhases in the
// commands package.
func ExecuteBuild(ctx context.Context, opts BuildOptions) error {
	// TODO: Implement shared build execution logic
	// This will extract the build phase execution from CLI_BUILD and make it reusable
	// by CLI_DEPLOY's build phase.
	//
	// For now, the build command (internal/cli/commands/build.go) implements the
	// build logic directly using executeBuildPhases.
	_ = ctx
	_ = opts
	return nil
}
