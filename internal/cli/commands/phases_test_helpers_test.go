// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"github.com/spf13/cobra"
)

// newTestRootCommand is defined in init_test.go but accessible to all test files in this package.

// Feature: CLI_PHASE_EXECUTION_COMMON
// Spec: spec/core/phase-execution-common.md

// commandSetupFunc is a function that sets up a command with custom PhaseFns.
type commandSetupFunc func(fns PhaseFns) *cobra.Command

// executeWithPhasesCustom executes a command with custom PhaseFns.
// This allows tests to inject phase behavior without using global state.
//
// Parameters:
//   - setupCommand: A function that creates a cobra.Command with the provided PhaseFns
//   - fns: The PhaseFns to use for phase execution
//   - args: Command-line arguments to pass to the command
//
// Returns:
//   - error: Any error returned by the command execution
func executeWithPhasesCustom(setupCommand commandSetupFunc, fns PhaseFns, args ...string) error {
	root := newTestRootCommand()
	cmd := setupCommand(fns)
	root.AddCommand(cmd)
	root.SetArgs(args)
	return root.Execute()
}

// setupDeployCommand creates a deploy command with custom PhaseFns.
func setupDeployCommand(fns PhaseFns) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy application to environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeployWithPhases(cmd, args, fns)
		},
	}
	cmd.Flags().String("version", "", "Version to deploy (defaults to git SHA)")
	return cmd
}

// setupRollbackCommand creates a rollback command with custom PhaseFns.
func setupRollbackCommand(fns PhaseFns) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback environment to a previous release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRollbackWithPhases(cmd, args, fns)
		},
	}
	cmd.Flags().Bool("to-previous", false, "Rollback to immediately previous release")
	cmd.Flags().String("to-release", "", "Rollback to specific release ID")
	cmd.Flags().String("to-version", "", "Rollback to most recent release with matching version")
	return cmd
}
