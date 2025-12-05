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

	"stagecraft/internal/core"
	"stagecraft/internal/core/state"
	"stagecraft/pkg/logging"
)

// Feature: CLI_PHASE_EXECUTION_COMMON
// Spec: spec/core/phase-execution-common.md

// PhaseFns holds the phase execution functions for dependency injection.
// This allows tests to override phase behavior without mutating global state.
type PhaseFns struct {
	Build       func(context.Context, *core.Plan, logging.Logger) error
	Push        func(context.Context, *core.Plan, logging.Logger) error
	MigratePre  func(context.Context, *core.Plan, logging.Logger) error
	Rollout     func(context.Context, *core.Plan, logging.Logger) error
	MigratePost func(context.Context, *core.Plan, logging.Logger) error
	Finalize    func(context.Context, *core.Plan, logging.Logger) error
}

// defaultPhaseFns provides the default phase execution functions.
var defaultPhaseFns = PhaseFns{
	Build:       buildPhaseFn,
	Push:        pushPhaseFn,
	MigratePre:  migratePrePhaseFn,
	Rollout:     rolloutPhaseFn,
	MigratePost: migratePostPhaseFn,
	Finalize:    finalizePhaseFn,
}

// allPhasesCommon returns all deployment phases in canonical execution order.
func allPhasesCommon() []state.ReleasePhase {
	return []state.ReleasePhase{
		state.PhaseBuild,
		state.PhasePush,
		state.PhaseMigratePre,
		state.PhaseRollout,
		state.PhaseMigratePost,
		state.PhaseFinalize,
	}
}

// phaseFnFor returns the phase function for the given phase from PhaseFns.
func phaseFnFor(phase state.ReleasePhase, fns PhaseFns) (func(context.Context, *core.Plan, logging.Logger) error, error) {
	switch phase {
	case state.PhaseBuild:
		return fns.Build, nil
	case state.PhasePush:
		return fns.Push, nil
	case state.PhaseMigratePre:
		return fns.MigratePre, nil
	case state.PhaseRollout:
		return fns.Rollout, nil
	case state.PhaseMigratePost:
		return fns.MigratePost, nil
	case state.PhaseFinalize:
		return fns.Finalize, nil
	default:
		return nil, fmt.Errorf("unknown phase: %q", phase)
	}
}

// markDownstreamPhasesSkippedCommon marks all phases after the failed phase as skipped.
func markDownstreamPhasesSkippedCommon(
	ctx context.Context,
	stateMgr *state.Manager,
	releaseID string,
	failedPhase state.ReleasePhase,
	logger logging.Logger,
) error {
	allPhases := allPhasesCommon()

	// Find the index of the failed phase
	failedIndex := -1
	for i, phase := range allPhases {
		if phase == failedPhase {
			failedIndex = i
			break
		}
	}

	if failedIndex == -1 {
		logger.Debug("Failed phase not found in phase list", logging.NewField("phase", failedPhase))
		return nil
	}

	// Get current release to check phase statuses
	release, err := stateMgr.GetRelease(ctx, releaseID)
	if err != nil {
		return fmt.Errorf("getting release %q: %w", releaseID, err)
	}

	// Mark all downstream phases as skipped (only if they're still pending or running)
	for i := failedIndex + 1; i < len(allPhases); i++ {
		phase := allPhases[i]
		currentStatus, ok := release.Phases[phase]
		if !ok {
			// Phase not initialized, skip it
			continue
		}
		// Only mark as skipped if still pending or running
		if currentStatus == state.StatusPending || currentStatus == state.StatusRunning {
			if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusSkipped); err != nil {
				logger.Debug("Failed to mark phase as skipped",
					logging.NewField("phase", phase),
					logging.NewField("error", err.Error()),
				)
				// Continue marking other phases even if one fails
			}
		}
	}

	return nil
}

// markAllPhasesFailedCommon marks all phases as failed (used when planning fails before any execution).
// Errors during phase marking are logged but do not stop the process.
func markAllPhasesFailedCommon(
	ctx context.Context,
	stateMgr *state.Manager,
	releaseID string,
	logger logging.Logger,
) {
	for _, phase := range allPhasesCommon() {
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); err != nil {
			logger.Debug("Failed to mark phase as failed",
				logging.NewField("phase", phase),
				logging.NewField("error", err.Error()),
			)
			// Continue marking other phases even if one fails
		}
	}
}

// executePhasesCommon executes all deployment phases in order using the provided PhaseFns.
// This is the shared phase execution logic used by both deploy and rollback commands.
func executePhasesCommon(
	ctx context.Context,
	stateMgr *state.Manager,
	releaseID string,
	plan *core.Plan,
	logger logging.Logger,
	fns PhaseFns,
) error {
	phases := allPhasesCommon()

	for _, phase := range phases {
		phaseName := string(phase)

		// Log phase start
		logger.Info("Starting phase", logging.NewField("phase", phaseName))

		// Set phase status to running
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusRunning); err != nil {
			return fmt.Errorf("updating phase %q to running: %w", phaseName, err)
		}

		// Get phase function
		phaseFn, err := phaseFnFor(phase, fns)
		if err != nil {
			// This should never happen with valid phases, but handle it gracefully
			if updateErr := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); updateErr != nil {
				logger.Debug("Failed to update phase status", logging.NewField("error", updateErr.Error()))
			}
			return fmt.Errorf("getting phase function for %q: %w", phaseName, err)
		}

		// Execute phase
		err = phaseFn(ctx, plan, logger)
		if err != nil {
			// Mark current phase as failed
			if updateErr := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusFailed); updateErr != nil {
				logger.Debug("Failed to update phase status", logging.NewField("error", updateErr.Error()))
			}

			// Mark all downstream phases as skipped
			if skipErr := markDownstreamPhasesSkippedCommon(ctx, stateMgr, releaseID, phase, logger); skipErr != nil {
				logger.Debug("Failed to mark downstream phases as skipped", logging.NewField("error", skipErr.Error()))
			}

			return fmt.Errorf("phase %q failed: %w", phaseName, err)
		}

		// Mark phase as completed
		if err := stateMgr.UpdatePhase(ctx, releaseID, phase, state.StatusCompleted); err != nil {
			return fmt.Errorf("updating phase %q to completed: %w", phaseName, err)
		}

		logger.Info("Phase completed", logging.NewField("phase", phaseName))
	}

	return nil
}
