# Migration Engine Registry

- Feature ID: `CORE_MIGRATION_REGISTRY`
- Status: todo
- Depends on: `CORE_CONFIG`

## Goal

Provide a registry-based system for migration engines that:
- Eliminates hardcoded engine lists in validation
- Enables dynamic engine registration
- Supports multiple migration tools (Drizzle, Prisma, Knex, raw, etc.)
- Maintains thread-safe engine lookup

## Architecture

### Registry Pattern

Migration engines are registered at runtime and selected by ID from configuration.
The registry is the single source of truth for available engines.

### Interface

```go
// pkg/providers/migration/migration.go

package migration

import "context"

// Migration represents a single migration step.
type Migration struct {
    ID          string
    Description string
    Applied     bool
    // Additional fields as needed by specific engines
}

// PlanOptions contains options for planning migrations.
type PlanOptions struct {
    // Engine-specific configuration decoded from
    // databases[dbName].migrations in stagecraft.yml
    Config any
    
    // MigrationPath is the path to migration files
    MigrationPath string
    
    // ConnectionEnv is the environment variable name for DB connection
    ConnectionEnv string
    
    // WorkDir is the working directory
    WorkDir string
}

// RunOptions contains options for running migrations.
type RunOptions struct {
    // Config is the engine-specific configuration
    Config any
    
    // MigrationPath is the path to migration files
    MigrationPath string
    
    // ConnectionEnv is the environment variable name for DB connection
    ConnectionEnv string
    
    // WorkDir is the working directory
    WorkDir string
    
    // Direction specifies migration direction (up, down, etc.)
    Direction string
    
    // Steps limits the number of migrations to run (0 = all)
    Steps int
}

// Engine is the interface that all migration engines must implement.
type Engine interface {
    // ID returns the unique identifier for this engine (e.g., "drizzle", "prisma", "knex", "raw").
    ID() string
    
    // Plan analyzes migration files and returns a list of pending migrations.
    Plan(ctx context.Context, opts PlanOptions) ([]Migration, error)
    
    // Run executes migrations.
    Run(ctx context.Context, opts RunOptions) error
}
```

### Registry Implementation

```go
// pkg/providers/migration/registry.go

package migration

import (
    "fmt"
    "sync"
)

// Registry manages migration engine registration and lookup.
type Registry struct {
    mu      sync.RWMutex
    engines map[string]Engine
}

// NewRegistry creates a new empty registry.
func NewRegistry() *Registry {
    return &Registry{
        engines: make(map[string]Engine),
    }
}

// Register registers a migration engine.
// Panics if the engine ID is empty or already registered.
func (r *Registry) Register(e Engine) {
    r.mu.Lock()
    defer r.mu.Unlock()
    
    id := e.ID()
    if id == "" {
        panic("migration engine registration: empty ID")
    }
    if _, exists := r.engines[id]; exists {
        panic(fmt.Sprintf("migration engine registration: duplicate ID %q", id))
    }
    
    r.engines[id] = e
}

// Get retrieves an engine by ID.
// Returns an error if the engine is not found.
func (r *Registry) Get(id string) (Engine, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    e, ok := r.engines[id]
    if !ok {
        return nil, fmt.Errorf("unknown migration engine %q", id)
    }
    return e, nil
}

// Has checks if an engine with the given ID is registered.
func (r *Registry) Has(id string) bool {
    r.mu.RLock()
    defer r.mu.RUnlock()
    _, ok := r.engines[id]
    return ok
}

// IDs returns all registered engine IDs.
func (r *Registry) IDs() []string {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    ids := make([]string, 0, len(r.engines))
    for id := range r.engines {
        ids = append(ids, id)
    }
    return ids
}

// DefaultRegistry is the global default registry.
var DefaultRegistry = NewRegistry()

// Register registers an engine in the default registry.
func Register(e Engine) {
    DefaultRegistry.Register(e)
}

// Get retrieves an engine from the default registry.
func Get(id string) (Engine, error) {
    return DefaultRegistry.Get(id)
}

// Has checks if an engine exists in the default registry.
func Has(id string) bool {
    return DefaultRegistry.Has(id)
}
```

## Usage in Config Validation

Instead of hardcoded validation:

```go
// ❌ OLD: Hardcoded list
if engine != "drizzle" && engine != "prisma" && engine != "knex" && engine != "raw" {
    return fmt.Errorf("invalid migration engine: %s", engine)
}

// ✅ NEW: Registry-based
import mig "stagecraft/pkg/providers/migration"

func validateDatabaseMigrations(dbName string, dbCfg DatabaseConfig) error {
    engine := dbCfg.Migrations.Engine
    if engine == "" {
        return fmt.Errorf("databases.%s.migrations.engine is required", dbName)
    }
    
    if !mig.Has(engine) {
        return fmt.Errorf(
            "unknown migration engine %q for database %s; available engines: %v",
            engine, dbName, mig.DefaultRegistry.IDs(),
        )
    }
    
    // Validate path exists
    if dbCfg.Migrations.Path == "" {
        return fmt.Errorf("databases.%s.migrations.path is required", dbName)
    }
    
    return nil
}
```

## Engine Registration

Engines register themselves during initialization:

```go
// internal/providers/migration/drizzle/drizzle.go
func init() {
    migration.Register(&DrizzleEngine{})
}

// internal/providers/migration/prisma/prisma.go
func init() {
    migration.Register(&PrismaEngine{})
}

// internal/providers/migration/knex/knex.go
func init() {
    migration.Register(&KnexEngine{})
}

// internal/providers/migration/raw/raw.go
func init() {
    migration.Register(&RawEngine{})
}
```

## Thread Safety

The registry uses `sync.RWMutex` for thread-safe concurrent access:
- Multiple readers can access the registry simultaneously
- Writers (Register) acquire exclusive lock
- All operations are protected by appropriate locks

## Testing

Registry tests should cover:
- Registration of engines
- Duplicate registration panics
- Retrieval of registered engines
- Error handling for unknown engines
- Thread-safety under concurrent access
- IDs() returns all registered IDs

## Non-Goals

- Engine discovery from external sources (v1)
- Dynamic engine loading from plugins (v1)
- Engine versioning (v1)

## Related Features

- `MIGRATION_INTERFACE` - Migration engine interface definition
- `MIGRATION_CONFIG` - Migration configuration schema
- Individual engine implementations (Drizzle, Prisma, Knex, Raw)

