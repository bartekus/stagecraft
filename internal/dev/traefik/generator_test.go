// SPDX-License-Identifier: AGPL-3.0-or-later

package traefik

import (
	"testing"

	"gopkg.in/yaml.v3"

	"stagecraft/internal/dev/mkcert"
	"stagecraft/pkg/config"
)

// Feature: DEV_TRAEFIK
// Spec: spec/dev/traefik.md

func TestGenerator_GenerateConfig_MinimalHTTP(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	gen := NewGenerator()

	out, err := gen.GenerateConfig(
		cfg,
		"app.localdev.test",
		"frontend",
		"3000",
		"api.localdev.test",
		"backend",
		"4000",
		nil, // certCfg - HTTP only
	)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v, want nil", err)
	}

	if out == nil {
		t.Fatalf("GenerateConfig() = nil, want non-nil *Config")
		return
	}
	if out.Static == nil {
		t.Fatalf("GenerateConfig() Static = nil, want non-nil")
		return
	}
	if out.Dynamic == nil || out.Dynamic.HTTP == nil {
		t.Fatalf("GenerateConfig() Dynamic.HTTP = nil, want non-nil")
		return
	}

	httpCfg := out.Dynamic.HTTP

	if len(httpCfg.Routers) != 2 {
		t.Fatalf("routers length = %d, want 2", len(httpCfg.Routers))
	}
	if len(httpCfg.Services) != 2 {
		t.Fatalf("services length = %d, want 2", len(httpCfg.Services))
	}

	frontendRouter, ok := httpCfg.Routers["frontend"]
	if !ok {
		t.Fatalf("missing frontend router")
	}
	if frontendRouter.Rule != "Host(`app.localdev.test`)" {
		t.Errorf("frontend router rule = %q, want %q", frontendRouter.Rule, "Host(`app.localdev.test`)")
	}
	if frontendRouter.TLS != nil {
		t.Errorf("frontend TLS = %#v, want nil when HTTPS disabled", frontendRouter.TLS)
	}

	backendRouter, ok := httpCfg.Routers["backend"]
	if !ok {
		t.Fatalf("missing backend router")
	}
	if backendRouter.Rule != "Host(`api.localdev.test`)" {
		t.Errorf("backend router rule = %q, want %q", backendRouter.Rule, "Host(`api.localdev.test`)")
	}
	if backendRouter.TLS != nil {
		t.Errorf("backend TLS = %#v, want nil when HTTPS disabled", backendRouter.TLS)
	}

	frontendSvc, ok := httpCfg.Services["frontend"]
	if !ok {
		t.Fatalf("missing frontend service")
	}
	if frontendSvc.LoadBalancer == nil || len(frontendSvc.LoadBalancer.Servers) != 1 {
		t.Fatalf("frontend service servers = %#v, want 1 server", frontendSvc.LoadBalancer)
	}
	if frontendSvc.LoadBalancer.Servers[0].URL != "http://frontend:3000" {
		t.Errorf("frontend server URL = %q, want %q", frontendSvc.LoadBalancer.Servers[0].URL, "http://frontend:3000")
	}

	backendSvc, ok := httpCfg.Services["backend"]
	if !ok {
		t.Fatalf("missing backend service")
	}
	if backendSvc.LoadBalancer == nil || len(backendSvc.LoadBalancer.Servers) != 1 {
		t.Fatalf("backend service servers = %#v, want 1 server", backendSvc.LoadBalancer)
	}
	if backendSvc.LoadBalancer.Servers[0].URL != "http://backend:4000" {
		t.Errorf("backend server URL = %q, want %q", backendSvc.LoadBalancer.Servers[0].URL, "http://backend:4000")
	}
}

func TestGenerator_GenerateConfig_EnableHTTPS_TLSConfig(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	gen := NewGenerator()

	certCfg := &mkcert.CertConfig{
		Enabled:  true,
		CertFile: ".stagecraft/dev/certs/dev-local.pem",
		KeyFile:  ".stagecraft/dev/certs/dev-local-key.pem",
		Domains:  []string{"app.localdev.test", "api.localdev.test"},
	}

	out, err := gen.GenerateConfig(
		cfg,
		"app.localdev.test",
		"frontend",
		"3000",
		"api.localdev.test",
		"backend",
		"4000",
		certCfg,
	)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v, want nil", err)
	}

	httpCfg := out.Dynamic.HTTP

	for name, r := range httpCfg.Routers {
		if r.TLS == nil {
			t.Fatalf("router %q TLS = nil, want non-nil when HTTPS enabled", name)
		}
		if r.TLS.CertFile != "/certs/dev-local.pem" {
			t.Errorf("router %q TLS.CertFile = %q, want %q", name, r.TLS.CertFile, "/certs/dev-local.pem")
		}
		if r.TLS.KeyFile != "/certs/dev-local-key.pem" {
			t.Errorf("router %q TLS.KeyFile = %q, want %q", name, r.TLS.KeyFile, "/certs/dev-local-key.pem")
		}
	}

	// Sanity check YAML serialization doesn't panic and produces valid YAML.
	data, err := out.ToYAMLDynamic()
	if err != nil {
		t.Fatalf("ToYAMLDynamic() error = %v, want nil", err)
	}
	if len(data) == 0 {
		t.Fatalf("ToYAMLDynamic() returned empty output, want non-empty YAML")
	}

	var decoded map[string]any
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal(ToYAMLDynamic()) error = %v, want nil", err)
	}
	if _, ok := decoded["http"]; !ok {
		t.Fatalf("dynamic YAML missing http root key")
	}
}

func TestGenerator_GenerateConfig_HTTPSDisabled_NoTLS(t *testing.T) {
	t.Helper()

	cfg := &config.Config{}
	gen := NewGenerator()

	// CertConfig with Enabled=false should result in no TLS
	certCfg := &mkcert.CertConfig{
		Enabled: false,
	}

	out, err := gen.GenerateConfig(
		cfg,
		"app.localdev.test",
		"frontend",
		"3000",
		"api.localdev.test",
		"backend",
		"4000",
		certCfg,
	)
	if err != nil {
		t.Fatalf("GenerateConfig() error = %v, want nil", err)
	}

	httpCfg := out.Dynamic.HTTP

	for name, r := range httpCfg.Routers {
		if r.TLS != nil {
			t.Errorf("router %q TLS = %#v, want nil when CertConfig.Enabled=false", name, r.TLS)
		}
	}
}
