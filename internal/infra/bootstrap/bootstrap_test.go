// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Feature: INFRA_HOST_BOOTSTRAP
// Spec: spec/infra/bootstrap.md

package bootstrap

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"stagecraft/pkg/providers/network"
)

// fakeExecutor is a test implementation of CommandExecutor that records
// command calls and responds based on injected behavior.
type fakeExecutor struct {
	mu sync.Mutex

	// commands records all (host, command) pairs that were executed
	commands []struct {
		Host    Host
		Command string
	}

	// behavior determines the response for a given (host, command) pair
	behavior func(host Host, cmd string) (stdout, stderr string, err error)
}

//nolint:gocritic // hugeParam: host matches CommandExecutor interface signature
func (f *fakeExecutor) Run(ctx context.Context, host Host, command string) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.commands = append(f.commands, struct {
		Host    Host
		Command string
	}{Host: host, Command: command})

	if f.behavior != nil {
		return f.behavior(host, command)
	}

	return "", "", nil
}

func (f *fakeExecutor) getCommands() []struct {
	Host    Host
	Command string
} {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]struct {
		Host    Host
		Command string
	}, len(f.commands))
	copy(result, f.commands)
	return result
}

func TestBootstrap_DockerAlreadyInstalled(t *testing.T) {
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				return "Docker version 24.0.0", "", nil
			}
			return "", "", fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
		{ID: "host-2", Name: "app-2", PublicIP: "192.0.2.2"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(result.Hosts) != 2 {
		t.Fatalf("expected 2 host results, got %d", len(result.Hosts))
	}

	if !result.AllSucceeded() {
		t.Fatalf("expected all hosts to succeed, got failures: %v", result)
	}

	// Verify only detection commands were run (no install commands)
	commands := exec.getCommands()
	dockerVersionCount := 0
	for _, cmd := range commands {
		if cmd.Command == "docker version" {
			dockerVersionCount++
		}
		if strings.Contains(cmd.Command, "apt-get") || strings.Contains(cmd.Command, "systemctl") {
			t.Errorf("unexpected install command: %s", cmd.Command)
		}
	}

	if dockerVersionCount != 2 {
		t.Errorf("expected 2 'docker version' calls (one per host), got %d", dockerVersionCount)
	}
}

func TestBootstrap_DockerMissingInstallSuccess(t *testing.T) {
	callCount := 0
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				callCount++
				if callCount == 1 {
					// First call: Docker not found
					return "", "", fmt.Errorf("docker: command not found")
				}
				// Second call (after install): Docker found
				return "Docker version 24.0.0", "", nil
			}

			// Install commands should succeed
			if cmd == "apt-get update -y" || cmd == "apt-get install -y docker.io" || cmd == "systemctl enable --now docker" {
				return "", "", nil
			}

			return "", "", fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.AllSucceeded() {
		t.Fatalf("expected host to succeed after install, got: %v", result.Hosts[0])
	}

	// Verify command sequence
	commands := exec.getCommands()
	expectedCommands := []string{
		"docker version",                // Detection (fails)
		"apt-get update -y",             // Install step 1
		"apt-get install -y docker.io",  // Install step 2
		"systemctl enable --now docker", // Install step 3
		"docker version",                // Verification (succeeds)
	}

	if len(commands) != len(expectedCommands) {
		t.Fatalf("expected %d commands, got %d: %v", len(expectedCommands), len(commands), commands)
	}

	for i, cmd := range commands {
		if cmd.Command != expectedCommands[i] {
			t.Errorf("command %d: expected %q, got %q", i, expectedCommands[i], cmd.Command)
		}
	}
}

func TestBootstrap_DockerInstallFails(t *testing.T) {
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				// Docker not found
				return "", "", fmt.Errorf("docker: command not found")
			}

			if cmd == "apt-get install -y docker.io" {
				// Install fails
				return "", "E: Unable to locate package docker.io", fmt.Errorf("apt-get install failed")
			}

			// Other commands succeed
			if cmd == "apt-get update -y" {
				return "", "", nil
			}

			return "", "", fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	if result.AllSucceeded() {
		t.Fatalf("expected host to fail, got success")
	}

	if result.FailureCount() != 1 {
		t.Fatalf("expected 1 failure, got %d", result.FailureCount())
	}

	hostResult := result.Hosts[0]
	if hostResult.Success {
		t.Fatalf("expected host result to indicate failure")
	}

	if !strings.Contains(hostResult.Error, "docker install failed") {
		t.Errorf("expected error to mention 'docker install failed', got: %s", hostResult.Error)
	}
}

func TestBootstrap_DockerVerificationFailsAfterInstall(t *testing.T) {
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				// Docker always fails (even after "install")
				return "", "", fmt.Errorf("docker: command not found")
			}

			// Install commands "succeed" but Docker still doesn't work
			if cmd == "apt-get update -y" || cmd == "apt-get install -y docker.io" || cmd == "systemctl enable --now docker" {
				return "", "", nil
			}

			return "", "", fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	if result.AllSucceeded() {
		t.Fatalf("expected host to fail verification, got success")
	}

	hostResult := result.Hosts[0]
	if !strings.Contains(hostResult.Error, "docker verification failed") {
		t.Errorf("expected error to mention 'docker verification failed', got: %s", hostResult.Error)
	}
}

func TestBootstrap_MultipleHostsMixedResults(t *testing.T) {
	hostCallCounts := make(map[string]int)
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			hostID := host.ID
			if cmd == "docker version" {
				hostCallCounts[hostID]++
				count := hostCallCounts[hostID]

				// Host 1: Docker already installed
				if hostID == "host-1" {
					return "Docker version 24.0.0", "", nil
				}

				// Host 2: Docker missing, install succeeds
				if hostID == "host-2" {
					if count == 1 {
						return "", "", fmt.Errorf("docker: command not found")
					}
					return "Docker version 24.0.0", "", nil
				}

				// Host 3: Docker install fails
				if hostID == "host-3" {
					return "", "", fmt.Errorf("docker: command not found")
				}
			}

			// Install commands
			if strings.Contains(cmd, "apt-get") || strings.Contains(cmd, "systemctl") {
				if hostID == "host-3" && cmd == "apt-get install -y docker.io" {
					return "", "E: Unable to locate package", fmt.Errorf("install failed")
				}
				return "", "", nil
			}

			return "", "", fmt.Errorf("unexpected command: %s", cmd)
		},
	}

	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
		{ID: "host-2", Name: "app-2", PublicIP: "192.0.2.2"},
		{ID: "host-3", Name: "app-3", PublicIP: "192.0.2.3"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	if result.SuccessCount() != 2 {
		t.Errorf("expected 2 successes, got %d", result.SuccessCount())
	}

	if result.FailureCount() != 1 {
		t.Errorf("expected 1 failure, got %d", result.FailureCount())
	}

	if result.Hosts[0].Host.ID != "host-1" || !result.Hosts[0].Success {
		t.Errorf("expected host-1 to succeed")
	}

	if result.Hosts[1].Host.ID != "host-2" || !result.Hosts[1].Success {
		t.Errorf("expected host-2 to succeed")
	}

	if result.Hosts[2].Host.ID != "host-3" || result.Hosts[2].Success {
		t.Errorf("expected host-3 to fail")
	}
}

// fakeNetworkProvider is a test implementation of NetworkProvider for testing.
type fakeNetworkProvider struct {
	mu         sync.Mutex
	installed  []string
	joined     []string
	installErr error
	joinErr    error
}

func (f *fakeNetworkProvider) ID() string {
	return "fake-network"
}

func (f *fakeNetworkProvider) EnsureInstalled(ctx context.Context, opts network.EnsureInstalledOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.installed = append(f.installed, opts.Host)
	return f.installErr
}

func (f *fakeNetworkProvider) EnsureJoined(ctx context.Context, opts network.EnsureJoinedOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.joined = append(f.joined, opts.Host)
	return f.joinErr
}

func (f *fakeNetworkProvider) NodeFQDN(host string) (string, error) {
	return host + ".fake.net", nil
}

func (f *fakeNetworkProvider) getInstalled() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]string, len(f.installed))
	copy(result, f.installed)
	return result
}

func (f *fakeNetworkProvider) getJoined() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := make([]string, len(f.joined))
	copy(result, f.joined)
	return result
}

func TestBootstrap_TailscaleAllSuccess(t *testing.T) {
	// Docker executor always succeeds
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				return "Docker version 24.0.0", "", nil
			}
			return "", "", nil
		},
	}

	// Network provider succeeds
	np := &fakeNetworkProvider{}

	svc := NewService(exec, np)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1", Tags: []string{"tag:app"}},
		{ID: "host-2", Name: "app-2", PublicIP: "192.0.2.2", Tags: []string{"tag:app"}},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.AllSucceeded() {
		t.Fatalf("expected all hosts to succeed, got failures: %v", result)
	}

	// Verify network provider was called for each host
	installed := np.getInstalled()
	if len(installed) != 2 {
		t.Errorf("expected 2 EnsureInstalled calls, got %d: %v", len(installed), installed)
	}

	joined := np.getJoined()
	if len(joined) != 2 {
		t.Errorf("expected 2 EnsureJoined calls, got %d: %v", len(joined), joined)
	}

	// Verify hostnames used
	if installed[0] != "app-1" || installed[1] != "app-2" {
		t.Errorf("expected installed hosts to be app-1 and app-2, got: %v", installed)
	}

	if joined[0] != "app-1" || joined[1] != "app-2" {
		t.Errorf("expected joined hosts to be app-1 and app-2, got: %v", joined)
	}
}

func TestBootstrap_TailscaleInstallFails(t *testing.T) {
	// Docker executor always succeeds
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				return "Docker version 24.0.0", "", nil
			}
			return "", "", nil
		},
	}

	// Network provider install fails
	np := &fakeNetworkProvider{
		installErr: fmt.Errorf("tailscale install failed"),
	}

	svc := NewService(exec, np)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	if result.AllSucceeded() {
		t.Fatalf("expected host to fail, got success")
	}

	hostResult := result.Hosts[0]
	if !strings.Contains(hostResult.Error, "tailscale install failed") {
		t.Errorf("expected error to mention 'tailscale install failed', got: %s", hostResult.Error)
	}

	// Verify EnsureJoined was not called (install failed first)
	joined := np.getJoined()
	if len(joined) > 0 {
		t.Errorf("expected no EnsureJoined calls after install failure, got: %v", joined)
	}
}

func TestBootstrap_TailscaleJoinFails(t *testing.T) {
	// Docker executor always succeeds
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				return "Docker version 24.0.0", "", nil
			}
			return "", "", nil
		},
	}

	// Network provider install succeeds but join fails
	np := &fakeNetworkProvider{
		joinErr: fmt.Errorf("tailscale join failed"),
	}

	svc := NewService(exec, np)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no top-level error, got: %v", err)
	}

	if result.AllSucceeded() {
		t.Fatalf("expected host to fail, got success")
	}

	hostResult := result.Hosts[0]
	if !strings.Contains(hostResult.Error, "tailscale join failed") {
		t.Errorf("expected error to mention 'tailscale join failed', got: %s", hostResult.Error)
	}

	// Verify EnsureInstalled was called (install succeeded)
	installed := np.getInstalled()
	if len(installed) != 1 {
		t.Errorf("expected 1 EnsureInstalled call, got %d: %v", len(installed), installed)
	}
}

func TestBootstrap_NoNetworkProvider(t *testing.T) {
	// Docker executor always succeeds
	exec := &fakeExecutor{
		behavior: func(host Host, cmd string) (string, string, error) {
			if cmd == "docker version" {
				return "Docker version 24.0.0", "", nil
			}
			return "", "", nil
		},
	}

	// No network provider (nil)
	svc := NewService(exec, nil)
	hosts := []Host{
		{ID: "host-1", Name: "app-1", PublicIP: "192.0.2.1"},
	}

	cfg := Config{SSHUser: "root"}
	result, err := svc.Bootstrap(context.Background(), hosts, cfg)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !result.AllSucceeded() {
		t.Fatalf("expected all hosts to succeed when network provider is nil, got failures: %v", result)
	}
}
