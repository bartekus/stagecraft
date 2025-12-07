// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package main provides the spec-vs-cli tool for comparing specs to CLI implementation.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"stagecraft/internal/cli"
	"stagecraft/internal/tools/cliintrospect"
	"stagecraft/internal/tools/specschema"
	"stagecraft/internal/tools/specvscli"
)

func main() {
	specRoot := flag.String("spec-root", "spec", "root directory containing spec files")
	warnOnly := flag.Bool("warn-only", false, "treat mismatches as warnings instead of errors")
	flag.Parse()

	// Load specs
	specs, err := specschema.LoadAllSpecs(*specRoot)
	if err != nil {
		log.Fatalf("failed to load specs: %v", err)
	}

	// Introspect CLI
	rootCmd := cli.NewRootCommand()
	cliCommands := cliintrospect.Introspect(rootCmd)

	// Compare
	results := specvscli.CompareAllCommands(specs, cliCommands)

	// Report results
	hasErrors := false
	hasWarnings := false

	for _, result := range results {
		if len(result.Errors) > 0 {
			hasErrors = true
			fmt.Printf("ERROR: Command %q:\n", result.CommandName)
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
		if len(result.Warnings) > 0 {
			hasWarnings = true
			fmt.Printf("WARNING: Command %q:\n", result.CommandName)
			for _, warn := range result.Warnings {
				fmt.Printf("  - %s\n", warn)
			}
		}
	}

	if hasErrors {
		if *warnOnly {
			fmt.Fprintf(os.Stderr, "\n⚠ Flag alignment issues found (warn-only mode)\n")
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "\n✗ Flag alignment failed\n")
		os.Exit(1)
	}

	if hasWarnings {
		fmt.Printf("\n⚠ Flag alignment warnings (non-blocking)\n")
	} else {
		fmt.Printf("✓ Flag alignment check passed\n")
	}

	os.Exit(0)
}
