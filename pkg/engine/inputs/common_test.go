// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "testing"

// boolPtr is a test helper for creating bool pointers.
func boolPtr(b bool) *bool {
	return &b
}

func TestPathNormalize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			input:   "apps/backend",
			want:    "apps/backend",
			wantErr: false,
		},
		{
			name:    "normalizes backslashes",
			input:   "apps\\backend",
			want:    "apps/backend",
			wantErr: false,
		},
		{
			name:    "normalizes multiple slashes",
			input:   "apps//backend",
			want:    "apps/backend",
			wantErr: false,
		},
		{
			name:    "rejects absolute path",
			input:   "/apps/backend",
			wantErr: true,
		},
		{
			name:    "rejects Windows absolute path",
			input:   "C:\\apps\\backend",
			wantErr: true,
		},
		{
			name:    "rejects path with ..",
			input:   "apps/../backend",
			wantErr: true,
		},
		{
			name:    "rejects path with .",
			input:   "apps/./backend",
			wantErr: true,
		},
		{
			name:    "rejects empty path",
			input:   "",
			wantErr: true,
		},
		{
			name:    "rejects path starting with ~",
			input:   "~/apps/backend",
			wantErr: true,
		},
		{
			name:    "allows standalone .",
			input:   ".",
			want:    ".",
			wantErr: false,
		},
		{
			name:    "rejects standalone ..",
			input:   "..",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PathNormalize(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("PathNormalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("PathNormalize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateSha256Hex64(t *testing.T) {
	tests := []struct {
		name    string
		hash    string
		wantErr bool
	}{
		{
			name:    "valid sha256 hash",
			hash:    "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
			wantErr: false,
		},
		{
			name:    "rejects uppercase",
			hash:    "A3B2C1D4E5F6A7B8C9D0E1F2A3B4C5D6E7F8A9B0C1D2E3F4A5B6C7D8E9F0A1B2",
			wantErr: true,
		},
		{
			name:    "rejects wrong length",
			hash:    "a3b2c1",
			wantErr: true,
		},
		{
			name:    "rejects non-hex",
			hash:    "a3b2c1d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1bg",
			wantErr: true,
		},
		{
			name:    "rejects empty",
			hash:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSha256Hex64(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSha256Hex64() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
