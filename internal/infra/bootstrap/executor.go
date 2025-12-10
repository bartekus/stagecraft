// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Package bootstrap implements the INFRA_HOST_BOOTSTRAP engine.
//
// Feature: INFRA_HOST_BOOTSTRAP
// Spec: spec/infra/bootstrap.md

package bootstrap

import (
	"context"
)

// CommandExecutor defines the interface for executing commands on remote hosts.
//
// This abstraction allows the bootstrap Service to execute commands (e.g., Docker
// installation, Tailscale setup) without being coupled to a specific SSH implementation.
//
// v1 Slice 5: Interface is defined; implementations will be added in later slices
// (e.g., SSH-based executor using executil or a dedicated SSH library).
type CommandExecutor interface {
	// Run executes a command on the given host and returns stdout, stderr, and an error.
	//
	// The command is executed over SSH (or another remote execution mechanism)
	// using the host's PublicIP and the SSH user from the bootstrap Config.
	//
	// If the command fails, err should be non-nil. stdout and stderr should still
	// be populated with any output captured before the failure.
	Run(ctx context.Context, host Host, command string) (stdout string, stderr string, err error)
}

// NoopExecutor is a stub executor that does not perform any real operations.
//
// It is used for testing and as a placeholder until real SSH execution is implemented.
type NoopExecutor struct{}

// Run implements CommandExecutor by returning empty output and no error.
func (n *NoopExecutor) Run(_ context.Context, _ Host, _ string) (string, string, error) {
	return "", "", nil
}
