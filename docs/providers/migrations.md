# Migration Engines

Stagecraft uses an engine-based architecture for database migrations. This allows Stagecraft to work with any migration tool or framework.

## Architecture

### Engine Interface

All migration engines implement the `Engine` interface:

```go
type Engine interface {
    ID() string
    Plan(ctx context.Context, opts PlanOptions) ([]Migration, error)
    Run(ctx context.Context, opts RunOptions) error
}
```

### Engine Registry

Engines register themselves at package initialization:

```go
func init() {
    migration.Register(&MyEngine{})
}
```

The registry is the single source of truth for available engines. Config validation checks against the registry, not hardcoded lists.

## Configuration

### Database Migration Config

Migrations are configured per database:

```yaml
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw        # Engine ID (must be registered)
      path: ./migrations # Path to migration files
      strategy: pre_deploy # pre_deploy, post_deploy, or manual
```

### Migration Strategies

- **pre_deploy**: Run migrations before deploying new code (default)
- **post_deploy**: Run migrations after deploying new code (for data backfills)
- **manual**: Never auto-run; require explicit `stagecraft migrate run` command

## Available Engines

### Raw Engine

The raw engine executes SQL files directly.

**Config Example:**
```yaml
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
```

**Behavior:**
- Reads SQL files from the specified path
- Executes them in lexicographic order
- Suitable for any database that supports SQL

**File Naming:**
- Files should end in `.sql`
- Order is determined by filename (e.g., `001_initial.sql`, `002_add_users.sql`)

### Future Engines

- **Drizzle**: Drizzle ORM migrations
- **Prisma**: Prisma migrations
- **Knex**: Knex.js migrations

## Creating a Custom Engine

### Step 1: Implement the Interface

```go
package myengine

import (
    "context"
    "stagecraft/pkg/providers/migration"
)

type MyEngine struct{}

func (e *MyEngine) ID() string {
    return "my-engine"
}

func (e *MyEngine) Plan(ctx context.Context, opts migration.PlanOptions) ([]migration.Migration, error) {
    // Analyze migration files
    // Return list of pending migrations
    return []migration.Migration{}, nil
}

func (e *MyEngine) Run(ctx context.Context, opts migration.RunOptions) error {
    // Execute migrations
    return nil
}
```

### Step 2: Parse Config

```go
type Config struct {
    // Engine-specific configuration
    CustomOption string `yaml:"custom_option"`
}

func (e *MyEngine) parseConfig(cfg any) (*Config, error) {
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### Step 3: Register the Engine

```go
func init() {
    migration.Register(&MyEngine{})
}
```

### Step 4: Import in Config Package

Add to `pkg/config/config.go`:

```go
import (
    _ "stagecraft/internal/providers/migration/myengine"
)
```

This ensures the engine registers itself before config validation.

## Validation

### Core Validation (Stagecraft)

Stagecraft core validates:
- `migrations.engine` is non-empty
- Engine exists in registry
- `migrations.path` is non-empty
- `migrations.strategy` is valid (if present)

### Engine-Specific Validation

Engine implementations validate their own config and migration files:

```go
func (e *MyEngine) Plan(ctx context.Context, opts migration.PlanOptions) error {
    // Validate migration path exists
    // Validate migration files are readable
    // Check database connection
    // Return list of migrations
}
```

## Migration Execution Flow

1. **Plan Phase**: Engine analyzes migration files and returns pending migrations
2. **Run Phase**: Engine executes migrations in order
3. **Tracking**: Engine tracks which migrations have been applied (implementation-specific)

## Best Practices

1. **Idempotent Migrations**: Migrations should be safe to run multiple times
2. **Transaction Safety**: Wrap migrations in transactions where possible
3. **Clear Errors**: Provide helpful error messages with migration context
4. **State Tracking**: Track applied migrations to avoid re-running
5. **Rollback Support**: Consider supporting rollback (v2 feature)

## Related Documentation

- [Migration Engine Registry](../spec/core/migration-registry.md)
- [Migration Strategy](../blog/03-migration-strategies.md)
- [Raw Engine Implementation](../../internal/providers/migration/raw/raw.go)

