// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"stagecraft/internal/tools/features"
)

// ValidateSpecIntegrity validates that features.yaml and spec files are in sync.
// It checks:
//   - Every feature in features.yaml has a corresponding spec file
//   - Spec files match their feature IDs
func ValidateSpecIntegrity(featuresPath, specRoot string) error {
	// Load features
	graph, err := features.LoadGraph(featuresPath)
	if err != nil {
		return fmt.Errorf("failed to load features: %w", err)
	}

	// Load all specs
	specs, err := LoadAllSpecs(specRoot)
	if err != nil {
		return fmt.Errorf("failed to load specs: %w", err)
	}

	// Build spec lookup by feature ID
	specMap := make(map[string]Spec)
	for _, spec := range specs {
		specMap[spec.Frontmatter.Feature] = spec
	}

	var errors []string

	// Check: every feature in features.yaml should have a spec file
	for id, node := range graph.Nodes {
		// Skip features without spec path
		if node.Spec == "" {
			continue
		}

		// Check if spec file exists
		specPath := node.Spec
		if !strings.HasPrefix(specPath, "spec/") {
			specPath = filepath.Join("spec", specPath)
		}

		// Check if file exists
		if _, err := os.Stat(specPath); err != nil {
			if os.IsNotExist(err) {
				errors = append(errors, fmt.Sprintf("feature %q references spec file %q that does not exist", id, node.Spec))
			} else {
				errors = append(errors, fmt.Sprintf("feature %q references spec file %q that cannot be accessed: %v", id, node.Spec, err))
			}
			continue
		}

		// Check if spec file has matching feature ID
		if spec, exists := specMap[id]; exists {
			if spec.Frontmatter.Feature != id {
				errors = append(errors, fmt.Sprintf("feature %q spec file has mismatched feature ID %q", id, spec.Frontmatter.Feature))
			}
		} else {
			// Spec file exists but doesn't have frontmatter with matching ID
			// Try to load it
			spec, err := LoadSpec(specPath)
			if err != nil {
				errors = append(errors, fmt.Sprintf("feature %q spec file %q exists but cannot be loaded: %v", id, node.Spec, err))
				continue
			}
			if spec.Frontmatter.Feature != id {
				errors = append(errors, fmt.Sprintf("feature %q spec file %q has feature ID %q (expected %q)", id, node.Spec, spec.Frontmatter.Feature, id))
			}
		}
	}

	// Check: spec files should be referenced in features.yaml (warning only)
	// This is less strict - we allow orphaned spec files
	for _, spec := range specs {
		if _, exists := graph.Nodes[spec.Frontmatter.Feature]; !exists {
			// This is a warning, not an error - orphaned specs are allowed
			// But we could add a flag to make this strict
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("spec integrity validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}
