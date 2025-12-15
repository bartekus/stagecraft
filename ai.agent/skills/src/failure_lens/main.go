// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

// Package main implements the failure_lens skill for deterministic failure classification.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// FailureClass enum based on GOV_CLI_EXIT_CODES
type FailureClass string

const (
	Success              FailureClass = "success"
	UserInput            FailureClass = "user_input"
	ConfigInvalid        FailureClass = "config_invalid"
	ExternalDependency   FailureClass = "external_dependency"
	ProviderFailure      FailureClass = "provider_failure"
	TransientEnvironment FailureClass = "transient_environment"
	InternalInvariant    FailureClass = "internal_invariant"
	Unclassified         FailureClass = "unclassified"
)

// Inputs defines the expected JSON input
type Inputs struct {
	Command   string `json:"command"`
	ExitCode  int    `json:"exit_code"`
	Stderr    string `json:"stderr"`
	Subsystem string `json:"subsystem,omitempty"` // cli, engine, provider, runtime
}

// Outputs defines the deterministic JSON output
type Outputs struct {
	FailureClass  FailureClass `json:"failure_class"`
	ProbableCause string       `json:"probable_cause"`
	NextActions   []string     `json:"next_actions"`
	DecisionRefs  []string     `json:"decision_refs"`
	SpecRefs      []string     `json:"spec_refs"`
}

func main() {
	// Read inputs from stdin
	inputData, err := io.ReadAll(os.Stdin)
	if err != nil {
		fatalError(err)
	}

	var inputs Inputs
	if err := json.Unmarshal(inputData, &inputs); err != nil {
		fatalError(err)
	}

	// Calculate deterministic outputs
	outputs := analyze(inputs)

	// Write outputs to stdout
	if err := json.NewEncoder(os.Stdout).Encode(outputs); err != nil {
		fatalError(err)
	}
}

func analyze(in Inputs) Outputs {
	if in.ExitCode == 0 {
		return Outputs{
			FailureClass:  Success,
			ProbableCause: "Command executed successfully",
			NextActions:   []string{},
			DecisionRefs:  []string{},
			SpecRefs:      []string{},
		}
	}

	// Heuristics for classification
	// Determinism rule: InternalInvariant > Provider > Config/User > External > Unclassified
	// Most specific wins.

	stderrLower := strings.ToLower(in.Stderr)

	if isInternalError(stderrLower, in.ExitCode) {
		return Outputs{
			FailureClass:  InternalInvariant,
			ProbableCause: "Internal code invariant violated (panic or unexpected state)",
			NextActions: []string{
				"Run with STAGECRAFT_DEBUG=1",
				"Report bug with stacktrace",
			},
			DecisionRefs: []string{"DECISION-002"},
			SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
		}
	}

	if isProviderError(stderrLower) {
		cause := "External provider returned an error"
		if strings.Contains(stderrLower, "tailscale") {
			cause = "Tailscale API or daemon failure"
		} else if strings.Contains(stderrLower, "digitalocean") {
			cause = "DigitalOcean API failure"
		}

		return Outputs{
			FailureClass:  ProviderFailure,
			ProbableCause: cause,
			NextActions: []string{
				"Check provider credential validity",
				"Check provider service status",
			},
			DecisionRefs: []string{"DECISION-002"},
			SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
		}
	}

	if isDependencyError(stderrLower) {
		return Outputs{
			FailureClass:  ExternalDependency,
			ProbableCause: "Required external tool missing or failed",
			NextActions: []string{
				"Install missing dependency",
				"Check PATH configuration",
			},
			DecisionRefs: []string{"DECISION-002"},
			SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
		}
	}

	if isConfigError(stderrLower) {
		return Outputs{
			FailureClass:  ConfigInvalid,
			ProbableCause: "Configuration validation failed",
			NextActions: []string{
				"Check stagecraft.yml syntax",
				"Validate config schema",
			},
			DecisionRefs: []string{"DECISION-002"},
			SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
		}
	}

	if isUserError(stderrLower, in.ExitCode) {
		return Outputs{
			FailureClass:  UserInput,
			ProbableCause: "Invalid command arguments or flags",
			NextActions: []string{
				"Check help output for usage",
			},
			DecisionRefs: []string{"DECISION-002"},
			SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
		}
	}

	// Default/Fallback
	return Outputs{
		FailureClass:  Unclassified,
		ProbableCause: "Unknown error condition",
		NextActions: []string{
			"Check logs for more details",
		},
		DecisionRefs: []string{"DECISION-002"},
		SpecRefs:     []string{"spec/governance/GOV_CLI_EXIT_CODES.md"},
	}
}

func isInternalError(s string, code int) bool {
	if code == 3 {
		return true
	}
	if strings.Contains(s, "panic:") || strings.Contains(s, "segmentation fault") || strings.Contains(s, "nil pointer") {
		return true
	}
	return false
}

func isProviderError(s string) bool {
	keywords := []string{"tailscale", "digitalocean", "aws", "provider error", "api error", "401 unauthorized", "403 forbidden", "500 internal server error"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func isDependencyError(s string) bool {
	keywords := []string{"command not found", "executable not found", "no such file or directory", "docker is not running"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func isConfigError(s string) bool {
	keywords := []string{"yaml:", "cannot unmarshal", "config validation", "missing required field", "stagecraft.yml"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func isUserError(s string, code int) bool {
	if code == 1 {
		return true // By convention in existing tools, though config also uses 1
	}
	keywords := []string{"unknown flag", "unknown command", "invalid argument", "usage:"}
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func fatalError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}
