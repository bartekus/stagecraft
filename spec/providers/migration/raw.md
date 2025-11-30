# Raw Migration Engine

- Feature ID: `MIGRATION_ENGINE_RAW`
- Status: todo
- Depends on: `CORE_MIGRATION_REGISTRY`, `MIGRATION_INTERFACE`

## Goal

Provide a simple, framework-agnostic migration engine that:
- Executes raw SQL files directly
- Works with any SQL database (PostgreSQL, MySQL, SQLite, etc.)
- Requires no migration framework dependencies
- Serves as the baseline for migration-agnostic operation

## Use Cases

The raw engine is useful for:
- Projects using raw SQL migrations
- Teams transitioning from manual migration scripts
- Simple projects that don't need framework-specific features
- Baseline for testing the migration engine architecture

## Configuration

### Schema

```yaml
databases:
  main:
    connection_env: DATABASE_URL
    migrations:
      engine: raw
      path: ./migrations
      strategy: pre_deploy
```

### File Organization

Migrations are SQL files in the specified directory:

```
migrations/
  001_initial.sql
  002_add_users.sql
  003_add_posts.sql
```

Files are executed in lexicographic order (by filename).

## Implementation

### Interface Compliance

The raw engine implements `Engine`:

```go
type RawEngine struct{}

func (e *RawEngine) ID() string {
    return "raw"
}

func (e *RawEngine) Plan(ctx context.Context, opts PlanOptions) ([]Migration, error) {
    // List SQL files in migration directory
    // Return as Migration structs
}

func (e *RawEngine) Run(ctx context.Context, opts RunOptions) error {
    // Connect to database
    // Execute SQL files in order
    // Track applied migrations
}
```

### Plan Phase

1. Read migration directory
2. Filter for `.sql` files
3. Sort by filename (lexicographic)
4. Return as `[]Migration` with `Applied: false` (v1 doesn't track state)

### Run Phase (v1 Placeholder)

Currently returns an error indicating execution is not yet implemented.
Future implementation will:
1. Connect to database using `opts.ConnectionEnv`
2. Create migration tracking table if needed
3. Execute SQL files in order
4. Track applied migrations
5. Support rollback (v2)

## File Naming Conventions

Recommended naming patterns:
- `001_initial.sql`
- `002_add_users.sql`
- `003_add_posts.sql`
- `20240101_initial.sql` (date-based)
- `v1.0.0_initial.sql` (version-based)

The engine sorts by filename, so use consistent naming.

## SQL File Format

SQL files can contain:
- Multiple statements (separated by `;`)
- Comments (`--` or `/* */`)
- DDL (CREATE, ALTER, DROP)
- DML (INSERT, UPDATE, DELETE)

Example:
```sql
-- Migration: Add users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

## Database Support

The raw engine works with any SQL database:
- PostgreSQL
- MySQL/MariaDB
- SQLite
- SQL Server
- Oracle (with appropriate SQL)

Connection is via `DATABASE_URL` environment variable (standard format).

## Validation

### Core Validation (Stagecraft)

- `migrations.engine` must be "raw" (validated via registry)
- `migrations.path` must be non-empty
- Migration directory must exist

### Engine-Specific Validation

- Migration directory must be readable
- At least one `.sql` file must exist (for Run phase)
- SQL files must be parseable (basic syntax check, v2)

## Testing

Tests should cover:
- Plan phase: listing SQL files
- Plan phase: sorting order
- Plan phase: ignoring non-SQL files
- Run phase: directory validation
- Error handling for missing directories
- Error handling for empty directories

## Comparison with Other Engines

### vs Drizzle Engine

- **Raw**: Direct SQL execution, no ORM
- **Drizzle**: Uses Drizzle migration format, ORM-aware

### vs Prisma Engine

- **Raw**: Direct SQL execution
- **Prisma**: Uses Prisma migration format, schema-aware

### vs Knex Engine

- **Raw**: Direct SQL execution
- **Knex**: Uses Knex migration format, query builder-aware

## Non-Goals (v1)

- Migration state tracking (v1.1)
- Rollback support (v2)
- Migration dependency resolution (v2)
- Multi-database transactions (v2)
- Migration validation (syntax checking) (v2)

## Related Features

- `CORE_MIGRATION_REGISTRY` - Migration engine registry system
- `MIGRATION_INTERFACE` - Migration engine interface
- `MIGRATION_CONFIG` - Migration configuration schema
- `CLI_MIGRATE_PLAN` - Migration planning command
- `CLI_MIGRATE_RUN` - Migration execution command

