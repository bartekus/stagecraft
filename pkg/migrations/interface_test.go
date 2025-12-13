// SPDX-License-Identifier: AGPL-3.0-or-later
package migrations_test

import (
	"context"
	"testing"

	"stagecraft/pkg/migrations"
)

// Feature: MIGRATION_INTERFACE
// Spec: spec/migrations/interface.md

// TestInterfaceCompliance ensures that the package defines the required types and interfaces
// as specified in spec/migrations/interface.md.
//
// This test will fail to compile if the types or interface methods are missing or incorrect.
func TestInterfaceCompliance(t *testing.T) {
	var _ migrations.Engine = (*MockEngine)(nil)
	var _ migrations.ValidatingEngine = (*MockEngine)(nil)
}

// MockEngine implements migrations.Engine to verify method signatures.
type MockEngine struct{}

func (m *MockEngine) Name() string { return "mock" }

func (m *MockEngine) List(ctx context.Context, req *migrations.MigrationRequest) ([]migrations.Migration, error) {
	return nil, nil
}

func (m *MockEngine) Plan(ctx context.Context, req *migrations.MigrationRequest) (migrations.MigrationPlan, error) {
	return migrations.MigrationPlan{}, nil
}

func (m *MockEngine) Apply(ctx context.Context, req *migrations.MigrationRequest) (migrations.MigrationApplyResult, error) {
	return migrations.MigrationApplyResult{}, nil
}

func (m *MockEngine) Validate(ctx context.Context, req *migrations.MigrationRequest) (migrations.ValidationResult, error) {
	return migrations.ValidationResult{}, nil
}

func TestTypesExist(t *testing.T) {
	// Verify struct fields and tags exist by instantiating them
	_ = migrations.Migration{
		ID:          "test-id",
		Description: "test desc",
		Tags:        []string{"a", "b"},
		Source:      "sql:test",
		DependsOn:   []migrations.MigrationID{"other"},
	}

	_ = migrations.Selection{
		All:  true,
		IDs:  []migrations.MigrationID{"id1"},
		Tags: []string{"tag1"},
	}

	_ = migrations.MigrationRequest{
		Environment: "dev",
		Mode:        migrations.ModePlan,
		Selection:   migrations.Selection{All: true},
		FailFast:    true,
		AllowNoop:   false,
		DryRun:      true,
	}

	_ = migrations.MigrationStepResult{
		ID:       "id1",
		Outcome:  migrations.OutcomeApplied,
		Message:  "done",
		Warnings: []string{"warn"},
	}

	_ = migrations.PlanSummary{
		Total:      1,
		WouldApply: 1,
		WouldSkip:  0,
	}

	_ = migrations.ApplySummary{
		Total:   1,
		Applied: 1,
		Skipped: 0,
		Failed:  0,
	}
}
