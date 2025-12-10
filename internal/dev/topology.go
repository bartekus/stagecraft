// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Feature: CLI_DEV
Specs: spec/commands/dev.md
Docs:
  - docs/engine/analysis/CLI_DEV.md
  - docs/engine/outlines/CLI_DEV_IMPLEMENTATION_OUTLINE.md
*/

package dev

import (
	"fmt"

	devcompose "stagecraft/internal/dev/compose"
	devmkcert "stagecraft/internal/dev/mkcert"
	devtraefik "stagecraft/internal/dev/traefik"

	corecompose "stagecraft/internal/compose"

	backendproviders "stagecraft/pkg/providers/backend"
	frontendproviders "stagecraft/pkg/providers/frontend"

	"stagecraft/pkg/config"
)

// Domains holds the dev domains for frontend and backend.
type Domains struct {
	Frontend string
	Backend  string
}

// Topology represents the composed dev infrastructure for CLI_DEV.
//
// It intentionally focuses on the *outputs* that CLI_DEV needs to
// orchestrate:
//
//   - A deterministic Docker Compose model for services
//   - A deterministic Traefik configuration for routing
//
// Process management (starting/stopping containers, Traefik, mkcert,
// hosts file) is explicitly out of scope for this type.
type Topology struct {
	Config         *config.Config
	Compose        *corecompose.ComposeFile
	Traefik        *devtraefik.Config
	Domains        Domains
	Backend        *devcompose.ServiceDefinition
	Frontend       *devcompose.ServiceDefinition
	TraefikService *devcompose.ServiceDefinition
}

// Builder orchestrates DEV_COMPOSE_INFRA and DEV_TRAEFIK to produce a Topology.
type Builder struct {
	composeGen  *devcompose.Generator
	traefikGen  *devtraefik.Generator
	backendReg  *backendproviders.Registry
	frontendReg *frontendproviders.Registry
}

// NewBuilder creates a new dev topology builder.
//
// Callers are expected to provide generators and registries so tests can construct
// deterministic builders without hidden side effects.
func NewBuilder(
	composeGen *devcompose.Generator,
	traefikGen *devtraefik.Generator,
	backendReg *backendproviders.Registry,
	frontendReg *frontendproviders.Registry,
) *Builder {
	if composeGen == nil {
		composeGen = devcompose.NewGenerator()
	}
	if traefikGen == nil {
		traefikGen = devtraefik.NewGenerator()
	}
	if backendReg == nil {
		backendReg = backendproviders.DefaultRegistry
	}
	if frontendReg == nil {
		frontendReg = frontendproviders.DefaultRegistry
	}

	return &Builder{
		composeGen:  composeGen,
		traefikGen:  traefikGen,
		backendReg:  backendReg,
		frontendReg: frontendReg,
	}
}

// NewDefaultBuilder creates a builder with default generators and registries.
// This is the convenience constructor for production use.
func NewDefaultBuilder() *Builder {
	return NewBuilder(nil, nil, nil, nil)
}

// Build constructs the dev topology by:
//
//  1. Generating a Docker Compose model for backend, frontend, and Traefik
//     via DEV_COMPOSE_INFRA.
//  2. Generating a Traefik config that routes frontend/backend domains to
//     the appropriate internal service/port via DEV_TRAEFIK.
//
// For v1, the Traefik service URLs are based on the first port mapping
// of each service (container port), matching the internal Docker network
// topology ("http://<service-name>:<container-port>").
//
// certCfg is the certificate configuration from DEV_MKCERT. When certCfg != nil
// and certCfg.Enabled is true, Traefik config will include TLS configuration.
func (b *Builder) Build(
	cfg *config.Config,
	domains Domains,
	backend *devcompose.ServiceDefinition,
	frontend *devcompose.ServiceDefinition,
	traefikService *devcompose.ServiceDefinition,
	certCfg *devmkcert.CertConfig,
) (*Topology, error) {
	// DEV_COMPOSE_INFRA already validates that backend is required.
	composeFile, err := b.composeGen.GenerateCompose(
		cfg,
		backend,
		frontend,
		traefikService,
	)
	if err != nil {
		return nil, fmt.Errorf("dev topology: generate compose: %w", err)
	}

	// Generate Traefik config only if Traefik service is included
	var traefikCfg *devtraefik.Config
	if traefikService != nil {
		frontendPort := firstContainerPort(frontend)
		backendPort := firstContainerPort(backend)

		var err error
		traefikCfg, err = b.traefikGen.GenerateConfig(
			cfg,
			domains.Frontend,
			frontendName(frontend),
			frontendPort,
			domains.Backend,
			backendName(backend),
			backendPort,
			certCfg,
		)
		if err != nil {
			return nil, fmt.Errorf("dev topology: generate traefik config: %w", err)
		}
	}

	top := &Topology{
		Config:         cfg,
		Compose:        composeFile,
		Traefik:        traefikCfg,
		Domains:        domains,
		Backend:        backend,
		Frontend:       frontend,
		TraefikService: traefikService,
	}

	return top, nil
}

// ResolveServiceDefinitions resolves backend and frontend providers from config
// and extracts their dev service definitions.
//
// For v1, this extracts basic service info (name, ports from env vars).
// Future slices can enhance this to extract more detailed service configuration.
func (b *Builder) ResolveServiceDefinitions(
	cfg *config.Config,
	env string,
) (backend, frontend *devcompose.ServiceDefinition, err error) {
	// Resolve backend service
	if cfg.Backend != nil {
		backendProvider, err := b.backendReg.Get(cfg.Backend.Provider)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve backend provider %q: %w", cfg.Backend.Provider, err)
		}

		providerCfg, err := cfg.Backend.GetProviderConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("get backend provider config: %w", err)
		}

		backend = b.extractBackendServiceDefinition(backendProvider, providerCfg)
	}

	// Resolve frontend service
	if cfg.Frontend != nil {
		frontendProvider, err := b.frontendReg.Get(cfg.Frontend.Provider)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve frontend provider %q: %w", cfg.Frontend.Provider, err)
		}

		providerCfg, err := cfg.Frontend.GetProviderConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("get frontend provider config: %w", err)
		}

		frontend = b.extractFrontendServiceDefinition(frontendProvider, providerCfg)
	}

	return backend, frontend, nil
}

// extractBackendServiceDefinition extracts a ServiceDefinition from backend provider config.
// For v1, this is intentionally minimal - it extracts:
// - Service name: "backend"
// - Ports: from PORT env var if present in provider config
// - Environment: from provider config env section
//
// Future slices can enhance this to extract image, build, volumes, etc.
func (b *Builder) extractBackendServiceDefinition(
	_ backendproviders.BackendProvider,
	providerCfg any,
) *devcompose.ServiceDefinition {
	svc := &devcompose.ServiceDefinition{
		Name: "backend",
	}

	// For v1, extract basic info from generic provider config structure.
	// This is a simplified extraction that works for generic provider.
	// Future: providers could implement a DevServiceDefinition() method.
	if cfgMap, ok := providerCfg.(map[string]any); ok {
		if devCfg, ok := cfgMap["dev"].(map[string]any); ok {
			// Extract environment variables
			if envMap, ok := devCfg["env"].(map[string]any); ok {
				env := make(map[string]string)
				for k, v := range envMap {
					if vStr, ok := v.(string); ok {
						env[k] = vStr
					}
				}
				svc.Environment = env

				// Extract PORT from env to create port mapping
				if portStr, ok := env["PORT"]; ok && portStr != "" {
					svc.Ports = []devcompose.PortMapping{
						{
							Host:      portStr,
							Container: portStr,
							Protocol:  "tcp",
						},
					}
				}
			}
		}
	}

	return svc
}

// extractFrontendServiceDefinition extracts a ServiceDefinition from frontend provider config.
// For v1, this is intentionally minimal - it extracts:
// - Service name: "frontend"
// - Ports: from PORT env var or common frontend ports (3000, 5173)
// - Environment: from provider config env section
//
// Future slices can enhance this to extract image, build, volumes, etc.
func (b *Builder) extractFrontendServiceDefinition(
	_ frontendproviders.FrontendProvider,
	providerCfg any,
) *devcompose.ServiceDefinition {
	svc := &devcompose.ServiceDefinition{
		Name: "frontend",
	}

	// For v1, extract basic info from generic provider config structure.
	// This is a simplified extraction that works for generic provider.
	// Future: providers could implement a DevServiceDefinition() method.
	if cfgMap, ok := providerCfg.(map[string]any); ok {
		if devCfg, ok := cfgMap["dev"].(map[string]any); ok {
			// Extract environment variables
			if envMap, ok := devCfg["env"].(map[string]any); ok {
				env := make(map[string]string)
				for k, v := range envMap {
					if vStr, ok := v.(string); ok {
						env[k] = vStr
					}
				}
				svc.Environment = env
			}

			// Extract port from env or use defaults
			port := "3000" // default frontend port
			if envMap, ok := devCfg["env"].(map[string]any); ok {
				if portStr, ok := envMap["PORT"].(string); ok && portStr != "" {
					port = portStr
				}
			}

			// Check command for common port patterns (e.g., --port 5173)
			if cmd, ok := devCfg["command"].([]any); ok {
				for i, arg := range cmd {
					if argStr, ok := arg.(string); ok {
						if argStr == "--port" && i+1 < len(cmd) {
							if nextPort, ok := cmd[i+1].(string); ok {
								port = nextPort
								break
							}
						}
						// Check for --port=5173 format
						if len(argStr) > 7 && argStr[:7] == "--port=" {
							port = argStr[7:]
							break
						}
					}
				}
			}

			svc.Ports = []devcompose.PortMapping{
				{
					Host:      port,
					Container: port,
					Protocol:  "tcp",
				},
			}
		}
	}

	return svc
}

// firstContainerPort returns the first container port from the service
// definition, or an empty string if no ports are defined.
//
// For v1 this is intentionally simple; if a provider needs more
// control, the spec and generator APIs can be extended later.
func firstContainerPort(svc *devcompose.ServiceDefinition) string {
	if svc == nil {
		return ""
	}
	if len(svc.Ports) == 0 {
		return ""
	}
	return svc.Ports[0].Container
}

func frontendName(svc *devcompose.ServiceDefinition) string {
	if svc == nil || svc.Name == "" {
		return ""
	}
	return svc.Name
}

func backendName(svc *devcompose.ServiceDefinition) string {
	if svc == nil || svc.Name == "" {
		return ""
	}
	return svc.Name
}
