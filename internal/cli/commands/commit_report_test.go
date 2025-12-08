// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"stagecraft/internal/reports/commithealth"
)

func TestRunCommitReport_GeneratesReport(t *testing.T) {
	t.Parallel()

	// Create temporary repository structure
	tmpDir := t.TempDir()

	// Create .stagecraft/reports directory
	reportsDir := filepath.Join(tmpDir, ".stagecraft", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create minimal spec/features.yaml
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
`
	if err := os.WriteFile(filepath.Join(specDir, "features.yaml"), []byte(featuresYAML), 0o644); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWd)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// TODO: Use fake HistorySource instead of real git
	// For now, skip if git is not available
	t.Skip("requires fake git implementation or isolated test repo")
}

func TestRunCommitReport_ReportStructure(t *testing.T) {
	t.Parallel()

	// Test that the generated report has the expected structure
	// This is a unit test that doesn't require actual git

	commits := []commithealth.CommitMetadata{
		{
			SHA:     "abc123",
			Message: "feat(CLI_DEPLOY): add rollback support",
		},
	}

	knownFeatures := map[string]bool{
		"CLI_DEPLOY": true,
	}

	repoInfo := commithealth.RepoInfo{
		Name:          "stagecraft",
		DefaultBranch: "main",
	}

	rangeInfo := commithealth.CommitRange{
		From:        "origin/main",
		To:          "HEAD",
		Description: "origin/main..HEAD",
	}

	report, err := commithealth.GenerateCommitHealthReport(commits, knownFeatures, repoInfo, rangeInfo)
	if err != nil {
		t.Fatalf("GenerateCommitHealthReport failed: %v", err)
	}

	// Verify report structure
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Repo.Name != "stagecraft" {
		t.Errorf("expected repo.name=stagecraft, got %s", report.Repo.Name)
	}

	// Verify report can be marshaled to JSON
	_, err = json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}
}

func TestLoadFeatureRegistry(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create spec/features.yaml
	specDir := filepath.Join(tmpDir, "spec")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	featuresYAML := `features:
  - id: CLI_DEPLOY
    title: "Deploy command"
    status: done
  - id: CLI_PLAN
    title: "Plan command"
    status: done
`
	if err := os.WriteFile(filepath.Join(specDir, "features.yaml"), []byte(featuresYAML), 0o644); err != nil {
		t.Fatalf("failed to write features.yaml: %v", err)
	}

	registry, err := loadFeatureRegistry(tmpDir)
	if err != nil {
		t.Fatalf("loadFeatureRegistry failed: %v", err)
	}

	// Verify registry contains expected features
	if !registry["CLI_DEPLOY"] {
		t.Error("CLI_DEPLOY not found in registry")
	}
	if !registry["CLI_PLAN"] {
		t.Error("CLI_PLAN not found in registry")
	}
}
