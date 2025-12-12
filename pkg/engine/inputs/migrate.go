// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "fmt"

type MigrateInputs struct {
	Database string `json:"database"`
	Strategy string `json:"strategy"`
	Engine   string `json:"engine"`
	Path     string `json:"path"`
	ConnEnv  string `json:"conn_env"`

	TimeoutSeconds int      `json:"timeout_seconds,omitempty"`
	Args           []string `json:"args,omitempty"` // order significant; do not sort
}

func (in *MigrateInputs) Normalize() error {
	in.Database = NormalizeString(in.Database)
	in.Strategy = NormalizeString(in.Strategy)
	in.Engine = NormalizeString(in.Engine)
	in.Path = NormalizeString(in.Path)
	in.ConnEnv = NormalizeString(in.ConnEnv)
	if in.Args != nil {
		for i := range in.Args {
			in.Args[i] = NormalizeString(in.Args[i])
		}
	}
	var err error
	in.Path, err = PathNormalize(in.Path)
	if err != nil {
		return fmt.Errorf("path: %w", err)
	}
	return nil
}

func (in *MigrateInputs) Validate() error {
	if in.Database == "" {
		return fmt.Errorf("database is required")
	}
	if in.Strategy == "" {
		return fmt.Errorf("strategy is required")
	}
	if in.Engine == "" {
		return fmt.Errorf("engine is required")
	}
	if in.Path == "" {
		return fmt.Errorf("path is required")
	}
	if in.ConnEnv == "" {
		return fmt.Errorf("conn_env is required")
	}
	if in.TimeoutSeconds != 0 && in.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0 if present")
	}
	return nil
}
