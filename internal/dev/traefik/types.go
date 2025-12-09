// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Feature: DEV_TRAEFIK
Spec: spec/dev/traefik.md
*/

package traefik

// StaticConfig represents Traefik static configuration for dev.
type StaticConfig struct {
	EntryPoints map[string]EntryPointConfig `yaml:"entryPoints"`
	Providers   map[string]ProviderConfig   `yaml:"providers"`
}

// EntryPointConfig represents a single entry point (e.g., web, websecure).
type EntryPointConfig struct {
	Address string `yaml:"address"`
}

// ProviderConfig represents a generic provider configuration.
type ProviderConfig struct {
	Docker *DockerProviderConfig `yaml:"docker,omitempty"`
}

// DockerProviderConfig represents the Docker provider configuration.
type DockerProviderConfig struct {
	Endpoint         string `yaml:"endpoint"`
	ExposedByDefault bool   `yaml:"exposedByDefault"`
	Network          string `yaml:"network"`
}

// DynamicConfig represents Traefik dynamic HTTP configuration.
type DynamicConfig struct {
	HTTP *HTTPConfig `yaml:"http"`
}

// HTTPConfig contains HTTP routers, services, and middlewares.
type HTTPConfig struct {
	Routers     map[string]RouterConfig     `yaml:"routers"`
	Services    map[string]ServiceConfig    `yaml:"services"`
	Middlewares map[string]MiddlewareConfig `yaml:"middlewares"`
}

// RouterConfig represents a Traefik router.
type RouterConfig struct {
	Rule        string     `yaml:"rule"`
	Service     string     `yaml:"service"`
	EntryPoints []string   `yaml:"entryPoints"`
	TLS         *TLSConfig `yaml:"tls,omitempty"`
}

// ServiceConfig represents a Traefik service.
type ServiceConfig struct {
	LoadBalancer *LoadBalancerConfig `yaml:"loadBalancer"`
}

// LoadBalancerConfig represents load balancer configuration.
type LoadBalancerConfig struct {
	Servers []ServerConfig `yaml:"servers"`
}

// ServerConfig represents a backend server.
type ServerConfig struct {
	URL string `yaml:"url"`
}

// MiddlewareConfig is a placeholder for future middlewares.
type MiddlewareConfig struct {
	// v1: intentionally empty
}

// TLSConfig represents TLS configuration for a router.
type TLSConfig struct {
	CertFile string `yaml:"certFile,omitempty"`
	KeyFile  string `yaml:"keyFile,omitempty"`
}

// Config is the top-level container for generated Traefik config.
type Config struct {
	Static  *StaticConfig  `yaml:"-"`
	Dynamic *DynamicConfig `yaml:"-"`
}
