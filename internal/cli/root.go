// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package cli wires together the Stagecraft root Cobra command and global CLI options.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stagecraft/internal/cli/commands"
	// "stagecraft/spec" // optional; see note below
	// "github.com/bartekus/stagecraft/internal/cli/commands"
	// "github.com/bartekus/stagecraft/spec" // optional; see note below
)

// NOTE: The import above for github.com/bartekus/stagecraft/spec is optional.
// You can remove it for now or replace it with whatever you prefer.
// It’s shown here mainly to reinforce the idea that the CLI is aligned with spec/.

// NewRootCommand constructs the Stagecraft root Cobra command.
// This command wires subcommands like `init`, `plan`, `deploy`, etc.
//
// Feature: ARCH_OVERVIEW
// Spec: spec/overview.md
func NewRootCommand() *cobra.Command {
	version := os.Getenv("STAGECRAFT_VERSION")
	if version == "" {
		version = "0.0.0-dev"
	}

	cmd := &cobra.Command{
		Use:           "stagecraft",
		Short:         "Stagecraft – deployment and infrastructure orchestration CLI",
		Long:          "Stagecraft is a Go-based CLI that orchestrates application deployment and infrastructure workflows.",
		SilenceUsage:  true, // don't dump usage on user errors
		SilenceErrors: true, // centralize error printing in main()
	}

	// Global flags - registered in lexicographic order for deterministic help output
	cmd.PersistentFlags().StringP("config", "c", "", "path to stagecraft.yml")
	cmd.PersistentFlags().Bool("dry-run", false, "show actions without executing")
	cmd.PersistentFlags().StringP("env", "e", "", "target environment")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")

	// Version command – simple and explicit.
	cmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of Stagecraft",
		Run: func(cmd *cobra.Command, args []string) {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Stagecraft version %s\n", version)
		},
	})

	// Subcommands - keep registrations in lexicographic order by .Use
	// to ensure deterministic help output (see Agent.md determinism rules).
	cmd.AddCommand(commands.NewBuildCommand())
	cmd.AddCommand(commands.NewCommitReportCommand())
	cmd.AddCommand(commands.NewCommitSuggestCommand())
	cmd.AddCommand(commands.NewDeployCommand())
	cmd.AddCommand(commands.NewDevCommand())
	cmd.AddCommand(commands.NewFeatureTraceabilityCommand())
	cmd.AddCommand(commands.NewInitCommand())
	cmd.AddCommand(commands.NewMigrateCommand())
	cmd.AddCommand(commands.NewPlanCommand())
	cmd.AddCommand(commands.NewReleasesCommand())
	cmd.AddCommand(commands.NewRollbackCommand())

	return cmd
}
