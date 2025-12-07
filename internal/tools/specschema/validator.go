// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specschema

import (
	"fmt"
	"strings"
)

// ValidateAll validates all specs and returns an error if any validation fails.
func ValidateAll(specs []Spec) error {
	var errors []string

	for _, spec := range specs {
		if err := ValidateSpec(spec); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", spec.Path, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}

// ValidateSpec validates a single spec's frontmatter.
func ValidateSpec(spec Spec) error {
	fm := spec.Frontmatter

	// Check required fields
	if fm.Feature == "" {
		return fmt.Errorf("missing required field: feature")
	}

	if fm.Version == "" {
		return fmt.Errorf("missing required field: version")
	}

	if fm.Status == "" {
		return fmt.Errorf("missing required field: status")
	}

	if fm.Domain == "" {
		return fmt.Errorf("missing required field: domain")
	}

	// Validate status enum
	validStatuses := map[string]bool{
		"todo": true,
		"wip":  true,
		"done": true,
	}
	if !validStatuses[fm.Status] {
		return fmt.Errorf("invalid status: %s (must be one of: todo, wip, done)", fm.Status)
	}

	// Validate feature matches filename
	expectedID := ExpectedFeatureIDFromPath(spec.Path)
	if fm.Feature != expectedID {
		return fmt.Errorf("feature ID mismatch: frontmatter has %q but filename suggests %q", fm.Feature, expectedID)
	}

	// Validate flags if present
	for i, flag := range fm.Inputs.Flags {
		if flag.Name == "" {
			return fmt.Errorf("flag[%d]: name is required", i)
		}
	}

	// Validate exit codes if present
	for name, code := range fm.Outputs.ExitCodes {
		if code < 0 {
			return fmt.Errorf("exit code %q has negative value: %d", name, code)
		}
	}

	return nil
}
