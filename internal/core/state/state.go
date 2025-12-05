// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package state provides state management for tracking deployment history and release information.
//
// Note: State is local-file-based and not safe for concurrent modification from multiple processes.
// A single Stagecraft process should own the state file at any time.
package state

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Feature: CORE_STATE
// Spec: spec/core/state.md

// DefaultStatePath is the default path for the state file.
const DefaultStatePath = ".stagecraft/releases.json"

// ReleasePhase represents a deployment phase.
type ReleasePhase string

const (
	// PhaseBuild represents the Docker image build phase.
	PhaseBuild ReleasePhase = "build"
	// PhasePush represents the push to registry phase.
	PhasePush ReleasePhase = "push"
	// PhaseMigratePre represents the pre-deployment migrations phase.
	PhaseMigratePre ReleasePhase = "migrate_pre"
	// PhaseRollout represents the container rollout phase.
	PhaseRollout ReleasePhase = "rollout"
	// PhaseMigratePost represents the post-deployment migrations phase.
	PhaseMigratePost ReleasePhase = "migrate_post"
	// PhaseFinalize represents the finalization and cleanup phase.
	PhaseFinalize ReleasePhase = "finalize"
)

// allPhases is the ordered list of all deployment phases.
// Used for initialization and validation to prevent drift.
var allPhases = []ReleasePhase{
	PhaseBuild,
	PhasePush,
	PhaseMigratePre,
	PhaseRollout,
	PhaseMigratePost,
	PhaseFinalize,
}

// PhaseStatus represents the status of a phase.
type PhaseStatus string

const (
	// StatusPending represents a phase that has not started.
	StatusPending PhaseStatus = "pending"
	// StatusRunning represents a phase that is currently executing.
	StatusRunning PhaseStatus = "running"
	// StatusCompleted represents a phase that has finished successfully.
	StatusCompleted PhaseStatus = "completed"
	// StatusFailed represents a phase that has failed.
	StatusFailed PhaseStatus = "failed"
	// StatusSkipped represents a phase that was skipped.
	StatusSkipped PhaseStatus = "skipped"
)

// Release represents a single deployment release.
// Release values returned from Manager methods should be treated as read-only snapshots.
type Release struct {
	// ID is a unique identifier for this release (e.g., "rel-20250101-120000" or "rel-20250101-120000123")
	ID string `json:"id"`

	// Environment is the target environment
	Environment string `json:"environment"`

	// Version is the deployed version (e.g., "v1.2.3" or git SHA)
	Version string `json:"version"`

	// CommitSHA is the git commit SHA.
	// MAY be empty for non-git deployments.
	CommitSHA string `json:"commit_sha"`

	// Timestamp is when the release was created
	Timestamp time.Time `json:"timestamp"`

	// Phases tracks the status of each deployment phase
	Phases map[ReleasePhase]PhaseStatus `json:"phases"`

	// PreviousID is the ID of the previous release (for rollback)
	PreviousID string `json:"previous_id,omitempty"`
}

// stateFile represents the JSON structure of the state file.
type stateFile struct {
	Releases []*Release `json:"releases"`
}

// Manager manages release state for Stagecraft deployments.
// Manager is safe for concurrent use within a single process.
// Note: State is not safe for concurrent modification from multiple processes.
type Manager struct {
	stateFile string
	now       func() time.Time
	mu        sync.Mutex
}

// ErrReleaseNotFound is returned when a release is not found.
var ErrReleaseNotFound = errors.New("release not found")

// NewManager creates a new state manager.
func NewManager(stateFile string) *Manager {
	return &Manager{
		stateFile: stateFile,
		now:       time.Now,
	}
}

// NewDefaultManager creates a new state manager with the default state file path.
// The state file path can be overridden via the STAGECRAFT_STATE_FILE environment variable
// for testing purposes.
func NewDefaultManager() *Manager {
	// Allow tests to override state file path via environment variable
	if envPath := os.Getenv("STAGECRAFT_STATE_FILE"); envPath != "" {
		return NewManager(envPath)
	}
	return NewManager(DefaultStatePath)
}

// generateReleaseID generates a release ID in the format rel-YYYYMMDD-HHMMSSmmm.
// The millisecond suffix ensures uniqueness even for high-frequency operations
// while preserving lexicographic ordering that matches chronological ordering.
// Format: rel-{date}-{time}{milliseconds}
// Example: rel-20250101-120000123
func generateReleaseID(t time.Time) string {
	return fmt.Sprintf("rel-%s-%s%03d",
		t.Format("20060102"),
		t.Format("150405"),
		t.Nanosecond()/1e6)
}

// cloneRelease creates a deep copy of a Release to prevent accidental mutation.
func cloneRelease(r *Release) *Release {
	if r == nil {
		return nil
	}

	clone := *r

	// Deep copy the Phases map
	if r.Phases != nil {
		clone.Phases = make(map[ReleasePhase]PhaseStatus, len(r.Phases))
		for k, v := range r.Phases {
			clone.Phases[k] = v
		}
	}

	return &clone
}

// isValidPhase checks if a phase is in the allowed set.
func isValidPhase(phase ReleasePhase) bool {
	for _, allowed := range allPhases {
		if allowed == phase {
			return true
		}
	}
	return false
}

// loadState loads the state file and returns the releases.
func (m *Manager) loadState(ctx context.Context) (*stateFile, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// If file doesn't exist, return empty state
	if _, err := os.Stat(m.stateFile); os.IsNotExist(err) {
		return &stateFile{Releases: []*Release{}}, nil
	}

	//nolint:gosec // G304: stateFile path comes from trusted config
	data, err := os.ReadFile(m.stateFile)
	if err != nil {
		return nil, fmt.Errorf("reading state file: %w", err)
	}

	var state stateFile
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("parsing state file: %w", err)
	}

	// Ensure Phases map is initialized for each release
	for _, release := range state.Releases {
		if release.Phases == nil {
			release.Phases = make(map[ReleasePhase]PhaseStatus)
		}
	}

	return &state, nil
}

// saveState saves the state file atomically (write to temp, then rename).
// Temp file includes PID to reduce conflicts when multiple processes write concurrently.
func (m *Manager) saveState(ctx context.Context, state *stateFile) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(m.stateFile)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}

	// Marshal JSON with indentation for readability
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling state: %w", err)
	}

	// Write to temporary file with PID suffix to reduce multi-process conflicts
	tmpFile := fmt.Sprintf("%s.%d.tmp", m.stateFile, os.Getpid())
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return fmt.Errorf("writing temporary state file: %w", err)
	}

	// Atomically rename temp file to final location
	if err := os.Rename(tmpFile, m.stateFile); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpFile)
		return fmt.Errorf("renaming state file: %w", err)
	}

	return nil
}

// CreateRelease creates a new release record.
func (m *Manager) CreateRelease(ctx context.Context, env, version, commitSHA string) (*Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Normalize and validate input
	env = strings.TrimSpace(env)
	version = strings.TrimSpace(version)
	commitSHA = strings.TrimSpace(commitSHA)

	if env == "" {
		return nil, fmt.Errorf("environment must not be empty")
	}
	if version == "" {
		return nil, fmt.Errorf("version must not be empty")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return nil, err
	}

	// Generate release ID
	now := m.now()
	releaseID := generateReleaseID(now)

	// Find previous release for this environment (O(n) single pass)
	var previous *Release
	for _, r := range state.Releases {
		if r.Environment != env {
			continue
		}
		if previous == nil || r.Timestamp.After(previous.Timestamp) {
			previous = r
		}
	}

	var previousID string
	if previous != nil {
		previousID = previous.ID
	}

	// Create new release
	release := &Release{
		ID:          releaseID,
		Environment: env,
		Version:     version,
		CommitSHA:   commitSHA,
		Timestamp:   now,
		Phases:      make(map[ReleasePhase]PhaseStatus),
		PreviousID:  previousID,
	}

	// Initialize all phases as pending
	for _, phase := range allPhases {
		release.Phases[phase] = StatusPending
	}

	// Add to state
	state.Releases = append(state.Releases, release)

	// Save state
	if err := m.saveState(ctx, state); err != nil {
		return nil, err
	}

	// Return a clone to prevent mutation
	return cloneRelease(release), nil
}

// findReleaseByID finds a release by ID in the state.
func (s *stateFile) findReleaseByID(id string) *Release {
	for _, release := range s.Releases {
		if release.ID == id {
			return release
		}
	}
	return nil
}

// GetRelease retrieves a release by ID.
// Returns a read-only snapshot of the release.
func (m *Manager) GetRelease(ctx context.Context, id string) (*Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return nil, err
	}

	release := state.findReleaseByID(id)
	if release == nil {
		return nil, fmt.Errorf("%w: %q", ErrReleaseNotFound, id)
	}

	// Return a clone to prevent mutation
	return cloneRelease(release), nil
}

// GetCurrentRelease retrieves the current release for an environment.
// Returns a read-only snapshot of the release.
func (m *Manager) GetCurrentRelease(ctx context.Context, env string) (*Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return nil, err
	}

	var current *Release
	for _, release := range state.Releases {
		if release.Environment == env {
			if current == nil || release.Timestamp.After(current.Timestamp) {
				current = release
			}
		}
	}

	if current == nil {
		return nil, fmt.Errorf("%w: environment %q", ErrReleaseNotFound, env)
	}

	// Return a clone to prevent mutation
	return cloneRelease(current), nil
}

// UpdatePhase updates the status of a deployment phase.
func (m *Manager) UpdatePhase(ctx context.Context, releaseID string, phase ReleasePhase, status PhaseStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Validate phase
	if !isValidPhase(phase) {
		return fmt.Errorf("unknown phase %q", phase)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return err
	}

	release := state.findReleaseByID(releaseID)
	if release == nil {
		return fmt.Errorf("%w: %q", ErrReleaseNotFound, releaseID)
	}

	// Initialize Phases map if nil (shouldn't happen, but be defensive)
	if release.Phases == nil {
		release.Phases = make(map[ReleasePhase]PhaseStatus)
	}

	// Update phase status
	release.Phases[phase] = status

	// Save state
	return m.saveState(ctx, state)
}

// ListReleases lists all releases for an environment, sorted newest first.
// Returns read-only snapshots of the releases.
func (m *Manager) ListReleases(ctx context.Context, env string) ([]*Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return nil, err
	}

	var releases []*Release
	for _, release := range state.Releases {
		if release.Environment == env {
			releases = append(releases, release)
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].Timestamp.After(releases[j].Timestamp)
	})

	// Return clones to prevent mutation
	clones := make([]*Release, len(releases))
	for i, r := range releases {
		clones[i] = cloneRelease(r)
	}

	return clones, nil
}

// ListAllReleases lists all releases across all environments, sorted by environment (ascending),
// then by timestamp (newest first), then by ID (ascending) for determinism.
// Returns read-only snapshots of the releases.
func (m *Manager) ListAllReleases(ctx context.Context) ([]*Release, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, err := m.loadState(ctx)
	if err != nil {
		return nil, err
	}

	// Copy all releases
	releases := make([]*Release, len(state.Releases))
	copy(releases, state.Releases)

	// Sort by environment (ascending), then timestamp (newest first), then ID (ascending)
	sort.Slice(releases, func(i, j int) bool {
		ri, rj := releases[i], releases[j]
		if ri.Environment != rj.Environment {
			return ri.Environment < rj.Environment
		}
		if !ri.Timestamp.Equal(rj.Timestamp) {
			return ri.Timestamp.After(rj.Timestamp)
		}
		return ri.ID < rj.ID
	})

	// Return clones to prevent mutation
	clones := make([]*Release, len(releases))
	for i, r := range releases {
		clones[i] = cloneRelease(r)
	}

	return clones, nil
}
