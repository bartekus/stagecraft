// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import (
	"fmt"
	"sort"
)

// RolloutInputs defines inputs for a rollout step.
type RolloutInputs struct {
	Mode string `json:"mode"`

	BatchSize int      `json:"batch_size,omitempty"`
	Targets   []string `json:"targets,omitempty"`
}

// Normalize canonicalizes RolloutInputs fields.
func (in *RolloutInputs) Normalize() error {
	in.Mode = NormalizeString(in.Mode)
	if in.Targets != nil {
		for i := range in.Targets {
			in.Targets[i] = NormalizeString(in.Targets[i])
		}
		sort.Strings(in.Targets)
	}
	return nil
}

// Validate validates RolloutInputs according to v1 rules.
func (in *RolloutInputs) Validate() error {
	if in.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	if in.BatchSize != 0 && in.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be > 0 if present")
	}
	for _, t := range in.Targets {
		if t == "" {
			return fmt.Errorf("targets contains empty value")
		}
	}
	return nil
}
