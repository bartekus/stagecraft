// Package migrations defines the interface and types for migration engines.
//
// SPDX-License-Identifier: AGPL-3.0-or-later
package migrations

import "fmt"

// ErrorKind classifies the type of migration error.
type ErrorKind string

const (
	// ErrInvalidConfig indicates a configuration error.
	ErrInvalidConfig ErrorKind = "invalid_config"
	// ErrUnsupported indicates the operation is not supported by the engine.
	ErrUnsupported ErrorKind = "unsupported"
	// ErrDependencyMissing indicates a required dependency is missing.
	ErrDependencyMissing ErrorKind = "dependency_missing"
	// ErrConnectionFailed indicates a failure to connect to the target.
	ErrConnectionFailed ErrorKind = "connection_failed"
	// ErrMigrationFailed indicates a failure during migration execution.
	ErrMigrationFailed ErrorKind = "migration_failed"
	// ErrInternal indicates an internal error.
	ErrInternal ErrorKind = "internal"
)

// MigrationError is a structured error for migration operations.
type MigrationError struct {
	Kind    ErrorKind   `json:"kind"`
	Message string      `json:"message"`
	Cause   error       `json:"-"` // underlying error, not marshaled
	StepID  MigrationID `json:"step_id,omitempty"`
}

func (e *MigrationError) Error() string {
	if e.StepID != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Kind, e.StepID, e.Message)
	}
	return fmt.Sprintf("[%s] %s", e.Kind, e.Message)
}

func (e *MigrationError) Unwrap() error {
	return e.Cause
}
