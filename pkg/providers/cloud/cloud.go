// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package cloud provides interfaces and types for cloud providers.
package cloud

import "context"

// Feature: PROVIDER_CLOUD_INTERFACE
// Spec: spec/providers/cloud/interface.md

// HostSpec describes a host to be created or deleted.
type HostSpec struct {
	// Name is the hostname for the host
	Name string

	// Role is the role of the host (e.g., "gateway", "app", "db", "cache")
	Role string

	// Size is the instance size (e.g., "s-2vcpu-4gb" for DigitalOcean)
	Size string

	// Region is the region where the host should be created (e.g., "nyc1")
	Region string
}

// InfraPlan describes the infrastructure changes to be made.
type InfraPlan struct {
	// ToCreate are the hosts that should be created
	ToCreate []HostSpec

	// ToDelete are the hosts that should be deleted
	ToDelete []HostSpec
}

// PlanOptions contains options for planning infrastructure changes.
type PlanOptions struct {
	// Config is the provider-specific configuration decoded from
	// cloud.providers[providerID] in stagecraft.yml.
	// The provider implementation is responsible for unmarshaling this.
	Config any

	// Environment is the environment name (e.g., "staging", "prod")
	Environment string
}

// ApplyOptions contains options for applying infrastructure changes.
type ApplyOptions struct {
	// Config is the provider-specific configuration
	Config any

	// Environment is the environment name (e.g., "staging", "prod")
	Environment string

	// Plan is the infrastructure plan to apply
	Plan InfraPlan
}

// CloudProvider is the interface that all cloud providers must implement.
//
//nolint:revive // CloudProvider is the preferred name for clarity
type CloudProvider interface {
	// ID returns the unique identifier for this provider (e.g., "digitalocean", "aws").
	ID() string

	// Plan generates an infrastructure plan for the given environment.
	// This is a dry-run operation that does not modify infrastructure.
	Plan(ctx context.Context, opts PlanOptions) (InfraPlan, error)

	// Apply applies the given infrastructure plan, creating and deleting hosts as needed.
	Apply(ctx context.Context, opts ApplyOptions) error
}

// ProviderMetadata contains metadata about a provider.
type ProviderMetadata struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Experimental bool
}

// MetadataProvider is an optional interface that providers can implement
// to expose descriptive metadata.
type MetadataProvider interface {
	// Base provider interface
	CloudProvider

	// Metadata returns descriptive metadata about the provider.
	Metadata() ProviderMetadata
}
