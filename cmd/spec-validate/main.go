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

	fmt.Printf("✓ Validated %d spec file(s)\n", len(specs))
	os.Exit(0)
}
