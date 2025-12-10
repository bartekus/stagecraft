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
	"fmt"

	"stagecraft/pkg/executil"
)

// SSHExecutor implements CommandExecutor using SSH to run commands on remote hosts.
//
// It uses the executil package to execute SSH commands locally, which then connect
// to remote hosts and execute commands there.
type SSHExecutor struct {
	runner  executil.Runner
	sshUser string
}

// NewSSHExecutor creates a new SSHExecutor using the given SSH user.
// If runner is nil, a new executil.Runner is created.
func NewSSHExecutor(sshUser string, runner executil.Runner) *SSHExecutor {
	if runner == nil {
		runner = executil.NewRunner()
	}
	return &SSHExecutor{
		runner:  runner,
		sshUser: sshUser,
	}
}

// Run executes the given command on the remote host using ssh.
//
// It builds a command like:
//
//	ssh -o BatchMode=yes -o StrictHostKeyChecking=no user@IP "<command>"
//
// The command is executed via executil.Runner, which handles context cancellation
// and error handling.
func (e *SSHExecutor) Run(ctx context.Context, host Host, command string) (string, string, error) {
	if host.PublicIP == "" {
		return "", "", fmt.Errorf("missing PublicIP for host %q", host.ID)
	}

	user := e.sshUser
	if user == "" {
		user = "root"
	}

	target := fmt.Sprintf("%s@%s", user, host.PublicIP)

	args := []string{
		"-o", "BatchMode=yes",
		"-o", "StrictHostKeyChecking=no",
		target,
		command,
	}

	cmd := executil.NewCommand("ssh", args...)
	result, err := e.runner.Run(ctx, cmd)
	if err != nil {
		// Wrap error with host context
		return string(result.Stdout), string(result.Stderr), fmt.Errorf("ssh to %s failed: %w", target, err)
	}

	return string(result.Stdout), string(result.Stderr), nil
}
