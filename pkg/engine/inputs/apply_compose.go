// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "fmt"

type ApplyComposeInputs struct {
	Environment string `json:"environment"`
	ComposePath string `json:"compose_path"`
	ProjectName string `json:"project_name"`

	// Producer must set explicitly; pointers enforce "presence" across JSON.
	Pull   *bool `json:"pull"`
	Detach *bool `json:"detach"`

	Services []string `json:"services,omitempty"`

	ExpectedComposeHashAlg string `json:"expected_compose_hash_alg,omitempty"`
	ExpectedComposeHash    string `json:"expected_compose_hash,omitempty"`
}

func (in *ApplyComposeInputs) Normalize() error {
	in.Environment = NormalizeString(in.Environment)
	in.ComposePath = NormalizeString(in.ComposePath)
	in.ProjectName = NormalizeString(in.ProjectName)
	in.ExpectedComposeHashAlg = NormalizeString(in.ExpectedComposeHashAlg)
	in.ExpectedComposeHash = NormalizeString(in.ExpectedComposeHash)

	if in.Services != nil {
		for i := range in.Services {
			in.Services[i] = NormalizeString(in.Services[i])
		}
		NormalizeTags(in.Services) // just a lex sort
	}

	var err error
	in.ComposePath, err = PathNormalize(in.ComposePath)
	if err != nil {
		return fmt.Errorf("compose_path: %w", err)
	}

	return nil
}

func (in *ApplyComposeInputs) Validate() error {
	if in.Environment == "" {
		return fmt.Errorf("environment is required")
	}
	if in.ComposePath == "" {
		return fmt.Errorf("compose_path is required")
	}
	if in.ProjectName == "" {
		return fmt.Errorf("project_name is required")
	}
	if in.Pull == nil {
		return fmt.Errorf("pull is required (producer must set explicitly)")
	}
	if in.Detach == nil {
		return fmt.Errorf("detach is required (producer must set explicitly)")
	}

	for _, s := range in.Services {
		if s == "" {
			return fmt.Errorf("services contains empty value")
		}
	}

	if in.ExpectedComposeHashAlg != "" || in.ExpectedComposeHash != "" {
		if in.ExpectedComposeHashAlg != "sha256" {
			return fmt.Errorf("expected_compose_hash_alg must be 'sha256' in v1")
		}
		if err := ValidateSha256Hex64(in.ExpectedComposeHash); err != nil {
			return fmt.Errorf("expected_compose_hash: %w", err)
		}
	}

	return nil
}
