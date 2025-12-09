// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package compose provides dev Docker Compose infrastructure generation.
package compose

// Feature: DEV_COMPOSE_INFRA
// Spec: spec/dev/compose-infra.md

// ServiceDefinition represents a service that should be included
// in the dev Docker Compose model.
//
// It is intentionally provider-agnostic. Backend, frontend, and
// infra features map their own config into this type, and then
// DEV_COMPOSE_INFRA translates it into the core compose model.
type ServiceDefinition struct {
	// Name is the logical service name in the dev topology.
	// This will typically become the Compose service key.
	Name string

	// Image is the container image reference, if using a pre-built image.
	// Either Image or Build may be set, but not both.
	Image string

	// Build describes build configuration for image builds, if applicable.
	// This is a raw map to avoid leaking provider-specific structure into core.
	Build map[string]any

	// Ports describes host-to-container port mappings for dev.
	Ports []PortMapping

	// Volumes describes filesystem mounts for dev.
	Volumes []VolumeMapping

	// Environment contains environment variables for the service.
	Environment map[string]string

	// Networks lists the networks the service should join.
	Networks []string

	// DependsOn lists other service names this service depends on.
	DependsOn []string

	// Labels contains arbitrary labels attached to the service.
	Labels map[string]string
}

// PortMapping represents a single port mapping for a service.
//
// This will eventually be converted into the string form expected
// by Docker Compose (for example "8080:3000/tcp").
type PortMapping struct {
	// Host is the host port exposed for dev, for example "8080".
	Host string

	// Container is the container port, for example "3000".
	Container string

	// Protocol is the transport protocol, typically "tcp" or "udp".
	// If empty, "tcp" is assumed by higher level logic.
	Protocol string
}

// VolumeMapping represents a single volume or bind mount mapping.
//
// Type is intentionally coarse for v1 and may be refined in future
// iterations once DEV_COMPOSE_INFRA has tests driving the details.
type VolumeMapping struct {
	// Type is the volume type, for example "bind", "volume", or "tmpfs".
	Type string

	// Source is the host path or volume name.
	Source string

	// Target is the container path.
	Target string

	// ReadOnly indicates whether the mount is read-only.
	ReadOnly bool
}
