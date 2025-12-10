// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package env

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"stagecraft/pkg/config"
)

// Feature: CORE_ENV_RESOLUTION
// Spec: spec/core/env-resolution.md

func TestNewResolver(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {Driver: "digitalocean"},
		},
	}

	resolver := NewResolver(cfg)
	if resolver == nil {
		t.Fatal("NewResolver returned nil")
		return
	}
	if resolver.cfg != cfg {
		t.Error("NewResolver did not store config correctly")
	}
}

func TestResolver_Resolve_EnvironmentNotFound(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {Driver: "digitalocean"},
		},
	}

	resolver := NewResolver(cfg)
	_, err := resolver.Resolve(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent environment")
	}

	if !errors.Is(err, ErrEnvironmentNotFound) {
		t.Errorf("expected ErrEnvironmentNotFound, got %v", err)
	}
}

func TestResolver_Resolve_EmptyEnvironments(t *testing.T) {
	cfg := &config.Config{
		Project:      config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{},
	}

	resolver := NewResolver(cfg)
	_, err := resolver.Resolve(context.Background(), "dev")
	if err == nil {
		t.Fatal("expected error for empty environments map")
	}

	if !errors.Is(err, ErrEnvironmentNotFound) {
		t.Errorf("expected ErrEnvironmentNotFound, got %v", err)
	}
}

func TestResolver_Resolve_Basic(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: ".env.local",
			},
		},
	}

	resolver := NewResolver(cfg)
	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if ctx.Name != "dev" {
		t.Errorf("expected Name 'dev', got %q", ctx.Name)
	}

	if ctx.Config.Driver != "digitalocean" {
		t.Errorf("expected Driver 'digitalocean', got %q", ctx.Config.Driver)
	}

	if ctx.EnvFile == "" {
		t.Error("expected EnvFile to be set")
	}

	if ctx.Variables == nil {
		t.Error("expected Variables to be initialized")
	}
}

func TestResolver_Resolve_WithEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.test")

	envContent := `KEY1=value1
KEY2=value2
# This is a comment
KEY3=value3
`
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: filepath.Base(envFile), // Use relative path
			},
		},
	}

	resolver := NewResolver(cfg)
	resolver.SetWorkDir(tmpDir)

	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if ctx.Variables["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1', got %q", ctx.Variables["KEY1"])
	}

	if ctx.Variables["KEY2"] != "value2" {
		t.Errorf("expected KEY2='value2', got %q", ctx.Variables["KEY2"])
	}

	if ctx.Variables["KEY3"] != "value3" {
		t.Errorf("expected KEY3='value3', got %q", ctx.Variables["KEY3"])
	}

	// Comments should not be in variables
	if _, ok := ctx.Variables["#"]; ok {
		t.Error("comments should not be parsed as variables")
	}
}

func TestResolver_Resolve_SystemOverridesEnvFile(t *testing.T) {
	// Set a system environment variable
	originalValue := os.Getenv("TEST_ENV_VAR")
	defer func() {
		if originalValue == "" {
			if err := os.Unsetenv("TEST_ENV_VAR"); err != nil {
				t.Logf("failed to unset TEST_ENV_VAR: %v", err)
			}
		} else {
			if err := os.Setenv("TEST_ENV_VAR", originalValue); err != nil {
				t.Logf("failed to restore TEST_ENV_VAR: %v", err)
			}
		}
	}()

	if err := os.Setenv("TEST_ENV_VAR", "system-value"); err != nil {
		t.Fatalf("failed to set TEST_ENV_VAR: %v", err)
	}

	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.test")

	envContent := "TEST_ENV_VAR=file-value\n"
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: filepath.Base(envFile),
			},
		},
	}

	resolver := NewResolver(cfg)
	resolver.SetWorkDir(tmpDir)

	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// System environment variables should override file variables
	if ctx.Variables["TEST_ENV_VAR"] != "system-value" {
		t.Errorf("expected system value to override file value, got %q", ctx.Variables["TEST_ENV_VAR"])
	}
}

func TestResolver_Resolve_MissingEnvFile(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: ".env.nonexistent",
			},
		},
	}

	resolver := NewResolver(cfg)
	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve should not fail for missing env file, got: %v", err)
	}

	if ctx == nil {
		t.Fatal("expected context to be returned even if env file is missing")
	}
}

func TestResolver_ResolveFromFlags_Default(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {Driver: "digitalocean"},
		},
	}

	resolver := NewResolver(cfg)
	ctx, err := resolver.ResolveFromFlags(context.Background(), "")
	if err != nil {
		t.Fatalf("ResolveFromFlags failed: %v", err)
	}

	if ctx.Name != "dev" {
		t.Errorf("expected default environment 'dev', got %q", ctx.Name)
	}
}

func TestResolver_ResolveFromFlags_WithFlag(t *testing.T) {
	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"staging": {Driver: "digitalocean"},
		},
	}

	resolver := NewResolver(cfg)
	ctx, err := resolver.ResolveFromFlags(context.Background(), "staging")
	if err != nil {
		t.Fatalf("ResolveFromFlags failed: %v", err)
	}

	if ctx.Name != "staging" {
		t.Errorf("expected environment 'staging', got %q", ctx.Name)
	}
}

func TestResolver_InterpolateVariables_Nested(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.test")

	envContent := `BASE_URL=https://api.example.com
API_URL=${BASE_URL}/v1
NESTED=${API_URL}/users
`
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: filepath.Base(envFile),
			},
		},
	}

	resolver := NewResolver(cfg)
	resolver.SetWorkDir(tmpDir)

	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	if ctx.Variables["API_URL"] != "https://api.example.com/v1" {
		t.Errorf("expected API_URL to be interpolated, got %q", ctx.Variables["API_URL"])
	}

	if ctx.Variables["NESTED"] != "https://api.example.com/v1/users" {
		t.Errorf("expected NESTED to be interpolated, got %q", ctx.Variables["NESTED"])
	}
}

func TestResolver_InterpolateVariables_UnknownVar(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env.test")

	envContent := `UNKNOWN_VAR=${NONEXISTENT}/path
`
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &config.Config{
		Project: config.ProjectConfig{Name: "test"},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Driver:  "digitalocean",
				EnvFile: filepath.Base(envFile),
			},
		},
	}

	resolver := NewResolver(cfg)
	resolver.SetWorkDir(tmpDir)

	ctx, err := resolver.Resolve(context.Background(), "dev")
	if err != nil {
		t.Fatalf("Resolve failed: %v", err)
	}

	// Unknown variables should remain as-is
	if ctx.Variables["UNKNOWN_VAR"] != "${NONEXISTENT}/path" {
		t.Errorf("expected unknown variable to remain as-is, got %q", ctx.Variables["UNKNOWN_VAR"])
	}
}

func TestParseEnvFileInto_QuotedValues(t *testing.T) {
	env := make(map[string]string)
	content := `KEY1="value with spaces"
KEY2='single quoted'
KEY3="value with \"quotes\""
KEY4="value\nwith\ttabs"
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value with spaces" {
		t.Errorf("expected KEY1='value with spaces', got %q", env["KEY1"])
	}

	if env["KEY2"] != "single quoted" {
		t.Errorf("expected KEY2='single quoted', got %q", env["KEY2"])
	}

	if env["KEY3"] != "value with \"quotes\"" {
		t.Errorf("expected KEY3 to handle escaped quotes, got %q", env["KEY3"])
	}

	if env["KEY4"] != "value\nwith\ttabs" {
		t.Errorf("expected KEY4 to handle escape sequences, got %q", env["KEY4"])
	}
}

func TestParseEnvFileInto_Comments(t *testing.T) {
	env := make(map[string]string)
	content := `# Full line comment
KEY1=value1 # Inline comment
KEY2="value # not a comment"
KEY3=value3
`
	parseEnvFileInto(env, []byte(content))

	if _, ok := env["#"]; ok {
		t.Error("full line comments should not create variables")
	}

	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1' (inline comment removed), got %q", env["KEY1"])
	}

	if env["KEY2"] != "value # not a comment" {
		t.Errorf("expected KEY2 to preserve # in quoted string, got %q", env["KEY2"])
	}

	if env["KEY3"] != "value3" {
		t.Errorf("expected KEY3='value3', got %q", env["KEY3"])
	}
}

func TestParseEnvFileInto_ExportKeyword(t *testing.T) {
	env := make(map[string]string)
	content := `export KEY1=value1
KEY2=value2
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1', got %q", env["KEY1"])
	}

	if env["KEY2"] != "value2" {
		t.Errorf("expected KEY2='value2', got %q", env["KEY2"])
	}
}

func TestParseEnvFileInto_EmptyValues(t *testing.T) {
	env := make(map[string]string)
	content := `EMPTY1=
EMPTY2=""
KEY=value
`
	parseEnvFileInto(env, []byte(content))

	if env["EMPTY1"] != "" {
		t.Errorf("expected EMPTY1='', got %q", env["EMPTY1"])
	}

	if env["EMPTY2"] != "" {
		t.Errorf("expected EMPTY2='', got %q", env["EMPTY2"])
	}

	if env["KEY"] != "value" {
		t.Errorf("expected KEY='value', got %q", env["KEY"])
	}
}

func TestParseEnvFileInto_Whitespace(t *testing.T) {
	env := make(map[string]string)
	content := `KEY1 = value1
KEY2=" value2 "
KEY3 = " value3 "
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1', got %q", env["KEY1"])
	}

	if env["KEY2"] != " value2 " {
		t.Errorf("expected KEY2=' value2 ', got %q", env["KEY2"])
	}

	if env["KEY3"] != " value3 " {
		t.Errorf("expected KEY3=' value3 ', got %q", env["KEY3"])
	}
}
