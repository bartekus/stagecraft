# Generic Backend Provider

- Feature ID: `PROVIDER_BACKEND_GENERIC`
- Status: todo
- Depends on: `CORE_BACKEND_REGISTRY`, `PROVIDER_BACKEND_INTERFACE`

## Goal

Provide a generic, command-based backend provider that:
- Runs arbitrary commands for development
- Builds Docker images using standard Dockerfiles
- Requires no framework-specific knowledge
- Serves as the baseline for backend-agnostic operation

## Use Cases

The generic provider is useful for:
- Node.js backends (Express, Fastify, etc.)
- Go backends (standard Go applications)
- Python backends (Django, FastAPI, etc.)
- Any backend that can be run via a command and built with Docker

## Configuration

### Schema

```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]        # required; command to run
        workdir: "./backend"                   # optional; defaults to project root
        env:                                   # optional; environment variables
          NODE_ENV: development
          PORT: "4000"
      build:
        dockerfile: "./backend/Dockerfile"     # optional; defaults to "Dockerfile"
        context: "./backend"                    # optional; defaults to workdir or "."
```

### Examples

#### Node.js Backend

```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["npm", "run", "dev"]
        workdir: "./backend"
        env:
          NODE_ENV: development
          PORT: "4000"
      build:
        dockerfile: "./backend/Dockerfile"
        context: "./backend"
```

#### Go Backend

```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["go", "run", "main.go"]
        workdir: "./cmd/api"
        env:
          PORT: "4000"
      build:
        dockerfile: "./cmd/api/Dockerfile"
        context: "."
```

#### Python Backend

```yaml
backend:
  provider: generic
  providers:
    generic:
      dev:
        command: ["uvicorn", "main:app", "--reload", "--host", "0.0.0.0"]
        workdir: "./backend"
        env:
          PYTHONPATH: "./backend"
      build:
        dockerfile: "./backend/Dockerfile"
        context: "./backend"
```

## Implementation

### Interface Compliance

The generic provider implements `BackendProvider`:

```go
type GenericProvider struct{}

func (p *GenericProvider) ID() string {
    return "generic"
}

func (p *GenericProvider) Dev(ctx context.Context, opts DevOptions) error {
    // Parse config from opts.Config
    // Execute command with environment variables
    // Stream output
}

func (p *GenericProvider) BuildDocker(ctx context.Context, opts BuildDockerOptions) error {
    // Parse config from opts.Config
    // Run docker build with dockerfile and context
    // Return image tag
}
```

### Config Parsing

The provider unmarshals its config from `opts.Config`:

```go
type Config struct {
    Dev struct {
        Command []string          `yaml:"command"`
        WorkDir string            `yaml:"workdir"`
        Env     map[string]string `yaml:"env"`
    } `yaml:"dev"`
    
    Build struct {
        Dockerfile string `yaml:"dockerfile"`
        Context    string `yaml:"context"`
    } `yaml:"build"`
}
```

### Dev Mode Behavior

1. Parse config from `opts.Config`
2. Validate that `dev.command` is non-empty
3. Determine working directory (config > opts.WorkDir > ".")
4. Merge environment variables (config.env + opts.Env)
5. Execute command with context cancellation
6. Stream stdout/stderr to user

### Build Mode Behavior

1. Parse config from `opts.Config`
2. Determine dockerfile path (config > "Dockerfile")
3. Determine build context (config > opts.WorkDir > ".")
4. Execute `docker build -t <imageTag> -f <dockerfile> <context>`
5. Return image tag on success

## Validation

### Required Fields

- `dev.command`: Must be non-empty array
- `build.dockerfile`: Optional, defaults to "Dockerfile"
- `build.context`: Optional, defaults to workdir or "."

### Error Handling

- Missing `dev.command`: Return clear error message
- Invalid command: Let exec fail with standard error
- Docker build failure: Return error with build output

## Testing

Tests should cover:
- Config parsing (valid and invalid)
- Dev mode command execution
- Build mode Docker execution
- Environment variable merging
- Working directory resolution
- Error handling for missing fields

## Comparison with Other Providers

### vs Encore.ts Provider

- **Generic**: Framework-agnostic, requires manual Dockerfile
- **Encore.ts**: Framework-aware, uses `encore build docker`

### vs Future Providers

- **Generic**: Baseline implementation
- **Framework-specific**: Optimized for specific frameworks (e.g., Next.js, Rails)

## Non-Goals

- Automatic Dockerfile generation (v1)
- Hot reload detection (v1)
- Build optimization (v1)
- Multi-stage build support (v1)

## Related Features

- `CORE_BACKEND_REGISTRY` - Provider registry system
- `PROVIDER_BACKEND_INTERFACE` - BackendProvider interface
- `CORE_BACKEND_PROVIDER_CONFIG_SCHEMA` - Provider config structure
- `CLI_DEV` - Development command that uses providers
- `CLI_BUILD` - Build command that uses providers

