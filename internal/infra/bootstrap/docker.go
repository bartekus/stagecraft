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
)

// ensureDocker ensures Docker is installed and working on the host.
//
// It performs:
// 1. Detection: Check if Docker is already installed
// 2. Installation: If missing, install Docker via apt (Ubuntu 22.04)
// 3. Verification: Re-check Docker after installation
//
// Returns (true, nil) if Docker is working, (false, error) otherwise.
func (s *service) ensureDocker(ctx context.Context, host Host, cfg Config) (bool, error) {
	// First try detection
	if ok := s.hasDocker(ctx, host, cfg); ok {
		return true, nil
	}

	// Install path
	if err := s.installDocker(ctx, host, cfg); err != nil {
		return false, fmt.Errorf("docker install failed: %w", err)
	}

	// Re-check after installation
	if ok := s.hasDocker(ctx, host, cfg); !ok {
		return false, fmt.Errorf("docker verification failed after install")
	}

	return true, nil
}

// hasDocker checks if Docker is installed and working on the host.
//
// It runs "docker version" and returns true if the command succeeds.
func (s *service) hasDocker(ctx context.Context, host Host, cfg Config) bool {
	_, _, err := s.executor.Run(ctx, host, "docker version")
	return err == nil
}

// installDocker installs Docker on the host using apt (Ubuntu 22.04).
//
// The installation is idempotent - running it multiple times is safe.
// Steps:
// 1. apt-get update -y
// 2. apt-get install -y docker.io
// 3. systemctl enable --now docker
func (s *service) installDocker(ctx context.Context, host Host, cfg Config) error {
	// Step 1: Update package list
	stdout, stderr, err := s.executor.Run(ctx, host, "apt-get update -y")
	if err != nil {
		return fmt.Errorf("apt-get update failed: %w (stdout: %s, stderr: %s)", err, stdout, stderr)
	}

	// Step 2: Install Docker
	stdout, stderr, err = s.executor.Run(ctx, host, "apt-get install -y docker.io")
	if err != nil {
		return fmt.Errorf("apt-get install docker.io failed: %w (stdout: %s, stderr: %s)", err, stdout, stderr)
	}

	// Step 3: Enable and start Docker service
	stdout, stderr, err = s.executor.Run(ctx, host, "systemctl enable --now docker")
	if err != nil {
		return fmt.Errorf("systemctl enable --now docker failed: %w (stdout: %s, stderr: %s)", err, stdout, stderr)
	}

	return nil
}
