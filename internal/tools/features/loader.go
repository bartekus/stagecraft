// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package features

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadGraph loads features.yaml and constructs a dependency graph.
func LoadGraph(path string) (*Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read features.yaml: %w", err)
	}

	var featuresYAML FeaturesYAML
	if err := yaml.Unmarshal(data, &featuresYAML); err != nil {
		return nil, fmt.Errorf("failed to parse features.yaml: %w", err)
	}

	graph := NewGraph()

	// Add all nodes
	for i := range featuresYAML.Features {
		graph.AddNode(&featuresYAML.Features[i])
	}

	// Add edges (dependencies)
	for _, feature := range featuresYAML.Features {
		for _, depID := range feature.DependsOn {
			// Verify dependency exists
			if _, exists := graph.Nodes[depID]; !exists {
				return nil, fmt.Errorf("feature %q depends on unknown feature %q", feature.ID, depID)
			}
			// Add edge: feature depends on depID, so depID -> feature
			graph.AddEdge(feature.ID, depID)
		}
	}

	return graph, nil
}
