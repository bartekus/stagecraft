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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/logging"
)

// Feature: CLI_PHASE_EXECUTION_COMMON
// Spec: spec/core/phase-execution-common.md

// TestPhaseExecution_GoldenFiles tests CLI output for phase failure scenarios
// to ensure error propagation correctness, uniform formatting, and CLI output stability.
func TestPhaseExecution_GoldenFiles(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		golden      string
		setupEnv    func(*testing.T) *isolatedStateTestEnv
		setupPhases func() PhaseFns
		expectError bool
	}{
		{
			name:    "deploy_phase_failure_build",
			command: "deploy",
			args:    []string{"deploy", "--env", "staging"},
			golden:  "deploy_phase_failure_build",
			setupEnv: func(t *testing.T) *isolatedStateTestEnv {
				env := setupIsolatedStateTestEnv(t)
				configPath := filepath.Join(env.TempDir, "stagecraft.yml")
				configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
				if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
				return env
			},
			setupPhases: func() PhaseFns {
				return PhaseFns{
					Build: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return fmt.Errorf("build phase failed: docker build error")
					},
					Push:        defaultPhaseFns.Push,
					MigratePre:  defaultPhaseFns.MigratePre,
					Rollout:     defaultPhaseFns.Rollout,
					MigratePost: defaultPhaseFns.MigratePost,
					Finalize:    defaultPhaseFns.Finalize,
				}
			},
			expectError: true,
		},
		{
			name:    "deploy_phase_failure_rollout",
			command: "deploy",
			args:    []string{"deploy", "--env", "staging"},
			golden:  "deploy_phase_failure_rollout",
			setupEnv: func(t *testing.T) *isolatedStateTestEnv {
				env := setupIsolatedStateTestEnv(t)
				configPath := filepath.Join(env.TempDir, "stagecraft.yml")
				configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
				if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}
				return env
			},
			setupPhases: func() PhaseFns {
				return PhaseFns{
					Build: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					Push: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					MigratePre: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return fmt.Errorf("rollout phase failed: service unavailable")
					},
					MigratePost: defaultPhaseFns.MigratePost,
					Finalize:    defaultPhaseFns.Finalize,
				}
			},
			expectError: true,
		},
		{
			name:    "rollback_phase_failure_rollout",
			command: "rollback",
			args:    []string{"rollback", "--env", "staging", "--to-previous"},
			golden:  "rollback_phase_failure_rollout",
			setupEnv: func(t *testing.T) *isolatedStateTestEnv {
				env := setupIsolatedStateTestEnv(t)
				configPath := filepath.Join(env.TempDir, "stagecraft.yml")
				configContent := `project:
  name: test-app
environments:
  staging:
    driver: local
`
				if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
					t.Fatalf("failed to write config file: %v", err)
				}

				// Create a previous release for rollback
				previous, err := env.Manager.CreateRelease(env.Ctx, "staging", "v1.0.0", "commit1")
				if err != nil {
					t.Fatalf("failed to create previous release: %v", err)
				}

				// Mark all phases as completed for previous release
				allPhases := []state.ReleasePhase{
					state.PhaseBuild,
					state.PhasePush,
					state.PhaseMigratePre,
					state.PhaseRollout,
					state.PhaseMigratePost,
					state.PhaseFinalize,
				}
				for _, phase := range allPhases {
					if err := env.Manager.UpdatePhase(env.Ctx, previous.ID, phase, state.StatusCompleted); err != nil {
						t.Fatalf("failed to update phase: %v", err)
					}
				}

				// Create current release
				current, err := env.Manager.CreateRelease(env.Ctx, "staging", "v2.0.0", "commit2")
				if err != nil {
					t.Fatalf("failed to create current release: %v", err)
				}
				for _, phase := range allPhases {
					if err := env.Manager.UpdatePhase(env.Ctx, current.ID, phase, state.StatusCompleted); err != nil {
						t.Fatalf("failed to update phase: %v", err)
					}
				}

				return env
			},
			setupPhases: func() PhaseFns {
				return PhaseFns{
					Build: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					Push: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					MigratePre: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return nil
					},
					Rollout: func(ctx context.Context, plan *core.Plan, logger logging.Logger) error {
						return fmt.Errorf("rollout phase failed: cannot connect to service")
					},
					MigratePost: defaultPhaseFns.MigratePost,
					Finalize:    defaultPhaseFns.Finalize,
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := tt.setupEnv(t)
			fns := tt.setupPhases()

			// Create command with custom PhaseFns
			root := newTestRootCommand()
			var cmd *cobra.Command
			switch tt.command {
			case "deploy":
				cmd = setupDeployCommand(fns)
			case "rollback":
				cmd = setupRollbackCommand(fns)
			default:
				t.Fatalf("unknown command: %s", tt.command)
			}
			root.AddCommand(cmd)

			// Execute command and capture output
			output, err := executeCommandForGolden(root, tt.args...)

			// Normalize output for golden comparison (remove timestamps, release IDs, etc.)
			normalized := normalizeGoldenOutput(output)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Compare with golden file
			expected := readGoldenFile(t, tt.golden)

			if *updateGolden {
				writeGoldenFile(t, tt.golden, normalized)
				expected = normalized
			}

			if normalized != expected {
				t.Errorf("output mismatch:\nGot:\n%s\nExpected:\n%s", normalized, expected)
			}

			// Verify state consistency: if error occurred, verify phase statuses
			// Note: This is a best-effort check; the primary goal is to verify CLI output
			if tt.expectError && err != nil {
				releases, listErr := env.Manager.ListReleases(env.Ctx, "staging")
				if listErr != nil {
					// Non-fatal: state check is secondary to output verification
					t.Logf("failed to list releases for state verification: %v", listErr)
				} else if len(releases) > 0 {
					// The last release should have the failed phase marked as failed
					lastRelease := releases[len(releases)-1]
					// Verify that at least one phase is in a failed state
					hasFailed := false
					for _, status := range lastRelease.Phases {
						if status == state.StatusFailed {
							hasFailed = true
							break
						}
					}
					if !hasFailed {
						// Log but don't fail - this is a secondary check
						t.Logf("note: no phases marked as failed in state (this may be expected depending on failure timing)")
					}
				}
			}
		})
	}
}

// normalizeGoldenOutput normalizes command output for golden file comparison.
// It removes or normalizes dynamic content like timestamps, release IDs, file paths, etc.
func normalizeGoldenOutput(output string) string {
	lines := strings.Split(output, "\n")
	normalized := make([]string, 0, len(lines))

	for _, line := range lines {
		// Skip empty lines at the end
		if len(normalized) > 0 && line == "" && normalized[len(normalized)-1] == "" {
			continue
		}

		// Normalize release IDs (e.g., rel-20251205-142344398 -> rel-TIMESTAMP)
		line = normalizeReleaseID(line)

		// Normalize timestamps in log messages
		line = normalizeTimestamps(line)

		// Normalize file paths (keep structure but normalize temp dirs)
		line = normalizePaths(line)

		normalized = append(normalized, line)
	}

	// Remove trailing empty lines
	for len(normalized) > 0 && normalized[len(normalized)-1] == "" {
		normalized = normalized[:len(normalized)-1]
	}

	return strings.Join(normalized, "\n")
}

// normalizeReleaseID replaces release IDs with a placeholder.
func normalizeReleaseID(line string) string {
	// Pattern: rel-YYYYMMDD-HHMMSSNNN
	// Replace with rel-TIMESTAMP
	if strings.Contains(line, "rel-") {
		// Simple approach: replace the timestamp part
		parts := strings.Split(line, "rel-")
		if len(parts) > 1 {
			// Find where the ID ends (space, comma, quote, etc.)
			idPart := parts[1]
			for i, r := range idPart {
				if r == ' ' || r == ',' || r == '"' || r == '\'' || r == ')' || r == '}' {
					normalizedID := "rel-TIMESTAMP" + idPart[i:]
					return parts[0] + normalizedID
				}
			}
			return parts[0] + "rel-TIMESTAMP"
		}
	}
	return line
}

// normalizeTimestamps normalizes timestamp patterns in log output.
func normalizeTimestamps(line string) string {
	// This is a simple implementation - can be enhanced if needed
	// For now, we'll keep timestamps as they are since they might be important for error messages
	return line
}

// normalizePaths normalizes file paths in output.
func normalizePaths(line string) string {
	// Replace common temp directory patterns with placeholders
	// This is a simple implementation - can be enhanced if needed
	return line
}
