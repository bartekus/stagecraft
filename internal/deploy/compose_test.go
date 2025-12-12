// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

// Feature: DEPLOY_COMPOSE_GEN
// Spec: spec/deploy/compose-gen.md
package deploy

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stagecraft/pkg/config"
)

func TestComposeGenerator_DeterministicOutput(t *testing.T) {
	tmpDir := t.TempDir()
	baseComposePath := filepath.Join(tmpDir, "docker-compose.yml")

	composeContent := `version: "3.9"
services:
  api:
    image: myapp:latest
`
	if err := os.WriteFile(baseComposePath, []byte(composeContent), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"staging": {Driver: "local"},
		},
	}

	generator := NewComposeGenerator()

	// Generate same compose file twice and verify identical bytes + hash
	workdir1 := t.TempDir()
	path1, hash1, err := generator.Generate(cfg, "staging", baseComposePath, "myapp:v1.0.0", workdir1)
	if err != nil {
		t.Fatalf("First generation failed: %v", err)
	}

	workdir2 := t.TempDir()
	path2, hash2, err := generator.Generate(cfg, "staging", baseComposePath, "myapp:v1.0.0", workdir2)
	if err != nil {
		t.Fatalf("Second generation failed: %v", err)
	}

	// Read both files
	// #nosec G304 // path is test-controlled under TempDir.
	bytes1, err := os.ReadFile(path1)
	if err != nil {
		t.Fatalf("Failed to read first file: %v", err)
	}

	// #nosec G304 // path is test-controlled under TempDir.
	bytes2, err := os.ReadFile(path2)
	if err != nil {
		t.Fatalf("Failed to read second file: %v", err)
	}

	// Verify identical bytes
	if !bytes.Equal(bytes1, bytes2) {
		t.Errorf("Generated compose files differ\nFirst:\n%s\nSecond:\n%s",
			string(bytes1), string(bytes2))
	}

	// Verify identical hash
	if hash1 != hash2 {
		t.Errorf("Hashes differ: %q != %q", hash1, hash2)
	}
}

func TestComposeGenerator_PreservesVolumesNetworksEtc(t *testing.T) {
	tmpDir := t.TempDir()
	baseComposePath := filepath.Join(tmpDir, "docker-compose.yml")

	// Compose file with volumes, networks, configs, secrets, and service using named volume
	composeContent := `version: "3.9"
services:
  api:
    image: myapp:latest
    volumes:
      - app_data:/data
volumes:
  app_data:
    driver: local
networks:
  default:
    driver: bridge
configs:
  app_config:
    external: true
secrets:
  db_password:
    external: true
x-custom-extension:
  foo: bar
`
	if err := os.WriteFile(baseComposePath, []byte(composeContent), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"staging": {Driver: "local"},
		},
	}

	generator := NewComposeGenerator()
	outputPath, _, err := generator.Generate(
		cfg,
		"staging",
		baseComposePath,
		"myapp:v1.0.0",
		tmpDir,
	)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Read generated file
	// #nosec G304 // path is test-controlled under TempDir.
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	outputStr := string(outputBytes)

	// Verify volumes section exists
	if !strings.Contains(outputStr, "volumes:") {
		t.Error("Generated compose file missing volumes section")
	}
	if !strings.Contains(outputStr, "app_data:") {
		t.Error("Generated compose file missing app_data volume")
	}

	// Verify networks section exists
	if !strings.Contains(outputStr, "networks:") {
		t.Error("Generated compose file missing networks section")
	}

	// Verify configs section exists
	if !strings.Contains(outputStr, "configs:") {
		t.Error("Generated compose file missing configs section")
	}
	if !strings.Contains(outputStr, "app_config:") {
		t.Error("Generated compose file missing app_config")
	}

	// Verify secrets section exists
	if !strings.Contains(outputStr, "secrets:") {
		t.Error("Generated compose file missing secrets section")
	}
	if !strings.Contains(outputStr, "db_password:") {
		t.Error("Generated compose file missing db_password secret")
	}

	// Verify service still references named volume
	if !strings.Contains(outputStr, "app_data:/data") {
		t.Error("Generated compose file missing volume reference in service")
	}

	// Verify image was injected
	if !strings.Contains(outputStr, "myapp:v1.0.0") {
		t.Error("Generated compose file missing injected image tag")
	}

	// Verify x-* extension preserved
	if !strings.Contains(outputStr, "x-custom-extension:") {
		t.Error("Generated compose file missing x-* extension")
	}

	// Verify single document (no multi-doc YAML)
	if strings.Contains(outputStr, "\n---\n") {
		t.Error("Generated compose file contains multi-document separator (---)")
	}
}

func TestComposeGenerator_ImageTagInjection(t *testing.T) {
	tmpDir := t.TempDir()
	baseComposePath := filepath.Join(tmpDir, "docker-compose.yml")

	composeContent := `version: "3.9"
services:
  api:
    build:
      context: .
    image: old:tag
  worker:
    build:
      context: .
`
	if err := os.WriteFile(baseComposePath, []byte(composeContent), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"staging": {Driver: "local"},
		},
	}

	generator := NewComposeGenerator()
	outputPath, _, err := generator.Generate(
		cfg,
		"staging",
		baseComposePath,
		"myapp:v1.0.0",
		tmpDir,
	)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// #nosec G304 // path is test-controlled under TempDir.
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	outputStr := string(outputBytes)

	// Verify both services have the injected image tag
	if !strings.Contains(outputStr, "myapp:v1.0.0") {
		t.Error("Generated compose file missing injected image tag")
	}

	// Verify old tag is replaced
	if strings.Contains(outputStr, "old:tag") {
		t.Error("Generated compose file still contains old image tag")
	}
}

func TestComposeGenerator_EnvFileMerging(t *testing.T) {
	tmpDir := t.TempDir()
	baseComposePath := filepath.Join(tmpDir, "docker-compose.yml")
	envFilePath := filepath.Join(tmpDir, ".env.staging")

	composeContent := `version: "3.9"
services:
  api:
    image: myapp:latest
    environment:
      EXISTING_VAR: existing_value
`
	if err := os.WriteFile(baseComposePath, []byte(composeContent), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	envFileContent := `NEW_VAR=new_value
EXISTING_VAR=should_not_override
`
	if err := os.WriteFile(envFilePath, []byte(envFileContent), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"staging": {
				Driver:  "local",
				EnvFile: ".env.staging",
			},
		},
	}

	generator := NewComposeGenerator()
	outputPath, _, err := generator.Generate(
		cfg,
		"staging",
		baseComposePath,
		"myapp:v1.0.0",
		tmpDir,
	)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// #nosec G304 // path is test-controlled under TempDir.
	outputBytes, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	outputStr := string(outputBytes)

	// Verify new var is added
	if !strings.Contains(outputStr, "NEW_VAR") {
		t.Error("Generated compose file missing env_file variable")
	}

	// Verify existing var is preserved (not overridden)
	if !strings.Contains(outputStr, "EXISTING_VAR: existing_value") {
		t.Error("Generated compose file missing existing environment variable")
	}

	// Verify env_file value doesn't override existing
	if strings.Contains(outputStr, "EXISTING_VAR: should_not_override") {
		t.Error("Generated compose file incorrectly overrode existing environment variable")
	}
}

func TestComposeGenerator_MissingEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	baseComposePath := filepath.Join(tmpDir, "docker-compose.yml")

	composeContent := `version: "3.9"
services:
  api:
    image: myapp:latest
`
	if err := os.WriteFile(baseComposePath, []byte(composeContent), 0o600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	cfg := &config.Config{
		Environments: map[string]config.EnvironmentConfig{
			"staging": {
				Driver:  "local",
				EnvFile: ".env.missing", // File doesn't exist
			},
		},
	}

	generator := NewComposeGenerator()
	_, _, err := generator.Generate(
		cfg,
		"staging",
		baseComposePath,
		"myapp:v1.0.0",
		tmpDir,
	)
	// Should not fail on missing env file
	if err != nil {
		t.Fatalf("Generate should not fail on missing env file: %v", err)
	}
}
