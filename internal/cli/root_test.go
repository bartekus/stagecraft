// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"stagecraft/internal/cli/commands"
	"stagecraft/pkg/config"
)

// Feature: ARCH_OVERVIEW
// Spec: spec/overview.md
func TestNewRootCommand_HasExpectedBasics(t *testing.T) {
	cmd := NewRootCommand()

	if cmd.Use != "stagecraft" {
		t.Fatalf("expected Use to be 'stagecraft', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}

	// Ensure version subcommand exists
	versionCmd, _, err := cmd.Find([]string{"version"})
	if err != nil {
		t.Fatalf("expected to find 'version' subcommand, got error: %v", err)
	}

	if versionCmd.Use != "version" {
		t.Fatalf("expected 'version' command Use to be 'version', got %q", versionCmd.Use)
	}
}

func TestVersionCommand_PrintsVersion(t *testing.T) {
	cmd := NewRootCommand()

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Execute 'stagecraft version'
	cmd.SetArgs([]string{"version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("expected no error executing 'version' command, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Stagecraft version") {
		t.Fatalf("expected output to contain 'Stagecraft version', got: %q", out)
	}
}

// Feature: CLI_GLOBAL_FLAGS
// Spec: spec/core/global-flags.md

// parseFlagsForTesting parses flags for a command in tests.
func parseFlagsForTesting(cmd *cobra.Command, args []string) error {
	cmd.SetArgs(args)
	return cmd.ParseFlags(args)
}

func TestResolveFlags_CommandLineFlags(t *testing.T) {
	cmd := NewRootCommand()
	if err := parseFlagsForTesting(cmd, []string{"--env", "staging", "--config", "/custom/path.yml", "--verbose", "--dry-run", "version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	flags, err := commands.ResolveFlags(cmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	if flags.Env != "staging" {
		t.Errorf("expected Env to be 'staging', got %q", flags.Env)
	}
	if flags.Config != "/custom/path.yml" {
		t.Errorf("expected Config to be '/custom/path.yml', got %q", flags.Config)
	}
	if !flags.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !flags.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestResolveFlags_EnvironmentVariables(t *testing.T) {
	// Save original env values
	origEnv := os.Getenv("STAGECRAFT_ENV")
	origConfig := os.Getenv("STAGECRAFT_CONFIG")
	origVerbose := os.Getenv("STAGECRAFT_VERBOSE")
	origDryRun := os.Getenv("STAGECRAFT_DRY_RUN")

	// Set environment variables
	os.Setenv("STAGECRAFT_ENV", "prod")
	os.Setenv("STAGECRAFT_CONFIG", "/env/path.yml")
	os.Setenv("STAGECRAFT_VERBOSE", "true")
	os.Setenv("STAGECRAFT_DRY_RUN", "true")
	defer func() {
		// Restore original values
		if origEnv != "" {
			os.Setenv("STAGECRAFT_ENV", origEnv)
		} else {
			os.Unsetenv("STAGECRAFT_ENV")
		}
		if origConfig != "" {
			os.Setenv("STAGECRAFT_CONFIG", origConfig)
		} else {
			os.Unsetenv("STAGECRAFT_CONFIG")
		}
		if origVerbose != "" {
			os.Setenv("STAGECRAFT_VERBOSE", origVerbose)
		} else {
			os.Unsetenv("STAGECRAFT_VERBOSE")
		}
		if origDryRun != "" {
			os.Setenv("STAGECRAFT_DRY_RUN", origDryRun)
		} else {
			os.Unsetenv("STAGECRAFT_DRY_RUN")
		}
	}()

	cmd := NewRootCommand()
	if err := parseFlagsForTesting(cmd, []string{"version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	flags, err := commands.ResolveFlags(cmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	if flags.Env != "prod" {
		t.Errorf("expected Env to be 'prod', got %q", flags.Env)
	}
	if flags.Config != "/env/path.yml" {
		t.Errorf("expected Config to be '/env/path.yml', got %q", flags.Config)
	}
	if !flags.Verbose {
		t.Error("expected Verbose to be true")
	}
	if !flags.DryRun {
		t.Error("expected DryRun to be true")
	}
}

func TestResolveFlags_Defaults(t *testing.T) {
	// Ensure no env vars are set
	os.Unsetenv("STAGECRAFT_ENV")
	os.Unsetenv("STAGECRAFT_CONFIG")
	os.Unsetenv("STAGECRAFT_VERBOSE")
	os.Unsetenv("STAGECRAFT_DRY_RUN")

	cmd := NewRootCommand()
	if err := parseFlagsForTesting(cmd, []string{"version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	flags, err := commands.ResolveFlags(cmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	if flags.Env != "dev" {
		t.Errorf("expected Env default to be 'dev', got %q", flags.Env)
	}
	if flags.Config == "" {
		t.Error("expected Config to have default value")
	}
	if flags.Verbose {
		t.Error("expected Verbose default to be false")
	}
	if flags.DryRun {
		t.Error("expected DryRun default to be false")
	}
}

func TestResolveFlags_Precedence(t *testing.T) {
	// Set environment variables
	os.Setenv("STAGECRAFT_ENV", "env-value")
	os.Setenv("STAGECRAFT_VERBOSE", "true")
	defer func() {
		os.Unsetenv("STAGECRAFT_ENV")
		os.Unsetenv("STAGECRAFT_VERBOSE")
	}()

	// Command-line flags should override environment variables
	cmd := NewRootCommand()
	// For boolean flags, we test that flag=true overrides env=false
	// Note: --verbose=false doesn't work the same way in Cobra, so we test the positive case
	if err := parseFlagsForTesting(cmd, []string{"--env", "flag-value", "--verbose", "version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	flags, err := commands.ResolveFlags(cmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	if flags.Env != "flag-value" {
		t.Errorf("expected Env to be 'flag-value' (from flag), got %q", flags.Env)
	}
	if !flags.Verbose {
		t.Error("expected Verbose to be true (from flag), got false")
	}
}

func TestResolveFlags_EnvValidation(t *testing.T) {
	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"dev":     {Driver: "local"},
			"staging": {Driver: "local"},
		},
	}

	cmd := NewRootCommand()
	if err := parseFlagsForTesting(cmd, []string{"--env", "invalid-env", "version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	_, err := commands.ResolveFlags(cmd, cfg)
	if err == nil {
		t.Fatal("expected ResolveFlags() to return error for invalid environment")
	}

	if !strings.Contains(err.Error(), "invalid environment") {
		t.Errorf("expected error to mention 'invalid environment', got: %v", err)
	}
}

// Note: Config file validation is now done by commands that use the config,
// not in ResolveFlags, so this test is removed.

func TestResolveFlags_PersistentFlagsInherited(t *testing.T) {
	cmd := NewRootCommand()
	devCmd, _, err := cmd.Find([]string{"dev"})
	if err != nil {
		t.Fatalf("expected to find 'dev' subcommand, got error: %v", err)
	}

	// Parse flags on the subcommand (which inherits persistent flags from root)
	if err := parseFlagsForTesting(devCmd, []string{"--env", "staging", "--verbose"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	// Flags should be available on subcommand
	flags, err := commands.ResolveFlags(devCmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	if flags.Env != "staging" {
		t.Errorf("expected Env to be 'staging', got %q", flags.Env)
	}
	if !flags.Verbose {
		t.Error("expected Verbose to be true")
	}
}

func TestResolveFlags_BoolEnvParsing(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true", "true", true},
		{"True", "True", true},
		{"TRUE", "TRUE", true},
		{"1", "1", true},
		{"false", "false", false},
		{"False", "False", false},
		{"FALSE", "FALSE", false},
		{"0", "0", false},
		{"empty", "", false},
		{"invalid", "invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("STAGECRAFT_VERBOSE", tc.envValue)
			defer os.Unsetenv("STAGECRAFT_VERBOSE")

			cmd := NewRootCommand()
			if err := parseFlagsForTesting(cmd, []string{"version"}); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			flags, err := commands.ResolveFlags(cmd, nil)
			if err != nil {
				t.Fatalf("ResolveFlags() returned error: %v", err)
			}

			if flags.Verbose != tc.expected {
				t.Errorf("expected Verbose to be %v for env value %q, got %v", tc.expected, tc.envValue, flags.Verbose)
			}
		})
	}
}

func TestResolveFlags_ConfigDefaultPath(t *testing.T) {
	os.Unsetenv("STAGECRAFT_CONFIG")

	cmd := NewRootCommand()
	if err := parseFlagsForTesting(cmd, []string{"version"}); err != nil {
		t.Fatalf("failed to parse flags: %v", err)
	}

	flags, err := commands.ResolveFlags(cmd, nil)
	if err != nil {
		t.Fatalf("ResolveFlags() returned error: %v", err)
	}

	// Default config path should be relative to current directory
	expected := filepath.Join(".", "stagecraft.yml")
	if flags.Config != expected {
		t.Errorf("expected Config default to be %q, got %q", expected, flags.Config)
	}
}
