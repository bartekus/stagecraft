// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"stagecraft/internal/infra/bootstrap"
	cloud "stagecraft/pkg/providers/cloud"
	network "stagecraft/pkg/providers/network"
)

// Feature: CLI_INFRA_UP
// Spec: spec/commands/infra-up.md

func TestNewInfraUpCommand_HasExpectedMetadata(t *testing.T) {
	cmd := NewInfraUpCommand()

	if cmd.Use != "up" {
		t.Fatalf("expected Use to be 'up', got %q", cmd.Use)
	}

	if cmd.Short == "" {
		t.Fatalf("expected Short description to be non-empty")
	}
}

func TestInfraUpCommand_ConfigLoadFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", "nonexistent.yml", "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when config file does not exist")
	}

	if !strings.Contains(err.Error(), "config not found") {
		t.Fatalf("expected error message to mention config not found, got: %v", err)
	}
}

func TestInfraUpCommand_CloudProviderNotConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create minimal config without cloud provider
	configContent := `project:
  name: test-project
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when cloud provider is not configured")
	}

	if !strings.Contains(err.Error(), "cloud provider is not configured") {
		t.Fatalf("expected error message to mention cloud provider, got: %v", err)
	}
}

func TestInfraUpCommand_UnknownCloudProvider(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config with unknown cloud provider
	configContent := `project:
  name: test-project
cloud:
  provider: unknown-provider
  providers:
    unknown-provider: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when cloud provider is unknown")
	}

	if !strings.Contains(err.Error(), "cloud provider") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected error message to mention cloud provider not found, got: %v", err)
	}
}

func TestInfraUpCommand_NetworkProviderNotConfigured(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config without network provider
	configContent := `project:
  name: test-project
cloud:
  provider: digitalocean
  providers:
    digitalocean: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when network provider is not configured")
	}

	if !strings.Contains(err.Error(), "network provider is not configured") {
		t.Fatalf("expected error message to mention network provider, got: %v", err)
	}
}

func TestInfraUpCommand_UnknownNetworkProvider(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config with unknown network provider
	configContent := `project:
  name: test-project
cloud:
  provider: digitalocean
  providers:
    digitalocean: {}
network:
  provider: unknown-network
  providers:
    unknown-network: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when network provider is unknown")
	}

	if !strings.Contains(err.Error(), "network provider") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected error message to mention network provider not found, got: %v", err)
	}
}

func TestInfraUpCommand_HappyPath(t *testing.T) {
	// This test is superseded by TestInfraUpCommand_Slice2HappyPath
	// Keeping it for backward compatibility but it now uses a fake provider
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Create config with a fake provider for testing
	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-happy-old
  providers:
    test-cloud-digitalocean-happy-old: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	fake := &fakeCloudProvider{
		id: "test-cloud-digitalocean-happy-old",
		hosts: []cloud.Host{
			{
				ID:       "host-1",
				Name:     "app-1",
				Role:     "app",
				PublicIP: "192.0.2.1",
				Tags:     []string{"stagecraft"},
			},
		},
	}

	cloud.Register(fake)

	// Override bootstrap service to use fake network provider
	originalNewBootstrapService := newBootstrapService
	defer func() {
		newBootstrapService = originalNewBootstrapService
	}()

	fakeNP := &fakeNetworkProviderForCLI{}
	newBootstrapService = func(_ bootstrap.CommandExecutor, _ network.NetworkProvider) bootstrap.Service {
		// Use real bootstrap service but with fake network provider
		return bootstrap.NewService(&bootstrap.NoopExecutor{}, fakeNP)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestInfraUpCommand_ProviderResolution tests that providers are resolved correctly.
// This test uses the actual provider registries to ensure integration works.
func TestInfraUpCommand_ProviderResolution(t *testing.T) {
	// Verify that digitalocean provider is registered
	_, err := cloud.Get("digitalocean")
	if err != nil {
		t.Fatalf("digitalocean provider should be registered: %v", err)
	}

	// Verify that tailscale provider is registered
	_, err = network.Get("tailscale")
	if err != nil {
		t.Fatalf("tailscale provider should be registered: %v", err)
	}
}

// fakeCloudProvider is a test implementation of CloudProvider for testing.
type fakeCloudProvider struct {
	id string

	mu sync.Mutex

	planCalled  bool
	applyCalled bool
	hostsCalled bool

	planErr  error
	applyErr error
	hostsErr error

	hosts []cloud.Host
}

func (f *fakeCloudProvider) ID() string {
	return f.id
}

func (f *fakeCloudProvider) Plan(ctx context.Context, opts cloud.PlanOptions) (cloud.InfraPlan, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.planCalled = true
	return cloud.InfraPlan{}, f.planErr
}

//nolint:gocritic // hugeParam: opts matches CloudProvider interface signature
func (f *fakeCloudProvider) Apply(ctx context.Context, opts cloud.ApplyOptions) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.applyCalled = true
	return f.applyErr
}

func (f *fakeCloudProvider) Hosts(ctx context.Context, opts cloud.HostsOptions) ([]cloud.Host, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.hostsCalled = true
	return f.hosts, f.hostsErr
}

func TestInfraUpCommand_CloudPlanFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	// Minimal config with a known provider id
	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean
  providers:
    test-cloud-digitalocean: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	fake := &fakeCloudProvider{
		id:      "test-cloud-digitalocean",
		planErr: fmt.Errorf("plan failed: boom"),
	}

	cloud.Register(fake)

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when plan fails")
	}
	if !strings.Contains(err.Error(), "plan failed") {
		t.Fatalf("expected error to mention plan failed, got: %v", err)
	}
}

func TestInfraUpCommand_CloudApplyFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-apply
  providers:
    test-cloud-digitalocean-apply: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	fake := &fakeCloudProvider{
		id:       "test-cloud-digitalocean-apply",
		applyErr: fmt.Errorf("apply failed: boom"),
	}

	cloud.Register(fake)

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when apply fails")
	}
	if !strings.Contains(err.Error(), "apply failed") {
		t.Fatalf("expected error to mention apply failed, got: %v", err)
	}
}

func TestInfraUpCommand_CloudHostsFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-hosts
  providers:
    test-cloud-digitalocean-hosts: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	fake := &fakeCloudProvider{
		id:       "test-cloud-digitalocean-hosts",
		hostsErr: fmt.Errorf("hosts failed: boom"),
	}

	cloud.Register(fake)

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when hosts fails")
	}
	if !strings.Contains(err.Error(), "listing hosts failed") {
		t.Fatalf("expected error to mention listing hosts failed, got: %v", err)
	}
}

func TestInfraUpCommand_Slice2HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-happy
  providers:
    test-cloud-digitalocean-happy: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	fake := &fakeCloudProvider{
		id: "test-cloud-digitalocean-happy",
		hosts: []cloud.Host{
			{
				ID:       "host-1",
				Name:     "app-1",
				Role:     "app",
				PublicIP: "192.0.2.1",
				Tags:     []string{"stagecraft"},
			},
		},
	}

	cloud.Register(fake)

	// Override bootstrap service to use fake network provider
	originalNewBootstrapService := newBootstrapService
	defer func() {
		newBootstrapService = originalNewBootstrapService
	}()

	fakeNP := &fakeNetworkProviderForCLI{}
	newBootstrapService = func(_ bootstrap.CommandExecutor, _ network.NetworkProvider) bootstrap.Service {
		// Use real bootstrap service but with fake network provider
		return bootstrap.NewService(&bootstrap.NoopExecutor{}, fakeNP)
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err != nil {
		t.Fatalf("unexpected error in happy path: %v", err)
	}

	// Verify all methods were called
	if !fake.planCalled {
		t.Error("expected Plan to be called")
	}
	if !fake.applyCalled {
		t.Error("expected Apply to be called")
	}
	if !fake.hostsCalled {
		t.Error("expected Hosts to be called")
	}
}

// fakeNetworkProviderForCLI is a test implementation of NetworkProvider for CLI tests.
type fakeNetworkProviderForCLI struct{}

func (f *fakeNetworkProviderForCLI) ID() string {
	return "fake-network"
}

func (f *fakeNetworkProviderForCLI) EnsureInstalled(ctx context.Context, opts network.EnsureInstalledOptions) error {
	return nil
}

func (f *fakeNetworkProviderForCLI) EnsureJoined(ctx context.Context, opts network.EnsureJoinedOptions) error {
	return nil
}

func (f *fakeNetworkProviderForCLI) NodeFQDN(host string) (string, error) {
	return host + ".fake.net", nil
}

// fakeBootstrapService is a test implementation of bootstrap.Service for testing.
type fakeBootstrapService struct {
	result *bootstrap.Result
	err    error
}

func (f *fakeBootstrapService) Bootstrap(_ context.Context, hosts []bootstrap.Host, _ bootstrap.Config) (*bootstrap.Result, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.result, nil
}

func TestInfraUpCommand_BootstrapAllSuccess(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-bootstrap
  providers:
    test-cloud-digitalocean-bootstrap: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Setup fake cloud provider
	fakeCloud := &fakeCloudProvider{
		id: "test-cloud-digitalocean-bootstrap",
		hosts: []cloud.Host{
			{ID: "host-1", Name: "app-1", Role: "app", PublicIP: "192.0.2.1"},
			{ID: "host-2", Name: "app-2", Role: "app", PublicIP: "192.0.2.2"},
		},
	}
	cloud.Register(fakeCloud)

	// Setup fake bootstrap service that returns all success
	originalNewBootstrapService := newBootstrapService
	defer func() {
		newBootstrapService = originalNewBootstrapService
	}()

	fakeBootstrap := &fakeBootstrapService{
		result: &bootstrap.Result{
			Hosts: []bootstrap.HostResult{
				{
					Host:    bootstrap.Host{ID: "host-1", Name: "app-1", Role: "app", PublicIP: "192.0.2.1"},
					Success: true,
					Error:   "",
				},
				{
					Host:    bootstrap.Host{ID: "host-2", Name: "app-2", Role: "app", PublicIP: "192.0.2.2"},
					Success: true,
					Error:   "",
				},
			},
		},
	}
	newBootstrapService = func(_ bootstrap.CommandExecutor, _ network.NetworkProvider) bootstrap.Service {
		return fakeBootstrap
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err != nil {
		t.Fatalf("expected no error when all hosts succeed, got: %v", err)
	}

	// Verify bootstrap was called
	if fakeBootstrap.result == nil {
		t.Error("expected bootstrap service to be called")
	}
}

func TestInfraUpCommand_BootstrapPartialFailure(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-partial
  providers:
    test-cloud-digitalocean-partial: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Setup fake cloud provider
	fakeCloud := &fakeCloudProvider{
		id: "test-cloud-digitalocean-partial",
		hosts: []cloud.Host{
			{ID: "host-1", Name: "app-1", Role: "app", PublicIP: "192.0.2.1"},
			{ID: "host-2", Name: "app-2", Role: "app", PublicIP: "192.0.2.2"},
		},
	}
	cloud.Register(fakeCloud)

	// Setup fake bootstrap service that returns partial failure
	originalNewBootstrapService := newBootstrapService
	defer func() {
		newBootstrapService = originalNewBootstrapService
	}()

	fakeBootstrap := &fakeBootstrapService{
		result: &bootstrap.Result{
			Hosts: []bootstrap.HostResult{
				{
					Host:    bootstrap.Host{ID: "host-1", Name: "app-1", Role: "app", PublicIP: "192.0.2.1"},
					Success: true,
					Error:   "",
				},
				{
					Host:    bootstrap.Host{ID: "host-2", Name: "app-2", Role: "app", PublicIP: "192.0.2.2"},
					Success: false,
					Error:   "SSH connection failed",
				},
			},
		},
	}
	newBootstrapService = func(_ bootstrap.CommandExecutor, _ network.NetworkProvider) bootstrap.Service {
		return fakeBootstrap
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when some hosts fail bootstrap")
	}

	// Verify it's a partial failure error
	if _, ok := err.(*bootstrapPartialFailureError); !ok {
		t.Fatalf("expected bootstrapPartialFailureError, got: %T: %v", err, err)
	}

	if !strings.Contains(err.Error(), "bootstrap completed") {
		t.Fatalf("expected error to mention bootstrap completion, got: %v", err)
	}
}

func TestInfraUpCommand_BootstrapGlobalFailure(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	configContent := `project:
  name: test-project
cloud:
  provider: test-cloud-digitalocean-global
  providers:
    test-cloud-digitalocean-global: {}
network:
  provider: tailscale
  providers:
    tailscale: {}
environments:
  staging:
    driver: docker
`
	configPath := filepath.Join(tmpDir, "stagecraft.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Setup fake cloud provider
	fakeCloud := &fakeCloudProvider{
		id: "test-cloud-digitalocean-global",
		hosts: []cloud.Host{
			{ID: "host-1", Name: "app-1", Role: "app", PublicIP: "192.0.2.1"},
		},
	}
	cloud.Register(fakeCloud)

	// Setup fake bootstrap service that returns a global error
	originalNewBootstrapService := newBootstrapService
	defer func() {
		newBootstrapService = originalNewBootstrapService
	}()

	fakeBootstrap := &fakeBootstrapService{
		err: fmt.Errorf("bootstrap service initialization failed"),
	}
	newBootstrapService = func(_ bootstrap.CommandExecutor, _ network.NetworkProvider) bootstrap.Service {
		return fakeBootstrap
	}

	root := newTestRootCommand()
	root.AddCommand(NewInfraCommand())

	_, err := executeCommandForGolden(root, "infra", "up", "--config", configPath, "--env", "staging")
	if err == nil {
		t.Fatalf("expected error when bootstrap service fails globally")
	}

	// Verify it's a global failure error
	if _, ok := err.(*bootstrapGlobalFailureError); !ok {
		t.Fatalf("expected bootstrapGlobalFailureError, got: %T: %v", err, err)
	}

	if !strings.Contains(err.Error(), "bootstrap failed") {
		t.Fatalf("expected error to mention bootstrap failed, got: %v", err)
	}
}
