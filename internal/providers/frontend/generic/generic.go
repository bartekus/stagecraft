// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package generic provides a generic command-based frontend provider implementation.
package generic

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/providers/frontend"
)

// Feature: PROVIDER_FRONTEND_GENERIC
// Spec: spec/providers/frontend/generic.md

// GenericProvider implements a command-based frontend provider.
//
//nolint:revive // GenericProvider is the preferred name for clarity
type GenericProvider struct{}

// Ensure GenericProvider implements FrontendProvider
var _ frontend.FrontendProvider = (*GenericProvider)(nil)

// ID returns the provider identifier.
func (p *GenericProvider) ID() string {
	return "generic"
}

// Config represents the generic provider configuration.
type Config struct {
	Dev struct {
		Command      []string          `yaml:"command"`
		WorkDir      string            `yaml:"workdir"`
		Env          map[string]string `yaml:"env"`
		ReadyPattern string            `yaml:"ready_pattern"`
		Shutdown     struct {
			Signal    string `yaml:"signal"`
			TimeoutMS int    `yaml:"timeout_ms"`
		} `yaml:"shutdown"`
	} `yaml:"dev"`
}

// Dev runs the frontend in development mode.
func (p *GenericProvider) Dev(ctx context.Context, opts frontend.DevOptions) error {
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

	// Merge provider env with opts.Env (opts.Env takes precedence)
	env := make(map[string]string)
	for k, v := range cfg.Dev.Env {
		env[k] = v
	}
	for k, v := range opts.Env {
		env[k] = v
	}

	// Build command
	//nolint:gosec // commands and args are trusted operator config from stagecraft.yml, not user input
	cmd := exec.CommandContext(ctx, cfg.Dev.Command[0], cfg.Dev.Command[1:]...)
	cmd.Dir = workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// If ready_pattern is specified, we need to capture output and watch for it
	if cfg.Dev.ReadyPattern != "" {
		return p.runWithReadyPattern(ctx, cmd, cfg.Dev.ReadyPattern, cfg.Dev.Shutdown)
	}

	// Otherwise, just stream output directly
	// TODO: Consider using structured logging instead of direct stdout/stderr writes
	// per Agent.md guidance. For v1, direct streaming is acceptable for dev-only provider.
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Handle graceful shutdown
	return p.runWithShutdown(ctx, cmd, cfg.Dev.Shutdown)
}

// runWithReadyPattern runs the command and watches for a ready pattern.
func (p *GenericProvider) runWithReadyPattern(ctx context.Context, cmd *exec.Cmd, pattern string, shutdownCfg struct {
	Signal    string `yaml:"signal"`
	TimeoutMS int    `yaml:"timeout_ms"`
},
) error {
	// Compile regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid ready_pattern regex: %w", err)
	}

	// Create pipes for stdout and stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating stdout pipe: %w", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("creating stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting command: %w", err)
	}

	// Channel to signal when ready pattern is found
	readyCh := make(chan bool, 1)
	errCh := make(chan error, 1)
	var readyOnce sync.Once

	// Monitor stdout
	// TODO: Consider using structured logging instead of direct os.Stdout writes
	// per Agent.md guidance. For v1, direct streaming is acceptable for dev-only provider.
	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			line := scanner.Text()
			_, _ = os.Stdout.WriteString(line + "\n")
			if re.MatchString(line) {
				readyOnce.Do(func() {
					select {
					case readyCh <- true:
					default:
					}
				})
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("reading stdout: %w", err)
		}
	}()

	// Monitor stderr
	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			_, _ = os.Stderr.WriteString(line + "\n")
			if re.MatchString(line) {
				readyOnce.Do(func() {
					select {
					case readyCh <- true:
					default:
					}
				})
			}
		}
		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("reading stderr: %w", err)
		}
	}()

	// Wait for either ready pattern, context cancellation, or process exit
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	// Wait for ready pattern or process exit
	select {
	case <-readyCh:
		// Pattern found, continue streaming and wait for process exit or context cancellation
		select {
		case <-ctx.Done():
			return p.shutdownProcess(cmd, shutdownCfg)
		case err := <-doneCh:
			if err != nil {
				return fmt.Errorf("process exited: %w", err)
			}
			return nil
		}
	case err := <-errCh:
		// Error reading output
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return err
	case <-ctx.Done():
		// Context cancelled
		return p.shutdownProcess(cmd, shutdownCfg)
	case err := <-doneCh:
		// Process exited before ready pattern found
		if err != nil {
			return fmt.Errorf("process exited before ready pattern found: %w", err)
		}
		return fmt.Errorf("process exited before ready pattern found")
	}
}

// runWithShutdown runs the command with graceful shutdown handling.
func (p *GenericProvider) runWithShutdown(ctx context.Context, cmd *exec.Cmd, shutdownCfg struct {
	Signal    string `yaml:"signal"`
	TimeoutMS int    `yaml:"timeout_ms"`
},
) error {
	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting command: %w", err)
	}

	// Wait for either context cancellation or process completion
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		return p.shutdownProcess(cmd, shutdownCfg)
	case err := <-doneCh:
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return fmt.Errorf("command failed with exit code %d: %w", exitErr.ExitCode(), err)
			}
			return fmt.Errorf("command failed: %w", err)
		}
		return nil
	}
}

// shutdownProcess gracefully shuts down the process.
func (p *GenericProvider) shutdownProcess(cmd *exec.Cmd, shutdownCfg struct {
	Signal    string `yaml:"signal"`
	TimeoutMS int    `yaml:"timeout_ms"`
},
) error {
	if cmd.Process == nil {
		return nil
	}

	// Determine signal (default: SIGINT)
	sig := syscall.SIGINT
	if shutdownCfg.Signal != "" {
		switch strings.ToUpper(shutdownCfg.Signal) {
		case "SIGINT":
			sig = syscall.SIGINT
		case "SIGTERM":
			sig = syscall.SIGTERM
		case "SIGKILL":
			sig = syscall.SIGKILL
		default:
			// Try to parse as signal number or name
			// For now, default to SIGINT if unknown
			sig = syscall.SIGINT
		}
	}

	// Send signal
	if err := cmd.Process.Signal(sig); err != nil {
		// Process may have already exited
		if err.Error() == "os: process already finished" {
			return nil
		}
		return fmt.Errorf("sending shutdown signal: %w", err)
	}

	// Determine timeout (default: 10 seconds)
	timeout := 10 * time.Second
	if shutdownCfg.TimeoutMS > 0 {
		timeout = time.Duration(shutdownCfg.TimeoutMS) * time.Millisecond
	}

	// Wait for graceful shutdown or timeout
	doneCh := make(chan error, 1)
	go func() {
		doneCh <- cmd.Wait()
	}()

	select {
	case <-doneCh:
		// Process exited gracefully
		return nil
	case <-time.After(timeout):
		// Timeout reached, force kill
		if err := cmd.Process.Kill(); err != nil {
			return fmt.Errorf("force killing process: %w", err)
		}
		_ = cmd.Wait() // Clean up
		return fmt.Errorf("process did not exit within %v, force killed", timeout)
	}
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
	frontend.Register(&GenericProvider{})
}
