// SPDX-License-Identifier: AGPL-3.0-or-later

// Package traefik provides Traefik configuration generation for development environments.
package traefik

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"

	"stagecraft/internal/dev/mkcert"
	"stagecraft/pkg/config"
)

// Feature: DEV_TRAEFIK
// Spec: spec/dev/traefik.md

// Generator generates Traefik configuration for dev environments.
type Generator struct {
	// future options can be added here
}

// NewGenerator creates a new Traefik config generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateConfig generates Traefik static and dynamic configuration.
//
// This is a thin v1 slice that:
// - Configures web / websecure entry points
// - Configures docker provider bound to the "stagecraft-dev" network
// - Creates one frontend router+service and one backend router+service
// - Wires TLS configuration from certCfg when enabled
//
// certCfg is the certificate configuration from DEV_MKCERT. When certCfg != nil
// and certCfg.Enabled is true, TLS configuration will reference certCfg.CertFile
// and certCfg.KeyFile using container-relative paths.
func (g *Generator) GenerateConfig(
	cfg *config.Config,
	frontendDomain string,
	frontendService string,
	frontendPort string,
	backendDomain string,
	backendService string,
	backendPort string,
	certCfg *mkcert.CertConfig,
) (*Config, error) {
	_ = cfg // v1: config not yet used

	static := &StaticConfig{
		EntryPoints: map[string]EntryPointConfig{
			"web": {
				Address: ":80",
			},
			"websecure": {
				Address: ":443",
			},
		},
		Providers: map[string]ProviderConfig{
			"docker": {
				Docker: &DockerProviderConfig{
					Endpoint:         "unix:///var/run/docker.sock",
					ExposedByDefault: false,
					Network:          "stagecraft-dev",
				},
			},
		},
	}

	// Build TLS config if HTTPS enabled.
	var tlsCfg *TLSConfig
	if certPath, keyPath, ok := certPathsFromConfig(certCfg); ok {
		tlsCfg = &TLSConfig{
			CertFile: certPath,
			KeyFile:  keyPath,
		}
	}

	httpCfg := &HTTPConfig{
		Routers:     make(map[string]RouterConfig),
		Services:    make(map[string]ServiceConfig),
		Middlewares: make(map[string]MiddlewareConfig),
	}

	// Frontend router and service.
	if frontendDomain != "" && frontendService != "" && frontendPort != "" {
		httpCfg.Routers["frontend"] = RouterConfig{
			Rule:        fmt.Sprintf("Host(`%s`)", frontendDomain),
			Service:     "frontend",
			EntryPoints: []string{"web", "websecure"},
			TLS:         tlsCfg,
		}

		httpCfg.Services["frontend"] = ServiceConfig{
			LoadBalancer: &LoadBalancerConfig{
				Servers: []ServerConfig{
					{URL: fmt.Sprintf("http://%s:%s", frontendService, frontendPort)},
				},
			},
		}
	}

	// Backend router and service.
	if backendDomain != "" && backendService != "" && backendPort != "" {
		httpCfg.Routers["backend"] = RouterConfig{
			Rule:        fmt.Sprintf("Host(`%s`)", backendDomain),
			Service:     "backend",
			EntryPoints: []string{"web", "websecure"},
			TLS:         tlsCfg,
		}

		httpCfg.Services["backend"] = ServiceConfig{
			LoadBalancer: &LoadBalancerConfig{
				Servers: []ServerConfig{
					{URL: fmt.Sprintf("http://%s:%s", backendService, backendPort)},
				},
			},
		}
	}

	// Deterministic ordering of entry points and maps will be enforced at
	// YAML serialization time by sorting keys where needed.
	sortEntryPoints(static)
	sortHTTPConfig(httpCfg)

	cfgOut := &Config{
		Static: &StaticConfig{
			EntryPoints: static.EntryPoints,
			Providers:   static.Providers,
		},
		Dynamic: &DynamicConfig{
			HTTP: httpCfg,
		},
	}

	return cfgOut, nil
}

// ToYAMLStatic encodes the static config to YAML in a deterministic way.
func (c *Config) ToYAMLStatic() ([]byte, error) {
	if c == nil || c.Static == nil {
		return nil, nil
	}

	node := &yaml.Node{}
	if err := node.Encode(c.Static); err != nil {
		return nil, err
	}

	// yaml.v3 already sorts map keys; indentation defaults are acceptable.
	buf := &outBuffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	_ = enc.Close()

	return buf.Bytes(), nil
}

// ToYAMLDynamic encodes the dynamic config to YAML in a deterministic way.
func (c *Config) ToYAMLDynamic() ([]byte, error) {
	if c == nil || c.Dynamic == nil || c.Dynamic.HTTP == nil {
		return nil, nil
	}

	node := &yaml.Node{}
	if err := node.Encode(c.Dynamic); err != nil {
		return nil, err
	}

	buf := &outBuffer{}
	enc := yaml.NewEncoder(buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	_ = enc.Close()

	return buf.Bytes(), nil
}

// outBuffer is a small helper to capture yaml.Encoder output without importing bytes.
type outBuffer struct {
	data []byte
}

func (b *outBuffer) Write(p []byte) (int, error) {
	b.data = append(b.data, p...)
	return len(p), nil
}

func (b *outBuffer) Bytes() []byte {
	return b.data
}

// sortEntryPoints ensures deterministic ordering in EntryPoints map
// by recreating the map in sorted-key order.
func sortEntryPoints(static *StaticConfig) {
	if static == nil || len(static.EntryPoints) == 0 {
		return
	}

	keys := make([]string, 0, len(static.EntryPoints))
	for k := range static.EntryPoints {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ordered := make(map[string]EntryPointConfig, len(static.EntryPoints))
	for _, k := range keys {
		ordered[k] = static.EntryPoints[k]
	}
	static.EntryPoints = ordered
}

// sortHTTPConfig ensures deterministic ordering of routers and services.
// (Map key iteration order is non-deterministic; re-building maps gives us
// stable behavior when encoded with yaml.v3.)
func sortHTTPConfig(httpCfg *HTTPConfig) {
	if httpCfg == nil {
		return
	}

	if len(httpCfg.Routers) > 0 {
		keys := make([]string, 0, len(httpCfg.Routers))
		for k := range httpCfg.Routers {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		ordered := make(map[string]RouterConfig, len(httpCfg.Routers))
		for _, k := range keys {
			rc := httpCfg.Routers[k]

			// Sort entry points inside each router.
			if len(rc.EntryPoints) > 0 {
				sort.Strings(rc.EntryPoints)
			}

			ordered[k] = rc
		}
		httpCfg.Routers = ordered
	}

	if len(httpCfg.Services) > 0 {
		keys := make([]string, 0, len(httpCfg.Services))
		for k := range httpCfg.Services {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		ordered := make(map[string]ServiceConfig, len(httpCfg.Services))
		for _, k := range keys {
			svc := httpCfg.Services[k]

			if svc.LoadBalancer != nil && len(svc.LoadBalancer.Servers) > 0 {
				sort.Slice(svc.LoadBalancer.Servers, func(i, j int) bool {
					return svc.LoadBalancer.Servers[i].URL < svc.LoadBalancer.Servers[j].URL
				})
			}

			ordered[k] = svc
		}
		httpCfg.Services = ordered
	}

	// Middlewares map is currently unused; when we start populating it,
	// we will apply the same sorted-key pattern.
}

const (
	// certsMountPath is the container path where certificates are mounted.
	// DEV_COMPOSE_INFRA mounts .stagecraft/dev/certs/ to this path.
	certsMountPath = "/certs"
)

// certPathsFromConfig converts mkcert.CertConfig paths to container-relative paths
// for use in Traefik TLS configuration.
//
// Returns container paths (e.g., "/certs/dev-local.pem", "/certs/dev-local-key.pem")
// when certCfg is enabled, or empty strings and false when HTTPS is disabled.
func certPathsFromConfig(certCfg *mkcert.CertConfig) (certPath, keyPath string, ok bool) {
	if certCfg == nil || !certCfg.Enabled {
		return "", "", false
	}

	// We trust mkcert.CertFile/KeyFile to be `.stagecraft/dev/certs/dev-local*.pem`.
	// DEV_COMPOSE_INFRA's mount is responsible for mapping that directory to /certs.
	// Using filepath.Base ensures we only reference the filename, making the config
	// stable regardless of the exact host path structure.
	return path.Join(certsMountPath, filepath.Base(certCfg.CertFile)),
		path.Join(certsMountPath, filepath.Base(certCfg.KeyFile)),
		true
}
