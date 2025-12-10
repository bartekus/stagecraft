// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Feature: DEV_PROCESS_MGMT
Specs: spec/dev/process-mgmt.md
Docs:
  - docs/engine/analysis/DEV_PROCESS_MGMT.md
  - docs/engine/outlines/DEV_PROCESS_MGMT_IMPLEMENTATION_OUTLINE.md
*/

// Package process contains dev process lifecycle management for CLI_DEV.
package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Options captures process management settings.
type Options struct {
	DevDir    string
	NoTraefik bool
	Detach    bool
	Verbose   bool
}

// Writer is the minimal writer abstraction used by Command.
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Command abstracts running a command.
//
// It is intentionally small so tests can provide fakes without depending
// on *exec.Cmd directly.
type Command interface {
	Run() error
	Start() error
	Wait() error
	SetStdout(w Writer)
	SetStderr(w Writer)
}

// ExecCommander abstracts command construction for testability.
type ExecCommander interface {
	CommandContext(ctx context.Context, name string, args ...string) Command
}

// Logger is a minimal logging abstraction.
//
// Default implementation writes to stderr without timestamps to preserve
// determinism of log formatting.
type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

// Runner manages the dev process lifecycle.
type Runner struct {
	exec ExecCommander
	log  Logger
}

// NewRunner constructs a Runner with default exec and logger.
func NewRunner() *Runner {
	return &Runner{
		exec: defaultExecCommander{},
		log:  defaultLogger{},
	}
}

// NewRunnerWithDeps constructs a Runner with explicit dependencies.
//
// This is primarily used for tests.
func NewRunnerWithDeps(execCmd ExecCommander, logger Logger) *Runner {
	if execCmd == nil {
		execCmd = defaultExecCommander{}
	}
	if logger == nil {
		logger = defaultLogger{}
	}
	return &Runner{
		exec: execCmd,
		log:  logger,
	}
}

// Run starts the dev stack using the dev compose file and handles lifecycle
// according to Options.
//
// In foreground mode, it blocks until context is cancelled or the compose
// command exits. On context cancellation it attempts a deterministic teardown
// via `docker compose down`.
//
// In detached mode, it runs `docker compose up -d` and returns when the
// command completes.
func (r *Runner) Run(ctx context.Context, opts Options) error {
	if opts.DevDir == "" {
		return fmt.Errorf("dev: dev dir must not be empty")
	}

	composePath := filepath.Join(opts.DevDir, "compose.yaml")

	if _, err := os.Stat(composePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("dev: compose file not found at %s", composePath)
		}
		return fmt.Errorf("dev: stat compose file at %s: %w", composePath, err)
	}

	if opts.Verbose {
		r.log.Infof("dev: using dev dir %s", opts.DevDir)
		r.log.Infof("dev: using compose file %s", composePath)
	}

	if opts.Detach {
		return r.runDetached(ctx, composePath, opts)
	}

	return r.runForeground(ctx, composePath, opts)
}

// runDetached runs `docker compose up -d [...]` and returns when the command
// completes.
func (r *Runner) runDetached(ctx context.Context, composePath string, opts Options) error {
	args := buildUpArgs(composePath, opts, true)

	if opts.Verbose {
		r.log.Infof("dev: running (detached): docker %s", strings.Join(args, " "))
	}

	cmd := r.exec.CommandContext(ctx, "docker", args...)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dev: docker compose up -d failed: %w", err)
	}

	r.log.Infof("dev: dev stack started in background")
	return nil
}

// runForeground runs `docker compose up [...]` attached and tears down the
// stack with `docker compose down` when the context is cancelled.
func (r *Runner) runForeground(ctx context.Context, composePath string, opts Options) error {
	args := buildUpArgs(composePath, opts, false)

	if opts.Verbose {
		r.log.Infof("dev: running (foreground): docker %s", strings.Join(args, " "))
	}

	cmd := r.exec.CommandContext(ctx, "docker", args...)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("dev: docker compose up failed to start: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		// User initiated interruption - tear down the stack.
		r.log.Infof("dev: context cancelled; tearing down dev stack with docker compose down")

		downErr := r.runDown(composePath, opts)
		if downErr != nil {
			r.log.Errorf("dev: teardown failed: %v", downErr)
		}

		// Do not mask the user initiated cancellation.
		return ctx.Err()

	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("dev: docker compose up failed: %w", err)
		}
		return nil
	}
}

// runDown issues `docker compose -f <path> down` using a background context.
//
// It is best-effort; errors are returned so the caller can decide how
// to handle them.
func (r *Runner) runDown(composePath string, opts Options) error {
	args := []string{
		"compose",
		"-f", composePath,
		"down",
	}

	if opts.Verbose {
		r.log.Infof("dev: running teardown: docker %s", strings.Join(args, " "))
	}

	ctx := context.Background()
	cmd := r.exec.CommandContext(ctx, "docker", args...)
	// For teardown, stdout is usually not critical, but we still wire it
	// to stderr/stdout for transparency.
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dev: docker compose down failed: %w", err)
	}

	return nil
}

// buildUpArgs builds arguments for `docker compose up`, with optional -d
// and Traefik scaling based on options.
func buildUpArgs(composePath string, opts Options, detach bool) []string {
	args := []string{
		"compose",
		"-f", composePath,
		"up",
	}

	if detach {
		args = append(args, "-d")
	}

	if opts.NoTraefik {
		args = append(args, "--scale", "traefik=0")
	}

	return args
}

// defaultExecCommander is the production ExecCommander backed by os/exec.
type defaultExecCommander struct{}

// CommandContext implements ExecCommander using exec.CommandContext.
func (defaultExecCommander) CommandContext(ctx context.Context, name string, args ...string) Command {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // name and args are fully controlled
	return &cmdWrapper{cmd: cmd}
}

// cmdWrapper adapts *exec.Cmd to the Command interface.
type cmdWrapper struct {
	cmd *exec.Cmd
}

func (c *cmdWrapper) Run() error {
	return c.cmd.Run()
}

func (c *cmdWrapper) Start() error {
	return c.cmd.Start()
}

func (c *cmdWrapper) Wait() error {
	return c.cmd.Wait()
}

func (c *cmdWrapper) SetStdout(w Writer) {
	c.cmd.Stdout = w
}

func (c *cmdWrapper) SetStderr(w Writer) {
	c.cmd.Stderr = w
}

// defaultLogger writes to stderr without timestamps.
type defaultLogger struct{}

func (defaultLogger) Infof(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (defaultLogger) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}
