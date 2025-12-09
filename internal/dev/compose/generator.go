// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package compose

import (
	"errors"
	"sort"
	"strconv"

	"gopkg.in/yaml.v3"

	corecompose "stagecraft/internal/compose"

	"stagecraft/pkg/config"
)

// Feature: DEV_COMPOSE_INFRA
// Spec: spec/dev/compose-infra.md

// ErrBackendServiceRequired is returned when GenerateCompose is called
// without a backend service definition. DEV_COMPOSE_INFRA requires at
// least a backend service for the dev topology.
var ErrBackendServiceRequired = errors.New("dev compose infra: backend service is required")

const (
	// devNetworkName is the deterministic network name for all dev services.
	devNetworkName = "stagecraft-dev"

	// traefikServiceName is the service name for Traefik.
	traefikServiceName = "traefik"

	// traefikImage is the deterministic Traefik image version for v1.
	traefikImage = "traefik:v2.11"
)

// Generator generates dev Docker Compose models by merging services
// from backend, frontend, and infrastructure features.
//
// Behaviour here is intentionally minimal for the first DEV_COMPOSE_INFRA
// slice; tests will drive the concrete implementation in later commits.
type Generator struct {
	// future configuration or options can be added here
}

// NewGenerator creates a new dev compose generator.
//
// At this stage it does not take any configuration. Future iterations
// may add options once tests and specs require them.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateCompose synthesizes a dev compose model from the supplied
// config and service definitions.
//
// For the initial slice this provides the smallest possible behaviour
// needed by the first DEV_COMPOSE_INFRA test: it returns a non-nil
// *corecompose.ComposeFile and a nil error. The structure of the
// compose file will be refined once tests define the required fields
// and deterministic ordering rules.
func (g *Generator) GenerateCompose(
	cfg *config.Config,
	backendService *ServiceDefinition,
	frontendService *ServiceDefinition,
	traefikService *ServiceDefinition,
) (*corecompose.ComposeFile, error) {
	// Basic topology validation – v1 requires at least a backend service.
	if backendService == nil {
		return nil, ErrBackendServiceRequired
	}

	// Use parameters to avoid linter complaints while behaviour is
	// still minimal and not all inputs are wired through.
	_ = cfg

	// Build services map
	services := make(map[string]any)

	// Add backend service
	backendServiceMap := g.buildServiceMap(backendService)
	services[backendService.Name] = backendServiceMap

	// Add frontend service if provided
	if frontendService != nil {
		frontendServiceMap := g.buildServiceMap(frontendService)
		services[frontendService.Name] = frontendServiceMap
	}

	// Add Traefik service if provided
	// For v1, DEV_COMPOSE_INFRA owns the Traefik service definition structure.
	// When traefikService != nil, we generate a complete Traefik service with
	// hardcoded image, ports, volumes, and command.
	if traefikService != nil {
		traefikServiceMap := g.generateTraefikService()
		services[traefikServiceName] = traefikServiceMap
	}

	// ensureDevNetwork is now handled in buildServiceMap, but we call it
	// here as a safety check for any services added directly (like Traefik)
	g.ensureDevNetwork(services)

	// Sort services lexicographically for deterministic ordering
	sortedServices := g.sortServices(services)

	// Create compose file data structure
	data := map[string]any{
		"version":  "3.8",
		"services": sortedServices,
	}

	// Always create stagecraft-dev network
	networks := map[string]any{
		devNetworkName: map[string]any{
			"name": devNetworkName,
		},
	}
	data["networks"] = networks

	return corecompose.NewComposeFile(data), nil
}

// convertPorts converts PortMapping slice to compose ports format.
// Ports are returned as []any where each element is a string in format
// "host:container/protocol". Ports are sorted deterministically by host port
// (numeric), then container port, then protocol.
func (g *Generator) convertPorts(portMappings []PortMapping) []any {
	if len(portMappings) == 0 {
		return nil
	}

	// Create a sortable slice for deterministic ordering
	type portEntry struct {
		portMapping PortMapping
		portString  string
	}

	entries := make([]portEntry, len(portMappings))
	for i, pm := range portMappings {
		protocol := pm.Protocol
		if protocol == "" {
			protocol = "tcp"
		}
		portStr := pm.Host + ":" + pm.Container + "/" + protocol
		entries[i] = portEntry{
			portMapping: pm,
			portString:  portStr,
		}
	}

	// Sort by host port (numeric), then container port, then protocol
	sort.Slice(entries, func(i, j int) bool {
		pi, pj := entries[i].portMapping, entries[j].portMapping

		// Compare host ports numerically
		hostI, errI := strconv.Atoi(pi.Host)
		hostJ, errJ := strconv.Atoi(pj.Host)
		if errI == nil && errJ == nil {
			if hostI != hostJ {
				return hostI < hostJ
			}
		} else {
			// Fallback to lexicographic if not numeric
			if pi.Host != pj.Host {
				return pi.Host < pj.Host
			}
		}

		// Compare container ports
		containerI, errI := strconv.Atoi(pi.Container)
		containerJ, errJ := strconv.Atoi(pj.Container)
		if errI == nil && errJ == nil {
			if containerI != containerJ {
				return containerI < containerJ
			}
		} else {
			if pi.Container != pj.Container {
				return pi.Container < pj.Container
			}
		}

		// Compare protocols
		protocolI := pi.Protocol
		if protocolI == "" {
			protocolI = "tcp"
		}
		protocolJ := pj.Protocol
		if protocolJ == "" {
			protocolJ = "tcp"
		}
		return protocolI < protocolJ
	})

	// Convert to []any using yaml.Node to force quoting
	result := make([]any, len(entries))
	for i, entry := range entries {
		// Create a yaml.Node with Style=DoubleQuoted to force quoting
		node := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: entry.portString,
			Style: yaml.DoubleQuotedStyle,
			Tag:   "!!str",
		}
		result[i] = node
	}

	return result
}

// convertEnvironment converts environment map to compose environment format.
// Returns map[string]any for compose compatibility. Keys are sorted
// lexicographically for determinism when serializing to YAML.
func (g *Generator) convertEnvironment(env map[string]string) map[string]any {
	if len(env) == 0 {
		return nil
	}

	// Sort keys for deterministic ordering
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create result map (Go maps don't preserve order, but YAML encoder
	// will use sorted keys when marshaling if we ensure deterministic access)
	result := make(map[string]any, len(env))
	for _, k := range keys {
		result[k] = env[k]
	}

	return result
}

// convertBuild converts a generic build map to a compose-compatible
// map[string]any. Keys are sorted lexicographically for deterministic
// YAML serialization. Nested values are passed through as-is.
func (g *Generator) convertBuild(build map[string]any) map[string]any {
	if len(build) == 0 {
		return nil
	}

	keys := make([]string, 0, len(build))
	for k := range build {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(map[string]any, len(build))
	for _, k := range keys {
		result[k] = build[k]
	}

	return result
}

// convertLabels converts label map to compose labels format.
// Returns map[string]any for compose compatibility. Keys are sorted
// lexicographically for determinism when serializing to YAML.
func (g *Generator) convertLabels(labels map[string]string) map[string]any {
	if len(labels) == 0 {
		return nil
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make(map[string]any, len(labels))
	for _, k := range keys {
		result[k] = labels[k]
	}

	return result
}

// convertDependsOn converts a slice of dependency service names to the
// compose depends_on format. It returns []any of strings and sorts the
// names lexicographically for deterministic ordering.
func (g *Generator) convertDependsOn(depends []string) []any {
	if len(depends) == 0 {
		return nil
	}

	names := make([]string, len(depends))
	copy(names, depends)

	sort.Strings(names)

	result := make([]any, len(names))
	for i, name := range names {
		result[i] = name
	}

	return result
}

// convertNetworks converts a slice of network names to the compose
// networks format. It returns []any of strings and sorts the names
// lexicographically for deterministic ordering.
func (g *Generator) convertNetworks(networks []string) []any {
	if len(networks) == 0 {
		return nil
	}

	names := make([]string, len(networks))
	copy(names, networks)

	sort.Strings(names)

	result := make([]any, len(names))
	for i, name := range names {
		result[i] = name
	}

	return result
}

// convertVolumes converts VolumeMapping slice to compose volumes format.
// Volumes are returned as []any where each element is a string in format
// "source:target" or "source:target:ro" for read-only mounts.
// Volumes are sorted deterministically by Target (lexicographically),
// then Source, then ReadOnly (read-write before read-only).
func (g *Generator) convertVolumes(vols []VolumeMapping) []any {
	if len(vols) == 0 {
		return nil
	}

	entries := make([]VolumeMapping, len(vols))
	copy(entries, vols)

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Target != entries[j].Target {
			return entries[i].Target < entries[j].Target
		}
		if entries[i].Source != entries[j].Source {
			return entries[i].Source < entries[j].Source
		}
		// Put read-write before read-only
		if entries[i].ReadOnly == entries[j].ReadOnly {
			return false
		}
		return !entries[i].ReadOnly && entries[j].ReadOnly
	})

	result := make([]any, len(entries))
	for i, v := range entries {
		suffix := ""
		if v.ReadOnly {
			suffix = ":ro"
		}
		result[i] = v.Source + ":" + v.Target + suffix
	}

	return result
}

// buildServiceMap converts a ServiceDefinition to a compose service map.
// It handles all fields and ensures the service joins stagecraft-dev network.
func (g *Generator) buildServiceMap(svc *ServiceDefinition) map[string]any {
	serviceMap := make(map[string]any)

	// Add image if provided
	if svc.Image != "" {
		serviceMap["image"] = svc.Image
	}

	// Add build if provided
	if len(svc.Build) > 0 {
		serviceMap["build"] = g.convertBuild(svc.Build)
	}

	// Add ports if provided
	if len(svc.Ports) > 0 {
		ports := g.convertPorts(svc.Ports)
		serviceMap["ports"] = ports
	}

	// Add environment if provided
	if len(svc.Environment) > 0 {
		env := g.convertEnvironment(svc.Environment)
		serviceMap["environment"] = env
	}

	// Add volumes if provided
	if len(svc.Volumes) > 0 {
		volumes := g.convertVolumes(svc.Volumes)
		serviceMap["volumes"] = volumes
	}

	// Add labels if provided
	if len(svc.Labels) > 0 {
		labels := g.convertLabels(svc.Labels)
		serviceMap["labels"] = labels
	}

	// Add depends_on if provided
	if len(svc.DependsOn) > 0 {
		dependsOn := g.convertDependsOn(svc.DependsOn)
		serviceMap["depends_on"] = dependsOn
	}

	// Add networks: ensure stagecraft-dev is included
	networks := append([]string{}, svc.Networks...)
	if !containsString(networks, devNetworkName) {
		networks = append(networks, devNetworkName)
	}
	if len(networks) > 0 {
		serviceMap["networks"] = g.convertNetworks(networks)
	}

	return serviceMap
}

// containsString checks if a string slice contains a value.
func containsString(slice []string, value string) bool {
	for _, s := range slice {
		if s == value {
			return true
		}
	}
	return false
}

// generateTraefikService generates a complete Traefik service definition
// with hardcoded values for v1. This matches the spec requirements:
// - Image: traefik:v2.11
// - Ports: 80:80, 443:443
// - Volumes: .stagecraft/dev/certs:/certs:ro, .stagecraft/dev/traefik:/etc/traefik:ro
// - Command: file provider flags
// - Networks: stagecraft-dev
func (g *Generator) generateTraefikService() map[string]any {
	service := make(map[string]any)

	// Image
	service["image"] = traefikImage

	// Ports: 80:80, 443:443
	// Use yaml.Node to force string quoting for consistency with other ports
	service["ports"] = []any{
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "80:80",
			Style: yaml.DoubleQuotedStyle,
			Tag:   "!!str",
		},
		&yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "443:443",
			Style: yaml.DoubleQuotedStyle,
			Tag:   "!!str",
		},
	}

	// Volumes: certs + traefik config
	service["volumes"] = []any{
		"./.stagecraft/dev/certs:/certs:ro",
		"./.stagecraft/dev/traefik:/etc/traefik:ro",
	}

	// Command flags for file provider
	service["command"] = []any{
		"--configfile=/etc/traefik/traefik-static.yaml",
		"--providers.file.directory=/etc/traefik",
		"--providers.file.watch=true",
	}

	// Networks: must include stagecraft-dev
	service["networks"] = []any{devNetworkName}

	return service
}

// ensureDevNetwork ensures all services join the stagecraft-dev network.
// If a service already has networks, stagecraft-dev is added to the list
// (deduplicated). If a service has no networks, stagecraft-dev is added.
func (g *Generator) ensureDevNetwork(services map[string]any) {
	for _, serviceData := range services {
		serviceMap, ok := serviceData.(map[string]any)
		if !ok {
			continue
		}

		// Get existing networks
		existingNetworks, ok := serviceMap["networks"].([]any)
		if !ok {
			// No networks, add stagecraft-dev
			serviceMap["networks"] = []any{devNetworkName}
			continue
		}

		// Check if stagecraft-dev is already present
		hasDevNetwork := false
		for _, net := range existingNetworks {
			if netStr, ok := net.(string); ok && netStr == devNetworkName {
				hasDevNetwork = true
				break
			}
		}

		// Add stagecraft-dev if not present
		if !hasDevNetwork {
			// Build networks list with stagecraft-dev added
			allNetworks := make([]any, len(existingNetworks)+1)
			copy(allNetworks, existingNetworks)
			allNetworks[len(existingNetworks)] = devNetworkName
			// Sort networks for determinism
			netStrs := make([]string, len(allNetworks))
			for i, n := range allNetworks {
				if ns, ok := n.(string); ok {
					netStrs[i] = ns
				}
			}
			sort.Strings(netStrs)
			sortedNetworks := make([]any, len(netStrs))
			for i, ns := range netStrs {
				sortedNetworks[i] = ns
			}
			serviceMap["networks"] = sortedNetworks
		} else {
			// Already present, but ensure it's sorted
			netStrs := make([]string, len(existingNetworks))
			for i, n := range existingNetworks {
				if ns, ok := n.(string); ok {
					netStrs[i] = ns
				}
			}
			sort.Strings(netStrs)
			sortedNetworks := make([]any, len(netStrs))
			for i, ns := range netStrs {
				sortedNetworks[i] = ns
			}
			serviceMap["networks"] = sortedNetworks
		}
	}
}

// sortServices returns a new map with services sorted lexicographically by name.
// This ensures deterministic service ordering in the compose file.
func (g *Generator) sortServices(services map[string]any) map[string]any {
	// Get service names and sort them
	names := make([]string, 0, len(services))
	for name := range services {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build sorted map
	sorted := make(map[string]any, len(services))
	for _, name := range names {
		sorted[name] = services[name]
	}

	return sorted
}
