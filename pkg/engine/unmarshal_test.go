// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package engine

import (
	"strings"
	"testing"
)

func TestUnmarshalStrictHostPlan_RejectsUnknownFields(t *testing.T) {
	// Valid HostPlan JSON with an extra unknown field
	jsonBytes := []byte(`{
		"version": "v1",
		"planId": "test-plan",
		"host": {"logicalId": "host-a"},
		"steps": [],
		"unknown_field": "should be rejected"
	}`)

	var plan HostPlan
	err := UnmarshalStrictHostPlan(jsonBytes, &plan, "test-plan")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestUnmarshalStrictHostPlan_RejectsTrailingTokens(t *testing.T) {
	// Valid HostPlan JSON followed by extra tokens
	jsonBytes := []byte(`{
		"version": "v1",
		"planId": "test-plan",
		"host": {"logicalId": "host-a"},
		"steps": []
	} extra tokens`)

	var plan HostPlan
	err := UnmarshalStrictHostPlan(jsonBytes, &plan, "test-plan")
	if err == nil {
		t.Fatal("expected error for trailing tokens")
	}
}

func TestUnmarshalStrictHostPlan_AcceptsValid(t *testing.T) {
	jsonBytes := []byte(`{
		"version": "v1",
		"planId": "test-plan",
		"host": {"logicalId": "host-a"},
		"steps": []
	}`)

	var plan HostPlan
	err := UnmarshalStrictHostPlan(jsonBytes, &plan, "test-plan")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.PlanID != "test-plan" {
		t.Errorf("planId = %q, want %q", plan.PlanID, "test-plan")
	}
	if plan.Host.LogicalID != "host-a" {
		t.Errorf("host.logicalId = %q, want %q", plan.Host.LogicalID, "host-a")
	}
}

func TestUnmarshalStrictHostPlan_ErrorContext(t *testing.T) {
	// Invalid JSON with planID in context
	jsonBytes := []byte(`{
		"version": "v1",
		"planId": "my-plan-123",
		"host": {"logicalId": "host-a"},
		"steps": [],
		"bad_field": true
	}`)

	var plan HostPlan
	err := UnmarshalStrictHostPlan(jsonBytes, &plan, "my-plan-123")
	if err == nil {
		t.Fatal("expected error for unknown field")
	}

	// Error should include planID context
	errStr := err.Error()
	if errStr == "" {
		t.Fatal("error message should not be empty")
	}
	// Error should mention planId in context
	if !strings.Contains(errStr, "planId") || !strings.Contains(errStr, "my-plan-123") {
		t.Errorf("error message should contain planId context, got: %q", errStr)
	}
}

func TestUnmarshalStrictPlan_RejectsUnknownFields(t *testing.T) {
	jsonBytes := []byte(`{
		"version": "v1",
		"id": "test-plan",
		"steps": [],
		"unknown_field": "should be rejected"
	}`)

	var plan Plan
	err := UnmarshalStrictPlan(jsonBytes, &plan)
	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestUnmarshalStrictPlan_RejectsTrailingTokens(t *testing.T) {
	jsonBytes := []byte(`{
		"version": "v1",
		"id": "test-plan",
		"steps": []
	} extra tokens`)

	var plan Plan
	err := UnmarshalStrictPlan(jsonBytes, &plan)
	if err == nil {
		t.Fatal("expected error for trailing tokens")
	}
}

func TestUnmarshalStrictPlan_AcceptsValid(t *testing.T) {
	jsonBytes := []byte(`{
		"version": "v1",
		"id": "test-plan",
		"steps": []
	}`)

	var plan Plan
	err := UnmarshalStrictPlan(jsonBytes, &plan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan.ID != "test-plan" {
		t.Errorf("id = %q, want %q", plan.ID, "test-plan")
	}
}
