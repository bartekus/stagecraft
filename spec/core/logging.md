---
feature: CORE_LOGGING
version: v1
status: done
domain: core
inputs:
  flags: []
outputs:
  exit_codes: {}
---
# Core Logging – Structured Logging Helpers

- Feature ID: `CORE_LOGGING`
- Status: todo

## Goal

Provide a consistent, structured logging system for all Stagecraft commands and components.

The logging system should:
- Support multiple log levels (Debug, Info, Warn, Error)
- Be configurable via global `--verbose` flag
- Provide structured output for machine parsing (optional)
- Support progress indicators and spinners
- Be testable and mockable

## API Design

### Package: `pkg/logging`

```go
package logging

type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    WithFields(fields ...Field) Logger
}

type Field struct {
    Key   string
    Value interface{}
}

func NewLogger(verbose bool) Logger
func NewField(key string, value interface{}) Field
```

### Usage Example

```go
logger := logging.NewLogger(verbose)
logger.Info("Starting deployment", 
    logging.NewField("env", "prod"),
    logging.NewField("version", "v1.2.3"))
```

## Behavior

### Log Levels

- **Debug**: Detailed information for debugging (only with `--verbose`)
- **Info**: General informational messages (default)
- **Warn**: Warning messages that don't stop execution
- **Error**: Error messages (always shown)

### Global Flag Integration

- `--verbose` flag enables Debug level logging
- Default level is Info
- Errors are always logged regardless of verbosity

### Output Format

- Human-readable by default
- Structured JSON format (optional, for future machine parsing)
- Progress indicators for long-running operations

## Non-Goals (initial version)

- File logging (stdout/stderr only)
- Log rotation
- Remote log shipping
- Complex structured formats (keep it simple)

## Tests

See `spec/features.yaml` entry for `CORE_LOGGING`:
- `pkg/logging/logging_test.go` – unit tests for:
  - Log level filtering
  - Field attachment
  - Verbose flag behavior

