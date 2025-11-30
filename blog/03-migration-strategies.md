## Migration Strategy

> **Related Documents:**
> - [`spec/core/config.md`](../spec/core/config.md) - Full config schema including `databases` section
> - [`docs/implementation-roadmap.md`](../docs/implementation-roadmap.md) - Phase 6: Migration System
> - [`spec/features.yaml`](../spec/features.yaml) - Migration-related features (MIGRATION_*)

### 1. **Migration approach**
- Treat migrations as first-class deployment steps, not just framework concerns
- Run migrations via one-off containers using the app image (similar to Kamal)
- Support multiple migration engines (Drizzle, Prisma, Knex, raw, etc.) without hardcoding

### 2. **Configuration schema (`stagecraft.yml`)**
- Use a YAML config with top-level `databases` and environment-specific overrides
- Support three migration strategies:
    - `pre_deploy`: Run before rollout (default)
    - `post_deploy`: Run after rollout (for data backfills)
    - `manual`: Never auto-run; require explicit command
- Support multiple databases per app with per-database migration config
- Use environment variable interpolation (`${VAR}`) in config values
  - **Note**: v1 supports basic `${VAR}` interpolation for migration config values only
  - Full environment variable interpolation across all config fields is deferred to v2

### 3. **Deployment pipeline order**
Fixed sequence:
1. Build image
2. Push to registry
3. Pre-deploy migrations
4. Rollout app containers
5. Post-deploy migrations
6. Finalize (record release)

### 4. **Architecture abstractions**
- `Migrator` interface: `Plan()` and `Run()` methods
- `ContainerRunner` interface: Generic "run one container on a host"
- `Deployer` struct: Orchestrates the full pipeline using interfaces
- Strategy pattern for builders, pushers, and rollout strategies

### 5. **Config loading and merging**
- Load `stagecraft.yml` with env var expansion
- Merge top-level `databases` into environment-specific configs (env overrides)
- Validate that each environment has at least one host with `primary` role

### 6. **Release state tracking**
- Store release history in `.stagecraft/releases.json` (file-based, v1)
- Append-only releases with phase status tracking
- Track: build, push, migrate_pre, rollout, migrate_post phases
- Include `previous_id` for rollback support

### 7. **Rollback strategy (v1)**
- Rollback only changes the application image, not database migrations
- Explicitly documented: "Database migrations are not rolled back"
- Avoids complexity of reversible migrations for initial version

### 8. **CLI commands**
- `stagecraft deploy --env <name>`: Full pipeline
- `stagecraft migrate plan --env <name>`: Show pending migrations
- `stagecraft migrate run --env <name>`: Run migrations with filters
- `stagecraft releases list/show`: Inspect release history
- `stagecraft rollback --env <name> --to <release-id>`: Rollback to previous image

### 9. **Implementation details**
- Go structs mirror the YAML schema
- Stateless migrator (can be extended later to check DB state)
- Simple container runner abstraction (SSH + Docker)
- Testable via interfaces (mocks for Builder, Pusher, RolloutStrategy, Migrator)

### 10. **Design principles**
- Explicit over implicit: migrations are explicit steps in the pipeline
- Config-driven: Everything defined in `stagecraft.yml`
- Environment-aware: Per-environment overrides with global defaults
- Extensible: Interfaces allow swapping implementations
- Simple v1: File-based state, image-only rollback, stateless migrations

These decisions provide a clear architecture for handling migrations as part of the deployment lifecycle, with a path for future enhancements.
