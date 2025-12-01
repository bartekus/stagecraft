// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package cli

import (
	"bytes"
	"strings"
	"testing"
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
