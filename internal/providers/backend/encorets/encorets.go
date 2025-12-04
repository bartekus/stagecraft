// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package encorets provides the Encore.ts backend provider implementation.
package encorets

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/logging"
	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

// EncoreTsProvider implements the Encore.ts backend provider.
//
//nolint:revive // EncoreTsProvider is the preferred name for clarity
type EncoreTsProvider struct{}

// Ensure EncoreTsProvider implements BackendProvider
var _ backend.BackendProvider = (*EncoreTsProvider)(nil)

// ID returns the provider identifier.
func (p *EncoreTsProvider) ID() string {
	return "encore-ts"
}

// Config represents the Encore.ts provider configuration.
type Config struct {
	Dev struct {
		EnvFile          string `yaml:"env_file"`            // required for dev
		Listen           string `yaml:"listen"`              // required
		WorkDir          string `yaml:"workdir"`             // optional
		EntryPoint       string `yaml:"entrypoint"`          // optional
		DisableTelemetry bool   `yaml:"disable_telemetry"`   // optional
		NodeExtraCACerts string `yaml:"node_extra_ca_certs"` // optional
		EncoreSecrets    struct {
			Types   []string `yaml:"types"`
			FromEnv []string `yaml:"from_env"`
		} `yaml:"encore_secrets"`
	} `yaml:"dev"`

	Build struct {
		WorkDir         string `yaml:"workdir"`           // optional
		ImageName       string `yaml:"image_name"`        // optional; default "api"
		DockerTagSuffix string `yaml:"docker_tag_suffix"` // optional
	} `yaml:"build"`
}

// Dev runs the Encore.ts backend in development mode.
func (p *EncoreTsProvider) Dev(ctx context.Context, opts backend.DevOptions) error {
	// Check if encore is available
	if err := p.checkEncoreAvailable(); err != nil {
		return err
	}

	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return err
	}

	if err := p.validateDevConfig(cfg); err != nil {
		return err
	}

	// Initialize logger
	// Note: verbose flag would come from opts if available, for now default to false
	logger := logging.NewLogger(false)
	logger = logger.WithFields(
		logging.NewField("provider", "encore-ts"),
		logging.NewField("operation", "dev"),
		logging.NewField("feature", "PROVIDER_BACKEND_ENCORE"),
	)

	// Resolve workdir
	workDir := cfg.Dev.WorkDir
	if workDir == "" {
		workDir = opts.WorkDir
	}
	if workDir == "" {
		workDir = "."
	}

	// Prepare base environment from opts.Env
	env := make(map[string]string)
	for k, v := range opts.Env {
		env[k] = v
	}

	// Load env_file if specified
	if cfg.Dev.EnvFile != "" {
		envFilePath := cfg.Dev.EnvFile
		if !filepath.IsAbs(envFilePath) {
			// Resolve relative to workDir
			envFilePath = filepath.Join(workDir, envFilePath)
		}

		if _, err := os.Stat(envFilePath); err == nil {
			// File exists, read and parse it
			// Simple dotenv parsing (key=value format)
			//nolint:gosec // G304: envFilePath comes from trusted stagecraft.yml config, not user input
			data, err := os.ReadFile(envFilePath)
			if err != nil {
				logger.Warn("Failed to read env_file",
					logging.NewField("path", envFilePath),
					logging.NewField("error", err.Error()),
				)
			} else {
				// Parse dotenv format
				lines := strings.Split(string(data), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" || strings.HasPrefix(line, "#") {
						continue
					}
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						// Remove quotes if present
						if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
							(value[0] == '\'' && value[len(value)-1] == '\'')) {
							value = value[1 : len(value)-1]
						}
						env[key] = value
					}
				}
			}
		} else {
			logger.Warn("env_file does not exist",
				logging.NewField("path", envFilePath),
			)
		}
	}

	// Apply telemetry and CA certs configuration
	if cfg.Dev.DisableTelemetry {
		env["DISABLE_ENCORE_TELEMETRY"] = "1"
	}

	if cfg.Dev.NodeExtraCACerts != "" {
		caPath := cfg.Dev.NodeExtraCACerts
		if !filepath.IsAbs(caPath) {
			caPath = filepath.Join(workDir, caPath)
		}
		env["NODE_EXTRA_CA_CERTS"] = caPath
	}

	// Sync secrets if configured
	if len(cfg.Dev.EncoreSecrets.FromEnv) > 0 {
		types := cfg.Dev.EncoreSecrets.Types
		if len(types) == 0 {
			types = []string{"dev", "preview", "local"}
		}

		for _, secretName := range cfg.Dev.EncoreSecrets.FromEnv {
			secretValue, exists := env[secretName]
			if !exists || secretValue == "" {
				logger.Warn("Missing environment variable for secret sync",
					logging.NewField("secret_name", secretName),
				)
				continue
			}

			// Sync to each secret type
			for _, secretType := range types {
				args := []string{"secret", "set", "--type", secretType, secretName}

				//nolint:gosec // encore CLI args and secret names are controlled by operator config/env, not end-user input
				cmd := exec.CommandContext(ctx, "encore", args...)
				cmd.Dir = workDir
				cmd.Stdin = strings.NewReader(secretValue)

				// Capture output for error reporting
				var stdout, stderr strings.Builder
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				if err := cmd.Run(); err != nil {
					// Truncate stderr for error detail
					detail := stderr.String()
					if len(detail) > 500 {
						detail = detail[:500] + "..."
					}

					return &ProviderError{
						Category:  ErrSecretSyncFailed,
						Provider:  "encore-ts",
						Operation: "dev",
						Message:   fmt.Sprintf("failed to set encore secret %s for type %s", secretName, secretType),
						Detail:    detail,
						Err:       err,
					}
				}
			}
		}
	}

	// Run encore dev server
	logger.Info("Starting Encore dev server",
		logging.NewField("listen", cfg.Dev.Listen),
		logging.NewField("workdir", workDir),
	)

	args := []string{"run", "--watch", "--listen", cfg.Dev.Listen}
	if cfg.Dev.EntryPoint != "" {
		args = append(args, "--entrypoint", cfg.Dev.EntryPoint)
	}

	//nolint:gosec // encore CLI args come from trusted stagecraft.yml and env
	cmd := exec.CommandContext(ctx, "encore", args...)
	cmd.Dir = workDir

	// Build environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Stream output through logger (use Debug level for command output)
	cmd.Stdout = &logWriter{logger: logger}
	cmd.Stderr = &logWriter{logger: logger}

	if err := cmd.Run(); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Process exited with error
		var exitCode int
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}

		return &ProviderError{
			Category:  ErrDevServerFailed,
			Provider:  "encore-ts",
			Operation: "dev",
			Message:   "encore dev server failed",
			Detail:    fmt.Sprintf("exit code: %d", exitCode),
			Err:       err,
		}
	}

	return nil
}

// logWriter implements io.Writer for streaming command output to logger
type logWriter struct {
	logger logging.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline for cleaner log output
	msg := strings.TrimRight(string(p), "\n\r")
	if msg == "" {
		return len(p), nil
	}

	// Use Debug level for streaming command output
	w.logger.Debug(msg)

	return len(p), nil
}

// BuildDocker builds a Docker image using Encore.
func (p *EncoreTsProvider) BuildDocker(ctx context.Context, opts backend.BuildDockerOptions) (string, error) {
	// Check if encore is available
	if err := p.checkEncoreAvailable(); err != nil {
		return "", err
	}

	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return "", err
	}

	// Initialize logger
	logger := logging.NewLogger(false)
	logger = logger.WithFields(
		logging.NewField("provider", "encore-ts"),
		logging.NewField("operation", "build"),
		logging.NewField("feature", "PROVIDER_BACKEND_ENCORE"),
	)

	// Resolve workdir
	workDir := cfg.Build.WorkDir
	if workDir == "" {
		workDir = opts.WorkDir
	}
	if workDir == "" {
		workDir = "."
	}

	// Resolve image reference
	// If opts.ImageTag already contains a registry (has /), use it as-is
	// Otherwise, construct from build.image_name and opts.ImageTag
	imageRef := opts.ImageTag
	if !strings.Contains(opts.ImageTag, "/") {
		// opts.ImageTag is just the tag part, construct full reference
		imageName := cfg.Build.ImageName
		if imageName == "" {
			imageName = "api"
		}

		// Construct: <image_name>:<tag><suffix>
		tag := opts.ImageTag
		if cfg.Build.DockerTagSuffix != "" {
			tag += cfg.Build.DockerTagSuffix
		}
		imageRef = fmt.Sprintf("%s:%s", imageName, tag)
	} else if cfg.Build.DockerTagSuffix != "" {
		// opts.ImageTag is full reference, but we may need to add suffix
		// Insert suffix before the tag part
		parts := strings.SplitN(imageRef, ":", 2)
		if len(parts) == 2 {
			imageRef = fmt.Sprintf("%s:%s%s", parts[0], parts[1], cfg.Build.DockerTagSuffix)
		}
	}

	logger.Info("Building Encore Docker image",
		logging.NewField("image", imageRef),
		logging.NewField("workdir", workDir),
	)

	// Run encore build docker
	args := []string{"build", "docker", imageRef}

	//nolint:gosec // encore CLI args come from trusted config (image tag)
	cmd := exec.CommandContext(ctx, "encore", args...)
	cmd.Dir = workDir

	// Stream output through logger
	cmd.Stdout = &logWriter{logger: logger}
	cmd.Stderr = &logWriter{logger: logger}

	if err := cmd.Run(); err != nil {
		// Check if context was cancelled
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Build failed
		var exitCode int
		var stderrOutput string
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			// Note: stderr was already streamed to logger, but we could capture it here if needed
		}

		detail := fmt.Sprintf("exit code: %d", exitCode)
		if stderrOutput != "" && len(stderrOutput) > 500 {
			detail += ", " + stderrOutput[:500] + "..."
		}

		return "", &ProviderError{
			Category:  ErrBuildFailed,
			Provider:  "encore-ts",
			Operation: "build",
			Message:   "encore build docker failed",
			Detail:    detail,
			Err:       err,
		}
	}

	logger.Info("Successfully built Docker image",
		logging.NewField("image", imageRef),
	)

	return imageRef, nil
}

// parseConfig unmarshals and validates the provider config.
func (p *EncoreTsProvider) parseConfig(cfg any) (*Config, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, &ProviderError{
			Category:  ErrInvalidConfig,
			Provider:  "encore-ts",
			Operation: "parse",
			Message:   "failed to marshal config",
			Err:       err,
		}
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, &ProviderError{
			Category:  ErrInvalidConfig,
			Provider:  "encore-ts",
			Operation: "parse",
			Message:   "invalid encore-ts provider config",
			Detail:    err.Error(),
			Err:       err,
		}
	}

	// Set defaults
	if config.Build.ImageName == "" {
		config.Build.ImageName = "api"
	}

	return &config, nil
}

// checkEncoreAvailable checks if the encore binary is available.
func (p *EncoreTsProvider) checkEncoreAvailable() error {
	_, err := exec.LookPath("encore")
	if err != nil {
		return &ProviderError{
			Category:  ErrProviderNotAvailable,
			Provider:  "encore-ts",
			Operation: "check",
			Message:   "encore binary not found",
			Detail:    "encore CLI must be installed and available in PATH",
			Err:       err,
		}
	}
	return nil
}

// validateDevConfig validates dev-specific config requirements.
func (p *EncoreTsProvider) validateDevConfig(cfg *Config) error {
	if cfg.Dev.Listen == "" {
		return &ProviderError{
			Category:  ErrInvalidConfig,
			Provider:  "encore-ts",
			Operation: "dev",
			Message:   "dev.listen is required",
		}
	}

	// Note: env_file is required per spec, but we'll check existence when reading
	// If it doesn't exist, we'll log a warning but continue (opts.Env may have values)

	return nil
}

func init() {
	backend.Register(&EncoreTsProvider{})
}
