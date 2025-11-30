# Stagecraft Architecture

## Layers

1. **CLI Layer (`cmd/`, `internal/cli/`)**
    - Responsibilities:
        - Parse user input (flags, args).
        - Provide help texts and guidance.
        - Present logs and progress.
    - Non-responsibilities:
        - Deployment logic.
        - Platform-specific behaviour.

2. **Core Layer (`internal/core/`)**
    - Responsibilities:
        - Interpret configuration.
        - Build deployment plans.
        - Track state transitions and outcomes.
    - Non-responsibilities:
        - Network I/O to specific providers.
        - CLI details.

3. **Driver Layer (`internal/drivers/`)**
    - Responsibilities:
        - Apply plans to concrete platforms (DigitalOcean, GitHub, etc.).
    - Non-responsibilities:
        - Plan construction.
        - CLI UX.

4. **Support Libraries (`pkg/`)**
    - Responsibilities:
        - Config loading and schema.
        - Plugin interfaces.
        - Shared utilities.

## Data Flow (High Level)

1. User invokes a CLI command (e.g. `stagecraft deploy --env=prod`).
2. CLI layer:
    - Parses flags and arguments.
    - Loads configuration via `pkg/config`.
3. Core layer:
    - Translates intent into a deployment plan.
4. Driver layer:
    - Executes the plan against the chosen provider(s).
5. CLI layer:
    - Streams progress and results back to the user.

For more details, see the individual command specs under `spec/commands/` and driver specs under `spec/drivers/`.
