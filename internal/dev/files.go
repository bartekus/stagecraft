// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Feature: CLI_DEV
Specs:
  - spec/commands/dev.md
Docs:
  - docs/engine/analysis/CLI_DEV.md
  - docs/engine/outlines/CLI_DEV_IMPLEMENTATION_OUTLINE.md
*/

package dev

import (
	"fmt"
	"os"
	"path/filepath"
)

// DevFiles describes the paths of generated dev config files.
//
//nolint:revive // DevFiles is clear and descriptive in the dev package context
type DevFiles struct {
	ComposePath        string
	TraefikStaticPath  string
	TraefikDynamicPath string
}

// WriteFiles writes the dev topology artifacts to disk under devDir.
//
// devDir is usually "<project-root>/.stagecraft/dev".
func WriteFiles(devDir string, topo *Topology) (DevFiles, error) {
	if topo == nil {
		return DevFiles{}, fmt.Errorf("dev files: topology is nil")
	}
	if topo.Compose == nil {
		return DevFiles{}, fmt.Errorf("dev files: compose model is nil")
	}
	// Traefik config is optional when --no-traefik is used
	if topo.Traefik != nil && (topo.Traefik.Static == nil || topo.Traefik.Dynamic == nil) {
		return DevFiles{}, fmt.Errorf("dev files: traefik config is incomplete")
	}

	composePath := filepath.Join(devDir, "compose.yaml")

	// Ensure directories exist.
	// #nosec G301 -- dev directory needs 0755 for docker compose access
	if err := os.MkdirAll(devDir, 0o755); err != nil {
		return DevFiles{}, fmt.Errorf("dev files: create dev dir: %w", err)
	}

	// Compose YAML.
	composeBytes, err := topo.Compose.ToYAML()
	if err != nil {
		return DevFiles{}, fmt.Errorf("dev files: marshal compose yaml: %w", err)
	}
	// #nosec G306 -- compose.yaml needs 0644 for docker compose to read it
	if err := os.WriteFile(composePath, composeBytes, 0o644); err != nil {
		return DevFiles{}, fmt.Errorf("dev files: write compose yaml: %w", err)
	}

	files := DevFiles{
		ComposePath: composePath,
	}

	// Traefik config files (only if Traefik is enabled)
	if topo.Traefik != nil {
		traefikDir := filepath.Join(devDir, "traefik")
		staticPath := filepath.Join(traefikDir, "traefik-static.yaml")
		dynamicPath := filepath.Join(traefikDir, "traefik-dynamic.yaml")

		// #nosec G301 -- traefik directory needs 0755 for docker compose access
		if err := os.MkdirAll(traefikDir, 0o755); err != nil {
			return DevFiles{}, fmt.Errorf("dev files: create traefik dir: %w", err)
		}

		// Traefik static YAML.
		staticBytes, err := topo.Traefik.ToYAMLStatic()
		if err != nil {
			return DevFiles{}, fmt.Errorf("dev files: marshal traefik static yaml: %w", err)
		}
		// #nosec G306 -- traefik config files need 0644 for docker compose to read them
		if err := os.WriteFile(staticPath, staticBytes, 0o644); err != nil {
			return DevFiles{}, fmt.Errorf("dev files: write traefik static yaml: %w", err)
		}

		// Traefik dynamic YAML.
		dynamicBytes, err := topo.Traefik.ToYAMLDynamic()
		if err != nil {
			return DevFiles{}, fmt.Errorf("dev files: marshal traefik dynamic yaml: %w", err)
		}
		// #nosec G306 -- traefik config files need 0644 for docker compose to read them
		if err := os.WriteFile(dynamicPath, dynamicBytes, 0o644); err != nil {
			return DevFiles{}, fmt.Errorf("dev files: write traefik dynamic yaml: %w", err)
		}

		files.TraefikStaticPath = staticPath
		files.TraefikDynamicPath = dynamicPath
	}

	return files, nil
}
