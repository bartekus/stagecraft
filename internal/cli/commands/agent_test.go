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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stagecraft/pkg/engine"
)

func TestRunAgentRun_ErrorIncludesFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	hostplanPath := filepath.Join(tmpDir, "test-hostplan.json")

	// Write invalid HostPlan JSON (unknown field)
	invalidJSON := `{
		"version": "v1",
		"planId": "test-plan",
		"host": {"logicalId": "host-a"},
		"steps": [],
		"unknown_field": "should be rejected"
	}`

	if err := os.WriteFile(hostplanPath, []byte(invalidJSON), 0o644); err != nil {
		t.Fatalf("failed to write test hostplan: %v", err)
	}

	// Create a command with the hostplan flag set
	cmd := NewAgentRunCommand()
	cmd.SetArgs([]string{"--hostplan", hostplanPath})

	// Execute and capture error
	err := cmd.Execute()

	if err == nil {
		t.Fatal("expected error for invalid hostplan")
	}

	// Error should include the file path
	errStr := err.Error()
	if !strings.Contains(errStr, hostplanPath) {
		t.Errorf("error message should contain file path %q, got: %q", hostplanPath, errStr)
	}
}

func TestRunAgentRun_RejectsEmptyLogicalID(t *testing.T) {
	tmpDir := t.TempDir()
	hostplanPath := filepath.Join(tmpDir, "test-hostplan.json")

	// Write HostPlan with empty LogicalID
	hostPlan := engine.HostPlan{
		Version: engine.HostPlanSchemaVersion,
		PlanID:  "test-plan",
		Host:    engine.HostRef{LogicalID: ""}, // Empty - should be rejected
		Steps:   []engine.HostPlanStep{},
	}

	jsonBytes, err := json.Marshal(hostPlan)
	if err != nil {
		t.Fatalf("failed to marshal hostplan: %v", err)
	}

	if err := os.WriteFile(hostplanPath, jsonBytes, 0o644); err != nil {
		t.Fatalf("failed to write test hostplan: %v", err)
	}

	// Create a command with the hostplan flag set
	cmd := NewAgentRunCommand()
	cmd.SetArgs([]string{"--hostplan", hostplanPath})

	// Execute and capture error
	err = cmd.Execute()

	if err == nil {
		t.Fatal("expected error for empty host.logicalId")
	}

	// Error should mention empty logicalId
	errStr := err.Error()
	if !strings.Contains(errStr, "empty") || !strings.Contains(errStr, "logicalId") {
		t.Errorf("error message should mention empty logicalId, got: %q", errStr)
	}

	// Error should include the file path
	if !strings.Contains(errStr, hostplanPath) {
		t.Errorf("error message should contain file path %q, got: %q", hostplanPath, errStr)
	}
}

func TestRunAgentRun_AcceptsValidHostPlan(t *testing.T) {
	tmpDir := t.TempDir()
	hostplanPath := filepath.Join(tmpDir, "test-hostplan.json")

	// Write valid HostPlan
	hostPlan := engine.HostPlan{
		Version: engine.HostPlanSchemaVersion,
		PlanID:  "test-plan",
		Host:    engine.HostRef{LogicalID: "host-a"},
		Steps: []engine.HostPlanStep{
			{
				ID:     "step-1",
				Index:  0,
				Action: engine.StepActionBuild,
				Target: engine.ResourceRef{
					Kind:     "image",
					Name:     "test",
					Provider: "stagecraft",
				},
				Inputs: json.RawMessage(`{"provider": "generic", "workdir": "apps/backend", "dockerfile": "Dockerfile", "context": "."}`),
			},
		},
	}

	jsonBytes, err := json.Marshal(hostPlan)
	if err != nil {
		t.Fatalf("failed to marshal hostplan: %v", err)
	}

	if err := os.WriteFile(hostplanPath, jsonBytes, 0o644); err != nil {
		t.Fatalf("failed to write test hostplan: %v", err)
	}

	// Create a command with the hostplan flag set and output to temp file
	outputPath := filepath.Join(tmpDir, "report.json")
	cmd := NewAgentRunCommand()
	cmd.SetArgs([]string{"--hostplan", hostplanPath, "--output", outputPath})

	// Execute - should succeed
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("unexpected error executing valid hostplan: %v", err)
	}

	// Verify report was written
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected execution report to be written to %s: %v", outputPath, err)
	}
}
