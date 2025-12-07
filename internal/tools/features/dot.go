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
	"strings"
)

// ToDOT generates a DOT format representation of the feature dependency graph.
func ToDOT(g *Graph) string {
	var sb strings.Builder
	sb.WriteString("digraph feature_dependencies {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")

	// Add nodes with status-based colors
	for id, node := range g.Nodes {
		color := getStatusColor(node.Status)
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\\n[%s]\" fillcolor=\"%s\" style=filled];\n",
			id, id, node.Status, color))
	}

	sb.WriteString("\n")

	// Add edges (dependencies)
	for _, node := range g.Nodes {
		for _, depID := range node.DependsOn {
			sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\";\n", depID, node.ID))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// getStatusColor returns a color for a feature status.
func getStatusColor(status string) string {
	switch status {
	case "done":
		return "lightgreen"
	case "wip":
		return "lightyellow"
	case "todo":
		return "lightgray"
	default:
		return "white"
	}
}
