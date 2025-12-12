// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "testing"

func TestBuildInputs_NormalizeSorts(t *testing.T) {
	in := &BuildInputs{
		Provider:   "generic",
		Workdir:    "apps/backend",
		Dockerfile: "Dockerfile",
		Context:    ".",
		Tags:       []string{"z-tag", "a-tag", "m-tag"},
		BuildArgs: []BuildArg{
			{Key: "Z_VAR", Value: "z-value"},
			{Key: "A_VAR", Value: "a-value"},
			{Key: "M_VAR", Value: "m-value"},
		},
		Labels: []BuildLabel{
			{Key: "z-label", Value: "z-value"},
			{Key: "a-label", Value: "a-value"},
			{Key: "m-label", Value: "m-value"},
		},
	}

	if err := in.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}

	// Tags should be sorted
	expectedTags := []string{"a-tag", "m-tag", "z-tag"}
	if len(in.Tags) != len(expectedTags) {
		t.Fatalf("tags length mismatch: got %d, want %d", len(in.Tags), len(expectedTags))
	}
	for i, want := range expectedTags {
		if in.Tags[i] != want {
			t.Errorf("tags[%d] = %q, want %q", i, in.Tags[i], want)
		}
	}

	// BuildArgs should be sorted by key
	if in.BuildArgs[0].Key != "A_VAR" {
		t.Errorf("build_args[0].key = %q, want %q", in.BuildArgs[0].Key, "A_VAR")
	}
	if in.BuildArgs[1].Key != "M_VAR" {
		t.Errorf("build_args[1].key = %q, want %q", in.BuildArgs[1].Key, "M_VAR")
	}
	if in.BuildArgs[2].Key != "Z_VAR" {
		t.Errorf("build_args[2].key = %q, want %q", in.BuildArgs[2].Key, "Z_VAR")
	}

	// Labels should be sorted by key
	if in.Labels[0].Key != "a-label" {
		t.Errorf("labels[0].key = %q, want %q", in.Labels[0].Key, "a-label")
	}
	if in.Labels[1].Key != "m-label" {
		t.Errorf("labels[1].key = %q, want %q", in.Labels[1].Key, "m-label")
	}
	if in.Labels[2].Key != "z-label" {
		t.Errorf("labels[2].key = %q, want %q", in.Labels[2].Key, "z-label")
	}
}

func TestBuildInputs_Validate(t *testing.T) {
	tests := []struct {
		name    string
		in      *BuildInputs
		wantErr bool
	}{
		{
			name: "valid inputs",
			in: &BuildInputs{
				Provider:   "generic",
				Workdir:    "apps/backend",
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
			wantErr: false,
		},
		{
			name: "missing provider",
			in: &BuildInputs{
				Workdir:    "apps/backend",
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
			wantErr: true,
		},
		{
			name: "missing dockerfile",
			in: &BuildInputs{
				Provider: "generic",
				Workdir:  "apps/backend",
				Context:  ".",
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
