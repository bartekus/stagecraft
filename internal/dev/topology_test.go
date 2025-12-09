// SPDX-License-Identifier: AGPL-3.0-or-later

package dev

import (
	"testing"

	devcompose "stagecraft/internal/dev/compose"
	devmkcert "stagecraft/internal/dev/mkcert"
	devtraefik "stagecraft/internal/dev/traefik"

	"stagecraft/pkg/config"
)

// Feature: CLI_DEV
// Spec: spec/commands/dev.md

func TestBuilder_Build_MinimalTopology(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	domains := Domains{
		Frontend: "app.localdev.test",
		Backend:  "api.localdev.test",
	}

	backend := &devcompose.ServiceDefinition{
		Name: "backend",
		Ports: []devcompose.PortMapping{
			{Host: "8080", Container: "4000", Protocol: "tcp"},
		},
	}
	frontend := &devcompose.ServiceDefinition{
		Name: "frontend",
		Ports: []devcompose.PortMapping{
			{Host: "3000", Container: "3000", Protocol: "tcp"},
		},
	}
	traefikSvc := &devcompose.ServiceDefinition{
		Name: "traefik",
	}

	builder := NewBuilder(
		devcompose.NewGenerator(),
		devtraefik.NewGenerator(),
		nil, // backend registry - use default
		nil, // frontend registry - use default
	)

	top, err := builder.Build(
		cfg,
		domains,
		backend,
		frontend,
		traefikSvc,
		nil, // certCfg - HTTP only
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	if top == nil {
		t.Fatalf("Build() = nil, want non-nil *Topology")
	}

	if top.Compose == nil {
		t.Fatalf("Topology.Compose = nil, want non-nil *ComposeFile")
	}

	if top.Traefik == nil {
		t.Fatalf("Topology.Traefik = nil, want non-nil *devtraefik.Config")
	}

	if top.Domains.Frontend != "app.localdev.test" {
		t.Errorf("Topology.Domains.Frontend = %q, want %q", top.Domains.Frontend, "app.localdev.test")
	}

	if top.Domains.Backend != "api.localdev.test" {
		t.Errorf("Topology.Domains.Backend = %q, want %q", top.Domains.Backend, "api.localdev.test")
	}
}

func TestBuilder_Build_ComposeAndTraefikIntegration(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	domains := Domains{
		Frontend: "app.localdev.test",
		Backend:  "api.localdev.test",
	}

	backend := &devcompose.ServiceDefinition{
		Name: "backend",
		Ports: []devcompose.PortMapping{
			{Host: "8080", Container: "4000", Protocol: "tcp"},
		},
	}
	frontend := &devcompose.ServiceDefinition{
		Name: "frontend",
		Ports: []devcompose.PortMapping{
			{Host: "3000", Container: "3000", Protocol: "tcp"},
		},
	}
	traefikSvc := &devcompose.ServiceDefinition{
		Name: "traefik",
	}

	builder := NewBuilder(
		devcompose.NewGenerator(),
		devtraefik.NewGenerator(),
		nil, // backend registry - use default
		nil, // frontend registry - use default
	)

	certCfg := &devmkcert.CertConfig{
		Enabled:  true,
		CertFile: ".stagecraft/dev/certs/dev-local.pem",
		KeyFile:  ".stagecraft/dev/certs/dev-local-key.pem",
		Domains:  []string{"app.localdev.test", "api.localdev.test"},
	}

	top, err := builder.Build(
		cfg,
		domains,
		backend,
		frontend,
		traefikSvc,
		certCfg,
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	if top.Traefik == nil || top.Traefik.Dynamic == nil || top.Traefik.Dynamic.HTTP == nil {
		t.Fatalf("Topology.Traefik or HTTP config is nil, want non-nil")
	}

	httpCfg := top.Traefik.Dynamic.HTTP

	frontendRouter, ok := httpCfg.Routers["frontend"]
	if !ok {
		t.Fatalf("Traefik routers missing frontend")
	}
	if frontendRouter.Rule != "Host(`app.localdev.test`)" {
		t.Errorf("frontend router rule = %q, want %q", frontendRouter.Rule, "Host(`app.localdev.test`)")
	}
	if frontendRouter.TLS == nil {
		t.Fatalf("frontend router TLS = nil, want non-nil when HTTPS enabled")
	}
	if frontendRouter.TLS.CertFile != "/certs/dev-local.pem" {
		t.Errorf("frontend router TLS.CertFile = %q, want %q", frontendRouter.TLS.CertFile, "/certs/dev-local.pem")
	}
	if frontendRouter.TLS.KeyFile != "/certs/dev-local-key.pem" {
		t.Errorf("frontend router TLS.KeyFile = %q, want %q", frontendRouter.TLS.KeyFile, "/certs/dev-local-key.pem")
	}

	backendRouter, ok := httpCfg.Routers["backend"]
	if !ok {
		t.Fatalf("Traefik routers missing backend")
	}
	if backendRouter.Rule != "Host(`api.localdev.test`)" {
		t.Errorf("backend router rule = %q, want %q", backendRouter.Rule, "Host(`api.localdev.test`)")
	}
	if backendRouter.TLS == nil {
		t.Fatalf("backend router TLS = nil, want non-nil when HTTPS enabled")
	}
	if backendRouter.TLS.CertFile != "/certs/dev-local.pem" {
		t.Errorf("backend router TLS.CertFile = %q, want %q", backendRouter.TLS.CertFile, "/certs/dev-local.pem")
	}
	if backendRouter.TLS.KeyFile != "/certs/dev-local-key.pem" {
		t.Errorf("backend router TLS.KeyFile = %q, want %q", backendRouter.TLS.KeyFile, "/certs/dev-local-key.pem")
	}

	frontendSvc, ok := httpCfg.Services["frontend"]
	if !ok {
		t.Fatalf("Traefik services missing frontend")
	}
	if frontendSvc.LoadBalancer == nil || len(frontendSvc.LoadBalancer.Servers) != 1 {
		t.Fatalf("frontend LoadBalancer servers = %#v, want 1 server", frontendSvc.LoadBalancer)
	}
	if frontendSvc.LoadBalancer.Servers[0].URL != "http://frontend:3000" {
		t.Errorf("frontend server URL = %q, want %q", frontendSvc.LoadBalancer.Servers[0].URL, "http://frontend:3000")
	}

	backendSvc, ok := httpCfg.Services["backend"]
	if !ok {
		t.Fatalf("Traefik services missing backend")
	}
	if backendSvc.LoadBalancer == nil || len(backendSvc.LoadBalancer.Servers) != 1 {
		t.Fatalf("backend LoadBalancer servers = %#v, want 1 server", backendSvc.LoadBalancer)
	}
	if backendSvc.LoadBalancer.Servers[0].URL != "http://backend:4000" {
		t.Errorf("backend server URL = %q, want %q", backendSvc.LoadBalancer.Servers[0].URL, "http://backend:4000")
	}
}

func TestBuilder_Build_NoTraefik(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	domains := Domains{
		Frontend: "app.localdev.test",
		Backend:  "api.localdev.test",
	}

	backend := &devcompose.ServiceDefinition{
		Name: "backend",
		Ports: []devcompose.PortMapping{
			{Host: "8080", Container: "4000", Protocol: "tcp"},
		},
	}
	frontend := &devcompose.ServiceDefinition{
		Name: "frontend",
		Ports: []devcompose.PortMapping{
			{Host: "3000", Container: "3000", Protocol: "tcp"},
		},
	}

	builder := NewDefaultBuilder()

	top, err := builder.Build(
		cfg,
		domains,
		backend,
		frontend,
		nil, // traefikSvc - nil means no Traefik
		nil, // certCfg
	)
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}

	if top == nil {
		t.Fatalf("Build() = nil, want non-nil *Topology")
	}

	if top.Compose == nil {
		t.Fatalf("Topology.Compose = nil, want non-nil *ComposeFile")
	}

	// Traefik should be nil when traefikSvc is nil
	if top.Traefik != nil {
		t.Errorf("Topology.Traefik = %v, want nil when traefikSvc is nil", top.Traefik)
	}

	if top.TraefikService != nil {
		t.Errorf("Topology.TraefikService = %v, want nil when traefikSvc is nil", top.TraefikService)
	}
}
