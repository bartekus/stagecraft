<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<!--

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

-->

# Stagecraft Examples

This directory contains example projects demonstrating how to use Stagecraft.

## Available Examples

### `basic-node/`

A minimal Node.js application example demonstrating:
- Basic Stagecraft configuration (`stagecraft.yml`)
- Backend service setup
- Database migrations
- Docker Compose integration

**Use case**: Getting started with Stagecraft, understanding basic configuration.

**To try it**:
```bash
cd examples/basic-node
stagecraft init  # If not already initialized
stagecraft dev   # Start development environment
```

**What it includes**:
- Simple Express.js backend
- PostgreSQL database
- SQL migration example
- Basic `stagecraft.yml` configuration

---

## Using Examples

### For Learning

Examples are designed to be:
- **Self-contained** - Each example works independently
- **Well-documented** - Each has its own README with setup instructions
- **Minimal** - Focus on specific Stagecraft features, not full applications

### For Development

Examples are useful for:
- Testing Stagecraft features
- Verifying configuration syntax
- Understanding provider integrations
- Debugging issues with specific setups

### For Contributors

When adding new Stagecraft features:
- Consider adding an example demonstrating the feature
- Keep examples minimal and focused
- Document setup and usage clearly
- Test examples regularly to ensure they still work

---

## Example Structure

Each example typically includes:

```
example-name/
├── README.md              # Setup and usage instructions
├── stagecraft.yml         # Stagecraft configuration
├── docker-compose.yml     # Docker Compose services (if needed)
├── backend/               # Backend service code
├── frontend/              # Frontend code (if applicable)
├── migrations/            # Database migrations (if applicable)
└── ...                   # Other example-specific files
```

---

## Adding New Examples

When creating a new example:

1. **Create a descriptive directory name** (e.g., `basic-node`, `encore-backend`)

2. **Include a README.md** with:
   - What the example demonstrates
   - Prerequisites
   - Setup instructions
   - How to run it
   - What to expect

3. **Keep it minimal** - Focus on demonstrating specific Stagecraft features

4. **Test it** - Ensure the example works with current Stagecraft version

5. **Document it** - Update this README with the new example

---

## Example Best Practices

- **Keep examples simple** - They should demonstrate concepts, not be production-ready apps
- **Use realistic but minimal configs** - Show real-world patterns without complexity
- **Document assumptions** - Note any prerequisites or setup requirements
- **Test regularly** - Examples should work with the current Stagecraft version
- **Version compatibility** - Note if an example requires specific Stagecraft features

---

## Questions?

- For general Stagecraft usage, see [docs/guides/getting-started.md](../docs/guides/getting-started.md)
- For configuration details, see [docs/stagecraft-spec.md](../docs/stagecraft-spec.md)
- For contributing examples, see [CONTRIBUTING.md](../CONTRIBUTING.md)
