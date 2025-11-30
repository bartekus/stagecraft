# Getting Started with Stagecraft

This guide will help you get started with Stagecraft, a deployment and infrastructure orchestration CLI.

## Prerequisites

- **Go 1.22+** (or recent stable version)
- **Docker** (for future dev/deploy functionality)
- **Git** (for version control)

## Installation

### From Source

```bash
git clone https://github.com/your-org/stagecraft.git
cd stagecraft
go build ./cmd/stagecraft
```

The binary will be created in the current directory as `stagecraft`.

### Add to PATH (Optional)

```bash
# Add to your shell profile (~/.zshrc, ~/.bashrc, etc.)
export PATH="$PATH:/path/to/stagecraft"
```

## Quick Start

### 1. Initialize a Project

Navigate to your project directory and run:

```bash
stagecraft init
```

This will:
- Create a `stagecraft.yml` configuration file
- Guide you through initial setup (interactive mode)
- Validate your project structure

For non-interactive mode with defaults:

```bash
stagecraft init --non-interactive
```

### 2. Configure Your Project

Edit `stagecraft.yml` to match your project structure:

```yaml
project:
  name: "my-app"

environments:
  dev:
    driver: "digitalocean"
  staging:
    driver: "digitalocean"
  prod:
    driver: "digitalocean"
```

See the [Configuration Guide](../spec/core/config.md) for detailed configuration options.

### 3. Verify Configuration

```bash
# Check that your config is valid
stagecraft --help
```

## Next Steps

- **Plan a deployment**: `stagecraft plan --env=staging` (coming soon)
- **Deploy**: `stagecraft deploy --env=staging` (coming soon)
- **Check status**: `stagecraft status --env=staging` (coming soon)

## Project Structure

After initialization, your project should have:

```
my-project/
├── stagecraft.yml          # Stagecraft configuration
├── docker-compose.yml      # (optional) Docker Compose services
└── ...
```

## Common Commands

| Command | Description |
|---------|-------------|
| `stagecraft init` | Initialize Stagecraft in your project |
| `stagecraft version` | Show version information |
| `stagecraft --help` | Show help for all commands |

## Getting Help

- Run `stagecraft <command> --help` for command-specific help
- See [CLI Reference](../reference/cli.md) for detailed command documentation
- Check [Architecture Documentation](../architecture.md) for system design
- Review [Feature Specifications](../spec/) for detailed behavior

## Troubleshooting

### Config File Not Found

If you see an error about a missing config file:

```bash
# Ensure you're in the project root
pwd

# Check if stagecraft.yml exists
ls -la stagecraft.yml

# Re-run init if needed
stagecraft init
```

### Build Errors

If you encounter build errors:

```bash
# Ensure Go is installed and up to date
go version

# Clean and rebuild
go clean ./...
go build ./cmd/stagecraft
```

## Development

If you're contributing to Stagecraft:

1. Read [Agent.md](../../Agent.md) for development guidelines
2. Review [Architecture Decision Records](adr/) for design decisions
3. Check [Implementation Status](implementation-status.md) for feature progress
4. Follow the spec-first, test-first workflow

## Further Reading

- [Stagecraft Specification](../stagecraft-spec.md) - Complete feature specification
- [Architecture Overview](../architecture.md) - System architecture
- [ADR 0001: Architecture](../adr/0001-architecture.md) - Initial architecture decisions

