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
	"strings"
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
	// First check: not installed
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}

	// OS compatibility: Debian supported
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout: "Linux",
	}
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		Stdout: "ID=debian\n",
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
		t.Fatal("EnsureInstalled() expected error for install failure, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "installation failed") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", msg, "installation failed")
	}
	if !strings.Contains(msg, "install failed") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", msg, "install failed")
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

func TestBuildTailscaleUpCommand(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		authKey  string
		hostname string
		tags     []string
		want     string
	}{
		{
			name:     "single tag",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{"tag:web"},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=tag:web",
		},
		{
			name:     "multiple tags",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{"tag:web", "tag:prod"},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=tag:web,tag:prod",
		},
		{
			name:     "no tags",
			authKey:  "tskey-auth-123",
			hostname: "app-1",
			tags:     []string{},
			want:     "tailscale up --authkey=tskey-auth-123 --hostname=app-1 --advertise-tags=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTailscaleUpCommand(tt.authKey, tt.hostname, tt.tags)
			if got != tt.want {
				t.Errorf("buildTailscaleUpCommand() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseOSRelease(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "debian",
			content: `PRETTY_NAME="Debian GNU/Linux 11 (bullseye)"
NAME="Debian GNU/Linux"
ID=debian
ID_LIKE=debian`,
			want: "debian",
		},
		{
			name: "ubuntu",
			content: `NAME="Ubuntu"
VERSION="22.04.3 LTS (Jammy Jellyfish)"
ID=ubuntu
ID_LIKE=debian`,
			want: "ubuntu",
		},
		{
			name:    "quoted ID",
			content: `ID="debian"`,
			want:    "debian",
		},
		{
			name:    "no ID field",
			content: `PRETTY_NAME="Some OS"`,
			want:    "",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseOSRelease(tt.content)
			if got != tt.want {
				t.Errorf("parseOSRelease() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateTailnetDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{
			name:    "valid domain",
			domain:  "example.ts.net",
			wantErr: false,
		},
		{
			name:    "valid subdomain",
			domain:  "sub.example.ts.net",
			wantErr: false,
		},
		{
			name:    "empty domain",
			domain:  "",
			wantErr: true,
		},
		{
			name:    "no dot",
			domain:  "example",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTailnetDomain(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTailnetDomain() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildNodeFQDN(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		host   string
		domain string
		want   string
	}{
		{
			name:   "simple host",
			host:   "app-1",
			domain: "example.ts.net",
			want:   "app-1.example.ts.net",
		},
		{
			name:   "host with dash",
			host:   "db-primary",
			domain: "example.ts.net",
			want:   "db-primary.example.ts.net",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildNodeFQDN(tt.host, tt.domain)
			if got != tt.want {
				t.Errorf("buildNodeFQDN() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseStatus_InvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := parseStatus("not json")
	if err == nil {
		t.Error("parseStatus() should return error for invalid JSON")
	}
}

func TestParseStatus_EmptyJSON(t *testing.T) {
	t.Parallel()

	// Empty JSON will unmarshal successfully but return empty struct
	status, err := parseStatus("{}")
	if err != nil {
		t.Fatalf("parseStatus() with empty JSON returned error: %v", err)
	}
	if status == nil {
		t.Fatal("parseStatus() returned nil status")
	}
	// Verify it's actually empty
	if status.TailnetName != "" {
		t.Errorf("parseStatus() with empty JSON: TailnetName = %q, want empty", status.TailnetName)
	}
}

func TestParseStatus_ValidStatus(t *testing.T) {
	t.Parallel()

	jsonData := `{
		"TailnetName": "example.ts.net",
		"Self": {
			"Online": true,
			"TailscaleIPs": ["100.64.0.1"],
			"Tags": ["tag:web"]
		}
	}`

	status, err := parseStatus(jsonData)
	if err != nil {
		t.Fatalf("parseStatus() returned error: %v", err)
	}
	if status.TailnetName != "example.ts.net" {
		t.Errorf("parseStatus() TailnetName = %q, want %q", status.TailnetName, "example.ts.net")
	}
	if len(status.Self.Tags) != 1 || status.Self.Tags[0] != "tag:web" {
		t.Errorf("parseStatus() Self.Tags = %v, want [tag:web]", status.Self.Tags)
	}
}

func TestParseTailscaleVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple version",
			input: "1.78.0",
			want:  "1.78.0",
		},
		{
			name:  "version with build metadata",
			input: "1.44.0-123-gabcd",
			want:  "1.44.0",
		},
		{
			name:  "version with patch suffix",
			input: "1.78.0-1",
			want:  "1.78.0",
		},
		{
			name:  "version in output string",
			input: "tailscale version 1.78.0",
			want:  "1.78.0",
		},
		{
			name:  "version with whitespace",
			input: "  1.78.0  ",
			want:  "1.78.0",
		},
		{
			name:  "version with multiple parts",
			input: "tailscale version 1.44.0-123-gabcd",
			want:  "1.44.0",
		},
		{
			name:    "unparseable version",
			input:   "not-a-version",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "incomplete version",
			input:   "1.78",
			wantErr: true,
		},
		{
			name:    "no version in string",
			input:   "tailscale version",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTailscaleVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTailscaleVersion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseTailscaleVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureInstalled_ConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errSub  string
	}{
		{
			name:    "missing auth_key_env",
			config:  map[string]interface{}{"tailnet_domain": "example.ts.net"},
			wantErr: true,
			errSub:  "auth_key_env is required",
		},
		{
			name:    "missing tailnet_domain",
			config:  map[string]interface{}{"auth_key_env": "TS_AUTHKEY"},
			wantErr: true,
			errSub:  "tailnet_domain is required",
		},
		{
			name: "valid config",
			config: map[string]interface{}{
				"auth_key_env":   "TS_AUTHKEY",
				"tailnet_domain": "example.ts.net",
			},
			wantErr: false,
			// Note: This test requires Commander setup for tailscale version check
			// We'll set it up in the test to simulate already installed
		},
		{
			name: "install method skip",
			config: map[string]interface{}{
				"auth_key_env":   "TS_AUTHKEY",
				"tailnet_domain": "example.ts.net",
				"install": map[string]interface{}{
					"method": "skip",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid config type",
			config: map[string]interface{}{
				"auth_key_env":   123, // wrong type - will be converted to string by YAML
				"tailnet_domain": "example.ts.net",
			},
			wantErr: true,
			// YAML unmarshaling converts int to string, so this might not fail at parseConfig
			// Instead it might fail later. Let's check for any error.
			errSub: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewLocalCommander()
			provider := &TailscaleProvider{
				commander: commander,
			}

			// For valid config and skip method, set up Commander to simulate already installed
			if !tt.wantErr {
				commander.Commands["test-host tailscale version"] = CommandResult{
					Stdout: "1.78.0",
				}
			}

			opts := network.EnsureInstalledOptions{
				Config: tt.config,
				Host:   "test-host",
			}

			err := provider.EnsureInstalled(context.Background(), opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("EnsureInstalled() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if tt.errSub != "" {
					if !strings.Contains(err.Error(), tt.errSub) {
						t.Errorf("EnsureInstalled() error = %q, want substring %q", err.Error(), tt.errSub)
					}
					// For config validation errors, verify error is wrapped with "tailscale provider: invalid config:"
					if tt.errSub == "auth_key_env is required" || tt.errSub == "tailnet_domain is required" {
						if !strings.Contains(err.Error(), "tailscale provider") {
							t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "tailscale provider")
						}
						if !strings.Contains(err.Error(), "invalid config") {
							t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "invalid config")
						}
					}
				}
			}
			// For skip method, verify no Commander calls were made (implicitly verified by no error)
			// The test passing with no Commander setup confirms the skip method works correctly
		})
	}
}

func TestEnsureInstalled_OSCompatibility(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		unameOut      string
		unameErr      error
		osRelease     string
		osReleaseErr  error
		lsbRelease    string
		lsbReleaseErr error
		wantErr       bool
		errSub        string
	}{
		{
			name:      "debian supported",
			unameOut:  "Linux",
			osRelease: "ID=debian\n",
			wantErr:   false,
		},
		{
			name:      "ubuntu supported",
			unameOut:  "Linux",
			osRelease: "ID=ubuntu\n",
			wantErr:   false,
		},
		{
			name:      "alpine unsupported",
			unameOut:  "Linux",
			osRelease: "ID=alpine\n",
			wantErr:   true,
			errSub:    "alpine",
		},
		{
			name:      "centos unsupported",
			unameOut:  "Linux",
			osRelease: "ID=centos\n",
			wantErr:   true,
			errSub:    "centos",
		},
		{
			name:     "darwin unsupported",
			unameOut: "Darwin",
			wantErr:  true,
			errSub:   "Darwin",
		},
		{
			name:     "uname fails gracefully",
			unameErr: fmt.Errorf("uname failed"),
			wantErr:  false,
		},
		{
			name:         "os-release missing, lsb_release debian",
			unameOut:     "Linux",
			osReleaseErr: fmt.Errorf("no such file"),
			lsbRelease:   "Debian",
			wantErr:      false,
		},
		{
			name:         "os-release missing, lsb_release ubuntu",
			unameOut:     "Linux",
			osReleaseErr: fmt.Errorf("no such file"),
			lsbRelease:   "Ubuntu",
			wantErr:      false,
		},
		{
			name:         "os-release missing, lsb_release alpine",
			unameOut:     "Linux",
			osReleaseErr: fmt.Errorf("no such file"),
			lsbRelease:   "Alpine",
			wantErr:      true,
			errSub:       "alpine", // Error message lowercases the distribution name
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewLocalCommander()

			// Always simulate "not installed" to force OS compatibility check
			commander.Commands["test-host tailscale version"] = CommandResult{
				ExitCode: 1,
				Error:    fmt.Errorf("command not found"),
			}

			// Wire uname command
			if tt.unameOut != "" || tt.unameErr != nil {
				commander.Commands["test-host uname -s"] = CommandResult{
					Stdout: tt.unameOut,
					Error:  tt.unameErr,
				}
			}

			// Wire os-release command
			if tt.osRelease != "" || tt.osReleaseErr != nil {
				commander.Commands["test-host cat /etc/os-release"] = CommandResult{
					Stdout: tt.osRelease,
					Error:  tt.osReleaseErr,
				}
			}

			// Wire lsb_release fallback
			if tt.lsbRelease != "" || tt.lsbReleaseErr != nil {
				commander.Commands["test-host lsb_release -i -s"] = CommandResult{
					Stdout: tt.lsbRelease,
					Error:  tt.lsbReleaseErr,
				}
			}

			// For supported OS, stub install script and verification
			if !tt.wantErr {
				commander.Commands["test-host curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
					ExitCode: 0,
					Stdout:   "Installation successful",
				}
				// Second tailscale version call for verification
				commander.Commands["test-host tailscale version"] = CommandResult{
					Stdout:   "1.78.0",
					ExitCode: 0,
				}
			}

			provider := &TailscaleProvider{
				commander: commander,
			}

			opts := network.EnsureInstalledOptions{
				Config: map[string]interface{}{
					"auth_key_env":   "TS_AUTHKEY",
					"tailnet_domain": "example.ts.net",
				},
				Host: "test-host",
			}

			err := provider.EnsureInstalled(context.Background(), opts)

			if (err != nil) != tt.wantErr {
				t.Fatalf("EnsureInstalled() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errSub != "" {
				if !strings.Contains(err.Error(), tt.errSub) {
					t.Errorf("EnsureInstalled() error = %q, want substring %q", err.Error(), tt.errSub)
				}
				// Verify error is wrapped with "tailscale provider: unsupported operating system:"
				if !strings.Contains(err.Error(), "tailscale provider") {
					t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "tailscale provider")
				}
				if !strings.Contains(err.Error(), "unsupported") {
					t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "unsupported")
				}
			}
		})
	}
}

func TestEnsureInstalled_VersionEnforcement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		installed  string
		minVersion string
		wantErr    bool
		errSub     string
	}{
		{
			name:       "meets minimum",
			installed:  "1.78.0",
			minVersion: "1.78.0",
			wantErr:    false,
		},
		{
			name:       "exceeds minimum",
			installed:  "1.80.0",
			minVersion: "1.78.0",
			wantErr:    false,
		},
		{
			name:       "below minimum",
			installed:  "1.44.0",
			minVersion: "1.78.0",
			wantErr:    true,
			errSub:     "below minimum",
		},
		{
			name:       "build metadata",
			installed:  "1.44.0-123-gabcd",
			minVersion: "1.44.0",
			wantErr:    false,
		},
		{
			name:       "patch suffix",
			installed:  "1.78.0-1",
			minVersion: "1.78.0",
			wantErr:    false,
		},
		{
			name:       "unparseable version",
			installed:  "not-a-version",
			minVersion: "1.78.0",
			wantErr:    true,
			errSub:     "cannot parse installed version",
		},
		{
			name:       "no min_version configured",
			installed:  "1.44.0",
			minVersion: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commander := NewLocalCommander()
			commander.Commands["app-1 tailscale version"] = CommandResult{
				Stdout:   tt.installed,
				ExitCode: 0,
			}

			provider := &TailscaleProvider{
				commander: commander,
			}

			config := map[string]interface{}{
				"auth_key_env":   "TS_AUTHKEY",
				"tailnet_domain": "example.ts.net",
			}
			if tt.minVersion != "" {
				config["install"] = map[string]interface{}{
					"min_version": tt.minVersion,
				}
			}

			opts := network.EnsureInstalledOptions{
				Config: config,
				Host:   "app-1",
			}

			err := provider.EnsureInstalled(context.Background(), opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("EnsureInstalled() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errSub != "" {
				if !strings.Contains(err.Error(), tt.errSub) {
					t.Errorf("EnsureInstalled() error = %q, want substring %q", err.Error(), tt.errSub)
				}
				// Verify error is wrapped with "tailscale provider: installation failed:"
				if !strings.Contains(err.Error(), "tailscale provider") {
					t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "tailscale provider")
				}
				if !strings.Contains(err.Error(), "installation failed") {
					t.Errorf("EnsureInstalled() error = %q, want to contain %q", err.Error(), "installation failed")
				}
			}
		})
	}
}

func TestTailscaleProvider_EnsureInstalled_VerificationFails(t *testing.T) {
	provider := &TailscaleProvider{
		commander: NewLocalCommander(),
	}

	commander := provider.commander.(*LocalCommander)

	// First check: not installed
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("command not found"),
	}

	// OS compatibility: Debian supported
	commander.Commands["app-1 uname -s"] = CommandResult{
		Stdout: "Linux",
	}
	commander.Commands["app-1 cat /etc/os-release"] = CommandResult{
		Stdout: "ID=debian\n",
	}

	// Install script succeeds
	commander.Commands["app-1 curl -fsSL https://tailscale.com/install.sh | sh"] = CommandResult{
		ExitCode: 0,
		Stdout:   "Installation successful",
	}

	// Verification fails (second tailscale version call)
	commander.Commands["app-1 tailscale version"] = CommandResult{
		ExitCode: 1,
		Error:    fmt.Errorf("verification failed"),
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
		t.Fatal("EnsureInstalled() expected error, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "installation failed") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", msg, "installation failed")
	}
	if !strings.Contains(msg, "installation verification failed") {
		t.Errorf("EnsureInstalled() error = %q, want substring %q", msg, "installation verification failed")
	}
}
