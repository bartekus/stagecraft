// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package main

import (
	"fmt"
	"os"

	"stagecraft/internal/cli"
)

func main() {
	rootCmd := cli.NewRootCommand()

	if err := rootCmd.Execute(); err != nil {
		// We deliberately avoid printing Cobra's default error twice
		// and centralize exit code handling here.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
