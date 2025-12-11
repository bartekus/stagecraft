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
	"sort"
	"strings"

	"stagecraft/pkg/providers/network"
)

// TailscaleProvider implements the NetworkProvider interface for Tailscale.
//
//nolint:revive // TailscaleProvider is intentionally named for clarity in provider package
type TailscaleProvider struct {
	commander Commander
	config    *Config
}

// Ensure TailscaleProvider implements NetworkProvider
var _ network.NetworkProvider = (*TailscaleProvider)(nil)

// ID returns the provider identifier.
func (p *TailscaleProvider) ID() string {
	return "tailscale"
}

// EnsureInstalled ensures Tailscale is installed on the given host.
func (p *TailscaleProvider) EnsureInstalled(ctx context.Context, opts network.EnsureInstalledOptions) error {
	// Parse config
	config, err := parseConfig(opts.Config)
	if err != nil {
		return fmt.Errorf("tailscale provider: %w", err)
	}

	// Store config for NodeFQDN
	p.config = config

	// Check if install should be skipped
	if config.Install.Method == "skip" {
		return nil
	}

	// Get commander (use default SSH commander if not set)
	commander := p.commander
	if commander == nil {
		commander = NewSSHCommander()
	}

	// Check if Tailscale is already installed
	stdout, _, err := commander.Run(ctx, opts.Host, "tailscale", "version")
	if err == nil && stdout != "" {
		// Already installed, check version if min_version is set
		if config.Install.MinVersion != "" {
			// For v1, we do a simple string comparison
			// In production, we'd parse semantic versions properly
			if strings.Contains(stdout, config.Install.MinVersion) {
				return nil
			}
			// Version doesn't meet minimum, but for v1 we'll accept it
			// Future: could upgrade or error here
			return nil
		}
		// Any version is acceptable
		return nil
	}

	// Check OS compatibility before attempting install
	if err := checkOSCompatibility(ctx, commander, opts.Host); err != nil {
		return err
	}

	// Install Tailscale using official install script
	installCmd := "curl -fsSL https://tailscale.com/install.sh | sh"
	_, stderr, err := commander.Run(ctx, opts.Host, "sh", "-c", installCmd)
	if err != nil {
		return fmt.Errorf("tailscale provider: %w: %s", ErrInstallFailed, stderr)
	}

	// Verify installation
	_, _, err = commander.Run(ctx, opts.Host, "tailscale", "version")
	if err != nil {
		return fmt.Errorf("tailscale provider: %w: installation verification failed", ErrInstallFailed)
	}

	return nil
}

// EnsureJoined ensures the host is joined to the Tailscale tailnet with the configured tags.
func (p *TailscaleProvider) EnsureJoined(ctx context.Context, opts network.EnsureJoinedOptions) error {
	// Parse config
	config, err := parseConfig(opts.Config)
	if err != nil {
		return fmt.Errorf("tailscale provider: %w", err)
	}

	// Store config for NodeFQDN
	p.config = config

	// Get auth key from environment
	authKey, err := getEnvVar(config.AuthKeyEnv)
	if err != nil {
		return fmt.Errorf("tailscale provider: %w", err)
	}

	// Compute tags: union of default_tags, role_tags, and opts.Tags
	tags := computeTags(config, opts.Tags)

	// Get commander
	commander := p.commander
	if commander == nil {
		commander = NewSSHCommander()
	}

	// Check current status
	stdout, _, err := commander.Run(ctx, opts.Host, "tailscale", "status", "--json")
	if err == nil {
		status, err := parseStatus(stdout)
		if err == nil {
			// Check if already joined correctly
			if status.TailnetName == config.TailnetDomain ||
				strings.HasSuffix(status.TailnetName, "."+config.TailnetDomain) {
				// Check tags match
				if tagsMatch(status.Self.Tags, tags) {
					return nil // Already joined correctly
				}
			} else {
				// Wrong tailnet
				return fmt.Errorf("tailscale provider: %w: host is in tailnet %q, expected %q",
					ErrTailnetMismatch, status.TailnetName, config.TailnetDomain)
			}
		}
	}

	// Join to tailnet
	joinCmd := buildTailscaleUpCommand(authKey, opts.Host, tags)

	_, stderr, err := commander.Run(ctx, opts.Host, "sh", "-c", joinCmd)
	if err != nil {
		// Check if it's an auth key error
		if strings.Contains(stderr, "invalid") || strings.Contains(stderr, "expired") {
			return fmt.Errorf("tailscale provider: %w", ErrAuthKeyInvalid)
		}
		return fmt.Errorf("tailscale provider: join failed: %s", stderr)
	}

	// Re-check status to verify join succeeded
	stdout, _, err = commander.Run(ctx, opts.Host, "tailscale", "status", "--json")
	if err != nil {
		return fmt.Errorf("tailscale provider: failed to verify join: %w", err)
	}

	status, err := parseStatus(stdout)
	if err != nil {
		return fmt.Errorf("tailscale provider: failed to parse status: %w", err)
	}

	// Validate final state
	if status.TailnetName != config.TailnetDomain &&
		!strings.HasSuffix(status.TailnetName, "."+config.TailnetDomain) {
		return fmt.Errorf("tailscale provider: %w: host is in tailnet %q, expected %q",
			ErrTailnetMismatch, status.TailnetName, config.TailnetDomain)
	}

	if !tagsMatch(status.Self.Tags, tags) {
		return fmt.Errorf("tailscale provider: %w: host tags %v do not match expected %v",
			ErrTagMismatch, status.Self.Tags, tags)
	}

	return nil
}

// NodeFQDN returns the FQDN for a node in the Tailscale mesh network.
// Note: This requires config to be set via EnsureInstalled or EnsureJoined first.
// If config is not available, returns an error.
func (p *TailscaleProvider) NodeFQDN(host string) (string, error) {
	// If config is not set, we can't generate FQDN
	// In practice, EnsureInstalled or EnsureJoined should be called first
	if p.config == nil {
		return "", fmt.Errorf("tailscale provider: %w: config not available (call EnsureInstalled or EnsureJoined first)", ErrConfigInvalid)
	}

	if err := validateTailnetDomain(p.config.TailnetDomain); err != nil {
		return "", err
	}
	return buildNodeFQDN(host, p.config.TailnetDomain), nil
}

// buildTailscaleUpCommand builds the Tailscale "up" command string.
// This is a pure function that takes explicit inputs and returns a command string.
func buildTailscaleUpCommand(authKey, hostname string, tags []string) string {
	tagArgs := strings.Join(tags, ",")
	return fmt.Sprintf("tailscale up --authkey=%s --hostname=%s --advertise-tags=%s",
		authKey, hostname, tagArgs)
}

// parseOSRelease parses the ID field from /etc/os-release content.
// Returns the distribution ID (e.g., "debian", "ubuntu") or empty string if not found.
// This is a pure function that operates on string content only.
func parseOSRelease(osReleaseContent string) string {
	lines := strings.Split(osReleaseContent, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "ID=") {
			continue
		}
		id := strings.TrimPrefix(line, "ID=")
		id = strings.Trim(id, `"`)
		id = strings.ToLower(id)
		return id
	}
	return ""
}

// validateTailnetDomain validates that a Tailnet domain is non-empty and has valid format.
// Returns an error if the domain is invalid.
func validateTailnetDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("tailscale provider: %w: tailnet_domain is required", ErrConfigInvalid)
	}
	// Basic validation: should contain at least one dot
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("tailscale provider: %w: tailnet_domain %q must contain a dot", ErrConfigInvalid, domain)
	}
	return nil
}

// buildNodeFQDN builds the FQDN for a Tailscale node.
// This is a pure function: host + domain = FQDN.
func buildNodeFQDN(host, domain string) string {
	return fmt.Sprintf("%s.%s", host, domain)
}

// computeTags computes the union of default tags, role tags, and provided tags.
func computeTags(config *Config, providedTags []string) []string {
	tagMap := make(map[string]bool)

	// Add default tags
	for _, tag := range config.DefaultTags {
		tagMap[tag] = true
	}

	// Add provided tags
	for _, tag := range providedTags {
		tagMap[tag] = true
	}

	// Convert to sorted slice for determinism
	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	return tags
}

// tagsMatch checks if the actual tags match the expected tags.
func tagsMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}

	actualMap := make(map[string]bool)
	for _, tag := range actual {
		actualMap[tag] = true
	}

	for _, tag := range expected {
		if !actualMap[tag] {
			return false
		}
	}

	return true
}

// checkOSCompatibility checks if the host OS is supported (Linux Debian/Ubuntu for v1).
func checkOSCompatibility(ctx context.Context, commander Commander, host string) error {
	// Check if Linux
	unameOut, _, err := commander.Run(ctx, host, "uname", "-s")
	if err != nil {
		// If uname fails, we'll proceed and let the install script handle it
		return nil
	}
	if !strings.Contains(strings.ToLower(unameOut), "linux") {
		return fmt.Errorf("tailscale provider: %w: detected OS %q, v1 supports Linux (Debian/Ubuntu) only",
			ErrUnsupportedOS, strings.TrimSpace(unameOut))
	}

	// Check for Debian/Ubuntu via /etc/os-release
	osRelease, _, err := commander.Run(ctx, host, "cat", "/etc/os-release")
	if err != nil {
		// If os-release doesn't exist, try lsb_release as fallback
		lsbOut, _, err2 := commander.Run(ctx, host, "lsb_release", "-i", "-s")
		if err2 != nil {
			// If both fail, we'll proceed and let the install script handle it
			return nil
		}
		distro := strings.ToLower(strings.TrimSpace(lsbOut))
		if distro == "debian" || distro == "ubuntu" {
			return nil
		}
		return fmt.Errorf("tailscale provider: %w: detected distribution %q, v1 supports Debian/Ubuntu only",
			ErrUnsupportedOS, distro)
	}

	// Parse os-release for ID
	id := parseOSRelease(osRelease)
	if id == "debian" || id == "ubuntu" {
		return nil
	}
	if id != "" {
		return fmt.Errorf("tailscale provider: %w: detected distribution %q, v1 supports Debian/Ubuntu only",
			ErrUnsupportedOS, id)
	}

	// If we can't determine the distribution, proceed (install script will handle it)
	return nil
}

func init() {
	network.Register(&TailscaleProvider{})
}
