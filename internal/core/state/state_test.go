// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package state

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// Feature: CORE_STATE
// Spec: spec/core/state.md

// newTestManager creates a test manager with a deterministic clock.
// Timestamps increase by 1 second per call, ensuring deterministic ordering.
func newTestManager(stateFile string) *Manager {
	t0 := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	current := t0
	return &Manager{
		stateFile: stateFile,
		now: func() time.Time {
			// Monotonically increasing timestamps for each call
			result := current
			current = current.Add(time.Second)
			return result
		},
	}
}

func TestNewManager(t *testing.T) {
	mgr := NewManager(".stagecraft/releases.json")
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}
	if mgr.stateFile != ".stagecraft/releases.json" {
		t.Errorf("expected stateFile '.stagecraft/releases.json', got %q", mgr.stateFile)
	}
	if mgr.now == nil {
		t.Error("expected now function to be initialized")
	}
}

func TestNewDefaultManager(t *testing.T) {
	mgr := NewDefaultManager()
	if mgr == nil {
		t.Fatal("NewDefaultManager returned nil")
	}
	if mgr.stateFile != DefaultStatePath {
		t.Errorf("expected stateFile %q, got %q", DefaultStatePath, mgr.stateFile)
	}
}

func TestManager_CreateRelease(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	release, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	if release.ID == "" {
		t.Error("expected release ID to be set")
	}

	if !strings.HasPrefix(release.ID, "rel-") {
		t.Errorf("expected release ID to start with 'rel-', got %q", release.ID)
	}

	if release.Environment != "prod" {
		t.Errorf("expected environment 'prod', got %q", release.Environment)
	}

	if release.Version != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got %q", release.Version)
	}

	if release.CommitSHA != "abc123" {
		t.Errorf("expected commit SHA 'abc123', got %q", release.CommitSHA)
	}

	if release.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}

	// Check that all phases are initialized as pending
	for _, phase := range allPhases {
		status, ok := release.Phases[phase]
		if !ok {
			t.Errorf("expected phase %q to be initialized", phase)
		} else if status != StatusPending {
			t.Errorf("expected phase %q to be pending, got %q", phase, status)
		}
	}

	// Verify file was created
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("expected state file to be created")
	}
}

func TestManager_CreateRelease_InputValidation(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Test empty environment
	_, err := mgr.CreateRelease(context.Background(), "", "v1.0.0", "abc123")
	if err == nil {
		t.Fatal("expected error for empty environment")
	}
	if !strings.Contains(err.Error(), "environment must not be empty") {
		t.Errorf("expected error about empty environment, got %v", err)
	}

	// Test empty version
	_, err = mgr.CreateRelease(context.Background(), "prod", "", "abc123")
	if err == nil {
		t.Fatal("expected error for empty version")
	}
	if !strings.Contains(err.Error(), "version must not be empty") {
		t.Errorf("expected error about empty version, got %v", err)
	}

	// Test whitespace-only environment
	_, err = mgr.CreateRelease(context.Background(), "  ", "v1.0.0", "abc123")
	if err == nil {
		t.Fatal("expected error for whitespace-only environment")
	}

	// Test whitespace-only version
	_, err = mgr.CreateRelease(context.Background(), "prod", "\t\n", "abc123")
	if err == nil {
		t.Fatal("expected error for whitespace-only version")
	}

	// Test trimmed whitespace (should succeed)
	release, err := mgr.CreateRelease(context.Background(), "  prod  ", "  v1.0.0  ", "  abc123  ")
	if err != nil {
		t.Fatalf("expected success with trimmed whitespace, got: %v", err)
	}
	if release.Environment != "prod" {
		t.Errorf("expected environment to be trimmed, got %q", release.Environment)
	}
	if release.Version != "v1.0.0" {
		t.Errorf("expected version to be trimmed, got %q", release.Version)
	}
}

func TestManager_CreateRelease_WithPreviousRelease(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create first release
	release1, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Create second release (deterministic clock ensures different timestamp)
	release2, err := mgr.CreateRelease(context.Background(), "prod", "v1.1.0", "def456")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	if release2.PreviousID != release1.ID {
		t.Errorf("expected previous ID %q, got %q", release1.ID, release2.PreviousID)
	}
}

func TestManager_CreateRelease_MultipleEnvironments(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create releases for different environments
	prodRelease, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	stagingRelease, err := mgr.CreateRelease(context.Background(), "staging", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Each environment should have no previous release initially
	if prodRelease.PreviousID != "" {
		t.Errorf("expected prod release to have no previous ID, got %q", prodRelease.PreviousID)
	}

	if stagingRelease.PreviousID != "" {
		t.Errorf("expected staging release to have no previous ID, got %q", stagingRelease.PreviousID)
	}
}

func TestManager_GetRelease(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a release
	created, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Retrieve it
	retrieved, err := mgr.GetRelease(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetRelease failed: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Errorf("expected ID %q, got %q", created.ID, retrieved.ID)
	}

	if retrieved.Environment != "prod" {
		t.Errorf("expected environment 'prod', got %q", retrieved.Environment)
	}

	if retrieved.Version != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got %q", retrieved.Version)
	}
}

func TestManager_GetRelease_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	_, err := mgr.GetRelease(context.Background(), "rel-nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent release")
	}

	if !errors.Is(err, ErrReleaseNotFound) {
		t.Errorf("expected ErrReleaseNotFound, got %v", err)
	}
}

func TestManager_GetRelease_ReadOnlySnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a release
	created, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Retrieve it
	retrieved, err := mgr.GetRelease(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetRelease failed: %v", err)
	}

	// Mutate the returned release (should not affect stored state)
	retrieved.Phases[PhaseBuild] = StatusCompleted
	retrieved.Version = "mutated"

	// Retrieve again - should be unchanged
	retrieved2, err := mgr.GetRelease(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetRelease failed: %v", err)
	}

	if retrieved2.Phases[PhaseBuild] != StatusPending {
		t.Error("expected mutation of returned release to not affect stored state")
	}

	if retrieved2.Version != "v1.2.3" {
		t.Error("expected mutation of returned release to not affect stored state")
	}
}

func TestManager_GetCurrentRelease(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create multiple releases for same environment
	release1, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	release2, err := mgr.CreateRelease(context.Background(), "prod", "v1.1.0", "def456")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Get current release (should be the most recent)
	current, err := mgr.GetCurrentRelease(context.Background(), "prod")
	if err != nil {
		t.Fatalf("GetCurrentRelease failed: %v", err)
	}

	if current.ID != release2.ID {
		t.Errorf("expected current release ID %q, got %q", release2.ID, current.ID)
	}

	// Verify it's not the older one
	if current.ID == release1.ID {
		t.Error("expected current release to be the newer one")
	}
}

func TestManager_GetCurrentRelease_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	_, err := mgr.GetCurrentRelease(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent environment")
	}

	if !errors.Is(err, ErrReleaseNotFound) {
		t.Errorf("expected ErrReleaseNotFound, got %v", err)
	}

	// Verify error message format matches other methods
	if !strings.Contains(err.Error(), "environment") {
		t.Errorf("expected error to mention environment, got %q", err.Error())
	}
}

func TestManager_UpdatePhase(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a release
	release, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Update a phase
	err = mgr.UpdatePhase(context.Background(), release.ID, PhaseBuild, StatusRunning)
	if err != nil {
		t.Fatalf("UpdatePhase failed: %v", err)
	}

	// Retrieve and verify
	updated, err := mgr.GetRelease(context.Background(), release.ID)
	if err != nil {
		t.Fatalf("GetRelease failed: %v", err)
	}

	if updated.Phases[PhaseBuild] != StatusRunning {
		t.Errorf("expected PhaseBuild to be running, got %q", updated.Phases[PhaseBuild])
	}

	// Other phases should still be pending
	if updated.Phases[PhasePush] != StatusPending {
		t.Errorf("expected PhasePush to be pending, got %q", updated.Phases[PhasePush])
	}
}

func TestManager_UpdatePhase_InvalidPhase(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a release
	release, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Try to update with invalid phase
	err = mgr.UpdatePhase(context.Background(), release.ID, ReleasePhase("invalid_phase"), StatusRunning)
	if err == nil {
		t.Fatal("expected error for invalid phase")
	}

	if !strings.Contains(err.Error(), "unknown phase") {
		t.Errorf("expected error about unknown phase, got %v", err)
	}
}

func TestManager_UpdatePhase_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	err := mgr.UpdatePhase(context.Background(), "rel-nonexistent", PhaseBuild, StatusRunning)
	if err == nil {
		t.Fatal("expected error for nonexistent release")
	}

	if !errors.Is(err, ErrReleaseNotFound) {
		t.Errorf("expected ErrReleaseNotFound, got %v", err)
	}
}

func TestManager_ListReleases(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create multiple releases for same environment
	release1, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	release2, err := mgr.CreateRelease(context.Background(), "prod", "v1.1.0", "def456")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	release3, err := mgr.CreateRelease(context.Background(), "prod", "v1.2.0", "ghi789")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// List releases
	releases, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(releases) != 3 {
		t.Errorf("expected 3 releases, got %d", len(releases))
	}

	// Should be sorted newest first
	if releases[0].ID != release3.ID {
		t.Errorf("expected newest release first, got %q", releases[0].ID)
	}

	if releases[1].ID != release2.ID {
		t.Errorf("expected second newest release second, got %q", releases[1].ID)
	}

	if releases[2].ID != release1.ID {
		t.Errorf("expected oldest release last, got %q", releases[2].ID)
	}
}

func TestManager_ListReleases_ReadOnlySnapshots(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a release
	_, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// List releases
	releases, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(releases))
	}

	// Mutate the returned release
	releases[0].Version = "mutated"

	// List again - should be unchanged
	releases2, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if releases2[0].Version != "v1.0.0" {
		t.Error("expected mutation of returned release to not affect stored state")
	}
}

func TestManager_ListReleases_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	releases, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(releases) != 0 {
		t.Errorf("expected empty list, got %d releases", len(releases))
	}
}

func TestManager_ListReleases_FiltersByEnvironment(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create releases for different environments
	_, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	_, err = mgr.CreateRelease(context.Background(), "staging", "v1.0.0", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// List only prod releases
	prodReleases, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(prodReleases) != 1 {
		t.Errorf("expected 1 prod release, got %d", len(prodReleases))
	}

	if prodReleases[0].Environment != "prod" {
		t.Errorf("expected environment 'prod', got %q", prodReleases[0].Environment)
	}
}

func TestManager_StateFilePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr1 := newTestManager(stateFile)

	// Create a release with first manager
	release, err := mgr1.CreateRelease(context.Background(), "prod", "v1.2.3", "abc123")
	if err != nil {
		t.Fatalf("CreateRelease failed: %v", err)
	}

	// Create a new manager (simulating a new process)
	mgr2 := newTestManager(stateFile)

	// Should be able to retrieve the release
	retrieved, err := mgr2.GetRelease(context.Background(), release.ID)
	if err != nil {
		t.Fatalf("GetRelease failed: %v", err)
	}

	if retrieved.ID != release.ID {
		t.Errorf("expected ID %q, got %q", release.ID, retrieved.ID)
	}
}

func TestManager_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create multiple releases to trigger multiple writes
	for i := 0; i < 5; i++ {
		_, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
		if err != nil {
			t.Fatalf("CreateRelease failed: %v", err)
		}
	}

	// Verify file is valid JSON
	//nolint:gosec // G304: stateFile is from t.TempDir() and is safe
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}

	var state struct {
		Releases []*Release `json:"releases"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("failed to parse state file as JSON: %v", err)
	}

	if len(state.Releases) != 5 {
		t.Errorf("expected 5 releases, got %d", len(state.Releases))
	}
}

func TestManager_LoadState_CorruptJSON(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	// Write corrupt JSON
	if err := os.WriteFile(stateFile, []byte(`{ not: "json"`), 0o600); err != nil {
		t.Fatalf("failed to write corrupt state file: %v", err)
	}

	mgr := newTestManager(stateFile)

	_, err := mgr.loadState(context.Background())
	if err == nil {
		t.Fatal("expected error for corrupt JSON")
	}

	if !strings.Contains(err.Error(), "parsing state file") {
		t.Errorf("expected error about parsing state file, got %v", err)
	}
}

func TestManager_LoadState_InitializesMissingPhases(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	// Minimal state with missing phases
	data := []byte(`{"releases":[{"id":"rel-1","environment":"prod","version":"v1","commit_sha":"sha","timestamp":"2025-01-01T00:00:00Z"}]}`)
	if err := os.WriteFile(stateFile, data, 0o600); err != nil {
		t.Fatalf("failed to write state file: %v", err)
	}

	mgr := newTestManager(stateFile)

	state, err := mgr.loadState(context.Background())
	if err != nil {
		t.Fatalf("loadState failed: %v", err)
	}

	if len(state.Releases) != 1 {
		t.Fatalf("expected 1 release, got %d", len(state.Releases))
	}

	rel := state.Releases[0]
	if rel.Phases == nil {
		t.Fatal("expected Phases map to be initialized")
	}
}

func TestManager_LoadState_NonexistentFile(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "nonexistent", "releases.json")

	mgr := newTestManager(stateFile)

	// Should not error when file doesn't exist
	state, err := mgr.loadState(context.Background())
	if err != nil {
		t.Fatalf("loadState should not error for nonexistent file, got: %v", err)
	}

	if state == nil {
		t.Fatal("expected state to be initialized")
	}

	if len(state.Releases) != 0 {
		t.Errorf("expected empty releases list, got %d", len(state.Releases))
	}
}

func TestGenerateReleaseID(t *testing.T) {
	now := time.Date(2025, 1, 15, 14, 30, 45, 123456789, time.UTC)
	id := generateReleaseID(now)

	expected := "rel-20250115-143045123"
	if id != expected {
		t.Errorf("expected release ID %q, got %q", expected, id)
	}

	// Verify lexicographic ordering matches chronological ordering
	earlier := time.Date(2025, 1, 15, 14, 30, 44, 999000000, time.UTC)
	later := time.Date(2025, 1, 15, 14, 30, 46, 0, time.UTC)

	id1 := generateReleaseID(earlier)
	id2 := generateReleaseID(later)

	if id1 >= id2 {
		t.Error("expected earlier release ID to be lexicographically less than later ID")
	}

	// Test same second, different milliseconds
	t1 := time.Date(2025, 1, 15, 14, 30, 45, 100000000, time.UTC)
	t2 := time.Date(2025, 1, 15, 14, 30, 45, 200000000, time.UTC)

	id3 := generateReleaseID(t1)
	id4 := generateReleaseID(t2)

	if id3 >= id4 {
		t.Error("expected earlier millisecond to produce lexicographically earlier ID")
	}
}

func TestManager_ContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	mgr := newTestManager(stateFile)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// All operations should respect context cancellation
	_, err := mgr.CreateRelease(ctx, "prod", "v1.0.0", "abc123")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	_, err = mgr.GetRelease(ctx, "rel-1")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	_, err = mgr.GetCurrentRelease(ctx, "prod")
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	err = mgr.UpdatePhase(ctx, "rel-1", PhaseBuild, StatusRunning)
	if err == nil {
		t.Error("expected error for cancelled context")
	}

	_, err = mgr.ListReleases(ctx, "prod")
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestManager_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	// Use real manager (not test manager) to test actual concurrency
	mgr := NewManager(stateFile)

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Create releases concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			defer wg.Done()
			_, err := mgr.CreateRelease(context.Background(), "prod", "v1.0.0", "abc123")
			if err != nil {
				t.Errorf("CreateRelease failed in goroutine %d: %v", n, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify state file is valid JSON
	//nolint:gosec // G304: stateFile is from t.TempDir() and is safe
	data, err := os.ReadFile(stateFile)
	if err != nil {
		t.Fatalf("failed to read state file: %v", err)
	}

	var state struct {
		Releases []*Release `json:"releases"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("failed to parse state file as JSON: %v", err)
	}

	// Should have exactly numGoroutines releases
	if len(state.Releases) != numGoroutines {
		t.Errorf("expected %d releases, got %d", numGoroutines, len(state.Releases))
	}
}

func TestManager_SpecExampleJSONRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	stateFile := filepath.Join(tmpDir, "releases.json")

	// JSON exactly matching the spec example
	specJSON := `{
  "releases": [
    {
      "id": "rel-20250101-120000",
      "environment": "prod",
      "version": "v1.2.3",
      "commit_sha": "abc123def456",
      "timestamp": "2025-01-01T12:00:00Z",
      "phases": {
        "build": "completed",
        "push": "completed",
        "migrate_pre": "completed",
        "rollout": "completed",
        "migrate_post": "completed",
        "finalize": "completed"
      },
      "previous_id": "rel-20241231-120000"
    }
  ]
}`

	if err := os.WriteFile(stateFile, []byte(specJSON), 0o600); err != nil {
		t.Fatalf("failed to write spec JSON: %v", err)
	}

	mgr := NewManager(stateFile)

	// Should be able to load and retrieve
	release, err := mgr.GetCurrentRelease(context.Background(), "prod")
	if err != nil {
		t.Fatalf("GetCurrentRelease failed: %v", err)
	}

	// Verify fields match spec example
	if release.ID != "rel-20250101-120000" {
		t.Errorf("expected ID 'rel-20250101-120000', got %q", release.ID)
	}

	if release.Environment != "prod" {
		t.Errorf("expected environment 'prod', got %q", release.Environment)
	}

	if release.Version != "v1.2.3" {
		t.Errorf("expected version 'v1.2.3', got %q", release.Version)
	}

	if release.CommitSHA != "abc123def456" {
		t.Errorf("expected commit SHA 'abc123def456', got %q", release.CommitSHA)
	}

	if release.PreviousID != "rel-20241231-120000" {
		t.Errorf("expected previous ID 'rel-20241231-120000', got %q", release.PreviousID)
	}

	// Verify all phases are present
	for _, phase := range allPhases {
		if status, ok := release.Phases[phase]; !ok {
			t.Errorf("expected phase %q to be present", phase)
		} else if status != StatusCompleted {
			t.Errorf("expected phase %q to be completed, got %q", phase, status)
		}
	}

	// List releases should also work
	releases, err := mgr.ListReleases(context.Background(), "prod")
	if err != nil {
		t.Fatalf("ListReleases failed: %v", err)
	}

	if len(releases) != 1 {
		t.Errorf("expected 1 release, got %d", len(releases))
	}
}
