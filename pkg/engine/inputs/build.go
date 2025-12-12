// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import "fmt"

// BuildInputs defines inputs for a build step.
type BuildInputs struct {
	Provider   string       `json:"provider"`
	Workdir    string       `json:"workdir"`
	Target     string       `json:"target,omitempty"`
	Dockerfile string       `json:"dockerfile"`
	Context    string       `json:"context"`
	Tags       []string     `json:"tags,omitempty"`
	BuildArgs  []BuildArg   `json:"build_args,omitempty"`
	Labels     []BuildLabel `json:"labels,omitempty"`
}

// Normalize canonicalizes BuildInputs fields.
func (in *BuildInputs) Normalize() error {
	in.Provider = NormalizeString(in.Provider)
	in.Workdir = NormalizeString(in.Workdir)
	in.Target = NormalizeString(in.Target)
	in.Dockerfile = NormalizeString(in.Dockerfile)
	in.Context = NormalizeString(in.Context)

	if in.Tags != nil {
		NormalizeTags(in.Tags)
	}
	if in.BuildArgs != nil {
		// sort by key
		NormalizeKV(in.BuildArgs)
		for i := range in.BuildArgs {
			in.BuildArgs[i].Key = NormalizeString(in.BuildArgs[i].Key)
			in.BuildArgs[i].Value = NormalizeString(in.BuildArgs[i].Value)
		}
	}
	if in.Labels != nil {
		NormalizeKV(in.Labels)
		for i := range in.Labels {
			in.Labels[i].Key = NormalizeString(in.Labels[i].Key)
			in.Labels[i].Value = NormalizeString(in.Labels[i].Value)
		}
	}

	var err error
	if in.Workdir != "" {
		in.Workdir, err = PathNormalize(in.Workdir)
		if err != nil {
			return fmt.Errorf("workdir: %w", err)
		}
	}
	if in.Dockerfile != "" {
		in.Dockerfile, err = PathNormalize(in.Dockerfile)
		if err != nil {
			return fmt.Errorf("dockerfile: %w", err)
		}
	}
	if in.Context != "" {
		in.Context, err = PathNormalize(in.Context)
		if err != nil {
			return fmt.Errorf("context: %w", err)
		}
	}

	return nil
}

// Validate validates BuildInputs according to v1 rules.
func (in *BuildInputs) Validate() error {
	if in.Provider == "" {
		return fmt.Errorf("provider is required")
	}
	if in.Workdir == "" {
		return fmt.Errorf("workdir is required")
	}
	if in.Dockerfile == "" {
		return fmt.Errorf("dockerfile is required (producer must set explicitly)")
	}
	if in.Context == "" {
		return fmt.Errorf("context is required (producer must set explicitly)")
	}
	for _, t := range in.Tags {
		if NormalizeString(t) == "" {
			return fmt.Errorf("tags contains empty value")
		}
	}
	for _, a := range in.BuildArgs {
		if a.Key == "" {
			return fmt.Errorf("build_args.key is required")
		}
	}
	for _, l := range in.Labels {
		if l.Key == "" {
			return fmt.Errorf("labels.key is required")
		}
	}
	return nil
}
