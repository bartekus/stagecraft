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
	"path/filepath"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/plan"
	"stagecraft/pkg/config"
	"stagecraft/pkg/engine"
)

// NewPlanSliceCommand returns the `stagecraft plan slice` command.
func NewPlanSliceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "slice",
		Short: "Slice a plan into per-host HostPlans",
		Long:  "Loads an engine.Plan and slices it into HostPlans, optionally saving them to files",
		RunE:  runPlanSlice,
	}

	cmd.Flags().String("plan", "", "Path to plan JSON file (or use --env to generate)")
	cmd.Flags().StringP("env", "e", "", "Environment name (if generating plan)")
	cmd.Flags().String("output-dir", "", "Directory to write host plans (default: stdout)")

	return cmd
}

func runPlanSlice(cmd *cobra.Command, args []string) error {
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	planPath, _ := cmd.Flags().GetString("plan")
	envFlag, _ := cmd.Flags().GetString("env")
	outputDir, _ := cmd.Flags().GetString("output-dir")

	var enginePlan *engine.Plan

	if planPath != "" {
		// Load plan from file
		data, err := os.ReadFile(planPath)
		if err != nil {
			return fmt.Errorf("reading plan file: %w", err)
		}

		enginePlan = &engine.Plan{}
		if err := json.Unmarshal(data, enginePlan); err != nil {
			return fmt.Errorf("unmarshaling plan: %w", err)
		}
	} else if envFlag != "" {
		// Generate plan from environment
		cfg, err := config.Load(flags.Config)
		if err != nil {
			if err == config.ErrConfigNotFound {
				return fmt.Errorf("stagecraft config not found at %s", flags.Config)
			}
			return fmt.Errorf("loading config: %w", err)
		}

		planner := core.NewPlanner(cfg)
		corePlan, err := planner.PlanDeploy(envFlag)
		if err != nil {
			return fmt.Errorf("generating deployment plan: %w", err)
		}

		enginePlan, err = plan.ToEnginePlan(corePlan, envFlag)
		if err != nil {
			return fmt.Errorf("converting to engine plan: %w", err)
		}
	} else {
		return fmt.Errorf("either --plan or --env must be provided")
	}

	// Slice plan
	result, err := engine.SlicePlan(*enginePlan)
	if err != nil {
		return fmt.Errorf("slicing plan: %w", err)
	}

	// Output results
	if outputDir != "" {
		// Write host plans to files
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}

		for hostID, hostPlan := range result.HostPlans {
			filename := filepath.Join(outputDir, fmt.Sprintf("hostplan-%s.json", hostID))
			jsonBytes, err := json.MarshalIndent(hostPlan, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling host plan for %s: %w", hostID, err)
			}

			if err := os.WriteFile(filename, jsonBytes, 0o644); err != nil {
				return fmt.Errorf("writing host plan to %s: %w", filename, err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Host plan for %s written to %s\n", hostID, filename)
		}

		// Write global steps if any
		if len(result.GlobalSteps) > 0 {
			filename := filepath.Join(outputDir, "global-steps.json")
			jsonBytes, err := json.MarshalIndent(result.GlobalSteps, "", "  ")
			if err != nil {
				return fmt.Errorf("marshaling global steps: %w", err)
			}

			if err := os.WriteFile(filename, jsonBytes, 0o644); err != nil {
				return fmt.Errorf("writing global steps: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Global steps written to %s\n", filename)
		}
	} else {
		// Output to stdout as JSON
		output := map[string]interface{}{
			"host_plans":   result.HostPlans,
			"global_steps": result.GlobalSteps,
		}

		jsonBytes, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling result: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(jsonBytes))
	}

	return nil
}
