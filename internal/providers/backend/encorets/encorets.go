// internal/providers/backend/encorets/encorets.go
package encorets

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"

	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

// EncoreTsProvider implements the Encore.ts backend provider.
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
		Secrets struct {
			Types   []string `yaml:"types"`
			FromEnv []string `yaml:"from_env"`
		} `yaml:"secrets"`
		EntryPoint string   `yaml:"entrypoint"`
		EnvFrom    []string `yaml:"env_from"`
		Listen     string   `yaml:"listen"`
	} `yaml:"dev"`

	Build struct {
		// Encore-specific build options can be added here
	} `yaml:"build"`
}

// Dev runs the Encore.ts backend in development mode.
func (p *EncoreTsProvider) Dev(ctx context.Context, opts backend.DevOptions) error {
	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return fmt.Errorf("parsing encore-ts provider config: %w", err)
	}

	// Load environment files if specified
	env := make(map[string]string)
	for k, v := range opts.Env {
		env[k] = v
	}

	for _, envFile := range cfg.Dev.EnvFrom {
		// In a full implementation, we would load and merge env files
		// For now, we just note that they should be loaded
		_ = envFile
	}

	// Sync secrets if configured
	if len(cfg.Dev.Secrets.FromEnv) > 0 {
		for _, secretName := range cfg.Dev.Secrets.FromEnv {
			secretValue, exists := env[secretName]
			if !exists {
				// In a full implementation, we'd load from env files
				continue
			}

			// Run encore secret set
			types := cfg.Dev.Secrets.Types
			if len(types) == 0 {
				types = []string{"dev", "preview", "local"}
			}

			args := []string{"secret", "set", "--type", types[0]}
			for _, t := range types[1:] {
				args = append(args, "--type", t)
			}
			args = append(args, secretName)

			cmd := exec.CommandContext(ctx, "encore", args...)
			cmd.Stdin = strings.NewReader(secretValue)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				return fmt.Errorf("setting encore secret %s: %w", secretName, err)
			}
		}
	}

	// Run encore dev server
	listen := cfg.Dev.Listen
	if listen == "" {
		listen = "0.0.0.0:4000"
	}

	args := []string{"run", "--watch", "--listen", listen}
	if cfg.Dev.EntryPoint != "" {
		args = append(args, "--entrypoint", cfg.Dev.EntryPoint)
	}

	cmd := exec.CommandContext(ctx, "encore", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = os.Environ()
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// BuildDocker builds a Docker image using Encore.
func (p *EncoreTsProvider) BuildDocker(ctx context.Context, opts backend.BuildDockerOptions) (string, error) {
	cfg, err := p.parseConfig(opts.Config)
	if err != nil {
		return "", fmt.Errorf("parsing encore-ts provider config: %w", err)
	}

	_ = cfg // Config may contain build-specific options in the future

	// Run encore build docker
	args := []string{"build", "docker", opts.ImageTag}

	cmd := exec.CommandContext(ctx, "encore", args...)
	cmd.Dir = opts.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("encore build docker failed: %w", err)
	}

	return opts.ImageTag, nil
}

// parseConfig unmarshals the provider config.
func (p *EncoreTsProvider) parseConfig(cfg any) (*Config, error) {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("marshaling config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("invalid encore-ts provider config: %w", err)
	}

	return &config, nil
}

func init() {
	backend.Register(&EncoreTsProvider{})
}

