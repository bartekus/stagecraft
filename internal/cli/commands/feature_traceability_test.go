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

	"stagecraft/internal/reports/featuretrace"
)

func TestRunFeatureTraceability_GeneratesReport(t *testing.T) {
	t.Parallel()

	// Create temporary repository structure
	tmpDir := t.TempDir()

	// Create .stagecraft/reports directory
	reportsDir := filepath.Join(tmpDir, ".stagecraft", "reports")
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		t.Fatalf("failed to create reports directory: %v", err)
	}

	// Create spec file with Feature header
	specDir := filepath.Join(tmpDir, "spec", "commands")
	if err := os.MkdirAll(specDir, 0o755); err != nil {
		t.Fatalf("failed to create spec directory: %v", err)
	}

	specContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

# Deploy Command
`
	if err := os.WriteFile(filepath.Join(specDir, "deploy.md"), []byte(specContent), 0o644); err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	// Create implementation file
	implDir := filepath.Join(tmpDir, "internal", "core")
	if err := os.MkdirAll(implDir, 0o755); err != nil {
		t.Fatalf("failed to create internal directory: %v", err)
	}

	implContent := `// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

package core
`
	if err := os.WriteFile(filepath.Join(implDir, "deploy.go"), []byte(implContent), 0o644); err != nil {
		t.Fatalf("failed to write impl file: %v", err)
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

	// TODO: Call runFeatureTraceability and verify report is generated
	// For now, test the scan directly
	scanConfig := featuretrace.ScanConfig{
		RootDir: tmpDir,
	}

	features, err := featuretrace.ScanFeaturePresence(scanConfig)
	if err != nil {
		t.Fatalf("ScanFeaturePresence failed: %v", err)
	}

	// Generate report
	report, err := featuretrace.GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	// Verify report structure
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	// Verify report can be marshaled to JSON
	_, err = json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}
}

func TestRunFeatureTraceability_ReportStructure(t *testing.T) {
	t.Parallel()

	// Test that the generated report has the expected structure
	features := []featuretrace.FeaturePresence{
		{
			FeatureID:           "CLI_DEPLOY",
			Status:              featuretrace.FeatureStatusDone,
			HasSpec:             true,
			SpecPath:            "spec/commands/deploy.md",
			ImplementationFiles: []string{"internal/core/deploy.go"},
			TestFiles:           []string{"internal/core/deploy_test.go"},
			CommitSHAs:          []string{"abc123"},
		},
	}

	report, err := featuretrace.GenerateFeatureTraceabilityReport(features)
	if err != nil {
		t.Fatalf("GenerateFeatureTraceabilityReport failed: %v", err)
	}

	// Verify report structure
	if report.SchemaVersion != "1.0" {
		t.Errorf("expected schema_version=1.0, got %s", report.SchemaVersion)
	}

	if report.Summary.TotalFeatures != 1 {
		t.Errorf("expected total_features=1, got %d", report.Summary.TotalFeatures)
	}

	// Verify report can be marshaled to JSON
	_, err = json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal report to JSON: %v", err)
	}
}
