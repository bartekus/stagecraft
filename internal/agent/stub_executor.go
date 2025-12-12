// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package agent

import (
	"context"
	"fmt"

	"stagecraft/pkg/engine"
	"stagecraft/pkg/engine/inputs"
)

// StubExecutor is a stub executor that validates inputs but doesn't execute actions.
// Useful for testing the pipeline end-to-end.
type StubExecutor struct{}

// Execute validates inputs but doesn't perform the actual action.
func (s *StubExecutor) Execute(ctx context.Context, step engine.HostPlanStep, inputsJSON []byte) error {
	// Strict decode and validate inputs based on action
	switch step.Action {
	case engine.StepActionBuild:
		var in inputs.BuildInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid build inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("build inputs validation failed: %w", err)
		}
		// Stub: log that we would build
		return nil

	case engine.StepActionMigrate:
		var in inputs.MigrateInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid migrate inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("migrate inputs validation failed: %w", err)
		}
		// Stub: log that we would migrate
		return nil

	case engine.StepActionApplyCompose:
		var in inputs.ApplyComposeInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid apply_compose inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("apply_compose inputs validation failed: %w", err)
		}
		// Stub: log that we would apply compose
		return nil

	case engine.StepActionHealthCheck:
		var in inputs.HealthCheckInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid health_check inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("health_check inputs validation failed: %w", err)
		}
		// Stub: log that we would health check
		return nil

	case engine.StepActionRenderCompose:
		var in inputs.RenderComposeInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid render_compose inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("render_compose inputs validation failed: %w", err)
		}
		// Stub: log that we would render compose
		return nil

	case engine.StepActionRollout:
		var in inputs.RolloutInputs
		if err := inputs.UnmarshalStrict(inputsJSON, &in); err != nil {
			return fmt.Errorf("invalid rollout inputs: %w", err)
		}
		if err := in.Validate(); err != nil {
			return fmt.Errorf("rollout inputs validation failed: %w", err)
		}
		// Stub: log that we would rollout
		return nil

	default:
		// Unknown action - just validate JSON is valid
		return nil
	}
}
