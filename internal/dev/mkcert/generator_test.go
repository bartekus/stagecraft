// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_MKCERT
// Spec: spec/dev/mkcert.md

package mkcert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"stagecraft/pkg/config"
)

// fakeLogger captures log lines for assertions.
type fakeLogger struct {
	infos  []string
	errors []string
}

func (l *fakeLogger) Infof(format string, args ...any) {
	l.infos = append(l.infos, fmt.Sprintf(format, args...))
}

func (l *fakeLogger) Errorf(format string, args ...any) {
	l.errors = append(l.errors, fmt.Sprintf(format, args...))
}

// fakeCommand implements Command for tests.
type fakeCommand struct {
	runErr error

	stdout Writer
	stderr Writer

	runCalled bool
}

func (c *fakeCommand) Run() error {
	c.runCalled = true
	return c.runErr
}

func (c *fakeCommand) SetStdout(w Writer) {
	c.stdout = w
}

func (c *fakeCommand) SetStderr(w Writer) {
	c.stderr = w
}

// fakeExecCommander records the last command and returns a fakeCommand.
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

func TestEnsureCertificates_DisabledHTTPS_NoOps(t *testing.T) {
	execFake := &fakeExecCommander{}
	logFake := &fakeLogger{}

	gen := NewGeneratorWithDeps(execFake, logFake)

	cfg := &config.Config{} // minimal stub; not used in v1

	certCfg, err := gen.EnsureCertificates(context.Background(), cfg, Options{
		DevDir:      ".stagecraft/dev",
		Domains:     []string{"app.localdev.test", "api.localdev.test"},
		EnableHTTPS: false,
		Verbose:     true,
	})
	if err != nil {
		t.Fatalf("EnsureCertificates returned error for disabled HTTPS: %v", err)
	}

	if certCfg.Enabled {
		t.Fatalf("expected Enabled=false when HTTPS disabled, got true")
	}

	if certCfg.CertDir != "" || certCfg.CertFile != "" || certCfg.KeyFile != "" || len(certCfg.Domains) != 0 {
		t.Fatalf("expected empty cert fields when HTTPS disabled, got %+v", certCfg)
	}

	if execFake.lastName != "" {
		t.Fatalf("expected no mkcert invocation when HTTPS disabled, got command %q", execFake.lastName)
	}
}

func TestEnsureCertificates_ReusesExistingCerts(t *testing.T) {
	tmpDir := t.TempDir()
	certDir := filepath.Join(tmpDir, "certs")

	// #nosec G301 -- test directory permissions
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		t.Fatalf("mkdir certDir: %v", err)
	}

	certPath := filepath.Join(certDir, "dev-local.pem")
	keyPath := filepath.Join(certDir, "dev-local-key.pem")

	// #nosec G306 -- test file permissions
	if err := os.WriteFile(certPath, []byte("dummy-cert"), 0o644); err != nil {
		t.Fatalf("write cert: %v", err)
	}
	if err := os.WriteFile(keyPath, []byte("dummy-key"), 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}

	execFake := &fakeExecCommander{
		cmd: &fakeCommand{
			runErr: errors.New("mkcert should not be called"),
		},
	}
	logFake := &fakeLogger{}

	gen := NewGeneratorWithDeps(execFake, logFake)

	cfg := &config.Config{}

	domains := []string{"api.localdev.test", "app.localdev.test"}

	certCfg, err := gen.EnsureCertificates(context.Background(), cfg, Options{
		DevDir:      tmpDir,
		Domains:     domains,
		EnableHTTPS: true,
		Verbose:     true,
	})
	if err != nil {
		t.Fatalf("EnsureCertificates returned error: %v", err)
	}

	if !certCfg.Enabled {
		t.Fatalf("expected Enabled=true when HTTPS enabled, got false")
	}

	if certCfg.CertDir != certDir {
		t.Fatalf("expected CertDir=%q, got %q", certDir, certCfg.CertDir)
	}

	if certCfg.CertFile != certPath {
		t.Fatalf("expected CertFile=%q, got %q", certPath, certCfg.CertFile)
	}

	if certCfg.KeyFile != keyPath {
		t.Fatalf("expected KeyFile=%q, got %q", keyPath, certCfg.KeyFile)
	}

	// Domains must be deduplicated and sorted.
	wantDomains := append([]string(nil), domains...)
	sort.Strings(wantDomains)

	if len(certCfg.Domains) != len(wantDomains) {
		t.Fatalf("expected %d domains, got %d", len(wantDomains), len(certCfg.Domains))
	}
	for i, d := range wantDomains {
		if certCfg.Domains[i] != d {
			t.Fatalf("expected domain[%d]=%q, got %q", i, d, certCfg.Domains[i])
		}
	}

	if execFake.cmd.runCalled {
		t.Fatalf("expected mkcert not to be called when cert files already exist")
	}
}

func TestEnsureCertificates_MkcertMissing_ReturnsHelpfulError(t *testing.T) {
	tmpDir := t.TempDir()

	execFake := &fakeExecCommander{
		cmd: &fakeCommand{
			runErr: exec.ErrNotFound, // simulate missing binary
		},
	}
	logFake := &fakeLogger{}

	gen := NewGeneratorWithDeps(execFake, logFake)

	cfg := &config.Config{}

	_, err := gen.EnsureCertificates(context.Background(), cfg, Options{
		DevDir:      tmpDir,
		Domains:     []string{"app.localdev.test"},
		EnableHTTPS: true,
		Verbose:     false,
	})
	if err == nil {
		t.Fatal("expected error when mkcert is missing, got nil")
	}

	msg := err.Error()
	if !strings.Contains(msg, "mkcert") || !strings.Contains(msg, "not found") {
		t.Fatalf("expected error message to mention mkcert not found, got %q", msg)
	}
}

func TestEnsureCertificates_InvokesMkcertWithSortedDomains(t *testing.T) {
	tmpDir := t.TempDir()

	execFake := &fakeExecCommander{
		cmd: &fakeCommand{
			runErr: nil,
		},
	}
	logFake := &fakeLogger{}

	gen := NewGeneratorWithDeps(execFake, logFake)

	cfg := &config.Config{}

	domains := []string{
		"api.localdev.test",
		"app.localdev.test",
		"app.localdev.test", // duplicate
	}

	_, err := gen.EnsureCertificates(context.Background(), cfg, Options{
		DevDir:      tmpDir,
		Domains:     domains,
		EnableHTTPS: true,
		Verbose:     true,
	})
	if err != nil {
		t.Fatalf("EnsureCertificates returned error: %v", err)
	}

	if execFake.lastName == "" {
		t.Fatal("expected mkcert to be invoked")
	}
	if execFake.lastName != "mkcert" {
		t.Fatalf("expected command 'mkcert', got %q", execFake.lastName)
	}

	gotArgs := strings.Join(execFake.lastArgs, " ")

	if !strings.Contains(gotArgs, "-cert-file dev-local.pem") {
		t.Fatalf("expected mkcert args to contain -cert-file dev-local.pem, got %q", gotArgs)
	}
	if !strings.Contains(gotArgs, "-key-file dev-local-key.pem") {
		t.Fatalf("expected mkcert args to contain -key-file dev-local-key.pem, got %q", gotArgs)
	}

	// Domains must appear sorted and deduplicated.
	if strings.Count(gotArgs, "app.localdev.test") != 1 || strings.Count(gotArgs, "api.localdev.test") != 1 {
		t.Fatalf("expected both domains once in mkcert args, got %q", gotArgs)
	}

	// Ensure ordering: "api.localdev.test app.localdev.test" OR the lexicographic order you choose.
	// Here we assert lexicographic order.
	if strings.Index(gotArgs, "api.localdev.test") > strings.Index(gotArgs, "app.localdev.test") {
		t.Fatalf("expected domains to be sorted in mkcert args, got %q", gotArgs)
	}
}
