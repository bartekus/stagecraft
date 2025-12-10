// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_NETWORK_TAILSCALE
// Spec: spec/providers/network/tailscale.md

package tailscale

import (
	"context"
	"fmt"
	"os"
	"strings"

	"stagecraft/pkg/executil"
)

// Commander is an interface for executing commands on remote hosts.
// This abstraction allows testing without real SSH connections.
type Commander interface {
	// Run executes a command on the given host and returns stdout, stderr, and error.
	Run(ctx context.Context, host string, cmd string, args ...string) (stdout, stderr string, err error)
}

// SSHCommander implements Commander using SSH and executil.
// For v1, this is a simplified implementation that assumes SSH is configured
// and uses executil to run SSH commands.
type SSHCommander struct {
	// SSHUser is the SSH username (optional, defaults to current user)
	SSHUser string
	// SSHPort is the SSH port (optional, defaults to 22)
	SSHPort string
}

// NewSSHCommander creates a new SSH commander.
func NewSSHCommander() *SSHCommander {
	return &SSHCommander{}
}

// Run executes a command on the remote host via SSH.
func (c *SSHCommander) Run(ctx context.Context, host string, cmd string, args ...string) (string, string, error) {
	// Build SSH command: ssh [user@]host [command]
	sshArgs := []string{}

	// Add user if specified
	if c.SSHUser != "" {
		host = fmt.Sprintf("%s@%s", c.SSHUser, host)
	}

	// Add port if specified
	if c.SSHPort != "" {
		sshArgs = append(sshArgs, "-p", c.SSHPort)
	}

	// Add host
	sshArgs = append(sshArgs, host)

	// Add command to execute
	fullCmd := cmd
	if len(args) > 0 {
		fullCmd = cmd + " " + strings.Join(args, " ")
	}
	sshArgs = append(sshArgs, fullCmd)

	// Execute SSH command
	runner := executil.NewRunner()
	execCmd := executil.NewCommand("ssh", sshArgs...)

	result, err := runner.Run(ctx, execCmd)
	if err != nil {
		return string(result.Stdout), string(result.Stderr), err
	}

	return string(result.Stdout), string(result.Stderr), nil
}

// LocalCommander implements Commander for local execution (testing).
type LocalCommander struct {
	Commands map[string]CommandResult
}

// CommandResult represents the result of a command execution.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Error    error
}

// NewLocalCommander creates a new local commander for testing.
func NewLocalCommander() *LocalCommander {
	return &LocalCommander{
		Commands: make(map[string]CommandResult),
	}
}

// Run executes a command locally (for testing).
func (c *LocalCommander) Run(ctx context.Context, host string, cmd string, args ...string) (string, string, error) {
	// Build command key - try both with and without "sh -c" wrapper
	cmdKey := fmt.Sprintf("%s %s %s", host, cmd, strings.Join(args, " "))

	// Also try with sh -c wrapper for shell commands
	var cmdKeySh string
	if cmd == "sh" && len(args) > 0 && args[0] == "-c" {
		// Extract the actual command from "sh -c <command>"
		actualCmd := strings.Join(args[1:], " ")
		cmdKeySh = fmt.Sprintf("%s %s", host, actualCmd)
	}

	// Look up command result
	var result CommandResult
	var ok bool

	// Try exact match first
	result, ok = c.Commands[cmdKey]
	if !ok && cmdKeySh != "" {
		// Try with sh -c unwrapped
		result, ok = c.Commands[cmdKeySh]
	}

	if !ok {
		// Default: command not found
		return "", "", fmt.Errorf("command not found: %s", cmdKey)
	}

	if result.Error != nil {
		return result.Stdout, result.Stderr, result.Error
	}

	if result.ExitCode != 0 {
		return result.Stdout, result.Stderr, fmt.Errorf("command failed with exit code %d", result.ExitCode)
	}

	return result.Stdout, result.Stderr, nil
}

// getEnvVar retrieves an environment variable value.
func getEnvVar(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("%w: %s", ErrAuthKeyMissing, name)
	}
	return value, nil
}
