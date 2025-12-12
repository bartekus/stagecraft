// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "fmt"

// RenderComposeInputs defines inputs for rendering a compose file.
type RenderComposeInputs struct {
	Environment string `json:"environment"`

	// One of these must be provided
	BaseComposePath   string `json:"base_compose_path,omitempty"`
	BaseComposeInline string `json:"base_compose_inline,omitempty"`

	Overlays  []ComposeOverlay `json:"overlays,omitempty"`
	Variables []ComposeVar     `json:"variables,omitempty"`

	OutputPath string `json:"output_path"`

	ExpectedComposeHashAlg string `json:"expected_compose_hash_alg,omitempty"`
	ExpectedComposeHash    string `json:"expected_compose_hash,omitempty"`
}

// Normalize canonicalizes RenderComposeInputs fields.
func (in *RenderComposeInputs) Normalize() error {
	in.Environment = NormalizeString(in.Environment)
	in.BaseComposePath = NormalizeString(in.BaseComposePath)
	in.BaseComposeInline = stringsTrimIfSet(in.BaseComposeInline)
	in.OutputPath = NormalizeString(in.OutputPath)
	in.ExpectedComposeHashAlg = NormalizeString(in.ExpectedComposeHashAlg)
	in.ExpectedComposeHash = NormalizeString(in.ExpectedComposeHash)

	if in.Overlays != nil {
		NormalizeKV(in.Overlays)
		for i := range in.Overlays {
			in.Overlays[i].Name = NormalizeString(in.Overlays[i].Name)
			in.Overlays[i].Path = NormalizeString(in.Overlays[i].Path)
			p, err := PathNormalize(in.Overlays[i].Path)
			if err != nil {
				return fmt.Errorf("overlays[%d].path: %w", i, err)
			}
			in.Overlays[i].Path = p
		}
	}
	if in.Variables != nil {
		NormalizeKV(in.Variables)
		for i := range in.Variables {
			in.Variables[i].Key = NormalizeString(in.Variables[i].Key)
			in.Variables[i].Value = NormalizeString(in.Variables[i].Value)
		}
	}

	var err error
	if in.BaseComposePath != "" {
		in.BaseComposePath, err = PathNormalize(in.BaseComposePath)
		if err != nil {
			return fmt.Errorf("base_compose_path: %w", err)
		}
	}
	in.OutputPath, err = PathNormalize(in.OutputPath)
	if err != nil {
		return fmt.Errorf("output_path: %w", err)
	}

	return nil
}

// Validate validates RenderComposeInputs according to v1 rules.
func (in *RenderComposeInputs) Validate() error {
	if in.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if in.OutputPath == "" {
		return fmt.Errorf("output_path is required")
	}

	hasPath := in.BaseComposePath != ""
	hasInline := in.BaseComposeInline != ""
	if hasPath == hasInline {
		return fmt.Errorf("exactly one of base_compose_path or base_compose_inline must be provided")
	}

	if in.ExpectedComposeHashAlg != "" || in.ExpectedComposeHash != "" {
		if in.ExpectedComposeHashAlg != "sha256" {
			return fmt.Errorf("expected_compose_hash_alg must be 'sha256' in v1")
		}
		if err := ValidateSha256Hex64(in.ExpectedComposeHash); err != nil {
			return fmt.Errorf("expected_compose_hash: %w", err)
		}
	}

	for _, o := range in.Overlays {
		if o.Name == "" {
			return fmt.Errorf("overlays.name is required")
		}
		if o.Path == "" {
			return fmt.Errorf("overlays.path is required")
		}
	}
	for _, v := range in.Variables {
		if v.Key == "" {
			return fmt.Errorf("variables.key is required")
		}
	}

	return nil
}

// Inline compose can be multiline; trim only outer whitespace to satisfy "no leading/trailing".
func stringsTrimIfSet(s string) string {
	if s == "" {
		return ""
	}
	return NormalizeString(s)
}
