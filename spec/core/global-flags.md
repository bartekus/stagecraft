# Global Flags – CLI Global Flag Handling

- Feature ID: `CLI_GLOBAL_FLAGS`
- Status: todo

## Goal

Implement global flags that apply to all Stagecraft commands:
- `--env` - Target environment (dev, staging, prod)
- `--config` - Path to stagecraft.yml
- `--verbose` - Enable verbose output
- `--dry-run` - Show what would be done without executing

## Behavior

### Flag Precedence

1. Command-line flags (highest priority)
2. Environment variables (e.g., `STAGECRAFT_ENV`)
3. Config file defaults
4. Built-in defaults (lowest priority)

### Environment Variable Support

- `STAGECRAFT_ENV` → `--env`
- `STAGECRAFT_CONFIG` → `--config`
- `STAGECRAFT_VERBOSE` → `--verbose`
- `STAGECRAFT_DRY_RUN` → `--dry-run`

### Flag Validation

- `--env` must be a valid environment name from config
- `--config` must point to a valid file (if specified)
- `--verbose` is a boolean flag
- `--dry-run` is a boolean flag

### Integration with Commands

All commands should:
- Accept these flags via Cobra persistent flags
- Use the resolved values from the root command
- Respect `--dry-run` to show actions without executing
- Use `--verbose` to control logging level

## Implementation

### Root Command Setup

```go
// In internal/cli/root.go
cmd.PersistentFlags().StringP("env", "e", "", "target environment")
cmd.PersistentFlags().StringP("config", "c", "", "path to stagecraft.yml")
cmd.PersistentFlags().BoolP("verbose", "v", false, "enable verbose output")
cmd.PersistentFlags().Bool("dry-run", false, "show actions without executing")
```

### Flag Resolution

Create a helper to resolve flags with precedence:
- Check flag value
- Fall back to environment variable
- Fall back to config default
- Fall back to built-in default

## Tests

See `spec/features.yaml` entry for `CLI_GLOBAL_FLAGS`:
- `internal/cli/root_test.go` – tests for:
  - Flag parsing
  - Precedence rules
  - Environment variable support
  - Default values

