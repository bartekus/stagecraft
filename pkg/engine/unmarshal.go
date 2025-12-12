// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UnmarshalStrictPlan unmarshals a Plan with strict field validation (rejects unknown fields).
func UnmarshalStrictPlan(data []byte, plan *Plan) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(plan); err != nil {
		return fmt.Errorf("strict decode plan: %w", err)
	}
	// Ensure there's no trailing junk.
	if dec.More() {
		return fmt.Errorf("strict decode plan: trailing tokens")
	}
	return nil
}

// UnmarshalStrictHostPlan unmarshals a HostPlan with strict field validation (rejects unknown fields).
func UnmarshalStrictHostPlan(data []byte, plan *HostPlan) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(plan); err != nil {
		return fmt.Errorf("strict decode host plan: %w", err)
	}
	// Ensure there's no trailing junk.
	if dec.More() {
		return fmt.Errorf("strict decode host plan: trailing tokens")
	}
	return nil
}
