// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_MKCERT
// Spec: spec/dev/mkcert.md

// Package mkcert provides certificate generation for local development using mkcert.
package mkcert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"stagecraft/pkg/config"
)

// Options captures mkcert behavior.
type Options struct {
	DevDir       string   // Root dev directory, usually ".stagecraft/dev"
	Domains      []string // Dev domains to issue certs for
	EnableHTTPS  bool     // false when --no-https is set
	MkcertBinary string   // Optional override, default "mkcert"
	Verbose      bool
}

// CertConfig represents the resulting certificate configuration.
type CertConfig struct {
	Enabled bool
	CertDir string
	Domains []string

	// Paths to certificate and key files for Traefik (or other consumers).
	// For v1 we assume a single cert covering all domains.
	CertFile string
	KeyFile  string
}

// Writer is the minimal writer abstraction used by commands/loggers.
type Writer interface {
	Write(p []byte) (n int, err error)
}

// Command abstracts a running command.
type Command interface {
	Run() error
	SetStdout(w Writer)
	SetStderr(w Writer)
}

// ExecCommander abstracts external command construction for testability.
type ExecCommander interface {
	CommandContext(ctx context.Context, name string, args ...string) Command
}

// Logger is a minimal logging abstraction.
type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

// Generator provisions dev certificates using mkcert.
type Generator struct {
	exec ExecCommander
	log  Logger
}

// NewGenerator creates a generator with default dependencies.
func NewGenerator() *Generator {
	return &Generator{
		exec: &defaultExecCommander{},
		log:  defaultLogger{},
	}
}

// NewGeneratorWithDeps creates a generator with explicit dependencies (for tests).
func NewGeneratorWithDeps(execCmd ExecCommander, logger Logger) *Generator {
	if execCmd == nil {
		execCmd = &defaultExecCommander{}
	}
	if logger == nil {
		logger = defaultLogger{}
	}

	return &Generator{
		exec: execCmd,
		log:  logger,
	}
}

// EnsureCertificates ensures certificates exist for the given domains.
//
//   - If EnableHTTPS is false, returns a disabled CertConfig and does not call mkcert.
//   - If EnableHTTPS is true, ensures CertDir exists under DevDir and certificates
//     exist for all requested domains, invoking mkcert if necessary.
func (g *Generator) EnsureCertificates(
	ctx context.Context,
	_ *config.Config,
	opts Options,
) (*CertConfig, error) {
	// HTTPS disabled: pure no-op config.
	if !opts.EnableHTTPS {
		return &CertConfig{
			Enabled: false,
		}, nil
	}

	if opts.DevDir == "" {
		return nil, fmt.Errorf("dev: mkcert: dev dir must not be empty")
	}

	// Compute certificate directory and file paths.
	certDir := filepath.Join(opts.DevDir, "certs")
	certPath := filepath.Join(certDir, "dev-local.pem")
	keyPath := filepath.Join(certDir, "dev-local-key.pem")

	// Deduplicate and sort domains.
	domains := dedupeAndSortDomains(opts.Domains)
	if len(domains) == 0 {
		return nil, fmt.Errorf("dev: mkcert: no domains configured")
	}

	// Ensure certificate directory exists.
	// #nosec G301 -- cert directory needs 0755 for docker compose access
	if err := os.MkdirAll(certDir, 0o755); err != nil {
		return nil, fmt.Errorf("dev: mkcert: create cert dir %s: %w", certDir, err)
	}

	// If both cert and key files already exist, reuse them without invoking mkcert.
	if fileExists(certPath) && fileExists(keyPath) {
		return &CertConfig{
			Enabled:  true,
			CertDir:  certDir,
			Domains:  domains,
			CertFile: certPath,
			KeyFile:  keyPath,
		}, nil
	}

	// Need to invoke mkcert to generate certificates.
	binary := opts.MkcertBinary
	if binary == "" {
		binary = "mkcert"
	}

	args := []string{
		"-cert-file", "dev-local.pem",
		"-key-file", "dev-local-key.pem",
	}
	args = append(args, domains...)

	if opts.Verbose {
		g.log.Infof("dev: mkcert: generating certificates in %s for domains %v", certDir, domains)
	}

	// If using the default exec commander, set the working directory so mkcert
	// writes files into certDir while still using deterministic relative paths
	// in the arguments.
	if de, ok := g.exec.(*defaultExecCommander); ok {
		de.SetDir(certDir)
	}

	cmd := g.exec.CommandContext(ctx, binary, args...)
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("dev: mkcert binary not found; install mkcert or run with --no-https: %w", err)
		}
		return nil, fmt.Errorf("dev: mkcert failed: %w", err)
	}

	return &CertConfig{
		Enabled:  true,
		CertDir:  certDir,
		Domains:  domains,
		CertFile: certPath,
		KeyFile:  keyPath,
	}, nil
}

// defaultExecCommander is the production ExecCommander backed by os/exec.
type defaultExecCommander struct {
	dir string
}

// SetDir sets the working directory for subsequent commands.
func (d *defaultExecCommander) SetDir(dir string) {
	d.dir = dir
}

// CommandContext implements ExecCommander using exec.CommandContext.
func (d *defaultExecCommander) CommandContext(ctx context.Context, name string, args ...string) Command {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // inputs are controlled by Stagecraft
	if d.dir != "" {
		cmd.Dir = d.dir
	}
	return &cmdWrapper{cmd: cmd}
}

type cmdWrapper struct {
	cmd *exec.Cmd
}

func (c *cmdWrapper) Run() error {
	return c.cmd.Run()
}

func (c *cmdWrapper) SetStdout(w Writer) {
	c.cmd.Stdout = w
}

func (c *cmdWrapper) SetStderr(w Writer) {
	c.cmd.Stderr = w
}

func dedupeAndSortDomains(domains []string) []string {
	seen := make(map[string]struct{}, len(domains))
	for _, d := range domains {
		if d == "" {
			continue
		}
		if _, ok := seen[d]; ok {
			continue
		}
		seen[d] = struct{}{}
	}

	out := make([]string, 0, len(seen))
	for d := range seen {
		out = append(out, d)
	}

	sort.Strings(out)
	return out
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// defaultLogger writes to stderr without timestamps for deterministic logs.
type defaultLogger struct{}

func (defaultLogger) Infof(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func (defaultLogger) Errorf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
}
