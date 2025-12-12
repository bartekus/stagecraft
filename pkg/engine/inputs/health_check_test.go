// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "testing"

func TestHealthCheckInputs_OneOfEndpointsOrServices(t *testing.T) {
	tests := []struct {
		name    string
		in      *HealthCheckInputs
		wantErr bool
	}{
		{
			name: "valid with endpoints",
			in: &HealthCheckInputs{
				Environment: "prod",
				Endpoints: []HealthEndpoint{
					{Name: "api", URL: "http://localhost:8080/health", ExpectedStatus: 200, Method: "GET"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid with services",
			in: &HealthCheckInputs{
				Environment: "prod",
				Services:    []string{"api", "web"},
			},
			wantErr: false,
		},
		{
			name: "error: both provided",
			in: &HealthCheckInputs{
				Environment: "prod",
				Endpoints: []HealthEndpoint{
					{Name: "api", URL: "http://localhost:8080/health", ExpectedStatus: 200, Method: "GET"},
				},
				Services: []string{"api"},
			},
			wantErr: true,
		},
		{
			name: "error: neither provided",
			in: &HealthCheckInputs{
				Environment: "prod",
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
