// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

package main

// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractSpecReferences_ValidComment(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

// Feature: GOV_CORE
// Spec: spec/governance/GOV_CORE.md

func main() {
}
`)

	refs := extractSpecReferences("test.go", content)
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(refs))
	}

	if refs[0].Path != "spec/governance/GOV_CORE.md" {
		t.Errorf("expected path 'spec/governance/GOV_CORE.md', got %q", refs[0].Path)
	}
	if refs[0].Line != 4 {
		t.Errorf("expected line 4, got %d", refs[0].Line)
	}
}

func TestExtractSpecReferences_CaseInsensitive(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

// spec: spec/commands/deploy.md
// Spec: spec/core/state.md
// SPEC: spec/providers/backend.md
`)

	refs := extractSpecReferences("test.go", content)
	if len(refs) != 3 {
		t.Fatalf("expected 3 references, got %d", len(refs))
	}

	expected := []string{
		"spec/commands/deploy.md",
		"spec/core/state.md",
		"spec/providers/backend.md",
	}

	for i, exp := range expected {
		if refs[i].Path != exp {
			t.Errorf("refs[%d].Path = %q, want %q", i, refs[i].Path, exp)
		}
	}
}

func TestExtractSpecReferences_IgnoresNonComments(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

func test() {
	// This is a regular comment
	fmt.Printf("Spec: test/feature1.md")
	testFile := "Spec: test/feature2.md"
	spec := SpecInfo{
		Path: "spec/commands/deploy.md",
	}
	// Spec: spec/commands/deploy.md
}
`)

	refs := extractSpecReferences("test.go", content)
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference (only from comment), got %d", len(refs))
	}

	if refs[0].Path != "spec/commands/deploy.md" {
		t.Errorf("expected path 'spec/commands/deploy.md', got %q", refs[0].Path)
	}
}

func TestExtractSpecReferences_IgnoresDebugOutput(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

func test() {
	fmt.Printf("SpecInfo{%v}", spec)
	log.Printf("Spec: %s", path)
	// Spec: spec/commands/deploy.md
}
`)

	refs := extractSpecReferences("test.go", content)
	if len(refs) != 1 {
		t.Fatalf("expected 1 reference (only from comment), got %d", len(refs))
	}
}

func TestExtractSpecReferences_EmptyPath(t *testing.T) {
	t.Parallel()

	content := []byte(`package main

// Spec:
// Spec:
`)

	refs := extractSpecReferences("test.go", content)
	if len(refs) != 2 {
		t.Fatalf("expected 2 references (with empty paths), got %d", len(refs))
	}

	for i, ref := range refs {
		if ref.Path != "" {
			t.Errorf("refs[%d].Path = %q, want empty string", i, ref.Path)
		}
	}
}

func TestValidateSpecPath_ValidPaths(t *testing.T) {
	t.Parallel()

	validPaths := []string{
		"spec/commands/deploy.md",
		"spec/core/state.md",
		"spec/governance/GOV_CORE.md",
		"spec/providers/backend.md",
	}

	for _, path := range validPaths {
		if err := validateSpecPath(path); err != nil {
			t.Errorf("validateSpecPath(%q) = %v, want nil", path, err)
		}
	}
}

func TestValidateSpecPath_InvalidPrefix(t *testing.T) {
	t.Parallel()

	invalidPaths := []string{
		"test/feature1.md",
		"docs/guide.md",
		"commands/deploy.md",
		"/spec/commands/deploy.md",
	}

	for _, path := range invalidPaths {
		if err := validateSpecPath(path); err == nil {
			t.Errorf("validateSpecPath(%q) = nil, want error", path)
		} else if !strings.Contains(err.Error(), "must start with 'spec/'") {
			t.Errorf("validateSpecPath(%q) = %v, want error about prefix", path, err)
		}
	}
}

func TestValidateSpecPath_InvalidSuffix(t *testing.T) {
	t.Parallel()

	invalidPaths := []string{
		"spec/commands/deploy",
		"spec/commands/deploy.txt",
		"spec/commands/deploy.go",
	}

	for _, path := range invalidPaths {
		if err := validateSpecPath(path); err == nil {
			t.Errorf("validateSpecPath(%q) = nil, want error", path)
		} else if !strings.Contains(err.Error(), "must end with '.md'") {
			t.Errorf("validateSpecPath(%q) = %v, want error about suffix", path, err)
		}
	}
}

func TestValidateSpecPath_InvalidCharacters(t *testing.T) {
	t.Parallel()

	invalidPaths := []string{
		"spec/commands/deploy.md\n",
		"spec/commands/deploy.md\t",
		"spec/commands/deploy.md ",
		"spec/commands/{deploy}.md",
		"spec/commands/deploy.md\"",
		"spec/commands/deploy.md\r",
	}

	for _, path := range invalidPaths {
		if err := validateSpecPath(path); err == nil {
			t.Errorf("validateSpecPath(%q) = nil, want error", path)
		} else if !strings.Contains(err.Error(), "invalid characters") {
			t.Errorf("validateSpecPath(%q) = %v, want error about invalid characters", path, err)
		}
	}
}

func TestValidateSpecPath_EmptyPath(t *testing.T) {
	t.Parallel()

	if err := validateSpecPath(""); err == nil {
		t.Error("validateSpecPath(\"\") = nil, want error")
	} else if !strings.Contains(err.Error(), "empty path") {
		t.Errorf("validateSpecPath(\"\") = %v, want error about empty path", err)
	}
}

func TestWalkGoFiles_SkipsTestdata(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   valid.go
	//   testdata/
	//     ignored.go
	//   subdir/
	//     valid.go
	//     testdata/
	//       ignored.go

	valid1 := filepath.Join(tmpDir, "valid.go")
	if err := os.WriteFile(valid1, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create valid.go: %v", err)
	}

	testdata1 := filepath.Join(tmpDir, "testdata", "ignored.go")
	if err := os.MkdirAll(filepath.Dir(testdata1), 0o750); err != nil {
		t.Fatalf("failed to create testdata dir: %v", err)
	}
	if err := os.WriteFile(testdata1, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create ignored.go: %v", err)
	}

	subdir := filepath.Join(tmpDir, "subdir")
	valid2 := filepath.Join(subdir, "valid.go")
	if err := os.MkdirAll(subdir, 0o750); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}
	if err := os.WriteFile(valid2, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create subdir/valid.go: %v", err)
	}

	testdata2 := filepath.Join(subdir, "testdata", "ignored.go")
	if err := os.MkdirAll(filepath.Dir(testdata2), 0o750); err != nil {
		t.Fatalf("failed to create subdir/testdata: %v", err)
	}
	if err := os.WriteFile(testdata2, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create subdir/testdata/ignored.go: %v", err)
	}

	files, err := walkGoFiles(tmpDir)
	if err != nil {
		t.Fatalf("walkGoFiles failed: %v", err)
	}

	// Should only find the two valid.go files, not the ones in testdata
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d: %v", len(files), files)
	}

	for _, f := range files {
		if strings.Contains(f, "testdata") {
			t.Errorf("found file in testdata directory: %s", f)
		}
	}
}

func TestWalkGoFiles_SkipsE2E(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create structure:
	// tmpDir/
	//   valid.go
	//   test/
	//     e2e/
	//       ignored.go

	valid := filepath.Join(tmpDir, "valid.go")
	if err := os.WriteFile(valid, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create valid.go: %v", err)
	}

	e2eDir := filepath.Join(tmpDir, "test", "e2e")
	if err := os.MkdirAll(e2eDir, 0o750); err != nil {
		t.Fatalf("failed to create e2e dir: %v", err)
	}

	e2eFile := filepath.Join(e2eDir, "ignored.go")
	if err := os.WriteFile(e2eFile, []byte("package main"), 0o600); err != nil {
		t.Fatalf("failed to create e2e/ignored.go: %v", err)
	}

	files, err := walkGoFiles(tmpDir)
	if err != nil {
		t.Fatalf("walkGoFiles failed: %v", err)
	}

	// Should only find valid.go, not e2e/ignored.go
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d: %v", len(files), files)
	}

	if strings.Contains(files[0], "e2e") {
		t.Errorf("found file in e2e directory: %s", files[0])
	}
}

func TestRun_ValidReferences(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create spec file
	specDir := filepath.Join(tmpDir, "spec", "commands")
	if err := os.MkdirAll(specDir, 0o700); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}
	specFile := filepath.Join(specDir, "deploy.md")
	if err := os.WriteFile(specFile, []byte("# Deploy\n"), 0o600); err != nil {
		t.Fatalf("failed to create spec file: %v", err)
	}

	// Create Go file with valid reference
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := []byte(`package main

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

func main() {
}
`)
	if err := os.WriteFile(goFile, goContent, 0o600); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	// Run should succeed
	if err := run(tmpDir); err != nil {
		t.Errorf("run(%q) = %v, want nil", tmpDir, err)
	}
}

func TestRun_MissingSpecFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create Go file with reference to non-existent spec
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := []byte(`package main

// Feature: CLI_DEPLOY
// Spec: spec/commands/nonexistent.md

func main() {
}
`)
	if err := os.WriteFile(goFile, goContent, 0o600); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	// Run should fail
	if err := run(tmpDir); err == nil {
		t.Error("run() = nil, want error")
	} else if !strings.Contains(err.Error(), "spec reference validation failed") {
		t.Errorf("run() = %v, want error about spec reference validation", err)
	}
}

func TestRun_InvalidPathFormat(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create Go file with invalid spec reference
	goFile := filepath.Join(tmpDir, "main.go")
	goContent := []byte(`package main

// Feature: CLI_DEPLOY
// Spec: test/feature1.md

func main() {
}
`)
	if err := os.WriteFile(goFile, goContent, 0o600); err != nil {
		t.Fatalf("failed to create go file: %v", err)
	}

	// Run should fail with invalid path format error
	if err := run(tmpDir); err == nil {
		t.Error("run() = nil, want error")
	} else if !strings.Contains(err.Error(), "spec reference validation failed") {
		t.Errorf("run() = %v, want error about spec reference validation", err)
	}
}

func TestRun_IgnoresTestdata(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()

	// Create spec file that main.go will reference
	specDir := filepath.Join(tmpDir, "spec", "commands")
	if err := os.MkdirAll(specDir, 0o700); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}
	specFile := filepath.Join(specDir, "deploy.md")
	if err := os.WriteFile(specFile, []byte("# Deploy\n"), 0o600); err != nil {
		t.Fatalf("failed to create spec file: %v", err)
	}

	// Create valid Go file
	validFile := filepath.Join(tmpDir, "main.go")
	validContent := []byte(`package main

// Feature: CLI_DEPLOY
// Spec: spec/commands/deploy.md

func main() {
}
`)
	if err := os.WriteFile(validFile, validContent, 0o600); err != nil {
		t.Fatalf("failed to create valid.go: %v", err)
	}

	// Create testdata directory with invalid reference (should be ignored)
	testdataDir := filepath.Join(tmpDir, "testdata")
	if err := os.MkdirAll(testdataDir, 0o700); err != nil {
		t.Fatalf("failed to create testdata dir: %v", err)
	}

	testdataFile := filepath.Join(testdataDir, "fixture.go")
	testdataContent := []byte(`package testdata

// Spec: test/invalid.md
// Spec: spec/commands/nonexistent.md

func test() {
}
`)
	if err := os.WriteFile(testdataFile, testdataContent, 0o600); err != nil {
		t.Fatalf("failed to create testdata file: %v", err)
	}

	// Run should succeed because:
	// 1. main.go has a valid reference to an existing spec file
	// 2. testdata files are ignored (even though they have invalid references)
	if err := run(tmpDir); err != nil {
		t.Errorf("run(%q) = %v, want nil (testdata should be ignored)", tmpDir, err)
	}
}
