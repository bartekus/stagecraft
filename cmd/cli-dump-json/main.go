// SPDX-License-Identifier: AGPL-3.0-or-later

// Package main provides a tool to dump the CLI structure to JSON for governance checks.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"stagecraft/internal/cli"

	"github.com/bartekus/cortex/pkg/introspect"
)

func main() {
	rootCmd := cli.NewRootCommand()

	// Introspect the CLI command tree
	commands := introspect.Introspect(rootCmd)

	// Encode to JSON and print to stdout
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(commands); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode CLI definition: %v\n", err)
		os.Exit(1)
	}
}
