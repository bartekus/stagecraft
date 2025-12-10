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

	"stagecraft/pkg/providers/network"
)

// Host represents an infrastructure host to be bootstrapped.
//
// This mirrors the core fields from cloud.Host and will evolve as the
// INFRA_HOST_BOOTSTRAP spec grows.
type Host struct {
	// ID is the stable unique identifier from the cloud provider.
	ID string

	// Name is the human-readable name (e.g., "app-1").
	Name string

	// Role is the logical role (e.g., "app", "db", "proxy").
	Role string

	// PublicIP is the IPv4 address used for initial SSH connectivity.
	PublicIP string

	// Tags are provider or user-defined tags.
	Tags []string
}

// Config defines bootstrap-level configuration derived from stagecraft.yml.
//
// v1 Slice 3 keeps this intentionally minimal; fields will be expanded
// in later slices as INFRA_HOST_BOOTSTRAP is implemented.
type Config struct {
	// SSHUser is the user used for initial SSH connectivity (e.g., "root").
	SSHUser string
}

// HostResult captures the outcome of bootstrapping a single host.
type HostResult struct {
	Host    Host
	Success bool
	// Error is a human-readable error description when Success == false.
	Error string
}

// Result is the aggregate result of a single bootstrap run.
type Result struct {
	Hosts []HostResult
}

// SuccessCount returns the number of hosts that were successfully bootstrapped.
func (r *Result) SuccessCount() int {
	count := 0
	for _, hr := range r.Hosts {
		if hr.Success {
			count++
		}
	}
	return count
}

// FailureCount returns the number of hosts that failed to bootstrap.
func (r *Result) FailureCount() int {
	count := 0
	for _, hr := range r.Hosts {
		if !hr.Success {
			count++
		}
	}
	return count
}

// AllSucceeded returns true if all hosts were successfully bootstrapped.
func (r *Result) AllSucceeded() bool {
	return r.FailureCount() == 0
}

// HasFailures returns true if any host failed to bootstrap.
func (r *Result) HasFailures() bool {
	return r.FailureCount() > 0
}

// Service defines the INFRA_HOST_BOOTSTRAP engine interface.
//
// Bootstrap MUST be safe to call multiple times; v1 stub implementation
// simply reports success for all hosts.
type Service interface {
	Bootstrap(ctx context.Context, hosts []Host, cfg Config) (*Result, error)
}

// service is the default Service implementation.
type service struct {
	executor        CommandExecutor
	networkProvider network.NetworkProvider
}

// NewService creates a new bootstrap Service with the given command executor and network provider.
//
// v1 Slice 5: Service now accepts a CommandExecutor for executing commands on hosts.
// v1 Slice 6: Service uses executor to detect and install Docker on hosts.
// v1 Slice 7: Service uses network provider to ensure Tailscale is installed and joined.
//
// For production use, provide a real executor (e.g., SSH-based) and network provider.
// For testing, provide mocks or NoopExecutor/nil.
// If executor is nil, a NoopExecutor is used.
// If networkProvider is nil, network setup is skipped.
func NewService(executor CommandExecutor, networkProvider network.NetworkProvider) Service {
	if executor == nil {
		executor = &NoopExecutor{}
	}
	return &service{
		executor:        executor,
		networkProvider: networkProvider,
	}
}

// Bootstrap implements the Service interface.
//
// v1 Slice 6: Bootstrap now performs per-host Docker detection and installation.
// Hosts are processed sequentially in the order provided (which should be sorted
// deterministically by the caller).
func (s *service) Bootstrap(ctx context.Context, hosts []Host, cfg Config) (*Result, error) {
	results := make([]HostResult, len(hosts))
	for i, h := range hosts {
		results[i] = s.bootstrapHost(ctx, h, cfg)
	}

	return &Result{
		Hosts: results,
	}, nil
}

// bootstrapHost performs the bootstrap steps for a single host.
//
// v1 Slice 6: Ensures Docker is installed and working on the host.
// v1 Slice 7: Ensures Tailscale is installed and joined via NetworkProvider.
func (s *service) bootstrapHost(ctx context.Context, host Host, cfg Config) HostResult {
	// 1. Ensure Docker is present and working
	ok, err := s.ensureDocker(ctx, host, cfg)
	if !ok {
		return HostResult{
			Host:    host,
			Success: false,
			Error:   err.Error(),
		}
	}

	// 2. Ensure Tailscale is installed and joined (via NetworkProvider)
	if s.networkProvider != nil {
		ok, err := s.ensureTailscale(ctx, host, cfg)
		if !ok {
			return HostResult{
				Host:    host,
				Success: false,
				Error:   err.Error(),
			}
		}
	}

	return HostResult{
		Host:    host,
		Success: true,
		Error:   "",
	}
}
