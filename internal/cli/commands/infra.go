// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package commands contains Cobra subcommands for the Stagecraft CLI.
package commands

import (
	"github.com/spf13/cobra"
)

// Feature: CLI_INFRA_UP
// Spec: spec/commands/infra-up.md

// NewInfraCommand returns the `stagecraft infra` command group.
func NewInfraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "infra",
		Short: "Infrastructure management commands",
		Long:  "Commands for provisioning and managing infrastructure for deployment environments",
	}

	cmd.AddCommand(NewInfraUpCommand())

	return cmd
}
