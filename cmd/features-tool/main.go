// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"stagecraft/internal/tools/features"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: features-tool <graph|impact> [args]")
	}

	switch os.Args[1] {
	case "graph":
		runGraph()
	case "impact":
		runImpact(os.Args[2:])
	default:
		log.Fatalf("unknown subcommand: %s", os.Args[1])
	}
}

func runGraph() {
	fs := flag.NewFlagSet("graph", flag.ExitOnError)
	featuresPath := fs.String("features", "spec/features.yaml", "path to features.yaml")
	dot := fs.Bool("dot", false, "output in DOT format")
	_ = fs.Parse(os.Args[2:])

	g, err := features.LoadGraph(*featuresPath)
	if err != nil {
		log.Fatalf("failed to load graph: %v", err)
	}

	if err := features.ValidateDAG(g); err != nil {
		log.Fatalf("feature DAG invalid: %v", err)
	}

	if *dot {
		fmt.Println(features.ToDOT(g))
	} else {
		fmt.Printf("✓ Feature dependency graph is valid (acyclic)\n")
		fmt.Printf("  Total features: %d\n", len(g.Nodes))
	}
}

func runImpact(args []string) {
	fs := flag.NewFlagSet("impact", flag.ExitOnError)
	featuresPath := fs.String("features", "spec/features.yaml", "path to features.yaml")
	featureID := fs.String("feature", "", "feature id to analyze")
	_ = fs.Parse(args)

	if *featureID == "" {
		log.Fatalf("-feature is required")
	}

	g, err := features.LoadGraph(*featuresPath)
	if err != nil {
		log.Fatalf("failed to load graph: %v", err)
	}

	impacted := features.Impact(g, *featureID)
	if len(impacted) == 0 {
		fmt.Printf("No features depend on %s\n", *featureID)
	} else {
		fmt.Printf("Features that depend on %s:\n", *featureID)
		for _, id := range impacted {
			fmt.Printf("  - %s\n", id)
		}
	}
}
