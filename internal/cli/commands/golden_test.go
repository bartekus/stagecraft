// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

// updateGolden is a flag to update golden files during development.
// Usage: go test -update ./internal/cli/commands
var updateGolden = flag.Bool("update", false, "update golden files")

// readGoldenFile reads a golden file, or returns empty string if it doesn't exist.
func readGoldenFile(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("failed to read golden file %s: %v", path, err)
	}
	return string(data)
}

// writeGoldenFile writes a golden file.
func writeGoldenFile(t *testing.T, name string, content string) {
	t.Helper()
	dir := filepath.Join("testdata")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create testdata directory: %v", err)
	}
	path := filepath.Join(dir, name+".golden")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write golden file %s: %v", path, err)
	}
}

// executeCommandForGolden executes a command and returns its output for golden file comparison.
func executeCommandForGolden(cmd *cobra.Command, args ...string) (string, error) {
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	err := cmd.Execute()
	return buf.String(), err
}
