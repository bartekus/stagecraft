// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "testing"

func TestApplyComposeInputs_RequiresPullDetach(t *testing.T) {
	tests := []struct {
		name    string
		in      *ApplyComposeInputs
		wantErr bool
	}{
		{
			name: "valid with explicit pull and detach",
			in: &ApplyComposeInputs{
				Environment: "prod",
				ComposePath: "compose.yml",
				ProjectName: "test-app",
				Pull:        boolPtr(true),
				Detach:      boolPtr(true),
			},
			wantErr: false,
		},
		{
			name: "missing pull",
			in: &ApplyComposeInputs{
				Environment: "prod",
				ComposePath: "compose.yml",
				ProjectName: "test-app",
				Detach:      boolPtr(true),
			},
			wantErr: true,
		},
		{
			name: "missing detach",
			in: &ApplyComposeInputs{
				Environment: "prod",
				ComposePath: "compose.yml",
				ProjectName: "test-app",
				Pull:        boolPtr(true),
			},
			wantErr: true,
		},
		{
			name: "both missing",
			in: &ApplyComposeInputs{
				Environment: "prod",
				ComposePath: "compose.yml",
				ProjectName: "test-app",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.in.Normalize(); err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			err := tt.in.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyComposeInputs_SortsServices(t *testing.T) {
	in := &ApplyComposeInputs{
		Environment: "prod",
		ComposePath: "compose.yml",
		ProjectName: "test-app",
		Pull:        boolPtr(true),
		Detach:      boolPtr(true),
		Services:    []string{"z-service", "a-service", "m-service"},
	}

	if err := in.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	expected := []string{"a-service", "m-service", "z-service"}
	if len(in.Services) != len(expected) {
		t.Fatalf("services length mismatch: got %d, want %d", len(in.Services), len(expected))
	}
	for i, want := range expected {
		if in.Services[i] != want {
			t.Errorf("services[%d] = %q, want %q", i, in.Services[i], want)
		}
	}
}
