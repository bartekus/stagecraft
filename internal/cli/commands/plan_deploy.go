// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/plan"
	"stagecraft/pkg/config"
)

// NewPlanDeployCommand returns the `stagecraft plan deploy` command.
func NewPlanDeployCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Generate an engine.Plan for deployment",
		Long:  "Generates a deterministic engine.Plan from the core planner and outputs it as JSON",
		RunE:  runPlanDeploy,
	}

	cmd.Flags().StringP("env", "e", "", "Target environment (required)")
	cmd.Flags().String("json", "", "Output path for JSON plan (default: stdout)")
	_ = cmd.MarkFlagRequired("env")

	return cmd
}

func runPlanDeploy(cmd *cobra.Command, args []string) error {
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	cfg, err := config.Load(flags.Config)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", flags.Config)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	_, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	envFlag, _ := cmd.Flags().GetString("env")
	if envFlag == "" {
		return fmt.Errorf("environment is required; use --env flag")
	}

	// Generate core plan
	planner := core.NewPlanner(cfg)
	corePlan, err := planner.PlanDeploy(envFlag)
	if err != nil {
		return fmt.Errorf("generating deployment plan: %w", err)
	}

	// Convert to engine plan
	enginePlan, err := plan.ToEnginePlan(corePlan, envFlag)
	if err != nil {
		return fmt.Errorf("converting to engine plan: %w", err)
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(enginePlan, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling plan: %w", err)
	}

	// Output
	jsonPath, _ := cmd.Flags().GetString("json")
	if jsonPath != "" {
		if err := os.WriteFile(jsonPath, jsonBytes, 0o644); err != nil {
			return fmt.Errorf("writing plan to %s: %w", jsonPath, err)
		}
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Plan written to %s\n", jsonPath)
	} else {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(jsonBytes))
	}

	return nil
}
