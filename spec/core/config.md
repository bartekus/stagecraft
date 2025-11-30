# Core Config – Loading and Validation

- Feature ID: `CORE_CONFIG`
- Status: todo

## Goal

Provide a single, well-defined entrypoint for loading and validating Stagecraft configuration.

The config should be:

- **Human-friendly** (YAML, with sane defaults).
- **Machine-validated** (schema + semantic checks).
- **Easy to test** (pure functions where possible).

## Format (Initial)

Config file: `stagecraft.yml` (default in repo root).

Minimal initial structure:

```yaml
project:
  name: "my-app"

environments:
  dev:
    driver: "digitalocean"
    # future: region, registry, etc.
  # staging:
  # prod:
```

Where:
•	project.name – Human-readable name of the project.
•	environments – Map of environment name -> environment config.
•	driver – String key for the driver to use (e.g. digitalocean, noop).

Behaviour

Default Path
•	DefaultConfigPath() returns the default path:
•	./stagecraft.yml (just "stagecraft.yml").

Existence Check
•	Exists(path string) (bool, error):
•	Returns true, nil if the file exists and is a regular file.
•	Returns false, nil if the file does not exist.
•	Returns false, error for other I/O errors.

Loading
•	Load(path string) (*Config, error):
•	If the file does not exist:
•	Returns an error of type ErrConfigNotFound (a sentinel error).
•	If the file exists but is invalid YAML or fails validation:
•	Returns an error describing the problem.
•	On success:
•	Returns a populated Config.

Validation (initial)
•	For now:
•	project.name must be non-empty.
•	If environments is present:
•	Each environment key must be non-empty.
•	Each environment must have a non-empty driver.

Future: more detailed validation and cross-checks.

Non-Goals (initial version)
•	No remote config loading.
•	No environment variable interpolation.
•	No advanced schema evolution/migrations.

Tests

See spec/features.yaml entry for CORE_CONFIG:
•	pkg/config/config_test.go – unit tests for:
•	DefaultConfigPath
•	Exists
•	Load (config not found, invalid YAML, basic happy path).


