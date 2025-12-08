// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package main provides the gen-implementation-status tool for generating implementation status documentation.
package main

import (
	"flag"
	"log"
	"os"

	"stagecraft/internal/tools/docs"
)

func main() {
	featuresPath := flag.String("features", "spec/features.yaml", "path to features.yaml")
	outPath := flag.String("out", "docs/engine/status/implementation-status.md", "output path for implementation status document")
	flag.Parse()

	if err := docs.GenerateImplementationStatus(*featuresPath, *outPath); err != nil {
		log.Fatalf("failed to generate implementation status: %v", err)
	}

	log.Printf("✓ Generated implementation status at %s", *outPath)
	os.Exit(0)
}
