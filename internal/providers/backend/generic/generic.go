// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package generic

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_GENERIC
// Spec: spec/providers/backend/generic.md

// GenericProvider implements a command-based backend provider.
type GenericProvider struct{}

// Ensure GenericProvider implements BackendProvider
var _ backend.BackendProvider = (*GenericProvider)(nil)

// ID returns the provider identifier.
func (p *GenericProvider) ID() string {
	return "generic"
}

// Config represents the generic provider configuration.
type Config struct {
	Dev struct {
		Command []string          `yaml:"command"`
		WorkDir string            `yaml:"workdir"`
		Env     map[string]string `yaml:"env"`
	} `yaml:"dev"`

	Build struct {
		Dockerfile string `yaml:"dockerfile"`
		Context    string `yaml:"context"`
	} `yaml:"build"`
}

// Dev runs the backend in development mode.
func (p *GenericProvider) Dev(ctx context.Context, opts backend.DevOptions) error {
	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return fmt.Errorf("parsing generic provider config: %w", err)
	}

	if len(cfg.Dev.Command) == 0 {
		return fmt.Errorf("generic provider: dev.command is required")
	}

	workDir := cfg.Dev.WorkDir
	if workDir == "" {
		workDir = opts.WorkDir
	}
	if workDir == "" {
		workDir = "."
	}

	// Merge provider env with opts.Env
	env := make(map[string]string)
	for k, v := range opts.Env {
		env[k] = v
	}
	for k, v := range cfg.Dev.Env {
		env[k] = v
	}

	// Build command
	cmd := exec.CommandContext(ctx, cfg.Dev.Command[0], cfg.Dev.Command[1:]...)
	cmd.Dir = workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Stream output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// BuildDocker builds a Docker image.
func (p *GenericProvider) BuildDocker(ctx context.Context, opts backend.BuildDockerOptions) (string, error) {
	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return "", fmt.Errorf("parsing generic provider config: %w", err)
	}

	dockerfile := cfg.Build.Dockerfile
	if dockerfile == "" {
		dockerfile = "Dockerfile"
	}

	context := cfg.Build.Context
	if context == "" {
		context = opts.WorkDir
	}
	if context == "" {
		context = "."
	}

	// Build docker build command
	args := []string{
		"build",
		"-t", opts.ImageTag,
		"-f", dockerfile,
		context,
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker build failed: %w", err)
	}

	return opts.ImageTag, nil
}

// parseConfig unmarshals the provider config.
func (p *GenericProvider) parseConfig(cfg any) (*Config, error) {
	// Convert to YAML bytes and unmarshal
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid generic provider config: %w", err)
	}

	return &config, nil
}

func init() {
	backend.Register(&GenericProvider{})
}

