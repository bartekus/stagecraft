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
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"stagecraft/internal/agent"
	"stagecraft/pkg/engine"
)

// NewAgentCommand returns the `stagecraft agent` command.
func NewAgentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Agent commands for executing HostPlans",
		Long:  "Commands for running HostPlans locally (for testing the CLI → Engine → Agent pipeline)",
	}

	cmd.AddCommand(NewAgentRunCommand())

	return cmd
}

// NewAgentRunCommand returns the `stagecraft agent run` command.
func NewAgentRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Execute a HostPlan",
		Long:  "Loads a HostPlan JSON file and executes it step-by-step with strict input validation",
		RunE:  runAgentRun,
	}

	cmd.Flags().String("hostplan", "", "Path to HostPlan JSON file (required)")
	cmd.Flags().String("output", "", "Path to write execution report JSON (default: stdout)")
	_ = cmd.MarkFlagRequired("hostplan")

	return cmd
}

func runAgentRun(cmd *cobra.Command, args []string) error {
	hostplanPath, _ := cmd.Flags().GetString("hostplan")
	outputPath, _ := cmd.Flags().GetString("output")

	// Load HostPlan with strict validation
	data, err := os.ReadFile(hostplanPath)
	if err != nil {
		return fmt.Errorf("reading host plan file %q: %w", hostplanPath, err)
	}

	var hostPlan engine.HostPlan
	// Try to extract planID from JSON for better error context (best-effort)
	var planID string
	if tempPlan := struct {
		PlanID string `json:"planId"`
	}{}; json.Unmarshal(data, &tempPlan) == nil {
		planID = tempPlan.PlanID
	}

	if err := engine.UnmarshalStrictHostPlan(data, &hostPlan, planID); err != nil {
		// Wrap error with file path context for debugging
		return fmt.Errorf("unmarshaling host plan from %q: %w", hostplanPath, err)
	}

	// Validate HostPlan has non-empty LogicalID (required for HostPlans)
	if hostPlan.Host.LogicalID == "" {
		return fmt.Errorf("host plan from %q has empty host.logicalId (required for HostPlans)", hostplanPath)
	}

	// Create executor with stub executors for all actions
	executor := agent.NewExecutor()
	stubExecutor := &agent.StubExecutor{}

	// Register stub executors for all known actions
	executor.RegisterExecutor(engine.StepActionBuild, stubExecutor)
	executor.RegisterExecutor(engine.StepActionMigrate, stubExecutor)
	executor.RegisterExecutor(engine.StepActionApplyCompose, stubExecutor)
	executor.RegisterExecutor(engine.StepActionHealthCheck, stubExecutor)
	executor.RegisterExecutor(engine.StepActionRenderCompose, stubExecutor)
	executor.RegisterExecutor(engine.StepActionRollout, stubExecutor)

	// Execute plan
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	report, err := executor.ExecuteHostPlan(ctx, hostPlan)
	if err != nil {
		return fmt.Errorf("executing host plan: %w", err)
	}

	// Output report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling execution report: %w", err)
	}

	if outputPath != "" {
		if err := os.WriteFile(outputPath, reportJSON, 0o644); err != nil {
			return fmt.Errorf("writing execution report: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Execution report written to %s\n", outputPath)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(reportJSON))
	}

	return nil
}
