// SPDX-License-Identifier: AGPL-3.0-or-later

//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

// Command feature-dashboard prints a high-level governance summary of feature health.
package main

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md

import (
	"context"
	"flag"
	"log"

	"stagecraft/internal/tools/features"
)

func main() {
	root := flag.String("root", ".", "root directory to scan")
	featuresPath := flag.String("features", "spec/features.yaml", "path to spec/features.yaml")
	flag.Parse()

	r := &features.GovernanceDashboardRunner{
		RootDir:      *root,
		FeaturesPath: *featuresPath,
	}

	if err := r.Run(context.Background()); err != nil {
		log.Fatalf("dashboard failed: %v", err)
	}
}
