// SPDX-License-Identifier: AGPL-3.0-or-later

//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

// Command feature-map-check runs governance validation for the Feature Mapping
// Invariant defined by GOV_V1_CORE Phase 4.
//
// It is intended to be invoked from scripts/run-all-checks.sh and CI; it should
// not introduce non-deterministic behavior.
package main

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md

import (
	"context"
	"flag"
	"fmt"
	"os"

	"stagecraft/internal/tools/features"
)

func main() {
	var (
		rootDir      string
		featuresPath string
	)

	flag.StringVar(&rootDir, "root", ".", "root directory to scan")
	flag.StringVar(&featuresPath, "features", "spec/features.yaml", "path to spec/features.yaml (relative to root)")
	flag.Parse()

	ctx := context.Background()

	r := &features.Runner{
		RootDir:      rootDir,
		FeaturesPath: featuresPath,
		Out:          os.Stdout,
	}

	if err := r.Run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "feature-map-check:", err)
		os.Exit(1)
	}
}
