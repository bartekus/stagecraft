// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// UnmarshalStrict unmarshals JSON data with strict validation:
// - disallows unknown fields
// - rejects trailing tokens after the main object
func UnmarshalStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("strict decode: %w", err)
	}
	// Ensure there's no trailing junk.
	if dec.More() {
		return fmt.Errorf("strict decode: trailing tokens")
	}
	// A second Decode should hit EOF.
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return fmt.Errorf("strict decode: extra JSON after object")
	}
	return nil
}
