// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package featuretrace defines the data model for feature traceability reports.
//
// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: docs/context-handoff/COMMIT_DISCIPLINE_PHASE3C.md
package featuretrace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ScanConfig contains configuration for scanning feature presence.
type ScanConfig struct {
	RootDir string
}

// ScanFeaturePresence scans the repository tree and returns feature presence information.
// The result is sorted lexicographically by FeatureID.
func ScanFeaturePresence(cfg ScanConfig) ([]FeaturePresence, error) {
	// TODO: Implement deterministic repository traversal
	// - Walk directory tree starting at cfg.RootDir
	// - Use os.ReadDir with explicit sort (never raw FS order)
	// - Extract Feature IDs from:
	//   - Spec headers: // Feature: FEATURE_ID
	//   - Implementation headers: // Feature: FEATURE_ID
	//   - Test headers: // Feature: FEATURE_ID
	// - Build FeaturePresence structs
	// - Sort by FeatureID lexicographically
	// - Return sorted slice

	presenceMap := make(map[string]*FeaturePresence)

	err := filepath.Walk(cfg.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and .git
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "testdata") {
			return filepath.SkipDir
		}

		// TODO: Process files to extract Feature IDs
		// - Read file content
		// - Look for "// Feature: FEATURE_ID" pattern
		// - Determine file type (spec, implementation, test)
		// - Update presenceMap

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scanning repository: %w", err)
	}

	// Convert map to sorted slice
	result := make([]FeaturePresence, 0, len(presenceMap))
	for _, fp := range presenceMap {
		result = append(result, *fp)
	}

	// Sort by FeatureID lexicographically
	sort.Slice(result, func(i, j int) bool {
		return result[i].FeatureID < result[j].FeatureID
	})

	return result, nil
}

// extractFeatureIDFromFile extracts Feature ID from a file's header comment.
// Returns empty string if no Feature ID is found.
func extractFeatureIDFromFile(filePath string) (string, error) {
	// TODO: Implement feature ID extraction
	// - Read file (first N lines or until non-comment line)
	// - Look for pattern: // Feature: FEATURE_ID
	// - Return FEATURE_ID or empty string
	return "", fmt.Errorf("not implemented")
}

// isSpecFile determines if a file path represents a spec file.
func isSpecFile(path string) bool {
	// TODO: Implement spec file detection
	// - Check if path is under spec/ directory
	// - Check file extension (.md)
	return false
}

// isTestFile determines if a file path represents a test file.
func isTestFile(path string) bool {
	// TODO: Implement test file detection
	// - Check if filename ends with _test.go
	return false
}

// isImplementationFile determines if a file path represents an implementation file.
func isImplementationFile(path string) bool {
	// TODO: Implement implementation file detection
	// - Check if it's a .go file
	// - Check if it's NOT a test file
	// - Check if it's NOT under testdata/
	return false
}
