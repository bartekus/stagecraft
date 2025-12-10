// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package bootstrap

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"stagecraft/pkg/executil"
)

// fakeRunner is a test implementation of executil.Runner for testing SSHExecutor.
type fakeRunner struct {
	cmd      executil.Command
	result   *executil.Result
	err      error
	runCalls []executil.Command
}

func (f *fakeRunner) Run(ctx context.Context, cmd executil.Command) (*executil.Result, error) {
	f.runCalls = append(f.runCalls, cmd)
	f.cmd = cmd
	if f.err != nil {
		return f.result, f.err
	}
	return f.result, nil
}

func (f *fakeRunner) RunStream(ctx context.Context, cmd executil.Command, output io.Writer) error {
	return fmt.Errorf("RunStream not implemented in fakeRunner")
}

func TestSSHExecutor_Run_Success(t *testing.T) {
	fr := &fakeRunner{
		result: &executil.Result{
			ExitCode: 0,
			Stdout:   []byte("Docker version 24.0.0"),
			Stderr:   []byte(""),
		},
		err: nil,
	}

	exec := NewSSHExecutor("root", fr)
	host := Host{
		ID:       "host-1",
		Name:     "app-1",
		PublicIP: "192.0.2.1",
	}

	stdout, stderr, err := exec.Run(context.Background(), host, "docker version")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if stdout != "Docker version 24.0.0" {
		t.Errorf("expected stdout 'Docker version 24.0.0', got %q", stdout)
	}

	if stderr != "" {
		t.Errorf("expected empty stderr, got %q", stderr)
	}

	// Verify SSH command was built correctly
	if len(fr.runCalls) != 1 {
		t.Fatalf("expected 1 Run call, got %d", len(fr.runCalls))
	}

	cmd := fr.runCalls[0]
	if cmd.Name != "ssh" {
		t.Errorf("expected command name 'ssh', got %q", cmd.Name)
	}

	argsStr := strings.Join(cmd.Args, " ")
	if !strings.Contains(argsStr, "root@192.0.2.1") {
		t.Errorf("expected ssh target 'root@192.0.2.1' in args, got %q", argsStr)
	}
	if !strings.Contains(argsStr, "docker version") {
		t.Errorf("expected remote command 'docker version' in args, got %q", argsStr)
	}
	if !strings.Contains(argsStr, "-o BatchMode=yes") {
		t.Errorf("expected BatchMode option in args, got %q", argsStr)
	}
	if !strings.Contains(argsStr, "-o StrictHostKeyChecking=no") {
		t.Errorf("expected StrictHostKeyChecking option in args, got %q", argsStr)
	}
}

func TestSSHExecutor_Run_ErrorWrapped(t *testing.T) {
	fr := &fakeRunner{
		result: &executil.Result{
			ExitCode: 255,
			Stdout:   []byte(""),
			Stderr:   []byte("Permission denied (publickey)"),
		},
		err: fmt.Errorf("command failed with exit code 255"),
	}

	exec := NewSSHExecutor("ubuntu", fr)
	host := Host{
		ID:       "host-2",
		Name:     "app-2",
		PublicIP: "198.51.100.10",
	}

	_, stderr, err := exec.Run(context.Background(), host, "docker ps")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "ssh to ubuntu@198.51.100.10 failed") {
		t.Errorf("expected host info in error, got %q", err.Error())
	}

	if stderr != "Permission denied (publickey)" {
		t.Errorf("expected stderr 'Permission denied (publickey)', got %q", stderr)
	}
}

func TestSSHExecutor_Run_MissingPublicIP(t *testing.T) {
	exec := NewSSHExecutor("root", &fakeRunner{})

	host := Host{
		ID:   "host-3",
		Name: "app-3",
		// PublicIP empty
	}

	_, _, err := exec.Run(context.Background(), host, "docker ps")
	if err == nil {
		t.Fatalf("expected error for missing PublicIP, got nil")
	}

	if !strings.Contains(err.Error(), "missing PublicIP") {
		t.Errorf("expected error to mention missing PublicIP, got %q", err.Error())
	}

	if !strings.Contains(err.Error(), "host-3") {
		t.Errorf("expected error to mention host ID, got %q", err.Error())
	}
}

func TestSSHExecutor_Run_DefaultUser(t *testing.T) {
	fr := &fakeRunner{
		result: &executil.Result{
			ExitCode: 0,
			Stdout:   []byte("ok"),
			Stderr:   []byte(""),
		},
		err: nil,
	}

	// Create executor with empty SSH user (should default to "root")
	exec := NewSSHExecutor("", fr)
	host := Host{
		ID:       "host-4",
		Name:     "app-4",
		PublicIP: "203.0.113.5",
	}

	_, _, err := exec.Run(context.Background(), host, "whoami")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify default user was used
	cmd := fr.runCalls[0]
	argsStr := strings.Join(cmd.Args, " ")
	if !strings.Contains(argsStr, "root@203.0.113.5") {
		t.Errorf("expected default user 'root' in ssh target, got %q", argsStr)
	}
}

func TestSSHExecutor_Run_NonZeroExitCode(t *testing.T) {
	fr := &fakeRunner{
		result: &executil.Result{
			ExitCode: 1,
			Stdout:   []byte(""),
			Stderr:   []byte("docker: command not found"),
		},
		err: fmt.Errorf("command failed with exit code 1"),
	}

	exec := NewSSHExecutor("root", fr)
	host := Host{
		ID:       "host-5",
		Name:     "app-5",
		PublicIP: "192.0.2.5",
	}

	_, stderr, err := exec.Run(context.Background(), host, "docker version")
	if err == nil {
		t.Fatalf("expected error for non-zero exit code, got nil")
	}

	if !strings.Contains(err.Error(), "ssh to root@192.0.2.5 failed") {
		t.Errorf("expected host info in error, got %q", err.Error())
	}

	if stderr != "docker: command not found" {
		t.Errorf("expected stderr 'docker: command not found', got %q", stderr)
	}
}

func TestSSHExecutor_NewSSHExecutor_DefaultRunner(t *testing.T) {
	// Test that nil runner creates a default executil.Runner
	exec := NewSSHExecutor("root", nil)
	if exec == nil {
		t.Fatalf("expected non-nil executor")
	}

	if exec.sshUser != "root" {
		t.Errorf("expected sshUser 'root', got %q", exec.sshUser)
	}

	if exec.runner == nil {
		t.Fatalf("expected non-nil runner (should create default)")
	}
}
