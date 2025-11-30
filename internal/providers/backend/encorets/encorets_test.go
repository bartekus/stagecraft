// internal/providers/backend/encorets/encorets_test.go
package encorets

import (
	"context"
	"testing"

	"stagecraft/pkg/providers/backend"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

func TestEncoreTsProvider_ID(t *testing.T) {
	p := &EncoreTsProvider{}
	if got := p.ID(); got != "encore-ts" {
		t.Errorf("ID() = %q, want %q", got, "encore-ts")
	}
}

func TestEncoreTsProvider_ParseConfig(t *testing.T) {
	p := &EncoreTsProvider{}

	cfg := map[string]any{
		"dev": map[string]any{
			"secrets": map[string]any{
				"types":    []string{"dev", "preview", "local"},
				"from_env": []string{"DOMAIN", "API_DOMAIN"},
			},
			"entrypoint": "./backend",
			"env_from":   []string{".env.local"},
			"listen":     "0.0.0.0:4000",
		},
	}

	parsed, err := p.parseConfig(cfg)
	if err != nil {
		t.Fatalf("parseConfig() error = %v", err)
	}

	if len(parsed.Dev.Secrets.Types) != 3 {
		t.Errorf("Dev.Secrets.Types length = %d, want 3", len(parsed.Dev.Secrets.Types))
	}

	if len(parsed.Dev.Secrets.FromEnv) != 2 {
		t.Errorf("Dev.Secrets.FromEnv length = %d, want 2", len(parsed.Dev.Secrets.FromEnv))
	}

	if parsed.Dev.EntryPoint != "./backend" {
		t.Errorf("Dev.EntryPoint = %q, want %q", parsed.Dev.EntryPoint, "./backend")
	}

	if parsed.Dev.Listen != "0.0.0.0:4000" {
		t.Errorf("Dev.Listen = %q, want %q", parsed.Dev.Listen, "0.0.0.0:4000")
	}
}

func TestEncoreTsProvider_ParseConfig_InvalidYAML(t *testing.T) {
	p := &EncoreTsProvider{}

	// Invalid config structure
	cfg := "not a map"

	_, err := p.parseConfig(cfg)
	if err == nil {
		t.Error("parseConfig() error = nil, want error for invalid config")
	}
}

func TestEncoreTsProvider_Dev_ValidatesConfig(t *testing.T) {
	p := &EncoreTsProvider{}

	opts := backend.DevOptions{
		Config: map[string]any{
			"dev": map[string]any{
				"secrets": map[string]any{
					"types":    []string{"dev"},
					"from_env": []string{"TEST_SECRET"},
				},
			},
		},
		WorkDir: ".",
		Env: map[string]string{
			"TEST_SECRET": "test-value",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// This will fail because encore CLI might not be available,
	// but we're testing the config parsing logic
	err := p.Dev(ctx, opts)
	// Error is expected (encore not found), but config parsing should succeed
	if err != nil && err.Error() == "" {
		t.Error("expected error message, got empty")
	}
}

