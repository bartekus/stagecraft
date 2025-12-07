---
feature: CLI_MIGRATE_BASIC
version: v1
status: done
domain: commands
inputs:
  flags:
    - name: --database
      type: string
      default: "main"
      description: "Database name to migrate"
    - name: --plan
      type: bool
      default: "false"
      description: "Show migration plan without executing"
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
# `stagecraft migrate` – Basic Migration Command

- Feature ID: `CLI_MIGRATE_BASIC`
- Status: todo
- Depends on: `CORE_CONFIG`, `CORE_MIGRATION_REGISTRY`, `MIGRATION_ENGINE_RAW`

## Goal

Provide a minimal but functional `stagecraft migrate` command that:
- Loads and validates `stagecraft.yml` from the current directory
- Resolves the configured migration engine from the registry
- Supports both planning (dry-run) and execution modes
- Works end-to-end for `examples/basic-node` with raw SQL migrations

## User Story

As a developer,
I want to run `stagecraft migrate` in my project,
so that my database migrations are applied
using the configured migration engine (raw, drizzle, etc.).

## Behaviour

### Input

- Reads `stagecraft.yml` from current working directory (default)
- Database name (default: `"main"`, configurable via `--database` flag)
- Future: `--config` flag to specify alternative path

### Steps

1. Load config from `stagecraft.yml` (or path from `--config`)
2. Validate config (already wired to registries)
3. Resolve database configuration (default: `"main"`)
4. Check that database has migrations configured
5. Resolve migration engine from registry using `databases[dbName].migrations.engine`
6. Extract engine-specific config from migration config
7. Determine migration path (relative to working directory)
8. If `--plan` flag:
   - Call `Engine.Plan(ctx, PlanOptions{...})`
   - Display migration plan
9. Otherwise:
   - Call `Engine.Run(ctx, RunOptions{...})`
   - Execute migrations

### Output

#### Plan Mode (`--plan`)

- List of pending migrations with:
  - Migration ID (filename)
  - Description
  - Status (pending/applied)
- Non-zero exit if planning fails

#### Run Mode (default)

- Progress messages for each migration applied
- Non-zero exit if execution fails
- Useful log lines (when `--verbose` is set):
  - Selected engine ID
  - Absolute path to config file
  - Database name
  - Migration path

### Error Handling

- Config file not found: Clear error message
- Invalid config: Validation error with helpful details
- Database not found: Error with available database names
- No migrations configured: Error indicating missing config
- Unknown engine: Error with available engine list
- Missing connection env: Error indicating which env var is required
- Migration execution failure: Error from engine (with context)

## CLI Usage

```bash
# Show migration plan (dry-run)
stagecraft migrate --plan

# Execute migrations
stagecraft migrate

# Execute for specific database
stagecraft migrate --database main
```

### Flags

- `--database <name>`: Database name to migrate (default: `"main"`)
- `--plan`: Show migration plan without executing
- `--verbose` / `-v`: Enable verbose output
- Future: `--config <path>`: Specify config file path
- Future: `--steps <n>`: Limit number of migrations to run

## Examples

### Basic Node.js App

```bash
cd examples/basic-node
stagecraft migrate
```

With `stagecraft.yml`:
```yaml
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
```

Expected behavior:
- Loads config
- Resolves `raw` engine
- Connects to database via `DATABASE_URL`
- Executes SQL files from `./migrations` in order
- Tracks applied migrations

### Plan Mode

```bash
stagecraft migrate --plan
```

Output:
```
Migration plan (2 pending):
  - 001_initial.sql: SQL migration: 001_initial.sql [pending]
  - 002_add_users.sql: SQL migration: 002_add_users.sql [pending]
```

## Implementation

### Command Structure

```go
func NewMigrateCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "migrate",
        Short: "Run database migrations",
        RunE:  runMigrate,
    }
    cmd.Flags().String("database", "main", "Database name to migrate")
    cmd.Flags().Bool("plan", false, "Show migration plan without executing")
    return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
    // 1. Load config
    // 2. Resolve database
    // 3. Resolve engine
    // 4. Call engine.Plan() or engine.Run()
}
```

### Config Resolution

- Use `config.Load(config.DefaultConfigPath())`
- Handle `config.ErrConfigNotFound` with helpful message
- Validation happens automatically in `Load()`

### Database Resolution

- Default to `"main"` if not specified
- Check `cfg.Databases[dbName]` exists
- Error includes available database names if not found

### Engine Resolution

- Use `migrationengines.Get(dbCfg.Migrations.Engine)`
- Error includes available engines via `migrationengines.DefaultRegistry.IDs()`

### Migration Path Resolution

- Use `dbCfg.Migrations.Path`
- If relative, resolve against `os.Getwd()`
- If absolute, use as-is

### Connection Environment Variable

- Read from `dbCfg.ConnectionEnv`
- Pass to engine via `RunOptions.ConnectionEnv`
- Engine is responsible for reading env var and connecting

## Validation

### Required Config

- `databases[dbName]` must exist
- `databases[dbName].migrations` must be set
- `databases[dbName].migrations.engine` must be set
- `databases[dbName].migrations.path` must be set
- `databases[dbName].connection_env` must be set
- Engine must be registered in migration registry

### Error Messages

- Config not found: `"stagecraft config not found at stagecraft.yml"`
- Database not found: `"database 'foo' not found in config; available: [main]"`
- No migrations: `"database 'main' has no migrations configured"`
- Unknown engine: `"unknown migration engine 'foo' for database main; available engines: [raw]"`
- Missing connection env: `"connection environment variable 'DATABASE_URL' is not set"`

## Testing

Tests should cover:
- Config loading and validation
- Database resolution (known and unknown)
- Engine resolution (known and unknown)
- Plan mode output
- Run mode execution
- Error handling for all failure modes
- Integration with raw engine (using test fixtures)

See `spec/features.yaml` entry for `CLI_MIGRATE_BASIC`:
- `internal/cli/commands/migrate_test.go` – unit/CLI behaviour tests
- `test/e2e/migrate_smoke_test.go` – end-to-end smoke test with `examples/basic-node`

## Non-Goals (v1)

- Multi-database migration in single command (v2)
- Migration rollback (v2)
- Migration dependency resolution (v2)
- Migration validation (syntax checking) (v2)
- Migration state tracking UI (v2)

## Related Features

- `CORE_CONFIG` – Config loading and validation
- `CORE_MIGRATION_REGISTRY` – Migration engine registry
- `MIGRATION_ENGINE_RAW` – Raw SQL migration engine implementation
- `CLI_MIGRATE_PLAN` – Dedicated plan command (future)
- `CLI_MIGRATE_RUN` – Dedicated run command (future)

