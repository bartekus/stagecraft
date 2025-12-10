// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_CLOUD_DO
// Spec: spec/providers/cloud/digitalocean.md

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"stagecraft/pkg/providers/cloud"
)

// DigitalOceanProvider implements the CloudProvider interface for DigitalOcean.
//
//nolint:revive // DigitalOceanProvider is intentionally named for clarity in provider package
type DigitalOceanProvider struct {
	client APIClient
}

// Ensure DigitalOceanProvider implements CloudProvider
var _ cloud.CloudProvider = (*DigitalOceanProvider)(nil)

// NewDigitalOceanProvider creates a new DigitalOcean provider with default API client.
// For production use, this will create a real DigitalOcean API client.
// For testing, use NewDigitalOceanProviderWithClient.
func NewDigitalOceanProvider() *DigitalOceanProvider {
	// TODO: Create real DO client in Slice 2
	return &DigitalOceanProvider{
		client: nil, // Will be implemented in Slice 2
	}
}

// NewDigitalOceanProviderWithClient creates a new DigitalOcean provider with injected API client.
// This is used for testing with mocked API clients.
func NewDigitalOceanProviderWithClient(client APIClient) *DigitalOceanProvider {
	return &DigitalOceanProvider{
		client: client,
	}
}

// ID returns the provider identifier.
func (p *DigitalOceanProvider) ID() string {
	return "digitalocean"
}

// Plan generates an infrastructure plan for the given environment.
// This is a dry-run operation that does not modify infrastructure.
func (p *DigitalOceanProvider) Plan(ctx context.Context, opts cloud.PlanOptions) (cloud.InfraPlan, error) {
	config, err := parseConfig(opts.Config)
	if err != nil {
		return cloud.InfraPlan{}, err
	}

	// Get API token from environment
	token, ok := os.LookupEnv(config.TokenEnv)
	if !ok || token == "" {
		return cloud.InfraPlan{}, fmt.Errorf("%w: API token missing from environment variable %s", ErrTokenMissing, config.TokenEnv)
	}
	_ = token // Token validated but not used directly in Plan (only in Apply)

	// Validate SSH key exists
	if _, err := p.client.GetSSHKey(ctx, config.SSHKeyName); err != nil {
		if errors.Is(err, ErrSSHKeyNotFound) {
			return cloud.InfraPlan{}, fmt.Errorf("%w: SSH key %q not found in DigitalOcean account", ErrSSHKeyNotFound, config.SSHKeyName)
		}
		return cloud.InfraPlan{}, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	env := opts.Environment
	envHosts, ok := config.Hosts[env]
	if !ok || len(envHosts) == 0 {
		// Environment not configured; no hosts to create/delete
		return cloud.InfraPlan{}, nil
	}

	// List existing droplets for this environment
	droplets, err := p.client.ListDroplets(ctx, DropletFilter{
		NamePrefix: env + "-",
	})
	if err != nil {
		return cloud.InfraPlan{}, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	// Build desired hosts map
	desired := make(map[string]HostConfig, len(envHosts))
	for name, hostCfg := range envHosts {
		desired[name] = hostCfg
	}

	// Build actual droplets map (strip environment prefix)
	actual := make(map[string]Droplet, len(droplets))
	for _, d := range droplets {
		// Strip "{env}-" prefix to get logical hostname
		name := strings.TrimPrefix(d.Name, env+"-")
		actual[name] = d
	}

	var toCreate, toDelete []cloud.HostSpec

	// Hosts to create
	for name, hostCfg := range desired {
		if _, exists := actual[name]; !exists {
			toCreate = append(toCreate, cloud.HostSpec{
				Name:   name,
				Role:   hostCfg.Role,
				Size:   firstNonEmpty(hostCfg.Size, config.DefaultSize),
				Region: firstNonEmpty(hostCfg.Region, config.DefaultRegion),
			})
		}
	}

	// Hosts to delete
	for name, d := range actual {
		if _, exists := desired[name]; !exists {
			toDelete = append(toDelete, cloud.HostSpec{
				Name:   name,
				Role:   "", // Not needed for delete
				Size:   d.Size,
				Region: d.Region,
			})
		}
	}

	// Sort lexicographically by Name
	sort.Slice(toCreate, func(i, j int) bool {
		return toCreate[i].Name < toCreate[j].Name
	})
	sort.Slice(toDelete, func(i, j int) bool {
		return toDelete[i].Name < toDelete[j].Name
	})

	return cloud.InfraPlan{
		ToCreate: toCreate,
		ToDelete: toDelete,
	}, nil
}

// firstNonEmpty returns the first non-empty string from the given values.
func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// Apply applies the given infrastructure plan, creating and deleting droplets as needed.
//
//nolint:gocritic // hugeParam: opts matches interface signature
func (p *DigitalOceanProvider) Apply(ctx context.Context, opts cloud.ApplyOptions) error {
	config, err := parseConfig(opts.Config)
	if err != nil {
		return err
	}

	// Get API token from environment
	token, ok := os.LookupEnv(config.TokenEnv)
	if !ok || token == "" {
		return fmt.Errorf("%w: API token missing from environment variable %s", ErrTokenMissing, config.TokenEnv)
	}
	_ = token // Used by real client, not provider

	// Validate SSH key exists and get its ID
	sshKey, err := p.client.GetSSHKey(ctx, config.SSHKeyName)
	if err != nil {
		if errors.Is(err, ErrSSHKeyNotFound) {
			return fmt.Errorf("%w: SSH key %q not found in DigitalOcean account", ErrSSHKeyNotFound, config.SSHKeyName)
		}
		return fmt.Errorf("%w: %v", ErrAPIError, err)
	}
	sshKeyID := sshKey.ID

	env := opts.Environment

	// Process creates in deterministic order
	toCreate := append([]cloud.HostSpec(nil), opts.Plan.ToCreate...)
	sort.Slice(toCreate, func(i, j int) bool {
		return toCreate[i].Name < toCreate[j].Name
	})

	for _, host := range toCreate {
		fullName := env + "-" + host.Name

		existing, err := p.client.GetDroplet(ctx, fullName)
		if err != nil && !errors.Is(err, ErrDropletNotFound) {
			return fmt.Errorf("%w: %v", ErrAPIError, err)
		}

		if existing != nil {
			// Idempotent if matches spec
			if existing.Region == host.Region && existing.Size == host.Size {
				continue
			}
			return fmt.Errorf("%w: droplet %q already exists with different spec", ErrDropletExists, fullName)
		}

		req := CreateDropletRequest{
			Name:   fullName,
			Region: host.Region,
			Size:   host.Size,
			Image:  "ubuntu-22-04-x64",
			SSHKeys: []int{
				sshKeyID,
			},
			Tags: []string{
				"stagecraft",
				"stagecraft-env-" + env,
			},
		}

		droplet, err := p.client.CreateDroplet(ctx, req)
		if err != nil {
			if errors.Is(err, ErrRateLimit) {
				return fmt.Errorf("%w: %v", ErrRateLimit, err)
			}
			return fmt.Errorf("%w: %v", ErrDropletCreateFailed, err)
		}

		if err := p.client.WaitForDroplet(ctx, droplet.ID, "active"); err != nil {
			if errors.Is(err, ErrDropletTimeout) {
				return fmt.Errorf("%w: %v", ErrDropletTimeout, err)
			}
			return fmt.Errorf("%w: %v", ErrAPIError, err)
		}
	}

	// Process deletes in deterministic order
	toDelete := append([]cloud.HostSpec(nil), opts.Plan.ToDelete...)
	sort.Slice(toDelete, func(i, j int) bool {
		return toDelete[i].Name < toDelete[j].Name
	})

	for _, host := range toDelete {
		fullName := env + "-" + host.Name

		existing, err := p.client.GetDroplet(ctx, fullName)
		if err != nil {
			if errors.Is(err, ErrDropletNotFound) {
				// Already deleted, idempotent
				continue
			}
			return fmt.Errorf("%w: %v", ErrAPIError, err)
		}

		if err := p.client.DeleteDroplet(ctx, existing.ID); err != nil {
			if errors.Is(err, ErrDropletNotFound) {
				continue
			}
			return fmt.Errorf("%w: %v", ErrDropletDeleteFailed, err)
		}

		if err := p.client.WaitForDroplet(ctx, existing.ID, "deleted"); err != nil {
			if errors.Is(err, ErrDropletTimeout) {
				return fmt.Errorf("%w: %v", ErrDropletTimeout, err)
			}
			return fmt.Errorf("%w: %v", ErrAPIError, err)
		}
	}

	return nil
}

// Hosts returns the list of provisioned hosts for the given environment.
// This is a stub implementation for Slice 2; full implementation will come in later slices.
func (p *DigitalOceanProvider) Hosts(ctx context.Context, opts cloud.HostsOptions) ([]cloud.Host, error) {
	// TODO: Implement full Hosts method in later slices
	// For now, return empty list to satisfy interface
	return []cloud.Host{}, nil
}

// init registers the provider with the cloud registry.
func init() {
	cloud.Register(NewDigitalOceanProvider())
}
