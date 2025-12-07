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

	"stagecraft/internal/tools/specschema"
)

func main() {
	root := flag.String("root", "spec", "root directory containing spec files")
	featuresPath := flag.String("features", "spec/features.yaml", "path to features.yaml")
	checkIntegrity := flag.Bool("check-integrity", false, "also validate features.yaml ↔ spec file integrity")
	flag.Parse()

	specs, err := specschema.LoadAllSpecs(*root)
	if err != nil {
		log.Fatalf("failed to load specs: %v", err)
	}

	if len(specs) == 0 {
		fmt.Fprintf(os.Stderr, "warning: no spec files found in %s\n", *root)
		os.Exit(0)
	}

	if err := specschema.ValidateAll(specs); err != nil {
		log.Fatalf("spec validation failed: %v", err)
	}

	if *checkIntegrity {
		if err := specschema.ValidateSpecIntegrity(*featuresPath, *root); err != nil {
			log.Fatalf("spec integrity validation failed: %v", err)
		}
		fmt.Printf("✓ Spec integrity check passed\n")
	}

	fmt.Printf("✓ Validated %d spec file(s)\n", len(specs))
	os.Exit(0)
}
