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
	"errors"
	"fmt"
	"os"
	"testing"

	"stagecraft/pkg/providers/network"
)

func TestTailscaleProvider_ID(t *testing.T) {
	provider := &TailscaleProvider{}
	if got := provider.ID(); got != "tailscale" {
		t.Errorf("ID() = %q, want %q", got, "tailscale")
	}
}

func TestTailscaleProvider_NodeFQDN_AfterEnsureInstalled(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	// Call EnsureInstalled to set config
	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Fatalf("EnsureInstalled() error = %v", err)
	}

	// Now NodeFQDN should work
	fqdn, err := provider.NodeFQDN("db-1")
	if err != nil {
		t.Fatalf("NodeFQDN() error = %v", err)
	}
	if fqdn != "db-1.example.ts.net" {
		t.Errorf("NodeFQDN() = %q, want %q", fqdn, "db-1.example.ts.net")
	}

	// Test with different host
	fqdn2, err := provider.NodeFQDN("gateway-1")
	if err != nil {
		t.Fatalf("NodeFQDN() error = %v", err)
	}
	if fqdn2 != "gateway-1.example.ts.net" {
		t.Errorf("NodeFQDN() = %q, want %q", fqdn2, "gateway-1.example.ts.net")
	}
}

func TestTailscaleProvider_NodeFQDN_WithConfig(t *testing.T) {
	provider := &TailscaleProvider{
		config: &Config{
			TailnetDomain: "test.ts.net",
		},
	}

	fqdn, err := provider.NodeFQDN("app-1")
	if err != nil {
		t.Fatalf("NodeFQDN() error = %v", err)
	}
	if fqdn != "app-1.test.ts.net" {
		t.Errorf("NodeFQDN() = %q, want %q", fqdn, "app-1.test.ts.net")
	}
}

func TestTailscaleProvider_NodeFQDN_NoConfig(t *testing.T) {
	provider := &TailscaleProvider{}

	_, err := provider.NodeFQDN("app-1")
	if err == nil {
		t.Error("NodeFQDN() expected error when config is not set, got nil")
	}
	if !contains(err.Error(), "config not available") {
		t.Errorf("NodeFQDN() error = %q, want substring %q", err.Error(), "config not available")
	}
}

func TestTailscaleProvider_EnsureInstalled_AlreadyInstalled(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
		config: &Config{
			AuthKeyEnv:    "TS_AUTHKEY",
			TailnetDomain: "example.ts.net",
			Install: InstallConfig{
				Method: "auto",
			},
		},
	}

	commander := provider.commander.(*LocalCommander)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_InstallSucceeds(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// First check: not installed (tailscale version fails)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// which also fails
	commander.Commands["app-1 which tailscale"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// Install script succeeds (note: commander will unwrap "sh -c" commands)
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
	}
	// Verify installation (second call after install)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_InstallSkipped(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
		config: &Config{
			AuthKeyEnv:    "TS_AUTHKEY",
			TailnetDomain: "example.ts.net",
			Install: InstallConfig{
				Method: "skip",
			},
		},
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
			"install": map[string]interface{}{
				"method": "skip",
			},
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureJoined_AlreadyJoined(t *testing.T) {
	// Set up auth key env var
	_ = os.Setenv("TS_AUTHKEY", "test-auth-key")
	defer func() { _ = os.Unsetenv("TS_AUTHKEY") }()

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
		config: &Config{
			AuthKeyEnv:    "TS_AUTHKEY",
			TailnetDomain: "example.ts.net",
			DefaultTags:   []string{"tag:stagecraft"},
		},
	}

	commander := provider.commander.(*LocalCommander)
	// Status shows already joined correctly
	statusJSON := `{
		"TailnetName": "example.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": ["tag:stagecraft", "tag:app"]
		}
	}`
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		Stdout:   statusJSON,
		ExitCode: 0,
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
			"default_tags":   []string{"tag:stagecraft"},
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureJoined() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureJoined_AuthKeyMissing(t *testing.T) {
	// Ensure env var is not set
	_ = os.Unsetenv("TS_AUTHKEY")

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
		config: &Config{
			AuthKeyEnv:    "TS_AUTHKEY",
			TailnetDomain: "example.ts.net",
		},
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err == nil {
		t.Error("EnsureJoined() expected error, got nil")
	}
	if !contains(err.Error(), "auth key missing") {
		t.Errorf("EnsureJoined() error = %q, want substring %q", err.Error(), "auth key missing")
	}
}

func TestTailscaleProvider_EnsureJoined_WrongTailnet(t *testing.T) {
	_ = os.Setenv("TS_AUTHKEY", "test-auth-key")
	defer func() { _ = os.Unsetenv("TS_AUTHKEY") }()

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
		config: &Config{
			AuthKeyEnv:    "TS_AUTHKEY",
			TailnetDomain: "example.ts.net",
		},
	}

	commander := provider.commander.(*LocalCommander)
	// Status shows different tailnet
	statusJSON := `{
		"TailnetName": "other.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": []
		}
	}`
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		Stdout:   statusJSON,
		ExitCode: 0,
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err == nil {
		t.Error("EnsureJoined() expected error, got nil")
	}
	if !contains(err.Error(), "tailnet mismatch") {
		t.Errorf("EnsureJoined() error = %q, want substring %q", err.Error(), "tailnet mismatch")
	}
}

func TestTailscaleProvider_EnsureInstalled_ConfigError(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env": "TS_AUTHKEY",
			// Missing tailnet_domain
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err == nil {
		t.Error("EnsureInstalled() expected error for invalid config, got nil")
	}
	if !contains(err.Error(), "invalid config") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", err.Error(), "invalid config")
	}
}

func TestTailscaleProvider_EnsureInstalled_InstallFails(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// Install script fails
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 1,
		Stderr:   "install failed",
		Error:    fmt.Errorf("install failed"),
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err == nil {
		t.Error("EnsureInstalled() expected error for install failure, got nil")
	}
	if !contains(err.Error(), "installation failed") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", err.Error(), "installation failed")
	}
}

func TestTailscaleProvider_EnsureJoined_ConfigError(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env": "TS_AUTHKEY",
			// Missing tailnet_domain
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err == nil {
		t.Error("EnsureJoined() expected error for invalid config, got nil")
	}
	if !contains(err.Error(), "invalid config") {
		t.Errorf("EnsureJoined() error = %q, want substring %q", err.Error(), "invalid config")
	}
}

func TestTailscaleProvider_computeTags(t *testing.T) {
	config := &Config{
		DefaultTags: []string{"tag:stagecraft"},
		RoleTags: map[string][]string{
			"app": {"tag:app"},
			"db":  {"tag:db"},
		},
	}

	// computeTags only uses default_tags and provided tags, not role_tags
	// Role tags should be passed in the provided tags
	tags := computeTags(config, []string{"tag:app", "tag:custom"})

	// Should contain default tags + provided tags
	expectedTags := []string{"tag:custom", "tag:stagecraft", "tag:app"}
	if len(tags) != len(expectedTags) {
		t.Errorf("computeTags() returned %d tags, want %d", len(tags), len(expectedTags))
	}

	// Check all expected tags are present
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}
	for _, expected := range expectedTags {
		if !tagMap[expected] {
			t.Errorf("computeTags() missing tag %q, got %v", expected, tags)
		}
	}
}

func TestTailscaleProvider_tagsMatch(t *testing.T) {
	tests := []struct {
		name     string
		actual   []string
		expected []string
		want     bool
	}{
		{
			name:     "exact match",
			actual:   []string{"tag:a", "tag:b"},
			expected: []string{"tag:a", "tag:b"},
			want:     true,
		},
		{
			name:     "different order",
			actual:   []string{"tag:b", "tag:a"},
			expected: []string{"tag:a", "tag:b"},
			want:     true,
		},
		{
			name:     "missing tag",
			actual:   []string{"tag:a"},
			expected: []string{"tag:a", "tag:b"},
			want:     false,
		},
		{
			name:     "extra tag",
			actual:   []string{"tag:a", "tag:b", "tag:c"},
			expected: []string{"tag:a", "tag:b"},
			want:     false,
		},
		{
			name:     "empty",
			actual:   []string{},
			expected: []string{},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tagsMatch(tt.actual, tt.expected)
			if got != tt.want {
				t.Errorf("tagsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTailscaleProvider_EnsureJoined_MissingAuthKey(t *testing.T) {
	// Make sure env var is not set
	_ = os.Unsetenv("TS_AUTHKEY")

	p := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]any{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := p.EnsureJoined(context.Background(), opts)
	if err == nil {
		t.Fatalf("EnsureJoined() error = nil, want error for missing auth key")
	}
	if !errors.Is(err, ErrAuthKeyMissing) {
		t.Fatalf("EnsureJoined() error = %v, want ErrAuthKeyMissing", err)
	}
}

func TestParseConfig_MissingRequiredFields(t *testing.T) {
	// Missing both auth_key_env and tailnet_domain
	cfg := map[string]any{}

	_, err := parseConfig(cfg)
	if err == nil {
		t.Fatalf("parseConfig() error = nil, want error")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("parseConfig() error = %v, want ErrConfigInvalid", err)
	}
}

func TestParseConfig_MissingTailnetDomain(t *testing.T) {
	cfg := map[string]any{
		"auth_key_env": "TS_AUTHKEY",
		// Missing tailnet_domain
	}

	_, err := parseConfig(cfg)
	if err == nil {
		t.Fatalf("parseConfig() error = nil, want error")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("parseConfig() error = %v, want ErrConfigInvalid", err)
	}
	if !contains(err.Error(), "tailnet_domain is required") {
		t.Errorf("parseConfig() error = %q, want substring %q", err.Error(), "tailnet_domain is required")
	}
}

func TestParseConfig_MissingAuthKeyEnv(t *testing.T) {
	cfg := map[string]any{
		"tailnet_domain": "example.ts.net",
		// Missing auth_key_env
	}

	_, err := parseConfig(cfg)
	if err == nil {
		t.Fatalf("parseConfig() error = nil, want error")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Fatalf("parseConfig() error = %v, want ErrConfigInvalid", err)
	}
	if !contains(err.Error(), "auth_key_env is required") {
		t.Errorf("parseConfig() error = %q, want substring %q", err.Error(), "auth_key_env is required")
	}
}

func TestTailscaleProvider_EnsureInstalled_UnsupportedOS(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname returns non-Linux
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout:   "Darwin",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err == nil {
		t.Error("EnsureInstalled() expected error for unsupported OS, got nil")
		return
	}
	if !errors.Is(err, ErrUnsupportedOS) {
		t.Errorf("EnsureInstalled() error = %v, want ErrUnsupportedOS", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_UnsupportedDistribution(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname returns Linux
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout:   "Linux",
		ExitCode: 0,
	}
	// os-release shows non-Debian/Ubuntu distro
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		Stdout:   "ID=centos\n",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err == nil {
		t.Error("EnsureInstalled() expected error for unsupported distribution, got nil")
		return
	}
	if !errors.Is(err, ErrUnsupportedOS) {
		t.Errorf("EnsureInstalled() error = %v, want ErrUnsupportedOS", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_SupportedDistribution(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname returns Linux
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout:   "Linux",
		ExitCode: 0,
	}
	// os-release shows Ubuntu
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		Stdout:   "ID=ubuntu\n",
		ExitCode: 0,
	}
	// Install succeeds
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
	}
	// Verify installation
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_SupportedDistributionDebian(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname returns Linux
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout:   "Linux",
		ExitCode: 0,
	}
	// os-release shows Debian
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		Stdout:   "ID=debian\n",
		ExitCode: 0,
	}
	// Install succeeds
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
	}
	// Verify installation
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_OSCheckFailsGracefully(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname fails (should proceed gracefully)
	commander.Commands["app-1 uname -s"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("uname failed"),
	}
	// Install succeeds (OS check failed but we proceed)
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
	}
	// Verify installation
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil (should proceed when OS check fails)", err)
	}
}

func TestTailscaleProvider_EnsureInstalled_LSBReleaseFallback(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// tailscale version fails (not installed)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}
	// uname returns Linux
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout:   "Linux",
		ExitCode: 0,
	}
	// os-release doesn't exist
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("file not found"),
	}
	// lsb_release shows Ubuntu
	commander.Commands["app-1 lsb_release -i -s"] = CommandResult{
		Stdout:   "Ubuntu",
		ExitCode: 0,
	}
	// Install succeeds
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
	}
	// Verify installation
	commander.Commands["app-1 tailscale version"] = CommandResult{
		Stdout:   "1.78.0",
		ExitCode: 0,
	}

	opts := network.EnsureInstalledOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
		},
		Host: "app-1",
	}

	err := provider.EnsureInstalled(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureInstalled() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureJoined_JoinSucceeds(t *testing.T) {
	_ = os.Setenv("TS_AUTHKEY", "test-auth-key")
	defer func() { _ = os.Unsetenv("TS_AUTHKEY") }()

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// Status check fails (not joined)
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("not joined"),
	}
	// Join succeeds
	joinCmd := "tailscale up --authkey=test-auth-key --hostname=app-1 --advertise-tags=tag:stagecraft,tag:app"
	commander.Commands["app-1 sh -c "+joinCmd] = CommandResult{
		ExitCode: 0,
	}
	commander.Commands["app-1 "+joinCmd] = CommandResult{
		ExitCode: 0,
	}
	// Re-check status shows correct tailnet and tags
	statusJSON := `{
		"TailnetName": "example.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": ["tag:stagecraft", "tag:app"]
		}
	}`
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		Stdout:   statusJSON,
		ExitCode: 0,
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
			"default_tags":   []string{"tag:stagecraft"},
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureJoined() error = %v, want nil", err)
	}
}

func TestTailscaleProvider_EnsureJoined_StatusParseError(t *testing.T) {
	_ = os.Setenv("TS_AUTHKEY", "test-auth-key")
	defer func() { _ = os.Unsetenv("TS_AUTHKEY") }()

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// Status check returns invalid JSON
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		Stdout:   "invalid json",
		ExitCode: 0,
	}
	// Join succeeds
	joinCmd := "tailscale up --authkey=test-auth-key --hostname=app-1 --advertise-tags=tag:stagecraft,tag:app"
	commander.Commands["app-1 sh -c "+joinCmd] = CommandResult{
		ExitCode: 0,
	}
	commander.Commands["app-1 "+joinCmd] = CommandResult{
		ExitCode: 0,
	}
	// Re-check status returns valid JSON
	statusJSON := `{
		"TailnetName": "example.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": ["tag:stagecraft", "tag:app"]
		}
	}`
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		Stdout:   statusJSON,
		ExitCode: 0,
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
			"default_tags":   []string{"tag:stagecraft"},
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err != nil {
		t.Errorf("EnsureJoined() error = %v, want nil (should handle parse error and proceed)", err)
	}
}

func TestTailscaleProvider_EnsureJoined_JoinFails(t *testing.T) {
	_ = os.Setenv("TS_AUTHKEY", "test-auth-key")
	defer func() { _ = os.Unsetenv("TS_AUTHKEY") }()

	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)
	// Status check fails (not joined)
	commander.Commands["app-1 tailscale status --json"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("not joined"),
	}
	// Join fails
	joinCmd := "tailscale up --authkey=test-auth-key --hostname=app-1 --advertise-tags=tag:stagecraft,tag:app"
	commander.Commands["app-1 sh -c "+joinCmd] = CommandResult{
		ExitCode: 1,
		Stderr:   "join failed",
		Error:    fmt.Errorf("join failed"),
	}
	commander.Commands["app-1 "+joinCmd] = CommandResult{
		ExitCode: 1,
		Stderr:   "join failed",
		Error:    fmt.Errorf("join failed"),
	}

	opts := network.EnsureJoinedOptions{
		Config: map[string]interface{}{
			"auth_key_env":   "TS_AUTHKEY",
			"tailnet_domain": "example.ts.net",
			"default_tags":   []string{"tag:stagecraft"},
		},
		Host: "app-1",
		Tags: []string{"tag:app"},
	}

	err := provider.EnsureJoined(context.Background(), opts)
	if err == nil {
		t.Error("EnsureJoined() expected error for join failure, got nil")
	}
	if !contains(err.Error(), "join failed") {
		t.Errorf("EnsureJoined() error = %q, want substring %q", err.Error(), "join failed")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
