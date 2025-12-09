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
	"strconv"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/config"
)

// Feature: DEV_COMPOSE_INFRA
// Spec: spec/dev/compose-infra.md

// TestGenerateCompose_BackendOnly_Minimal ensures that the first thin
// slice of DEV_COMPOSE_INFRA returns a non-nil compose file and no
// error when invoked with a minimal backend service definition.
//
// This is intentionally a very small assertion; future tests will
// drive the detailed structure of the generated compose model.
func TestGenerateCompose_BackendOnly_Minimal(t *testing.T) {
	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}

	gen := NewGenerator()

	got, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if got == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}
}

func TestGenerator_GenerateCompose_BackendAndFrontend(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
	}

	gen := NewGenerator()

	got, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if got == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	// Verify both services are present
	services := got.GetServices()
	if len(services) != 2 {
		t.Fatalf("GenerateCompose() got %d services, want 2", len(services))
	}

	// Verify service names are present (order-agnostic; map iteration is non-deterministic)
	serviceMap := make(map[string]bool)
	for _, svc := range services {
		serviceMap[svc] = true
	}

	if !serviceMap["backend"] {
		t.Errorf("GenerateCompose() missing service: backend")
	}

	if !serviceMap["frontend"] {
		t.Errorf("GenerateCompose() missing service: frontend")
	}
}

func TestGenerator_GenerateCompose_DeterministicOrdering(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
	}

	gen := NewGenerator()

	first, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("first GenerateCompose() error = %v, want nil", err)
	}

	second, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("second GenerateCompose() error = %v, want nil", err)
	}

	if first == nil {
		t.Fatalf("first GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	if second == nil {
		t.Fatalf("second GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	firstServices := first.GetServices()
	secondServices := second.GetServices()

	if len(firstServices) != len(secondServices) {
		t.Fatalf(
			"GenerateCompose() service count mismatch: first=%d second=%d",
			len(firstServices),
			len(secondServices),
		)
	}

	for i := range firstServices {
		if firstServices[i] != secondServices[i] {
			t.Fatalf(
				"GenerateCompose() service ordering mismatch at index %d: first=%q second=%q",
				i,
				firstServices[i],
				secondServices[i],
			)
		}
	}
}

func TestGenerator_GenerateCompose_WithTraefikService(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
	}
	traefik := &ServiceDefinition{
		Name: "traefik",
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, traefik)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	services := composeFile.GetServices()
	if len(services) != 3 {
		t.Fatalf("GenerateCompose() services length = %d, want 3 (backend, frontend, traefik)", len(services))
	}

	var (
		foundBackend  bool
		foundFrontend bool
		foundTraefik  bool
	)

	for _, name := range services {
		switch name {
		case "backend":
			foundBackend = true
		case "frontend":
			foundFrontend = true
		case "traefik":
			foundTraefik = true
		}
	}

	if !foundBackend {
		t.Fatalf("GenerateCompose() services = %v, want to include backend", services)
	}

	if !foundFrontend {
		t.Fatalf("GenerateCompose() services = %v, want to include frontend", services)
	}

	if !foundTraefik {
		t.Fatalf("GenerateCompose() services = %v, want to include traefik", services)
	}
}

func TestGenerator_GenerateCompose_BackendWithPortsAndEnv(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		Ports: []PortMapping{
			{Host: "8080", Container: "3000", Protocol: "tcp"},
			{Host: "9090", Container: "4000", Protocol: "tcp"},
		},
		Environment: map[string]string{
			"B": "2",
			"A": "1",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	// Get backend service data
	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	// Verify ports
	ports, ok := backendService["ports"].([]any)
	if !ok {
		t.Fatalf("backend service ports = %T, want []any", backendService["ports"])
	}

	if len(ports) != 2 {
		t.Fatalf("backend service ports length = %d, want 2", len(ports))
	}

	// Verify port format and ordering (should be sorted by host port)
	// Ports are stored as yaml.Node to force quoting in YAML output
	portStrs := make([]string, len(ports))
	for i, p := range ports {
		var portStr string
		switch v := p.(type) {
		case string:
			portStr = v
		case *yaml.Node:
			portStr = v.Value
		default:
			t.Fatalf("port[%d] = %T, want string or *yaml.Node", i, p)
		}
		portStrs[i] = portStr
	}

	// Check first port (should be 8080:3000/tcp)
	if portStrs[0] != "8080:3000/tcp" {
		t.Errorf("port[0] = %q, want \"8080:3000/tcp\"", portStrs[0])
	}

	// Check second port (should be 9090:4000/tcp)
	if portStrs[1] != "9090:4000/tcp" {
		t.Errorf("port[1] = %q, want \"9090:4000/tcp\"", portStrs[1])
	}

	// Verify environment
	env, ok := backendService["environment"].(map[string]any)
	if !ok {
		t.Fatalf("backend service environment = %T, want map[string]any", backendService["environment"])
	}

	if len(env) != 2 {
		t.Fatalf("backend service environment length = %d, want 2", len(env))
	}

	// Verify environment keys are present (order doesn't matter for map, but values should be correct)
	if env["A"] != "1" {
		t.Errorf("environment[\"A\"] = %v, want \"1\"", env["A"])
	}

	if env["B"] != "2" {
		t.Errorf("environment[\"B\"] = %v, want \"2\"", env["B"])
	}
}

func TestGenerator_GenerateCompose_FrontendWithPortsAndEnv(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		Ports: []PortMapping{
			{Host: "8080", Container: "3000", Protocol: "tcp"},
			{Host: "9090", Container: "4000", Protocol: "tcp"},
		},
		Environment: map[string]string{
			"B": "2",
			"A": "1",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	// Verify ports
	ports, ok := frontendService["ports"].([]any)
	if !ok {
		t.Fatalf("frontend service ports = %T, want []any", frontendService["ports"])
	}

	if len(ports) != 2 {
		t.Fatalf("frontend service ports length = %d, want 2", len(ports))
	}

	portStrs := make([]string, len(ports))
	for i, p := range ports {
		var portStr string
		switch v := p.(type) {
		case string:
			portStr = v
		case *yaml.Node:
			portStr = v.Value
		default:
			t.Fatalf("port[%d] = %T, want string or *yaml.Node", i, p)
		}
		portStrs[i] = portStr
	}

	if portStrs[0] != "8080:3000/tcp" {
		t.Errorf("port[0] = %q, want \"8080:3000/tcp\"", portStrs[0])
	}

	if portStrs[1] != "9090:4000/tcp" {
		t.Errorf("port[1] = %q, want \"9090:4000/tcp\"", portStrs[1])
	}

	// Verify environment
	env, ok := frontendService["environment"].(map[string]any)
	if !ok {
		t.Fatalf("frontend service environment = %T, want map[string]any", frontendService["environment"])
	}

	if len(env) != 2 {
		t.Fatalf("frontend service environment length = %d, want 2", len(env))
	}

	if env["A"] != "1" {
		t.Errorf("environment[\"A\"] = %v, want \"1\"", env["A"])
	}

	if env["B"] != "2" {
		t.Errorf("environment[\"B\"] = %v, want \"2\"", env["B"])
	}
}

func TestGenerator_GenerateCompose_BackendWithVolumes(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		Volumes: []VolumeMapping{
			// Intentionally out of order to exercise deterministic sorting.
			{
				Type:     "bind",
				Source:   "./data",
				Target:   "/app/data",
				ReadOnly: false,
			},
			{
				Type:     "bind",
				Source:   "./config",
				Target:   "/app/config",
				ReadOnly: true,
			},
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	rawVolumes, ok := backendService["volumes"]
	if !ok {
		t.Fatalf("backend service missing volumes key")
	}

	volumes, ok := rawVolumes.([]any)
	if !ok {
		t.Fatalf("backend service volumes = %T, want []any", rawVolumes)
	}

	if len(volumes) != 2 {
		t.Fatalf("backend service volumes length = %d, want 2", len(volumes))
	}

	volStrs := make([]string, len(volumes))
	for i, v := range volumes {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("volume[%d] = %T, want string", i, v)
		}
		volStrs[i] = s
	}

	// Expect deterministic ordering by target path:
	// /app/config first, then /app/data.
	if volStrs[0] != "./config:/app/config:ro" {
		t.Errorf("volume[0] = %q, want \"./config:/app/config:ro\"", volStrs[0])
	}

	if volStrs[1] != "./data:/app/data" {
		t.Errorf("volume[1] = %q, want \"./data:/app/data\"", volStrs[1])
	}
}

func TestGenerator_GenerateCompose_FrontendWithVolumes(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		Volumes: []VolumeMapping{
			// Intentionally out of order to exercise deterministic sorting.
			{
				Type:     "bind",
				Source:   "./data",
				Target:   "/app/data",
				ReadOnly: false,
			},
			{
				Type:     "bind",
				Source:   "./config",
				Target:   "/app/config",
				ReadOnly: true,
			},
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	rawVolumes, ok := frontendService["volumes"]
	if !ok {
		t.Fatalf("frontend service missing volumes key")
	}

	volumes, ok := rawVolumes.([]any)
	if !ok {
		t.Fatalf("frontend service volumes = %T, want []any", rawVolumes)
	}

	if len(volumes) != 2 {
		t.Fatalf("frontend service volumes length = %d, want 2", len(volumes))
	}

	volStrs := make([]string, len(volumes))
	for i, v := range volumes {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("volume[%d] = %T, want string", i, v)
		}
		volStrs[i] = s
	}

	// Expect deterministic ordering by target path:
	// /app/config first, then /app/data.
	if volStrs[0] != "./config:/app/config:ro" {
		t.Errorf("volume[0] = %q, want \"./config:/app/config:ro\"", volStrs[0])
	}

	if volStrs[1] != "./data:/app/data" {
		t.Errorf("volume[1] = %q, want \"./data:/app/data\"", volStrs[1])
	}
}

func TestGenerator_GenerateCompose_BackendWithLabels(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		Labels: map[string]string{
			"traefik.http.routers.backend.rule": "Host(`api.localdev.test`)",
			"app":                               "backend",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	rawLabels, ok := backendService["labels"]
	if !ok {
		t.Fatalf("backend service missing labels key")
	}

	labels, ok := rawLabels.(map[string]any)
	if !ok {
		t.Fatalf("backend service labels = %T, want map[string]any", rawLabels)
	}

	if len(labels) != 2 {
		t.Fatalf("backend service labels length = %d, want 2", len(labels))
	}

	if labels["app"] != "backend" {
		t.Errorf("labels[\"app\"] = %v, want \"backend\"", labels["app"])
	}

	if labels["traefik.http.routers.backend.rule"] != "Host(`api.localdev.test`)" {
		t.Errorf(
			"labels[\"traefik.http.routers.backend.rule\"] = %v, want \"Host(`api.localdev.test`)\"",
			labels["traefik.http.routers.backend.rule"],
		)
	}
}

func TestGenerator_GenerateCompose_FrontendWithLabels(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		Labels: map[string]string{
			"traefik.http.routers.frontend.rule": "Host(`app.localdev.test`)",
			"app":                                "frontend",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	rawLabels, ok := frontendService["labels"]
	if !ok {
		t.Fatalf("frontend service missing labels key")
	}

	labels, ok := rawLabels.(map[string]any)
	if !ok {
		t.Fatalf("frontend service labels = %T, want map[string]any", rawLabels)
	}

	if len(labels) != 2 {
		t.Fatalf("frontend service labels length = %d, want 2", len(labels))
	}

	if labels["app"] != "frontend" {
		t.Errorf("labels[\"app\"] = %v, want \"frontend\"", labels["app"])
	}

	if labels["traefik.http.routers.frontend.rule"] != "Host(`app.localdev.test`)" {
		t.Errorf(
			"labels[\"traefik.http.routers.frontend.rule\"] = %v, want \"Host(`app.localdev.test`)\"",
			labels["traefik.http.routers.frontend.rule"],
		)
	}
}

func TestGenerator_GenerateCompose_BackendWithDependsOn(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		DependsOn: []string{
			"db",
			"cache",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	rawDependsOn, ok := backendService["depends_on"]
	if !ok {
		t.Fatalf("backend service missing depends_on key")
	}

	depends, ok := rawDependsOn.([]any)
	if !ok {
		t.Fatalf("backend service depends_on = %T, want []any", rawDependsOn)
	}

	if len(depends) != 2 {
		t.Fatalf("backend service depends_on length = %d, want 2", len(depends))
	}

	depStrs := make([]string, len(depends))
	for i, v := range depends {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("depends_on[%d] = %T, want string", i, v)
		}
		depStrs[i] = s
	}

	// Deterministic ordering: lexicographic, so "cache" then "db".
	if depStrs[0] != "cache" {
		t.Errorf("depends_on[0] = %q, want \"cache\"", depStrs[0])
	}

	if depStrs[1] != "db" {
		t.Errorf("depends_on[1] = %q, want \"db\"", depStrs[1])
	}
}

func TestGenerator_GenerateCompose_FrontendWithDependsOn(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		DependsOn: []string{
			"backend",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	rawDependsOn, ok := frontendService["depends_on"]
	if !ok {
		t.Fatalf("frontend service missing depends_on key")
	}

	depends, ok := rawDependsOn.([]any)
	if !ok {
		t.Fatalf("frontend service depends_on = %T, want []any", rawDependsOn)
	}

	if len(depends) != 1 {
		t.Fatalf("frontend service depends_on length = %d, want 1", len(depends))
	}

	s, ok := depends[0].(string)
	if !ok {
		t.Fatalf("depends_on[0] = %T, want string", depends[0])
	}

	if s != "backend" {
		t.Errorf("depends_on[0] = %q, want \"backend\"", s)
	}
}

func TestGenerator_GenerateCompose_BackendWithNetworks(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
		Networks: []string{
			"b-net",
			"a-net",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	rawNetworks, ok := backendService["networks"]
	if !ok {
		t.Fatalf("backend service missing networks key")
	}

	networks, ok := rawNetworks.([]any)
	if !ok {
		t.Fatalf("backend service networks = %T, want []any", rawNetworks)
	}

	// stagecraft-dev is automatically added, so we expect 3 networks total
	if len(networks) != 3 {
		t.Fatalf("backend service networks length = %d, want 3 (a-net, b-net, stagecraft-dev)", len(networks))
	}

	netStrs := make([]string, len(networks))
	for i, v := range networks {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("network[%d] = %T, want string", i, v)
		}
		netStrs[i] = s
	}

	// Deterministic ordering: lexicographic ("a-net", "b-net", "stagecraft-dev").
	if netStrs[0] != "a-net" {
		t.Errorf("network[0] = %q, want \"a-net\"", netStrs[0])
	}

	if netStrs[1] != "b-net" {
		t.Errorf("network[1] = %q, want \"b-net\"", netStrs[1])
	}

	if netStrs[2] != "stagecraft-dev" {
		t.Errorf("network[2] = %q, want \"stagecraft-dev\"", netStrs[2])
	}
}

func TestGenerator_GenerateCompose_FrontendWithNetworks(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
		Networks: []string{
			"frontend-net",
			"app-net",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	rawNetworks, ok := frontendService["networks"]
	if !ok {
		t.Fatalf("frontend service missing networks key")
	}

	networks, ok := rawNetworks.([]any)
	if !ok {
		t.Fatalf("frontend service networks = %T, want []any", rawNetworks)
	}

	// stagecraft-dev is automatically added, so we expect 3 networks total
	if len(networks) != 3 {
		t.Fatalf("frontend service networks length = %d, want 3 (app-net, frontend-net, stagecraft-dev)", len(networks))
	}

	netStrs := make([]string, len(networks))
	for i, v := range networks {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("network[%d] = %T, want string", i, v)
		}
		netStrs[i] = s
	}

	// Deterministic ordering: lexicographic ("app-net", "frontend-net", "stagecraft-dev").
	if netStrs[0] != "app-net" {
		t.Errorf("network[0] = %q, want \"app-net\"", netStrs[0])
	}

	if netStrs[1] != "frontend-net" {
		t.Errorf("network[1] = %q, want \"frontend-net\"", netStrs[1])
	}

	if netStrs[2] != "stagecraft-dev" {
		t.Errorf("network[2] = %q, want \"stagecraft-dev\"", netStrs[2])
	}
}

func TestGenerator_GenerateCompose_BackendWithImageAndBuild(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name:  "backend",
		Image: "ghcr.io/example/backend:dev",
		Build: map[string]any{
			"context":    "./backend",
			"dockerfile": "Dockerfile.dev",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	backendService := composeFile.GetServiceData("backend")
	if backendService == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	// Verify image
	img, ok := backendService["image"].(string)
	if !ok {
		t.Fatalf("backend service image = %T, want string", backendService["image"])
	}
	if img != "ghcr.io/example/backend:dev" {
		t.Errorf("backend service image = %q, want %q", img, "ghcr.io/example/backend:dev")
	}

	// Verify build
	rawBuild, ok := backendService["build"]
	if !ok {
		t.Fatalf("backend service missing build key")
	}

	build, ok := rawBuild.(map[string]any)
	if !ok {
		t.Fatalf("backend service build = %T, want map[string]any", rawBuild)
	}

	if len(build) != 2 {
		t.Fatalf("backend service build length = %d, want 2", len(build))
	}

	if build["context"] != "./backend" {
		t.Errorf("build[\"context\"] = %v, want \"./backend\"", build["context"])
	}

	if build["dockerfile"] != "Dockerfile.dev" {
		t.Errorf("build[\"dockerfile\"] = %v, want \"Dockerfile.dev\"", build["dockerfile"])
	}
}

func TestGenerator_GenerateCompose_FrontendWithImageAndBuild(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name:  "frontend",
		Image: "ghcr.io/example/frontend:dev",
		Build: map[string]any{
			"context":    "./frontend",
			"dockerfile": "Dockerfile.dev",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	frontendService := composeFile.GetServiceData("frontend")
	if frontendService == nil {
		t.Fatalf("GetServiceData(\"frontend\") = nil, want non-nil")
	}

	// Verify image
	img, ok := frontendService["image"].(string)
	if !ok {
		t.Fatalf("frontend service image = %T, want string", frontendService["image"])
	}
	if img != "ghcr.io/example/frontend:dev" {
		t.Errorf("frontend service image = %q, want %q", img, "ghcr.io/example/frontend:dev")
	}

	// Verify build
	rawBuild, ok := frontendService["build"]
	if !ok {
		t.Fatalf("frontend service missing build key")
	}

	build, ok := rawBuild.(map[string]any)
	if !ok {
		t.Fatalf("frontend service build = %T, want map[string]any", rawBuild)
	}

	if len(build) != 2 {
		t.Fatalf("frontend service build length = %d, want 2", len(build))
	}

	if build["context"] != "./frontend" {
		t.Errorf("build[\"context\"] = %v, want \"./frontend\"", build["context"])
	}

	if build["dockerfile"] != "Dockerfile.dev" {
		t.Errorf("build[\"dockerfile\"] = %v, want \"Dockerfile.dev\"", build["dockerfile"])
	}
}

func TestGenerator_GenerateCompose_TraefikWithImageBuildAndFields(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
	}
	traefik := &ServiceDefinition{
		Name:  "traefik",
		Image: "traefik:v3.0",
		Build: map[string]any{
			"context":    "./infra/traefik",
			"dockerfile": "Dockerfile.dev",
		},
		Ports: []PortMapping{
			{Host: "80", Container: "80", Protocol: "tcp"},
			{Host: "443", Container: "443", Protocol: "tcp"},
		},
		Environment: map[string]string{
			"TRAEFIK_LOG_LEVEL": "DEBUG",
			"TRAEFIK_API":       "true",
		},
		Volumes: []VolumeMapping{
			{
				Type:     "bind",
				Source:   "./traefik/config",
				Target:   "/etc/traefik",
				ReadOnly: true,
			},
		},
		Labels: map[string]string{
			"traefik.enable": "true",
		},
		DependsOn: []string{
			"backend",
			"frontend",
		},
		Networks: []string{
			"stagecraft-dev",
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, traefik)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	if composeFile == nil {
		t.Fatalf("GenerateCompose() got = nil, want non-nil *ComposeFile")
	}

	traefikService := composeFile.GetServiceData("traefik")
	if traefikService == nil {
		t.Fatalf("GetServiceData(\"traefik\") = nil, want non-nil")
	}

	// For v1, DEV_COMPOSE_INFRA owns the Traefik service definition.
	// The traefikService parameter is used to signal inclusion, but the
	// actual service structure is hardcoded. Verify the hardcoded values:

	// Image: should be traefik:v2.11 (hardcoded)
	img, ok := traefikService["image"].(string)
	if !ok {
		t.Fatalf("traefik service image = %T, want string", traefikService["image"])
	}
	if img != "traefik:v2.11" {
		t.Errorf("traefik service image = %q, want %q", img, "traefik:v2.11")
	}

	// Build: should not be present (we use image, not build)
	_, hasBuild := traefikService["build"]
	if hasBuild {
		t.Errorf("traefik service should not have build key (we use image)")
	}

	// Ports: should be 80:80 and 443:443 (hardcoded)
	rawPorts, ok := traefikService["ports"]
	if !ok {
		t.Fatalf("traefik service missing ports key")
	}

	ports, ok := rawPorts.([]any)
	if !ok {
		t.Fatalf("traefik service ports = %T, want []any", rawPorts)
	}
	if len(ports) != 2 {
		t.Fatalf("traefik service ports length = %d, want 2", len(ports))
	}

	portStrs := make([]string, len(ports))
	for i, p := range ports {
		switch v := p.(type) {
		case string:
			portStrs[i] = v
		case int:
			portStrs[i] = strconv.Itoa(v)
		case *yaml.Node:
			portStrs[i] = v.Value
		default:
			t.Fatalf("port[%d] = %T, want string, int, or *yaml.Node", i, p)
		}
	}

	if portStrs[0] != "80:80" {
		t.Errorf("port[0] = %q, want \"80:80\"", portStrs[0])
	}
	if portStrs[1] != "443:443" {
		t.Errorf("port[1] = %q, want \"443:443\"", portStrs[1])
	}

	// Volumes: should have certs and traefik config mounts (hardcoded)
	rawVolumes, ok := traefikService["volumes"]
	if !ok {
		t.Fatalf("traefik service missing volumes key")
	}

	volumes, ok := rawVolumes.([]any)
	if !ok {
		t.Fatalf("traefik service volumes = %T, want []any", rawVolumes)
	}
	if len(volumes) != 2 {
		t.Fatalf("traefik service volumes length = %d, want 2", len(volumes))
	}

	volStrs := make([]string, len(volumes))
	for i, v := range volumes {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("volume[%d] = %T, want string", i, v)
		}
		volStrs[i] = s
	}

	expectedVols := []string{
		"./.stagecraft/dev/certs:/certs:ro",
		"./.stagecraft/dev/traefik:/etc/traefik:ro",
	}
	for _, expected := range expectedVols {
		found := false
		for _, vol := range volStrs {
			if vol == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("traefik service missing volume: %q", expected)
		}
	}

	// Command: should have file provider flags (hardcoded)
	rawCommand, ok := traefikService["command"]
	if !ok {
		t.Fatalf("traefik service missing command key")
	}

	command, ok := rawCommand.([]any)
	if !ok {
		t.Fatalf("traefik service command = %T, want []any", rawCommand)
	}

	expectedCommands := []string{
		"--configfile=/etc/traefik/traefik-static.yaml",
		"--providers.file.directory=/etc/traefik",
		"--providers.file.watch=true",
	}
	if len(command) != len(expectedCommands) {
		t.Fatalf("traefik service command length = %d, want %d", len(command), len(expectedCommands))
	}

	cmdStrs := make([]string, len(command))
	for i, c := range command {
		s, ok := c.(string)
		if !ok {
			t.Fatalf("command[%d] = %T, want string", i, c)
		}
		cmdStrs[i] = s
	}

	for i, expected := range expectedCommands {
		if cmdStrs[i] != expected {
			t.Errorf("command[%d] = %q, want %q", i, cmdStrs[i], expected)
		}
	}

	// Networks: should include stagecraft-dev
	rawNetworks, ok := traefikService["networks"]
	if !ok {
		t.Fatalf("traefik service missing networks key")
	}

	networks, ok := rawNetworks.([]any)
	if !ok {
		t.Fatalf("traefik service networks = %T, want []any", rawNetworks)
	}

	if len(networks) != 1 {
		t.Fatalf("traefik service networks length = %d, want 1", len(networks))
	}

	if networks[0] != "stagecraft-dev" {
		t.Errorf("networks[0] = %v, want \"stagecraft-dev\"", networks[0])
	}
}

func TestGenerator_GenerateCompose_InvalidTopology(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	gen := NewGenerator()

	got, err := gen.GenerateCompose(cfg, nil, nil, nil)
	if err == nil {
		t.Fatalf("GenerateCompose() error = nil, want non-nil")
	}

	if !errors.Is(err, ErrBackendServiceRequired) {
		t.Fatalf("GenerateCompose() error = %v, want ErrBackendServiceRequired", err)
	}

	if got != nil {
		t.Fatalf("GenerateCompose() got = %v, want nil", got)
	}
}

// TestGenerator_GenerateCompose_NetworkCreation verifies that stagecraft-dev
// network is always created and all services join it.
func TestGenerator_GenerateCompose_NetworkCreation(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	frontend := &ServiceDefinition{
		Name: "frontend",
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, frontend, nil)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	// Verify network exists in compose file
	networksData := composeFile.GetServiceData("backend")
	if networksData == nil {
		t.Fatalf("GetServiceData(\"backend\") = nil, want non-nil")
	}

	// Check that backend has stagecraft-dev network
	networks, ok := networksData["networks"].([]any)
	if !ok {
		t.Fatalf("backend service networks = %T, want []any", networksData["networks"])
	}

	hasDevNetwork := false
	for _, net := range networks {
		if netStr, ok := net.(string); ok && netStr == "stagecraft-dev" {
			hasDevNetwork = true
			break
		}
	}

	if !hasDevNetwork {
		t.Errorf("backend service missing stagecraft-dev network")
	}

	// Verify network section exists in compose file
	// We can't directly access networks from ComposeFile, but we can verify
	// via YAML output
	yamlBytes, err := composeFile.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML() error = %v, want nil", err)
	}

	yamlStr := string(yamlBytes)
	if !contains(yamlStr, "stagecraft-dev:") {
		t.Errorf("compose YAML missing stagecraft-dev network definition")
	}

	if !contains(yamlStr, "name: stagecraft-dev") {
		t.Errorf("compose YAML missing stagecraft-dev network name")
	}
}

// TestGenerator_GenerateCompose_TraefikServiceGeneration verifies that
// when traefikService != nil, a complete Traefik service is generated
// with hardcoded values (image, ports, volumes, command).
func TestGenerator_GenerateCompose_TraefikServiceGeneration(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	backend := &ServiceDefinition{
		Name: "backend",
	}
	traefik := &ServiceDefinition{
		Name: "traefik",
		// These should be ignored - DEV_COMPOSE_INFRA owns Traefik service structure
		Image: "custom-traefik:latest",
		Ports: []PortMapping{
			{Host: "8080", Container: "8080"},
		},
	}

	gen := NewGenerator()

	composeFile, err := gen.GenerateCompose(cfg, backend, nil, traefik)
	if err != nil {
		t.Fatalf("GenerateCompose() error = %v, want nil", err)
	}

	traefikService := composeFile.GetServiceData("traefik")
	if traefikService == nil {
		t.Fatalf("GetServiceData(\"traefik\") = nil, want non-nil")
	}

	// Verify hardcoded image (not the one from traefik parameter)
	img, ok := traefikService["image"].(string)
	if !ok {
		t.Fatalf("traefik service image = %T, want string", traefikService["image"])
	}
	if img != "traefik:v2.11" {
		t.Errorf("traefik service image = %q, want \"traefik:v2.11\"", img)
	}

	// Verify hardcoded ports (80:80, 443:443, not 8080:8080)
	ports, ok := traefikService["ports"].([]any)
	if !ok {
		t.Fatalf("traefik service ports = %T, want []any", traefikService["ports"])
	}

	if len(ports) != 2 {
		t.Fatalf("traefik service ports length = %d, want 2", len(ports))
	}

	portStrs := make([]string, len(ports))
	for i, p := range ports {
		switch v := p.(type) {
		case string:
			portStrs[i] = v
		case *yaml.Node:
			portStrs[i] = v.Value
		default:
			t.Fatalf("port[%d] = %T, want string or *yaml.Node", i, p)
		}
	}

	if portStrs[0] != "80:80" {
		t.Errorf("port[0] = %q, want \"80:80\"", portStrs[0])
	}
	if portStrs[1] != "443:443" {
		t.Errorf("port[1] = %q, want \"443:443\"", portStrs[1])
	}

	// Verify hardcoded volumes
	volumes, ok := traefikService["volumes"].([]any)
	if !ok {
		t.Fatalf("traefik service volumes = %T, want []any", traefikService["volumes"])
	}

	if len(volumes) != 2 {
		t.Fatalf("traefik service volumes length = %d, want 2", len(volumes))
	}

	volStrs := make([]string, len(volumes))
	for i, v := range volumes {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("volume[%d] = %T, want string", i, v)
		}
		volStrs[i] = s
	}

	expectedVols := map[string]bool{
		"./.stagecraft/dev/certs:/certs:ro":         true,
		"./.stagecraft/dev/traefik:/etc/traefik:ro": true,
	}

	for _, vol := range volStrs {
		if !expectedVols[vol] {
			t.Errorf("traefik service unexpected volume: %q", vol)
		}
	}

	// Verify hardcoded command
	command, ok := traefikService["command"].([]any)
	if !ok {
		t.Fatalf("traefik service command = %T, want []any", traefikService["command"])
	}

	expectedCommands := []string{
		"--configfile=/etc/traefik/traefik-static.yaml",
		"--providers.file.directory=/etc/traefik",
		"--providers.file.watch=true",
	}

	if len(command) != len(expectedCommands) {
		t.Fatalf("traefik service command length = %d, want %d", len(command), len(expectedCommands))
	}

	for i, expected := range expectedCommands {
		cmdStr, ok := command[i].(string)
		if !ok {
			t.Fatalf("command[%d] = %T, want string", i, command[i])
		}
		if cmdStr != expected {
			t.Errorf("command[%d] = %q, want %q", i, cmdStr, expected)
		}
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		strings.Contains(s, substr))
}
