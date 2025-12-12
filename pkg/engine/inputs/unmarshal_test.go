// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "testing"

func TestUnmarshalStrict_RejectsUnknown(t *testing.T) {
	// Valid JSON with an extra unknown field
	jsonBytes := []byte(`{
		"provider": "generic",
		"workdir": "apps/backend",
		"dockerfile": "Dockerfile",
		"context": ".",
		"unknown_field": "should be rejected"
	}`)

	var in BuildInputs
	err := UnmarshalStrict(jsonBytes, &in)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestUnmarshalStrict_AcceptsValid(t *testing.T) {
	jsonBytes := []byte(`{
		"provider": "generic",
		"workdir": "apps/backend",
		"dockerfile": "Dockerfile",
		"context": "."
	}`)

	var in BuildInputs
	err := UnmarshalStrict(jsonBytes, &in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if in.Provider != "generic" {
		t.Errorf("provider = %q, want %q", in.Provider, "generic")
	}
}

func TestUnmarshalStrict_RejectsTrailingTokens(t *testing.T) {
	// Valid object followed by extra tokens
	jsonBytes := []byte(`{"provider": "generic", "workdir": "apps/backend", "dockerfile": "Dockerfile", "context": "."} extra tokens`)

	var in BuildInputs
	err := UnmarshalStrict(jsonBytes, &in)
	if err == nil {
		t.Fatal("expected error for trailing tokens")
	}
}
