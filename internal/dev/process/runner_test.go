// SPDX-License-Identifier: AGPL-3.0-or-later

package process

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeLogger captures log lines for assertions.
type fakeLogger struct {
	infos  []string
	errors []string
}

func (l *fakeLogger) Infof(format string, args ...any) {
	l.infos = append(l.infos, sprintf(format, args...))
}

func (l *fakeLogger) Errorf(format string, args ...any) {
	l.errors = append(l.errors, sprintf(format, args...))
}

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}

// fakeCommand implements Command for tests.
type fakeCommand struct {
	runErr   error
	startErr error
	waitErr  error

	stdout Writer
	stderr Writer

	startCalled bool
	waitCalled  bool
	runCalled   bool
}

func (c *fakeCommand) Run() error {
	c.runCalled = true
	return c.runErr
}

func (c *fakeCommand) Start() error {
	c.startCalled = true
	return c.startErr
}

func (c *fakeCommand) Wait() error {
	c.waitCalled = true
	return c.waitErr
}

func (c *fakeCommand) SetStdout(w Writer) {
	c.stdout = w
}

func (c *fakeCommand) SetStderr(w Writer) {
	c.stderr = w
}

// fakeExecCommander records the last command created.
type fakeExecCommander struct {
	lastName string
	lastArgs []string

	cmd *fakeCommand
}

func (f *fakeExecCommander) CommandContext(_ context.Context, name string, args ...string) Command {
	f.lastName = name
	f.lastArgs = append([]string(nil), args...)
	if f.cmd == nil {
		f.cmd = &fakeCommand{}
	}
	return f.cmd
}

func TestRunner_DetachedBuildsExpectedCommand(t *testing.T) {
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "compose.yaml")

	// #nosec G306 -- test file permissions
	if err := os.WriteFile(composePath, []byte("version: '3.8'\n"), 0o644); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	execFake := &fakeExecCommander{}
	logFake := &fakeLogger{}

	r := NewRunnerWithDeps(execFake, logFake)

	opts := Options{
		DevDir:    tmpDir,
		NoTraefik: true,
		Detach:    true,
		Verbose:   true,
	}

	err := r.Run(context.Background(), opts)
	if err != nil && !errors.Is(err, context.Canceled) && execFake.cmd.runErr == nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if execFake.lastName != "docker" {
		t.Fatalf("expected command 'docker', got %q", execFake.lastName)
	}

	got := strings.Join(execFake.lastArgs, " ")

	// Expected parts:
	expectParts := []string{
		"compose",
		"-f", composePath,
		"up",
		"-d",
		"--scale", "traefik=0",
	}

	for _, part := range expectParts {
		if !strings.Contains(got, part) {
			t.Errorf("expected args to contain %q, got %q", part, got)
		}
	}

	if !execFake.cmd.runCalled {
		t.Errorf("expected Run to be called in detached mode")
	}
}

func TestRunner_ForegroundBuildsExpectedCommand(t *testing.T) {
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "compose.yaml")

	// #nosec G306 -- test file permissions
	if err := os.WriteFile(composePath, []byte("version: '3.8'\n"), 0o644); err != nil {
		t.Fatalf("write compose file: %v", err)
	}

	execFake := &fakeExecCommander{
		cmd: &fakeCommand{
			waitErr: nil, // foreground completes successfully
		},
	}
	logFake := &fakeLogger{}

	r := NewRunnerWithDeps(execFake, logFake)

	opts := Options{
		DevDir:    tmpDir,
		NoTraefik: false,
		Detach:    false,
		Verbose:   true,
	}

	if err := r.Run(context.Background(), opts); err != nil {
		t.Fatalf("Run returned unexpected error: %v", err)
	}

	if execFake.lastName != "docker" {
		t.Fatalf("expected command 'docker', got %q", execFake.lastName)
	}

	got := strings.Join(execFake.lastArgs, " ")

	expectParts := []string{
		"compose",
		"-f", composePath,
		"up",
	}

	for _, part := range expectParts {
		if !strings.Contains(got, part) {
			t.Errorf("expected args to contain %q, got %q", part, got)
		}
	}

	if !execFake.cmd.startCalled || !execFake.cmd.waitCalled {
		t.Errorf("expected Start and Wait to be called in foreground mode")
	}
}

func TestRunner_MissingComposeFileFails(t *testing.T) {
	tmpDir := t.TempDir()

	execFake := &fakeExecCommander{}
	logFake := &fakeLogger{}

	r := NewRunnerWithDeps(execFake, logFake)

	opts := Options{
		DevDir:    tmpDir,
		NoTraefik: false,
		Detach:    false,
		Verbose:   false,
	}

	err := r.Run(context.Background(), opts)
	if err == nil {
		t.Fatal("expected error for missing compose file, got nil")
	}

	if !strings.Contains(err.Error(), "compose file not found") {
		t.Errorf("expected error about missing compose file, got %v", err)
	}
}

func TestRunner_EmptyDevDirFails(t *testing.T) {
	r := NewRunner()

	err := r.Run(context.Background(), Options{})
	if err == nil {
		t.Fatal("expected error for empty dev dir, got nil")
	}

	if !strings.Contains(err.Error(), "dev dir must not be empty") {
		t.Errorf("expected error about dev dir, got %v", err)
	}
}
