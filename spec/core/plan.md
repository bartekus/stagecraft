# Deployment Planning Engine

- Feature ID: `CORE_PLAN`
- Status: done
- Depends on: `CORE_CONFIG`

## Goal

Provide a deployment planning engine that creates structured plans from configuration. The planner translates configuration into a sequence of operations that can be executed by drivers.

## Architecture

### Plan Structure

A `Plan` consists of:
- Environment name
- List of operations to execute
- Operation dependencies (for future use)

### Operation Types

- `infra_provision` - Infrastructure provisioning
- `migration` - Database migrations (pre_deploy, post_deploy, manual)
- `build` - Building Docker images
- `deploy` - Deploying containers
- `health_check` - Health checks after deployment

### Planner

The `Planner` creates deployment plans by:
1. Validating environment exists in config
2. Adding migration operations (pre_deploy strategy)
3. Adding build operations
4. Adding deploy operations
5. Adding migration operations (post_deploy strategy)
6. Adding health check operations

## Implementation

See `internal/core/plan.go` for the implementation.

### Example Plan

```go
plan := &Plan{
    Environment: "prod",
    Operations: []Operation{
        {
            Type: OpTypeMigration,
            Description: "Run pre_deploy migrations for database main",
            Metadata: map[string]interface{}{
                "database": "main",
                "strategy": "pre_deploy",
                "engine": "raw",
            },
        },
        {
            Type: OpTypeBuild,
            Description: "Build backend using provider generic",
        },
        {
            Type: OpTypeDeploy,
            Description: "Deploy to environment prod",
        },
        {
            Type: OpTypeHealthCheck,
            Description: "Health check for environment prod",
        },
    },
}
```

## Migration Strategies

- `pre_deploy`: Run migrations before deployment
- `post_deploy`: Run migrations after deployment
- `manual`: Migrations must be run manually

## Future Enhancements

- Dependency resolution between operations
- Parallel execution of independent operations
- Rollback plan generation
- Plan validation and dry-run execution

## Related Features

- `CORE_CONFIG` - Configuration loading
- `CLI_PLAN` - Plan command (dry-run)
- `CLI_DEPLOY` - Deploy command execution

