// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Feature: PROVIDER_CLOUD_DO
// Spec: spec/providers/cloud/digitalocean.md

// Package digitalocean provides the DigitalOcean CloudProvider implementation.
package digitalocean

import "context"

// APIClient defines the interface for DigitalOcean API operations.
// This interface enables dependency injection for testing.
type APIClient interface {
	// ListDroplets lists all droplets matching the given filter.
	ListDroplets(ctx context.Context, filter DropletFilter) ([]Droplet, error)

	// GetDroplet retrieves a droplet by name.
	GetDroplet(ctx context.Context, name string) (*Droplet, error)

	// CreateDroplet creates a new droplet.
	CreateDroplet(ctx context.Context, req CreateDropletRequest) (*Droplet, error)

	// DeleteDroplet deletes a droplet by ID.
	DeleteDroplet(ctx context.Context, id int) error

	// ListSSHKeys lists all SSH keys in the account.
	ListSSHKeys(ctx context.Context) ([]SSHKey, error)

	// GetSSHKey retrieves an SSH key by name.
	GetSSHKey(ctx context.Context, name string) (*SSHKey, error)

	// WaitForDroplet waits for a droplet to reach the specified status.
	WaitForDroplet(ctx context.Context, id int, status string) error
}

// DropletFilter filters droplets for listing.
type DropletFilter struct {
	// NamePrefix filters droplets by name prefix (e.g., "staging-").
	NamePrefix string

	// Tags filters droplets by tags.
	Tags []string
}

// Droplet represents a DigitalOcean droplet.
type Droplet struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Region   string   `json:"region"`
	Size     string   `json:"size"`
	Status   string   `json:"status"`
	Networks Networks `json:"networks"`
}

// Networks represents droplet network configuration.
type Networks struct {
	V4 []NetworkV4 `json:"v4"`
}

// NetworkV4 represents an IPv4 network.
type NetworkV4 struct {
	IPAddress string `json:"ip_address"`
	Type      string `json:"type"`
}

// CreateDropletRequest represents a droplet creation request.
type CreateDropletRequest struct {
	Name    string
	Region  string
	Size    string
	Image   string // e.g., "ubuntu-22-04-x64"
	SSHKeys []int  // SSH key IDs
	Tags    []string
}

// SSHKey represents a DigitalOcean SSH key.
type SSHKey struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
