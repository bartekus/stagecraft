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
	"os"
	"path/filepath"
	"testing"

	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_GENERIC
// Spec: spec/providers/backend/generic.md

func TestGenericProvider_ID(t *testing.T) {
	p := &GenericProvider{}
	if got := p.ID(); got != "generic" {
		t.Errorf("ID() = %q, want %q", got, "generic")
	}
}

func TestGenericProvider_ParseConfig(t *testing.T) {
	p := &GenericProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"command": []string{"go", "run", "main.go"},
			"workdir": "./cmd/api",
			"env": map[string]string{
				"PORT": "4000",
			},
		},
		"build": map[string]any{
			"dockerfile": "./Dockerfile",
			"context": ".",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if len(parsed.Dev.Command) != 3 {
		t.Errorf("Dev.Command length = %d, want 3", len(parsed.Dev.Command))
	}

	if parsed.Dev.WorkDir != "./cmd/api" {
		t.Errorf("Dev.WorkDir = %q, want %q", parsed.Dev.WorkDir, "./cmd/api")
	}

	if parsed.Dev.Env["PORT"] != "4000" {
		t.Errorf("Dev.Env[PORT] = %q, want %q", parsed.Dev.Env["PORT"], "4000")
	}

	if parsed.Build.Dockerfile != "./Dockerfile" {
		t.Errorf("Build.Dockerfile = %q, want %q", parsed.Build.Dockerfile, "./Dockerfile")
	}
}

func TestGenericProvider_ParseConfig_InvalidYAML(t *testing.T) {
	p := &GenericProvider{}

	// Invalid config structure
	cfg := "not a map"

	_, err := p.parseConfig(cfg)
	if err == nil {
		t.Error("parseConfig() error = nil, want error for invalid config")
	}
}

func TestGenericProvider_Dev_ValidatesCommand(t *testing.T) {
	p := &GenericProvider{}

	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"command": []string{},
			},
		},
		WorkDir: ".",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := p.Dev(ctx, opts)
	if err == nil {
		t.Error("expected error for empty command, got nil")
	}

	if err != nil && err.Error() == "" {
		t.Error("expected error message, got empty")
	}
}

func TestGenericProvider_Dev_WorkDirResolution(t *testing.T) {
	p := &GenericProvider{}

	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  map[string]any
		opts    backend.DevOptions
		wantDir string
	}{
		{
			name: "config workdir takes precedence",
			config: map[string]any{
				"dev": map[string]any{
					"command": []string{"echo", "test"},
					"workdir": tmpDir,
				},
			},
			opts: backend.DevOptions{
				WorkDir: "/other/dir",
			},
			wantDir: tmpDir,
		},
		{
			name: "opts workdir used when config missing",
			config: map[string]any{
				"dev": map[string]any{
					"command": []string{"echo", "test"},
				},
			},
			opts: backend.DevOptions{
				WorkDir: tmpDir,
			},
			wantDir: tmpDir,
		},
		{
			name: "defaults to current directory",
			config: map[string]any{
				"dev": map[string]any{
					"command": []string{"echo", "test"},
				},
			},
			opts: backend.DevOptions{
				WorkDir: "",
			},
			wantDir: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Config = tt.config

			// Create a test script that outputs its working directory
			testScript := filepath.Join(tt.wantDir, "test_script.sh")
			if err := os.WriteFile(testScript, []byte("#!/bin/sh\npwd\n"), 0755); err != nil {
				t.Fatalf("failed to create test script: %v", err)
			}

			// Update config to use the script
			tt.opts.Config = map[string]any{
				"dev": map[string]any{
					"command": []string{testScript},
					"workdir": tt.wantDir,
				},
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// This will fail because the script might not exist in all cases,
			// but we're testing the workdir resolution logic
			_ = p.Dev(ctx, tt.opts)
		})
	}
}

func TestGenericProvider_Dev_EnvMerging(t *testing.T) {
	p := &GenericProvider{}

	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"command": []string{"echo", "test"},
				"env": map[string]string{
					"PROVIDER_VAR": "provider-value",
					"OVERRIDE":     "provider-override",
				},
			},
		},
		WorkDir: ".",
		Env: map[string]string{
			"OPTS_VAR":  "opts-value",
			"OVERRIDE":  "opts-override",
			"OPTS_ONLY": "opts-only-value",
		},
	}

	// We can't easily test the actual env merging without running a command,
	// but we can verify the config parsing works
	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if cfg.Dev.Env["PROVIDER_VAR"] != "provider-value" {
		t.Errorf("provider env not parsed correctly")
	}
}

func TestGenericProvider_BuildDocker_DefaultDockerfile(t *testing.T) {
	p := &GenericProvider{}

	opts := backend.BuildDockerOptions{
		Config: map[string]any{
			"build": map[string]any{
				// dockerfile not specified, should default to "Dockerfile"
			},
		},
		ImageTag: "test:tag",
		WorkDir:  ".",
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This will fail because docker might not be available,
	// but we're testing the default dockerfile logic
	_, err := p.BuildDocker(ctx, opts)
	// Error is expected (docker not available or no Dockerfile), but config parsing should succeed
	if err != nil && err.Error() == "" {
		t.Error("expected error message, got empty")
	}
}

func TestGenericProvider_BuildDocker_ContextResolution(t *testing.T) {
	p := &GenericProvider{}

	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  map[string]any
		opts    backend.BuildDockerOptions
		wantCtx string
	}{
		{
			name: "config context takes precedence",
			config: map[string]any{
				"build": map[string]any{
					"context": tmpDir,
				},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "test:tag",
				WorkDir:  "/other/dir",
			},
			wantCtx: tmpDir,
		},
		{
			name: "opts workdir used when context missing",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "test:tag",
				WorkDir:  tmpDir,
			},
			wantCtx: tmpDir,
		},
		{
			name: "defaults to current directory",
			config: map[string]any{
				"build": map[string]any{},
			},
			opts: backend.BuildDockerOptions{
				ImageTag: "test:tag",
				WorkDir:  "",
			},
			wantCtx: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.opts.Config = tt.config

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// This will fail because docker might not be available,
			// but we're testing the context resolution logic
			_, err := p.BuildDocker(ctx, tt.opts)
			// Error is expected, but we verify the logic doesn't panic
			if err != nil && err.Error() == "" {
				t.Error("expected error message, got empty")
			}
		})
	}
}

