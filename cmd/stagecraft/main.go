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
