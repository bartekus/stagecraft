// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package specvscli

import (
	"fmt"
	"strings"

	"stagecraft/internal/tools/cliintrospect"
	"stagecraft/internal/tools/specschema"
)

// DiffResult represents the result of comparing specs to CLI implementation.
type DiffResult struct {
	CommandName string
	Errors      []string
	Warnings    []string
}

// CompareFlags compares flags from a spec to flags from CLI introspection.
func CompareFlags(specFlags []specschema.CliFlag, cliFlags []cliintrospect.FlagInfo, commandName string) DiffResult {
	result := DiffResult{
		CommandName: commandName,
		Errors:      []string{},
		Warnings:    []string{},
	}

	// Build maps for easier lookup
	specFlagMap := make(map[string]specschema.CliFlag)
	for _, flag := range specFlags {
		// Normalize flag name (remove leading dashes)
		name := strings.TrimPrefix(flag.Name, "--")
		name = strings.TrimPrefix(name, "-")
		specFlagMap[name] = flag
	}

	cliFlagMap := make(map[string]cliintrospect.FlagInfo)
	for _, flag := range cliFlags {
		cliFlagMap[flag.Name] = flag
	}

	// Check: spec declares flag that doesn't exist in CLI
	for name, specFlag := range specFlagMap {
		if _, exists := cliFlagMap[name]; !exists {
			result.Errors = append(result.Errors, fmt.Sprintf("spec declares flag %q that does not exist in CLI", specFlag.Name))
		}
	}

	// Check: CLI has flag that's not in spec
	// Note: We skip global/persistent flags that are inherited from root
	for name, cliFlag := range cliFlagMap {
		if cliFlag.Persistent {
			// Skip persistent flags - they're inherited from root
			continue
		}
		if _, exists := specFlagMap[name]; !exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("CLI has flag %q that is not documented in spec", name))
		}
	}

	return result
}

// CompareAllCommands compares all CLI commands to their corresponding specs.
func CompareAllCommands(specs []specschema.Spec, cliCommands []cliintrospect.CommandInfo) []DiffResult {
	var results []DiffResult

	// Build spec lookup by feature ID
	specMap := make(map[string]specschema.Spec)
	for _, spec := range specs {
		specMap[spec.Frontmatter.Feature] = spec
	}

	// Compare each CLI command
	for _, cmd := range cliCommands {
		// Try to find matching spec
		// Command use is typically like "build", "deploy", etc.
		// Feature ID is typically like "CLI_BUILD", "CLI_DEPLOY"
		featureID := inferFeatureID(cmd.Use)
		spec, hasSpec := specMap[featureID]

		if !hasSpec {
			// No spec for this command - skip
			continue
		}

		// Compare flags
		diff := CompareFlags(spec.Frontmatter.Inputs.Flags, cmd.Flags, cmd.Use)
		if len(diff.Errors) > 0 || len(diff.Warnings) > 0 {
			results = append(results, diff)
		}

		// Recursively check subcommands
		for _, subcmd := range cmd.Subcommands {
			subResults := compareSubcommands(subcmd, specMap)
			results = append(results, subResults...)
		}
	}

	return results
}

// compareSubcommands recursively compares subcommands.
func compareSubcommands(cmd cliintrospect.CommandInfo, specMap map[string]specschema.Spec) []DiffResult {
	var results []DiffResult

	featureID := inferFeatureID(cmd.Use)
	if spec, hasSpec := specMap[featureID]; hasSpec {
		diff := CompareFlags(spec.Frontmatter.Inputs.Flags, cmd.Flags, cmd.Use)
		if len(diff.Errors) > 0 || len(diff.Warnings) > 0 {
			results = append(results, diff)
		}
	}

	for _, subcmd := range cmd.Subcommands {
		subResults := compareSubcommands(subcmd, specMap)
		results = append(results, subResults...)
	}

	return results
}

// inferFeatureID attempts to infer a feature ID from a command name.
// For example: "build" -> "CLI_BUILD", "deploy" -> "CLI_DEPLOY"
func inferFeatureID(commandUse string) string {
	// Convert to uppercase and prefix with CLI_
	upper := strings.ToUpper(commandUse)
	return "CLI_" + upper
}
