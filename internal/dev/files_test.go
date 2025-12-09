// SPDX-License-Identifier: AGPL-3.0-or-later

package dev

import (
	"os"
	"path/filepath"
	"testing"

	devcompose "stagecraft/internal/dev/compose"
	devtraefik "stagecraft/internal/dev/traefik"

	corecompose "stagecraft/internal/compose"
	"stagecraft/pkg/config"
)

func TestWriteFiles_WritesAllArtifacts(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")

	// Minimal config and services; you already know this pattern from topology tests.
	cfg := &config.Config{}
	backend := &devcompose.ServiceDefinition{Name: "backend"}
	frontend := &devcompose.ServiceDefinition{Name: "frontend"}
	traefikSvc := &devcompose.ServiceDefinition{Name: "traefik"}

	builder := NewBuilder(nil, nil)

	topo, err := builder.Build(
		cfg,
		Domains{
			Frontend: "app.localdev.test",
			Backend:  "api.localdev.test",
		},
		backend,
		frontend,
		traefikSvc,
		true,
		"",
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	files, err := WriteFiles(devDir, topo)
	if err != nil {
		t.Fatalf("WriteFiles() error = %v, want nil", err)
	}

	// Verify paths are inside devDir.
	if got, wantPrefix := files.ComposePath, devDir+string(os.PathSeparator); len(got) <= len(wantPrefix) || got[:len(wantPrefix)] != wantPrefix {
		t.Errorf("ComposePath = %q, want prefix %q", got, wantPrefix)
	}

	// Verify files exist and are non-empty.
	for _, path := range []string{
		files.ComposePath,
		files.TraefikStaticPath,
		files.TraefikDynamicPath,
	} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("expected file %q to exist: %v", path, err)
		}
		if info.Size() == 0 {
			t.Errorf("expected file %q to be non-empty", path)
		}
	}
}

// Optionally, a determinism test: write twice and compare bytes.
func TestWriteFiles_DeterministicOutput(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	devDir1 := filepath.Join(tmpDir, "run1", ".stagecraft", "dev")
	devDir2 := filepath.Join(tmpDir, "run2", ".stagecraft", "dev")

	cfg := &config.Config{}
	backend := &devcompose.ServiceDefinition{Name: "backend"}
	frontend := &devcompose.ServiceDefinition{Name: "frontend"}
	traefikSvc := &devcompose.ServiceDefinition{Name: "traefik"}

	builder := NewBuilder(nil, nil)

	topo, err := builder.Build(
		cfg,
		Domains{
			Frontend: "app.localdev.test",
			Backend:  "api.localdev.test",
		},
		backend,
		frontend,
		traefikSvc,
		true,
		"",
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	files1, err := WriteFiles(devDir1, topo)
	if err != nil {
		t.Fatalf("WriteFiles(run1) error = %v", err)
	}
	files2, err := WriteFiles(devDir2, topo)
	if err != nil {
		t.Fatalf("WriteFiles(run2) error = %v", err)
	}

	compare := func(p1, p2 string) {
		b1, err := os.ReadFile(p1)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", p1, err)
		}
		b2, err := os.ReadFile(p2)
		if err != nil {
			t.Fatalf("ReadFile(%q) error = %v", p2, err)
		}
		if string(b1) != string(b2) {
			t.Errorf("files %q and %q differ", p1, p2)
		}
	}

	compare(files1.ComposePath, files2.ComposePath)
	compare(files1.TraefikStaticPath, files2.TraefikStaticPath)
	compare(files1.TraefikDynamicPath, files2.TraefikDynamicPath)
}

func TestWriteFiles_NilTopologyFails(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")

	_, err := WriteFiles(devDir, nil)
	if err == nil {
		t.Fatalf("WriteFiles(nil) error = nil, want non-nil")
	}
}

func TestWriteFiles_ValidatesComposeExists(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")

	topo := &Topology{
		Compose: nil,
		Traefik: &devtraefik.Config{
			Static:  &devtraefik.StaticConfig{},
			Dynamic: &devtraefik.DynamicConfig{},
		},
	}

	_, err := WriteFiles(devDir, topo)
	if err == nil {
		t.Fatalf("WriteFiles(topology with nil Compose) error = nil, want non-nil")
	}
}

func TestWriteFiles_ValidatesTraefikExists(t *testing.T) {
	t.Helper()

	tmpDir := t.TempDir()
	devDir := filepath.Join(tmpDir, ".stagecraft", "dev")

	topo := &Topology{
		Compose: corecompose.NewComposeFile(map[string]any{}),
		Traefik: nil,
	}

	_, err := WriteFiles(devDir, topo)
	if err == nil {
		t.Fatalf("WriteFiles(topology with nil Traefik) error = nil, want non-nil")
	}
}
