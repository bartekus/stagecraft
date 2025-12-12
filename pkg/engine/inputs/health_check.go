// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "fmt"

type HealthCheckInputs struct {
	Environment string `json:"environment"`

	// One of these must be provided
	Endpoints []HealthEndpoint `json:"endpoints,omitempty"`
	Services  []string         `json:"services,omitempty"`

	TimeoutSeconds  int `json:"timeout_seconds,omitempty"`
	IntervalSeconds int `json:"interval_seconds,omitempty"`
	Retries         int `json:"retries,omitempty"`
}

func (in *HealthCheckInputs) Normalize() error {
	in.Environment = NormalizeString(in.Environment)

	if in.Services != nil {
		for i := range in.Services {
			in.Services[i] = NormalizeString(in.Services[i])
		}
		NormalizeTags(in.Services)
	}
	if in.Endpoints != nil {
		NormalizeKV(in.Endpoints) // by name
		for i := range in.Endpoints {
			in.Endpoints[i].Name = NormalizeString(in.Endpoints[i].Name)
			in.Endpoints[i].URL = NormalizeString(in.Endpoints[i].URL)
			in.Endpoints[i].Method = NormalizeString(in.Endpoints[i].Method)

			if in.Endpoints[i].Headers != nil {
				NormalizeKV(in.Endpoints[i].Headers)
				for j := range in.Endpoints[i].Headers {
					in.Endpoints[i].Headers[j].Key = NormalizeString(in.Endpoints[i].Headers[j].Key)
					in.Endpoints[i].Headers[j].Value = NormalizeString(in.Endpoints[i].Headers[j].Value)
				}
			}
		}
	}

	return nil
}

func (in *HealthCheckInputs) Validate() error {
	if in.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	hasEndpoints := len(in.Endpoints) > 0
	hasServices := len(in.Services) > 0
	if hasEndpoints == hasServices {
		return fmt.Errorf("exactly one of endpoints or services must be provided")
	}

	if in.TimeoutSeconds != 0 && in.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0 if present")
	}
	if in.IntervalSeconds != 0 && in.IntervalSeconds <= 0 {
		return fmt.Errorf("interval_seconds must be > 0 if present")
	}
	if in.Retries < 0 {
		return fmt.Errorf("retries must be >= 0 if present")
	}

	for _, s := range in.Services {
		if s == "" {
			return fmt.Errorf("services contains empty value")
		}
	}

	for _, ep := range in.Endpoints {
		if ep.Name == "" {
			return fmt.Errorf("endpoints.name is required")
		}
		if ep.URL == "" {
			return fmt.Errorf("endpoints.url is required")
		}
		if ep.ExpectedStatus <= 0 {
			return fmt.Errorf("endpoints.expected_status must be a valid HTTP status")
		}
		if ep.Method == "" {
			return fmt.Errorf("endpoints.method is required (producer must set explicitly)")
		}
		for _, h := range ep.Headers {
			if h.Key == "" {
				return fmt.Errorf("endpoints.headers.key is required")
			}
		}
	}

	return nil
}
