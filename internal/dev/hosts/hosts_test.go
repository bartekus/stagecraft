// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_HOSTS
// Spec: spec/dev/hosts.md

// Package hosts provides hosts file management for dev domains.
package hosts

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilePath(t *testing.T) {
	t.Helper()

	path := FilePath()
	if path == "" {
		t.Fatal("FilePath() returned empty string")
	}

	// Verify platform-specific paths
	// Note: This test runs on the current platform, so we can't test all platforms
	// but we can verify the path is reasonable
	if !strings.Contains(path, "hosts") {
		t.Errorf("FilePath() = %q, expected to contain 'hosts'", path)
	}
}

func TestParseFile_EmptyFile(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create empty file
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create empty hosts file: %v", err)
	}

	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 0 {
		t.Errorf("ParseFile() entries count = %d, want 0", len(file.Entries))
	}
}

func TestParseFile_MissingFile(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "nonexistent")

	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil for missing file", err)
	}

	if len(file.Entries) != 0 {
		t.Errorf("ParseFile() entries count = %d, want 0", len(file.Entries))
	}
}

func TestParseFile_ValidEntries(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	content := `127.0.0.1    localhost
::1    localhost
192.168.1.1    example.com
`
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write hosts file: %v", err)
	}

	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 3 {
		t.Fatalf("ParseFile() entries count = %d, want 3", len(file.Entries))
	}

	// Check first entry
	if file.Entries[0].IP != "127.0.0.1" {
		t.Errorf("Entry[0].IP = %q, want %q", file.Entries[0].IP, "127.0.0.1")
	}
	if len(file.Entries[0].Domains) != 1 || file.Entries[0].Domains[0] != "localhost" {
		t.Errorf("Entry[0].Domains = %v, want [localhost]", file.Entries[0].Domains)
	}
}

func TestParseFile_StagecraftManaged(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	content := `127.0.0.1    app.localdev.test api.localdev.test    # Stagecraft managed
127.0.0.1    example.com
`
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write hosts file: %v", err)
	}

	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 2 {
		t.Fatalf("ParseFile() entries count = %d, want 2", len(file.Entries))
	}

	// Check first entry is marked as managed
	if !file.Entries[0].Managed {
		t.Error("Entry[0].Managed = false, want true")
	}
	if file.Entries[0].Comment != "# Stagecraft managed" {
		t.Errorf("Entry[0].Comment = %q, want %q", file.Entries[0].Comment, "# Stagecraft managed")
	}

	// Check second entry is not managed
	if file.Entries[1].Managed {
		t.Error("Entry[1].Managed = true, want false")
	}
}

func TestParseFile_Comments(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	content := `# This is a comment
127.0.0.1    localhost
# Another comment
`
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write hosts file: %v", err)
	}

	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	// Should have 3 entries: comment, entry, comment
	if len(file.Entries) != 3 {
		t.Fatalf("ParseFile() entries count = %d, want 3", len(file.Entries))
	}

	// Check comments are preserved
	if file.Entries[0].Comment != "# This is a comment" {
		t.Errorf("Entry[0].Comment = %q, want %q", file.Entries[0].Comment, "# This is a comment")
	}
	if file.Entries[2].Comment != "# Another comment" {
		t.Errorf("Entry[2].Comment = %q, want %q", file.Entries[2].Comment, "# Another comment")
	}
}

func TestWriteFile_RoundTrip(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create initial file
	file := &File{
		Entries: []Entry{
			{IP: "127.0.0.1", Domains: []string{"localhost"}},
			{IP: "::1", Domains: []string{"localhost"}},
		},
	}

	if err := WriteFile(hostsPath, file); err != nil {
		t.Fatalf("WriteFile() error = %v, want nil", err)
	}

	// Parse it back
	parsed, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(parsed.Entries) != 2 {
		t.Fatalf("Parsed entries count = %d, want 2", len(parsed.Entries))
	}
}

func TestFile_AddManagedEntry(t *testing.T) {
	t.Helper()

	file := &File{Entries: []Entry{}}

	// Add entry
	file.AddManagedEntry([]string{"app.localdev.test", "api.localdev.test"})

	if len(file.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(file.Entries))
	}

	entry := file.Entries[0]
	if entry.IP != "127.0.0.1" {
		t.Errorf("Entry.IP = %q, want %q", entry.IP, "127.0.0.1")
	}
	if !entry.Managed {
		t.Error("Entry.Managed = false, want true")
	}
	if entry.Comment != StagecraftComment {
		t.Errorf("Entry.Comment = %q, want %q", entry.Comment, StagecraftComment)
	}

	// Domains should be sorted
	if len(entry.Domains) != 2 {
		t.Fatalf("Entry.Domains length = %d, want 2", len(entry.Domains))
	}
	if entry.Domains[0] != "api.localdev.test" {
		t.Errorf("Entry.Domains[0] = %q, want %q", entry.Domains[0], "api.localdev.test")
	}
	if entry.Domains[1] != "app.localdev.test" {
		t.Errorf("Entry.Domains[1] = %q, want %q", entry.Domains[1], "app.localdev.test")
	}
}

func TestFile_AddManagedEntry_Idempotent(t *testing.T) {
	t.Helper()

	file := &File{Entries: []Entry{}}

	domains := []string{"app.localdev.test", "api.localdev.test"}

	// Add entry twice
	file.AddManagedEntry(domains)
	file.AddManagedEntry(domains)

	// Should still have only one entry
	if len(file.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(file.Entries))
	}
}

func TestFile_RemoveManagedEntries(t *testing.T) {
	t.Helper()

	file := &File{
		Entries: []Entry{
			{IP: "127.0.0.1", Domains: []string{"app.localdev.test"}, Managed: true},
			{IP: "127.0.0.1", Domains: []string{"example.com"}, Managed: false},
			{IP: "127.0.0.1", Domains: []string{"api.localdev.test"}, Managed: true},
		},
	}

	file.RemoveManagedEntries()

	if len(file.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(file.Entries))
	}

	if file.Entries[0].Domains[0] != "example.com" {
		t.Errorf("Remaining entry domain = %q, want %q", file.Entries[0].Domains[0], "example.com")
	}
}

func TestManager_AddEntries(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create empty hosts file
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create hosts file: %v", err)
	}

	mgr := NewManagerWithOptions(Options{
		HostsFilePath: hostsPath,
	})

	ctx := context.Background()
	if err := mgr.AddEntries(ctx, []string{"app.localdev.test", "api.localdev.test"}); err != nil {
		t.Fatalf("AddEntries() error = %v, want nil", err)
	}

	// Verify entries were added
	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(file.Entries))
	}

	entry := file.Entries[0]
	if !entry.Managed {
		t.Error("Entry.Managed = false, want true")
	}
	if entry.IP != "127.0.0.1" {
		t.Errorf("Entry.IP = %q, want %q", entry.IP, "127.0.0.1")
	}
}

func TestManager_AddEntries_PreservesExisting(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file with existing entry
	content := `127.0.0.1    localhost
`
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write hosts file: %v", err)
	}

	mgr := NewManagerWithOptions(Options{
		HostsFilePath: hostsPath,
	})

	ctx := context.Background()
	if err := mgr.AddEntries(ctx, []string{"app.localdev.test"}); err != nil {
		t.Fatalf("AddEntries() error = %v, want nil", err)
	}

	// Verify both entries exist
	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 2 {
		t.Fatalf("Entries count = %d, want 2", len(file.Entries))
	}

	// First entry should be localhost (preserved)
	if file.Entries[0].Domains[0] != "localhost" {
		t.Errorf("Entry[0].Domains[0] = %q, want %q", file.Entries[0].Domains[0], "localhost")
	}

	// Second entry should be Stagecraft-managed
	if !file.Entries[1].Managed {
		t.Error("Entry[1].Managed = false, want true")
	}
}

func TestManager_Cleanup(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create hosts file with managed and unmanaged entries
	content := `127.0.0.1    app.localdev.test api.localdev.test    # Stagecraft managed
127.0.0.1    example.com
`
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write hosts file: %v", err)
	}

	mgr := NewManagerWithOptions(Options{
		HostsFilePath: hostsPath,
	})

	ctx := context.Background()
	if err := mgr.Cleanup(ctx); err != nil {
		t.Fatalf("Cleanup() error = %v, want nil", err)
	}

	// Verify only unmanaged entry remains
	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if len(file.Entries) != 1 {
		t.Fatalf("Entries count = %d, want 1", len(file.Entries))
	}

	if file.Entries[0].Domains[0] != "example.com" {
		t.Errorf("Remaining entry domain = %q, want %q", file.Entries[0].Domains[0], "example.com")
	}
}

func TestManager_AddEntries_Idempotent(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")

	// Create empty hosts file
	//nolint:gosec // G306: test file permissions are acceptable
	if err := os.WriteFile(hostsPath, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to create hosts file: %v", err)
	}

	mgr := NewManagerWithOptions(Options{
		HostsFilePath: hostsPath,
	})

	ctx := context.Background()
	domains := []string{"app.localdev.test", "api.localdev.test"}

	// Add entries twice
	if err := mgr.AddEntries(ctx, domains); err != nil {
		t.Fatalf("AddEntries() error = %v, want nil", err)
	}
	if err := mgr.AddEntries(ctx, domains); err != nil {
		t.Fatalf("AddEntries() error = %v, want nil", err)
	}

	// Verify only one entry exists
	file, err := ParseFile(hostsPath)
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	managedCount := 0
	for _, entry := range file.Entries {
		if entry.Managed {
			managedCount++
		}
	}

	if managedCount != 1 {
		t.Errorf("Managed entries count = %d, want 1", managedCount)
	}
}
