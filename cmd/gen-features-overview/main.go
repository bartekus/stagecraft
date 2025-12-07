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
	"log"
	"os"

	"stagecraft/internal/tools/docs"
)

func main() {
	featuresPath := flag.String("features", "spec/features.yaml", "path to features.yaml")
	specRoot := flag.String("spec-root", "spec", "root directory containing spec files")
	outPath := flag.String("out", "docs/features/OVERVIEW.md", "output path for overview document")
	flag.Parse()

	if err := docs.GenerateFeatureOverview(*featuresPath, *specRoot, *outPath); err != nil {
		log.Fatalf("failed to generate feature overview: %v", err)
	}

	log.Printf("✓ Generated feature overview at %s", *outPath)
	os.Exit(0)
}
