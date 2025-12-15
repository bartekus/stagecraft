// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

package main

import (
	"reflect"
	"testing"
)

func TestAnalyze(t *testing.T) {
	tests := []struct {
		name     string
		input    Inputs
		expected FailureClass
	}{
		{
			name:     "Success",
			input:    Inputs{ExitCode: 0, Stderr: ""},
			expected: Success,
		},
		{
			name:     "Internal Panic",
			input:    Inputs{ExitCode: 2, Stderr: "panic: nil pointer dereference"}, // Code shouldn't matter if panic found
			expected: InternalInvariant,
		},
		{
			name:     "Internal Exit Code",
			input:    Inputs{ExitCode: 3, Stderr: "something went wrong"},
			expected: InternalInvariant,
		},
		{
			name:     "Provider Error Generic",
			input:    Inputs{ExitCode: 1, Stderr: "api error: 500 internal server error"},
			expected: ProviderFailure,
		},
		{
			name:     "Provider Error Tailscale",
			input:    Inputs{ExitCode: 1, Stderr: "tailscale: socket not connected"},
			expected: ProviderFailure,
		},
		{
			name:     "Dependency Error",
			input:    Inputs{ExitCode: 127, Stderr: "docker: command not found"},
			expected: ExternalDependency,
		},
		{
			name:     "Config Error",
			input:    Inputs{ExitCode: 1, Stderr: "yaml: cannot unmarshal"},
			expected: ConfigInvalid,
		},
		{
			name:     "User Error Flag",
			input:    Inputs{ExitCode: 1, Stderr: "unknown flag: --foo"},
			expected: UserInput,
		},
		{
			name:     "Unclassified",
			input:    Inputs{ExitCode: 99, Stderr: "something weird happened"},
			expected: Unclassified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := analyze(tt.input)
			if got.FailureClass != tt.expected {
				t.Errorf("analyze() FailureClass = %v, want %v", got.FailureClass, tt.expected)
			}
			// Determinism check: verify fields are never nil
			if got.NextActions == nil {
				t.Error("analyze() NextActions is nil, want empty slice")
			}
			if got.DecisionRefs == nil {
				t.Error("analyze() DecisionRefs is nil, want empty slice")
			}
			if got.SpecRefs == nil {
				t.Error("analyze() SpecRefs is nil, want empty slice")
			}

			// Verify Decision-002 ref is present for errors
			if tt.expected != Success {
				hasDecision := false
				for _, d := range got.DecisionRefs {
					if d == "DECISION-002" {
						hasDecision = true
						break
					}
				}
				if !hasDecision {
					t.Error("analyze() DecisionRefs missing DECISION-002")
				}
			}
		})
	}
}

func TestAnalyzeOrdering(t *testing.T) {
	// Test priority: Internal > Provider
	input := Inputs{
		ExitCode: 3,                     // Internal
		Stderr:   "tailscale api error", // Provider keyword
	}
	got := analyze(input)
	if got.FailureClass != InternalInvariant {
		t.Errorf("Priority check failed: got %v, want %v", got.FailureClass, InternalInvariant)
	}
}

func TestOutputDeterminism(t *testing.T) {
	input := Inputs{ExitCode: 1, Stderr: "unknown flag"}
	out1 := analyze(input)
	out2 := analyze(input)

	if !reflect.DeepEqual(out1, out2) {
		t.Error("analyze() is not deterministic")
	}
}
