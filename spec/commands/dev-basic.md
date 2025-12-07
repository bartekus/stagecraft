---
feature: CLI_DEV_BASIC
version: v1
status: done
domain: commands
inputs:
  flags:
    - name: --verbose
      type: bool
      default: "false"
      description: "Enable verbose output"
    - name: -v
      type: bool
      default: "false"
      description: "Shorthand for --verbose"
outputs:
  exit_codes:
    success: 0
    error: 1
---
# `stagecraft dev` – Basic Development Command

- Feature ID: `CLI_DEV_BASIC`
- Status: todo
- Depends on: `CORE_CONFIG`, `CORE_BACKEND_REGISTRY`, `PROVIDER_BACKEND_GENERIC`

## Goal

Provide a minimal but functional `stagecraft dev` command that:
- Loads and validates `stagecraft.yml` from the current directory
- Resolves the configured backend provider from the registry
- Delegates to the provider's `Dev()` method
- Streams output to the user
- Works end-to-end for `examples/basic-node`

## User Story

As a developer,
I want to run `stagecraft dev` in my project,
so that my backend starts in development mode
using the configured provider (generic, encore-ts, etc.).

## Behaviour

### Input

- Reads `stagecraft.yml` from current working directory (default)
- Future: `--config` flag to specify alternative path

### Steps

1. Load config from `stagecraft.yml` (or path from `--config`)
2. Validate config (already wired to registries)
3. Check that `backend` section exists
4. Resolve backend provider from registry using `backend.provider`
5. Extract provider-specific config from `backend.providers[providerID]`
6. Determine working directory (current working directory)
7. Call `BackendProvider.Dev(ctx, DevOptions{...})`
8. Stream logs from provider (provider handles stdout/stderr)

### Output

- Non-zero exit code if any step fails
- Useful log lines (when `--verbose` is set):
  - Selected provider ID
  - Absolute path to config file
  - Working directory
- Provider-level status messages (from provider implementation)

### Error Handling

- Config file not found: Clear error message
- Invalid config: Validation error with helpful details
- Unknown provider: Error with available provider list
- Missing provider config: Error indicating which config key is missing
- Provider execution failure: Error from provider (with context)

## CLI Usage

```bash
stagecraft dev
```

### Flags

- `--verbose` / `-v`: Enable verbose output (shows provider ID, config path, etc.)
- Future: `--config <path>`: Specify config file path
- Future: `--env <name>`: Select environment

## Examples

### Basic Node.js App

```bash
cd examples/basic-node
stagecraft dev
```

With `stagecraft.yml`:
```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
```

Expected behavior:
- Loads config
- Resolves `generic` provider
- Executes `npm run dev` in `./backend` directory
- Streams output to terminal

## Implementation

### Command Structure

```go
func NewDevCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "dev",
        Short: "Start development environment",
        RunE:  runDev,
    }
    return cmd
}

func runDev(cmd *cobra.Command, args []string) error {
    // 1. Load config
    // 2. Validate
    // 3. Resolve provider
    // 4. Call provider.Dev()
}
```

### Config Resolution

- Use `config.Load(config.DefaultConfigPath())`
- Handle `config.ErrConfigNotFound` with helpful message
- Validation happens automatically in `Load()`

### Provider Resolution

- Use `backendproviders.Get(cfg.Backend.Provider)`
- Error includes available providers via `backendproviders.DefaultRegistry.IDs()`

### Provider Config Extraction

- Use `cfg.Backend.GetProviderConfig()`
- This returns `any` which is passed directly to `DevOptions.Config`

### Working Directory

- Use `os.Getwd()` for project root
- Provider implementations handle relative paths in their config

## Validation

### Required Config

- `backend.provider` must be set
- `backend.providers[backend.provider]` must exist
- Provider must be registered in backend registry

### Error Messages

- Config not found: `"stagecraft config not found at stagecraft.yml"`
- Unknown provider: `"unknown backend provider 'foo'; available providers: [generic, encore-ts]"`
- Missing provider config: `"backend.providers.generic is missing; provider-specific config is required"`

## Testing

Tests should cover:
- Config loading and validation
- Provider resolution (known and unknown)
- Provider config extraction
- Error handling for all failure modes
- Integration with generic provider (using test fixtures)

See `spec/features.yaml` entry for `CLI_DEV_BASIC`:
- `internal/cli/commands/dev_test.go` – unit/CLI behaviour tests
- `test/e2e/dev_smoke_test.go` – end-to-end smoke test with `examples/basic-node`

## Non-Goals (v1)

- Multi-service orchestration (v2)
- Frontend provider integration (v2)
- Docker Compose infra management (v2)
- Environment variable file loading (v2)
- Process lifecycle management (v2)
- Hot reload detection (v2)

## Related Features

- `CORE_CONFIG` – Config loading and validation
- `CORE_BACKEND_REGISTRY` – Backend provider registry
- `PROVIDER_BACKEND_GENERIC` – Generic backend provider implementation
- `PROVIDER_BACKEND_ENCORE` – Encore.ts backend provider (future)
- `CLI_DEV` – Full dev command with all features (future)

