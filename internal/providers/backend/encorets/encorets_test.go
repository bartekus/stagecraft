// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package encorets

import (
	"fmt"
	"strings"
	"testing"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

func TestEncoreTsProvider_ID(t *testing.T) {
	p := &EncoreTsProvider{}
	if got := p.ID(); got != "encore-ts" {
		t.Errorf("ID() = %q, want %q", got, "encore-ts")
	}
}

func TestEncoreTsProvider_ParseConfig(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"encore_secrets": map[string]any{
				"types":    []string{"dev", "preview", "local"},
				"from_env": []string{"DOMAIN", "API_DOMAIN"},
			},
			"entrypoint": "./backend",
			"env_file":   ".env.local",
			"listen":     "0.0.0.0:4000",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if len(parsed.Dev.EncoreSecrets.Types) != 3 {
		t.Errorf("Dev.EncoreSecrets.Types length = %d, want 3", len(parsed.Dev.EncoreSecrets.Types))
	}

	if len(parsed.Dev.EncoreSecrets.FromEnv) != 2 {
		t.Errorf("Dev.EncoreSecrets.FromEnv length = %d, want 2", len(parsed.Dev.EncoreSecrets.FromEnv))
	}

	if parsed.Dev.EntryPoint != "./backend" {
		t.Errorf("Dev.EntryPoint = %q, want %q", parsed.Dev.EntryPoint, "./backend")
	}

	if parsed.Dev.Listen != "0.0.0.0:4000" {
		t.Errorf("Dev.Listen = %q, want %q", parsed.Dev.Listen, "0.0.0.0:4000")
	}
}

func TestEncoreTsProvider_ParseConfig_InvalidYAML(t *testing.T) {
	p := &EncoreTsProvider{}

	// Invalid config structure
	cfg := "not a map"

	_, err := p.parseConfig(cfg)
	if err == nil {
		t.Error("parseConfig() error = nil, want error for invalid config")
	}
}

func TestEncoreTsProvider_ParseConfig_WithAllFields(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"env_file":            ".env.local",
			"listen":              "0.0.0.0:4000",
			"workdir":             "./backend",
			"entrypoint":          "./src/index.ts",
			"disable_telemetry":   true,
			"node_extra_ca_certs": "./certs/ca.pem",
			"encore_secrets": map[string]any{
				"types":    []string{"dev", "preview"},
				"from_env": []string{"SECRET1", "SECRET2"},
			},
		},
		"build": map[string]any{
			"workdir":           "./backend",
			"image_name":        "my-api",
			"docker_tag_suffix": "-encore",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if parsed.Dev.EnvFile != ".env.local" {
		t.Errorf("Dev.EnvFile = %q, want %q", parsed.Dev.EnvFile, ".env.local")
	}

	if parsed.Dev.WorkDir != "./backend" {
		t.Errorf("Dev.WorkDir = %q, want %q", parsed.Dev.WorkDir, "./backend")
	}

	if parsed.Dev.DisableTelemetry != true {
		t.Errorf("Dev.DisableTelemetry = %v, want true", parsed.Dev.DisableTelemetry)
	}

	if parsed.Dev.NodeExtraCACerts != "./certs/ca.pem" {
		t.Errorf("Dev.NodeExtraCACerts = %q, want %q", parsed.Dev.NodeExtraCACerts, "./certs/ca.pem")
	}

	if parsed.Build.WorkDir != "./backend" {
		t.Errorf("Build.WorkDir = %q, want %q", parsed.Build.WorkDir, "./backend")
	}

	if parsed.Build.ImageName != "my-api" {
		t.Errorf("Build.ImageName = %q, want %q", parsed.Build.ImageName, "my-api")
	}

	if parsed.Build.DockerTagSuffix != "-encore" {
		t.Errorf("Build.DockerTagSuffix = %q, want %q", parsed.Build.DockerTagSuffix, "-encore")
	}
}

func TestEncoreTsProvider_ParseConfig_DefaultImageName(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"build": map[string]any{},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if parsed.Build.ImageName != "api" {
		t.Errorf("Build.ImageName = %q, want %q (default)", parsed.Build.ImageName, "api")
	}
}

func TestEncoreTsProvider_ValidateDevConfig_MissingListen(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := &Config{}
	// Missing required field: Listen is empty

	err := p.validateDevConfig(cfg)
	if err == nil {
		t.Error("validateDevConfig() error = nil, want error for missing listen")
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("expected ProviderError, got nil")
	}

	if pe.Category != ErrInvalidConfig {
		t.Errorf("ProviderError.Category = %q, want %q", pe.Category, ErrInvalidConfig)
	}

	if pe.Operation != "dev" {
		t.Errorf("ProviderError.Operation = %q, want %q", pe.Operation, "dev")
	}
}

func TestEncoreTsProvider_ValidateDevConfig_Valid(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := &Config{}
	cfg.Dev.Listen = "0.0.0.0:4000"

	err := p.validateDevConfig(cfg)
	if err != nil {
		t.Errorf("validateDevConfig() error = %v, want nil", err)
	}
}

func TestProviderError_Error(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test error",
		Detail:    "test detail",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
	}

	if !strings.Contains(msg, "encore-ts") {
		t.Errorf("Error() message should contain provider ID, got %q", msg)
	}

	if !strings.Contains(msg, "dev") {
		t.Errorf("Error() message should contain operation, got %q", msg)
	}

	if !strings.Contains(msg, ErrInvalidConfig) {
		t.Errorf("Error() message should contain category, got %q", msg)
	}
}

func TestProviderError_Error_NoDetail(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test error",
	}

	msg := err.Error()
	if strings.Contains(msg, ":") && strings.Contains(msg, "test detail") {
		t.Error("Error() should not include detail when Detail is empty")
	}
}

func TestIsProviderError(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	if !IsProviderError(err) {
		t.Error("IsProviderError() = false, want true")
	}

	regularErr := fmt.Errorf("regular error")
	if IsProviderError(regularErr) {
		t.Error("IsProviderError() = true for regular error, want false")
	}
}

func TestGetProviderError(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	pe := GetProviderError(err)
	if pe == nil {
		t.Fatal("GetProviderError() = nil, want ProviderError")
	}

	if pe.Category != ErrInvalidConfig {
		t.Errorf("GetProviderError().Category = %q, want %q", pe.Category, ErrInvalidConfig)
	}

	regularErr := fmt.Errorf("regular error")
	if GetProviderError(regularErr) != nil {
		t.Error("GetProviderError() should return nil for regular error")
	}
}

func TestProviderError_Category(t *testing.T) {
	err := &ProviderError{
		Category:  ErrInvalidConfig,
		Provider:  "encore-ts",
		Operation: "dev",
		Message:   "test",
	}

	if got := err.Category; got != ErrInvalidConfig {
		t.Errorf("Category = %q, want %q", got, ErrInvalidConfig)
	}
}

func TestParseEnvFileInto(t *testing.T) {
	tests := []struct {
		name    string
		envFile string
		wantEnv map[string]string
	}{
		{
			name: "inline comments",
			envFile: `KEY1=value1 # inline comment
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "export keyword",
			envFile: `export KEY1=value1
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "quoted values with escapes",
			envFile: `KEY1="value with spaces"
KEY2="value with \"quotes\""
KEY3="value with\nnewline"
KEY4='single quoted'
KEY5=unquoted`,
			wantEnv: map[string]string{
				"KEY1": "value with spaces",
				"KEY2": "value with \"quotes\"",
				"KEY3": "value with\nnewline",
				"KEY4": "single quoted",
				"KEY5": "unquoted",
			},
		},
		{
			name: "empty values",
			envFile: `KEY1=
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "",
				"KEY2": "value2",
			},
		},
		{
			name: "preserve # inside quotes",
			envFile: `KEY1="value # not a comment"
KEY2=value # this is a comment`,
			wantEnv: map[string]string{
				"KEY1": "value # not a comment",
				"KEY2": "value",
			},
		},
		{
			name: "later values override earlier",
			envFile: `KEY=first
KEY=second`,
			wantEnv: map[string]string{
				"KEY": "second",
			},
		},
		{
			name: "blank lines and comments",
			envFile: `# This is a comment
KEY1=value1

KEY2=value2
# Another comment
KEY3=value3`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
		{
			name: "escape sequences in double quotes",
			envFile: `KEY1="tab\there"
KEY2="newline\nhere"
KEY3="backslash\\here"
KEY4="quote\"here"`,
			wantEnv: map[string]string{
				"KEY1": "tab\there",
				"KEY2": "newline\nhere",
				"KEY3": "backslash\\here",
				"KEY4": "quote\"here",
			},
		},
		{
			name: "malformed lines are skipped",
			envFile: `KEY1=value1
MALFORMED
KEY2=value2`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
		},
		{
			name: "empty keys are skipped",
			envFile: `KEY1=value1
=value2
KEY2=value3`,
			wantEnv: map[string]string{
				"KEY1": "value1",
				"KEY2": "value3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := make(map[string]string)
			parseEnvFileInto(env, []byte(tt.envFile))

			// Manual comparison
			if len(env) != len(tt.wantEnv) {
				t.Errorf("parseEnvFileInto() got %d keys, want %d", len(env), len(tt.wantEnv))
			}
			for k, wantV := range tt.wantEnv {
				if gotV, ok := env[k]; !ok {
					t.Errorf("parseEnvFileInto() missing key %q", k)
				} else if gotV != wantV {
					t.Errorf("parseEnvFileInto() key %q = %q, want %q", k, gotV, wantV)
				}
			}
			// Check for unexpected keys
			for k := range env {
				if _, ok := tt.wantEnv[k]; !ok {
					t.Errorf("parseEnvFileInto() unexpected key %q", k)
				}
			}
		})
	}
}
