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
	"sort"
	"strings"
	"testing"

	"stagecraft/pkg/providers/cloud"
)

func TestDigitalOceanProvider_ID(t *testing.T) {
	t.Parallel()

	provider := NewDigitalOceanProvider()
	if got := provider.ID(); got != "digitalocean" {
		t.Errorf("ID() = %q, want %q", got, "digitalocean")
	}
}

func TestDigitalOceanProvider_RegistryIntegration(t *testing.T) {
	t.Parallel()

	// Provider should be registered
	provider, err := cloud.Get("digitalocean")
	if err != nil {
		t.Fatalf("Get(\"digitalocean\") failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Get(\"digitalocean\") returned nil provider")
	}
	if got := provider.ID(); got != "digitalocean" {
		t.Errorf("provider.ID() = %q, want %q", got, "digitalocean")
	}
}

func TestParseConfig_ValidMinimalConfig(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() failed: %v", err)
	}
	if config.TokenEnv != "DO_TOKEN" {
		t.Errorf("config.TokenEnv = %q, want %q", config.TokenEnv, "DO_TOKEN")
	}
	if config.SSHKeyName != "my-ssh-key" {
		t.Errorf("config.SSHKeyName = %q, want %q", config.SSHKeyName, "my-ssh-key")
	}
	if config.Hosts == nil {
		t.Fatal("config.Hosts is nil")
	}
	if _, ok := config.Hosts["staging"]; !ok {
		t.Error("config.Hosts does not contain \"staging\"")
	}
}

func TestParseConfig_MissingTokenEnv(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if config != nil {
		t.Errorf("parseConfig() returned config %+v, want nil", config)
	}
	if err == nil {
		t.Fatal("parseConfig() returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "token_env is required") {
		t.Errorf("error message %q does not contain \"token_env is required\"", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider: invalid config") {
		t.Errorf("error message %q does not contain \"digitalocean provider: invalid config\"", err.Error())
	}
}

func TestParseConfig_MissingSSHKeyName(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env": "DO_TOKEN",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if config != nil {
		t.Errorf("parseConfig() returned config %+v, want nil", config)
	}
	if err == nil {
		t.Fatal("parseConfig() returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "ssh_key_name is required") {
		t.Errorf("error message %q does not contain \"ssh_key_name is required\"", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider: invalid config") {
		t.Errorf("error message %q does not contain \"digitalocean provider: invalid config\"", err.Error())
	}
}

func TestParseConfig_EmptyHosts(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts":        map[string]any{},
	}

	config, err := parseConfig(cfg)
	if config != nil {
		t.Errorf("parseConfig() returned config %+v, want nil", config)
	}
	if err == nil {
		t.Fatal("parseConfig() returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "hosts configuration is required") {
		t.Errorf("error message %q does not contain \"hosts configuration is required\"", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider: invalid config") {
		t.Errorf("error message %q does not contain \"digitalocean provider: invalid config\"", err.Error())
	}
}

func TestParseConfig_MissingHostRole(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					// Missing role
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if config != nil {
		t.Errorf("parseConfig() returned config %+v, want nil", config)
	}
	if err == nil {
		t.Fatal("parseConfig() returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "role is required") {
		t.Errorf("error message %q does not contain \"role is required\"", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider: invalid config") {
		t.Errorf("error message %q does not contain \"digitalocean provider: invalid config\"", err.Error())
	}
}

func TestParseConfig_OptionalFields(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env":      "DO_TOKEN",
		"ssh_key_name":   "my-ssh-key",
		"default_region": "nyc1",
		"default_size":   "s-2vcpu-4gb",
		"regions":        []string{"nyc1", "nyc3"},
		"sizes":          []string{"s-2vcpu-4gb", "s-4vcpu-8gb"},
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() failed: %v", err)
	}
	if config.DefaultRegion != "nyc1" {
		t.Errorf("config.DefaultRegion = %q, want %q", config.DefaultRegion, "nyc1")
	}
	if config.DefaultSize != "s-2vcpu-4gb" {
		t.Errorf("config.DefaultSize = %q, want %q", config.DefaultSize, "s-2vcpu-4gb")
	}
	if len(config.Regions) != 2 || config.Regions[0] != "nyc1" || config.Regions[1] != "nyc3" {
		t.Errorf("config.Regions = %v, want [nyc1 nyc3]", config.Regions)
	}
	if len(config.Sizes) != 2 || config.Sizes[0] != "s-2vcpu-4gb" || config.Sizes[1] != "s-4vcpu-8gb" {
		t.Errorf("config.Sizes = %v, want [s-2vcpu-4gb s-4vcpu-8gb]", config.Sizes)
	}
}

func TestParseConfig_HostDefaults(t *testing.T) {
	t.Parallel()

	cfg := map[string]any{
		"token_env":      "DO_TOKEN",
		"ssh_key_name":   "my-ssh-key",
		"default_region": "nyc1",
		"default_size":   "s-2vcpu-4gb",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
					// No size or region specified, should use defaults
				},
				"db-1": map[string]any{
					"role":   "db",
					"size":   "s-4vcpu-8gb",
					"region": "nyc3",
				},
			},
		},
	}

	config, err := parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() failed: %v", err)
	}

	stagingHosts := config.Hosts["staging"]
	if stagingHosts == nil {
		t.Fatal("config.Hosts[\"staging\"] is nil")
	}

	app1 := stagingHosts["app-1"]
	if app1.Role != "app" {
		t.Errorf("app1.Role = %q, want %q", app1.Role, "app")
	}
	if app1.Size != "" {
		t.Errorf("app1.Size = %q, want empty string (will use default at runtime)", app1.Size)
	}
	if app1.Region != "" {
		t.Errorf("app1.Region = %q, want empty string (will use default at runtime)", app1.Region)
	}

	db1 := stagingHosts["db-1"]
	if db1.Role != "db" {
		t.Errorf("db1.Role = %q, want %q", db1.Role, "db")
	}
	if db1.Size != "s-4vcpu-8gb" {
		t.Errorf("db1.Size = %q, want %q", db1.Size, "s-4vcpu-8gb")
	}
	if db1.Region != "nyc3" {
		t.Errorf("db1.Region = %q, want %q", db1.Region, "nyc3")
	}
}

func TestNewDigitalOceanProviderWithClient(t *testing.T) {
	t.Parallel()

	// Create a mock client
	mockClient := &mockAPIClient{}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	if provider == nil {
		t.Fatal("NewDigitalOceanProviderWithClient() returned nil")
	}
	if got := provider.ID(); got != "digitalocean" {
		t.Errorf("provider.ID() = %q, want %q", got, "digitalocean")
	}
	if provider.client == nil {
		t.Error("provider.client is nil")
	}
}

// mockAPIClient is a mock for testing Plan() and Apply() operations.
type mockAPIClient struct {
	droplets map[string]Droplet // keyed by droplet.Name
	sshKeys  map[string]SSHKey  // keyed by Name

	// Error injection
	getDropletErr    error
	createDropletErr error
	deleteDropletErr error
	waitErr          error
	listErr          error
	sshKeyErr        error

	// Operation tracking
	created []CreateDropletRequest
	deleted []int
	waited  []struct {
		id     int
		status string
	}
}

func (m *mockAPIClient) ListDroplets(ctx context.Context, filter DropletFilter) ([]Droplet, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}

	if m.droplets == nil {
		return nil, nil
	}

	// Filter by NamePrefix if specified
	var result []Droplet
	for _, d := range m.droplets {
		if filter.NamePrefix != "" && !strings.HasPrefix(d.Name, filter.NamePrefix) {
			continue
		}
		// Tags filtering can be added later if needed
		result = append(result, d)
	}
	return result, nil
}

func (m *mockAPIClient) GetDroplet(ctx context.Context, name string) (*Droplet, error) {
	if m.getDropletErr != nil {
		return nil, m.getDropletErr
	}

	if m.droplets == nil {
		return nil, ErrDropletNotFound
	}

	d, ok := m.droplets[name]
	if !ok {
		return nil, ErrDropletNotFound
	}
	return &d, nil
}

//nolint:gocritic // hugeParam: mock implementation matches interface signature
func (m *mockAPIClient) CreateDroplet(ctx context.Context, req CreateDropletRequest) (*Droplet, error) {
	if m.createDropletErr != nil {
		return nil, m.createDropletErr
	}

	if m.droplets == nil {
		m.droplets = make(map[string]Droplet)
	}

	id := len(m.droplets) + 1
	d := Droplet{
		ID:     id,
		Name:   req.Name,
		Region: req.Region,
		Size:   req.Size,
		Status: "new",
	}

	m.droplets[req.Name] = d
	m.created = append(m.created, req)
	return &d, nil
}

func (m *mockAPIClient) DeleteDroplet(ctx context.Context, id int) error {
	if m.deleteDropletErr != nil {
		return m.deleteDropletErr
	}

	if m.droplets == nil {
		return ErrDropletNotFound
	}

	for name, d := range m.droplets {
		if d.ID == id {
			delete(m.droplets, name)
			m.deleted = append(m.deleted, id)
			return nil
		}
	}
	return ErrDropletNotFound
}

func (m *mockAPIClient) ListSSHKeys(ctx context.Context) ([]SSHKey, error) {
	if m.sshKeys == nil {
		return nil, nil
	}

	var result []SSHKey
	for _, k := range m.sshKeys {
		result = append(result, k)
	}
	return result, nil
}

func (m *mockAPIClient) GetSSHKey(ctx context.Context, name string) (*SSHKey, error) {
	if m.sshKeyErr != nil {
		return nil, m.sshKeyErr
	}

	if m.sshKeys == nil {
		return nil, ErrSSHKeyNotFound
	}

	key, ok := m.sshKeys[name]
	if !ok {
		return nil, ErrSSHKeyNotFound
	}
	return &key, nil
}

func (m *mockAPIClient) WaitForDroplet(ctx context.Context, id int, status string) error {
	if m.waitErr != nil {
		return m.waitErr
	}

	m.waited = append(m.waited, struct {
		id     int
		status string
	}{id: id, status: status})
	return nil
}

func TestDigitalOceanProvider_Plan_HappyPath_NoExistingDroplets(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
				"db-1": map[string]any{
					"role": "db",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 1, Name: "my-ssh-key"},
		},
		droplets: nil,
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "staging",
	})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(plan.ToCreate) != 2 {
		t.Errorf("plan.ToCreate length = %d, want 2", len(plan.ToCreate))
	}
	if len(plan.ToDelete) != 0 {
		t.Errorf("plan.ToDelete length = %d, want 0", len(plan.ToDelete))
	}

	// Verify ToCreate is sorted lexicographically
	if !sort.SliceIsSorted(plan.ToCreate, func(i, j int) bool {
		return plan.ToCreate[i].Name < plan.ToCreate[j].Name
	}) {
		t.Error("plan.ToCreate is not sorted lexicographically")
	}

	// Verify hostnames
	names := make([]string, len(plan.ToCreate))
	for i, h := range plan.ToCreate {
		names[i] = h.Name
	}
	expectedNames := []string{"app-1", "db-1"}
	if len(names) != len(expectedNames) {
		t.Errorf("plan.ToCreate names = %v, want %v", names, expectedNames)
	} else {
		for i, name := range expectedNames {
			if names[i] != name {
				t.Errorf("plan.ToCreate[%d].Name = %q, want %q", i, names[i], name)
			}
		}
	}
}

func TestDigitalOceanProvider_Plan_EnvironmentNotDefined(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 1, Name: "my-ssh-key"},
		},
		droplets: nil,
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "prod", // Not defined in config
	})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	if len(plan.ToCreate) != 0 {
		t.Errorf("plan.ToCreate length = %d, want 0", len(plan.ToCreate))
	}
	if len(plan.ToDelete) != 0 {
		t.Errorf("plan.ToDelete length = %d, want 0", len(plan.ToDelete))
	}
}

func TestDigitalOceanProvider_Plan_Reconciliation(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
				"db-1": map[string]any{
					"role": "db",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 1, Name: "my-ssh-key"},
		},
		droplets: map[string]Droplet{
			"staging-app-1":    {ID: 1, Name: "staging-app-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
			"staging-orphan-1": {ID: 2, Name: "staging-orphan-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "staging",
	})
	if err != nil {
		t.Fatalf("Plan() failed: %v", err)
	}

	// Should create db-1 (app-1 already exists)
	if len(plan.ToCreate) != 1 {
		t.Errorf("plan.ToCreate length = %d, want 1", len(plan.ToCreate))
	} else if plan.ToCreate[0].Name != "db-1" {
		t.Errorf("plan.ToCreate[0].Name = %q, want %q", plan.ToCreate[0].Name, "db-1")
	}

	// Should delete orphan-1 (not in desired config)
	if len(plan.ToDelete) != 1 {
		t.Errorf("plan.ToDelete length = %d, want 1", len(plan.ToDelete))
	} else if plan.ToDelete[0].Name != "orphan-1" {
		t.Errorf("plan.ToDelete[0].Name = %q, want %q", plan.ToDelete[0].Name, "orphan-1")
	}

	// Verify both lists are sorted
	if !sort.SliceIsSorted(plan.ToCreate, func(i, j int) bool {
		return plan.ToCreate[i].Name < plan.ToCreate[j].Name
	}) {
		t.Error("plan.ToCreate is not sorted lexicographically")
	}
	if !sort.SliceIsSorted(plan.ToDelete, func(i, j int) bool {
		return plan.ToDelete[i].Name < plan.ToDelete[j].Name
	}) {
		t.Error("plan.ToDelete is not sorted lexicographically")
	}
}

func TestDigitalOceanProvider_Plan_MissingTokenEnv(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv() - but this test doesn't use Setenv, so could be parallel
	// Keeping consistent with other Plan tests

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 1, Name: "my-ssh-key"},
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	// Don't set DO_TOKEN env var

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "staging",
	})

	if err == nil {
		t.Fatal("Plan() returned nil error, want error")
	}
	if plan.ToCreate != nil || plan.ToDelete != nil {
		t.Error("Plan() returned non-empty plan on error")
	}

	if !errors.Is(err, ErrTokenMissing) {
		t.Errorf("error is not ErrTokenMissing: %v", err)
	}
	if !strings.Contains(err.Error(), "API token missing from environment variable DO_TOKEN") {
		t.Errorf("error message %q does not contain expected text", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider") {
		t.Errorf("error message %q does not contain provider prefix", err.Error())
	}
}

func TestDigitalOceanProvider_Plan_SSHKeyNotFound(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: nil, // No SSH keys
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "staging",
	})

	if err == nil {
		t.Fatal("Plan() returned nil error, want error")
	}
	if plan.ToCreate != nil || plan.ToDelete != nil {
		t.Error("Plan() returned non-empty plan on error")
	}

	if !errors.Is(err, ErrSSHKeyNotFound) {
		t.Errorf("error is not ErrSSHKeyNotFound: %v", err)
	}
	if !strings.Contains(err.Error(), "my-ssh-key") {
		t.Errorf("error message %q does not contain SSH key name", err.Error())
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message %q does not contain 'not found'", err.Error())
	}
	if !strings.Contains(err.Error(), "digitalocean provider") {
		t.Errorf("error message %q does not contain provider prefix", err.Error())
	}
}

func TestDigitalOceanProvider_Plan_APIErrorOnListDroplets(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 1, Name: "my-ssh-key"},
		},
		listErr: errors.New("network error"),
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan, err := provider.Plan(ctx, cloud.PlanOptions{
		Config:      cfg,
		Environment: "staging",
	})

	if err == nil {
		t.Fatal("Plan() returned nil error, want error")
	}
	if plan.ToCreate != nil || plan.ToDelete != nil {
		t.Error("Plan() returned non-empty plan on error")
	}

	if !errors.Is(err, ErrAPIError) {
		t.Errorf("error is not ErrAPIError: %v", err)
	}
	if !strings.Contains(err.Error(), "digitalocean provider: API error") {
		t.Errorf("error message %q does not contain expected prefix", err.Error())
	}
}

func TestDigitalOceanProvider_Apply_HappyPath_CreateDroplets(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":      "DO_TOKEN",
		"ssh_key_name":   "my-ssh-key",
		"default_region": "nyc1",
		"default_size":   "s-2vcpu-4gb",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
				"db-1": map[string]any{
					"role": "db",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: make(map[string]Droplet),
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-2vcpu-4gb", Region: "nyc1"},
			{Name: "db-1", Role: "db", Size: "s-2vcpu-4gb", Region: "nyc1"},
		},
		ToDelete: nil,
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify CreateDroplet was called twice
	if len(mockClient.created) != 2 {
		t.Errorf("CreateDroplet called %d times, want 2", len(mockClient.created))
	}

	// Verify droplet names
	createdNames := make(map[string]bool)
	for _, req := range mockClient.created {
		createdNames[req.Name] = true
		if req.Name != "staging-app-1" && req.Name != "staging-db-1" {
			t.Errorf("unexpected droplet name: %q", req.Name)
		}
		if req.Region != "nyc1" {
			t.Errorf("droplet %q has region %q, want %q", req.Name, req.Region, "nyc1")
		}
		if req.Size != "s-2vcpu-4gb" {
			t.Errorf("droplet %q has size %q, want %q", req.Name, req.Size, "s-2vcpu-4gb")
		}
		if len(req.SSHKeys) != 1 || req.SSHKeys[0] != 123 {
			t.Errorf("droplet %q has SSH keys %v, want [123]", req.Name, req.SSHKeys)
		}
		if req.Image != "ubuntu-22-04-x64" {
			t.Errorf("droplet %q has image %q, want %q", req.Name, req.Image, "ubuntu-22-04-x64")
		}
	}

	if !createdNames["staging-app-1"] {
		t.Error("staging-app-1 was not created")
	}
	if !createdNames["staging-db-1"] {
		t.Error("staging-db-1 was not created")
	}

	// Verify WaitForDroplet was called twice with "active" status
	if len(mockClient.waited) != 2 {
		t.Errorf("WaitForDroplet called %d times, want 2", len(mockClient.waited))
	}
	for _, w := range mockClient.waited {
		if w.status != "active" {
			t.Errorf("WaitForDroplet called with status %q, want %q", w.status, "active")
		}
	}
}

func TestDigitalOceanProvider_Apply_HappyPath_DeleteDroplets(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: map[string]Droplet{
			"staging-app-1": {ID: 1, Name: "staging-app-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
			"staging-db-1":  {ID: 2, Name: "staging-db-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: nil,
		ToDelete: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-2vcpu-4gb", Region: "nyc1"},
			{Name: "db-1", Role: "db", Size: "s-2vcpu-4gb", Region: "nyc1"},
		},
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify DeleteDroplet was called with correct IDs
	if len(mockClient.deleted) != 2 {
		t.Errorf("DeleteDroplet called %d times, want 2", len(mockClient.deleted))
	}

	deletedIDs := make(map[int]bool)
	for _, id := range mockClient.deleted {
		deletedIDs[id] = true
	}
	if !deletedIDs[1] {
		t.Error("droplet ID 1 was not deleted")
	}
	if !deletedIDs[2] {
		t.Error("droplet ID 2 was not deleted")
	}

	// Verify WaitForDroplet was called twice with "deleted" status
	if len(mockClient.waited) != 2 {
		t.Errorf("WaitForDroplet called %d times, want 2", len(mockClient.waited))
	}
	for _, w := range mockClient.waited {
		if w.status != "deleted" {
			t.Errorf("WaitForDroplet called with status %q, want %q", w.status, "deleted")
		}
	}

	// Verify droplets were removed from mock
	if len(mockClient.droplets) != 0 {
		t.Errorf("mock still has %d droplets, want 0", len(mockClient.droplets))
	}
}

func TestDigitalOceanProvider_Apply_IdempotentCreate_DropletAlreadyMatches(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: map[string]Droplet{
			"staging-app-1": {ID: 1, Name: "staging-app-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-2vcpu-4gb", Region: "nyc1"},
		},
		ToDelete: nil,
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify CreateDroplet was NOT called
	if len(mockClient.created) != 0 {
		t.Errorf("CreateDroplet called %d times, want 0 (idempotent)", len(mockClient.created))
	}

	// Verify WaitForDroplet was NOT called
	if len(mockClient.waited) != 0 {
		t.Errorf("WaitForDroplet called %d times, want 0 (idempotent)", len(mockClient.waited))
	}
}

func TestDigitalOceanProvider_Apply_ReconciliationError_DropletExistsButMismatched(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: map[string]Droplet{
			"staging-app-1": {ID: 1, Name: "staging-app-1", Region: "nyc1", Size: "s-2vcpu-4gb"},
		},
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-4vcpu-8gb", Region: "nyc1"},
		},
		ToDelete: nil,
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})

	if err == nil {
		t.Fatal("Apply() returned nil error, want error")
	}

	if !errors.Is(err, ErrDropletExists) {
		t.Errorf("error is not ErrDropletExists: %v", err)
	}

	if !strings.Contains(err.Error(), "staging-app-1") {
		t.Errorf("error message %q does not contain droplet name", err.Error())
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error message %q does not contain 'already exists'", err.Error())
	}

	// Verify CreateDroplet was NOT called
	if len(mockClient.created) != 0 {
		t.Error("CreateDroplet should not be called when droplet exists with mismatched spec")
	}
}

func TestDigitalOceanProvider_Apply_IdempotentDelete_DropletAlreadyGone(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: make(map[string]Droplet), // Empty - droplet already deleted
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: nil,
		ToDelete: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-2vcpu-4gb", Region: "nyc1"},
		},
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})
	if err != nil {
		t.Fatalf("Apply() failed: %v", err)
	}

	// Verify DeleteDroplet was NOT called (idempotent)
	if len(mockClient.deleted) != 0 {
		t.Errorf("DeleteDroplet called %d times, want 0 (idempotent)", len(mockClient.deleted))
	}

	// Verify WaitForDroplet was NOT called
	if len(mockClient.waited) != 0 {
		t.Errorf("WaitForDroplet called %d times, want 0 (idempotent)", len(mockClient.waited))
	}
}

func TestDigitalOceanProvider_Apply_WaitForDropletError(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv()

	ctx := context.Background()
	cfg := map[string]any{
		"token_env":    "DO_TOKEN",
		"ssh_key_name": "my-ssh-key",
		"hosts": map[string]any{
			"staging": map[string]any{
				"app-1": map[string]any{
					"role": "app",
				},
			},
		},
	}

	mockClient := &mockAPIClient{
		sshKeys: map[string]SSHKey{
			"my-ssh-key": {ID: 123, Name: "my-ssh-key"},
		},
		droplets: make(map[string]Droplet),
		waitErr:  ErrDropletTimeout,
	}

	provider := NewDigitalOceanProviderWithClient(mockClient)
	t.Setenv("DO_TOKEN", "dummy-token")

	plan := cloud.InfraPlan{
		ToCreate: []cloud.HostSpec{
			{Name: "app-1", Role: "app", Size: "s-2vcpu-4gb", Region: "nyc1"},
		},
		ToDelete: nil,
	}

	err := provider.Apply(ctx, cloud.ApplyOptions{
		Config:      cfg,
		Environment: "staging",
		Plan:        plan,
	})

	if err == nil {
		t.Fatal("Apply() returned nil error, want error")
	}

	if !errors.Is(err, ErrDropletTimeout) {
		t.Errorf("error is not ErrDropletTimeout: %v", err)
	}

	// Verify CreateDroplet was called
	if len(mockClient.created) != 1 {
		t.Errorf("CreateDroplet called %d times, want 1", len(mockClient.created))
	}
}
