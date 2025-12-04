// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package encorets provides the Encore.ts backend provider implementation.
package encorets

import (
	"errors"
	"fmt"
)

// Feature: PROVIDER_BACKEND_ENCORE
// Spec: spec/providers/backend/encore-ts.md

// Error categories
const (
	ErrProviderNotAvailable = "PROVIDER_NOT_AVAILABLE"
	ErrInvalidConfig        = "INVALID_CONFIG"
	ErrInvalidProject       = "INVALID_PROJECT"
	ErrSecretSyncFailed     = "SECRET_SYNC_FAILED"
	ErrDevServerFailed      = "DEV_SERVER_FAILED"
	ErrBuildFailed          = "BUILD_FAILED"
)

// ProviderError represents an error from the Encore.ts provider
type ProviderError struct {
	Category  string
	Provider  string
	Operation string
	Message   string
	Detail    string
	Err       error
}

func (e *ProviderError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s/%s/%s] %s: %s",
			e.Provider, e.Operation, e.Category, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s/%s/%s] %s",
		e.Provider, e.Operation, e.Category, e.Message)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// Helper functions

// IsProviderError checks if an error is a ProviderError
func IsProviderError(err error) bool {
	var pe *ProviderError
	return errors.As(err, &pe)
}

// GetProviderError extracts a ProviderError from an error chain
func GetProviderError(err error) *ProviderError {
	var pe *ProviderError
	if errors.As(err, &pe) {
		return pe
	}
	return nil
}

