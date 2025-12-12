// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Package core contains core domain logic such as planning deployment operations.
package core

import (
	"fmt"
	"sort"

	"stagecraft/pkg/config"
)

// Feature: CORE_PLAN
// Spec: spec/core/plan.md

// Plan represents a deployment plan with operations to execute.
type Plan struct {
	Environment string
	Operations  []Operation
	Metadata    map[string]interface{} // Deployment context (version, config path, etc.)
}

// Operation represents a single step in a deployment plan.
type Operation struct {
	ID           string // Stable operation identifier for dependency references (required)
	Type         OperationType
	Description  string
	Dependencies []string // IDs of operations that must complete first
	Metadata     map[string]interface{}
}

// OperationType represents the kind of operation.
type OperationType string

const (
	// OpTypeInfraProvision represents infrastructure provisioning operations.
	OpTypeInfraProvision OperationType = "infra_provision"
	// OpTypeMigration represents database migration operations.
	OpTypeMigration OperationType = "migration"
	// OpTypeBuild represents build operations.
	OpTypeBuild OperationType = "build"
	// OpTypeDeploy represents deploy operations.
	OpTypeDeploy OperationType = "deploy"
	// OpTypeHealthCheck represents health check operations.
	OpTypeHealthCheck OperationType = "health_check"
)

// Planner creates deployment plans from configuration.
type Planner struct {
	config *config.Config
}

// NewPlanner creates a new planner for the given config.
func NewPlanner(cfg *config.Config) *Planner {
	return &Planner{
		config: cfg,
	}
}

// PlanDeploy creates a deployment plan for the given environment.
func (p *Planner) PlanDeploy(envName string) (*Plan, error) {
	_, ok := p.config.Environments[envName]
	if !ok {
		return nil, fmt.Errorf("environment %q not found in config", envName)
	}

	plan := &Plan{
		Environment: envName,
		Operations:  []Operation{},
	}

	// Track pre-deploy migration IDs for deploy dependencies
	var preDeployMigrationIDs []string

	// Add migration operations (pre-deploy) - returns sorted IDs
	migrationIDs := p.addMigrationOps(plan, "pre_deploy")
	preDeployMigrationIDs = append(preDeployMigrationIDs, migrationIDs...)
	// Ensure deterministic ordering
	sort.Strings(preDeployMigrationIDs)

	// Add build operations
	p.addBuildOps(plan)

	// Add deploy operations (depends on build + pre-deploy migrations)
	p.addDeployOps(plan, preDeployMigrationIDs)

	// Add migration operations (post-deploy)
	p.addMigrationOps(plan, "post_deploy")

	// Add health check operations (depends on deploy)
	p.addHealthCheckOps(plan)

	// Defensive check: ensure all operations have IDs
	for i, op := range plan.Operations {
		if op.ID == "" {
			return nil, fmt.Errorf("planner produced empty operation id at index %d", i)
		}
	}

	return plan, nil
}

// addMigrationOps adds migration operations for the given strategy.
// Returns the IDs of created operations for dependency tracking.
// Database names are sorted to ensure deterministic operation order.
func (p *Planner) addMigrationOps(plan *Plan, strategy string) []string {
	var opIDs []string

	// Sort database names for deterministic iteration order
	dbNames := make([]string, 0, len(p.config.Databases))
	for name := range p.config.Databases {
		dbNames = append(dbNames, name)
	}
	sort.Strings(dbNames)

	for _, dbName := range dbNames {
		dbCfg := p.config.Databases[dbName]
		if dbCfg.Migrations == nil {
			continue
		}

		if dbCfg.Migrations.Strategy != strategy {
			continue
		}

		opID := fmt.Sprintf("migration_%s_%s", dbName, strategy)
		opIDs = append(opIDs, opID)

		plan.Operations = append(plan.Operations, Operation{
			ID:           opID,
			Type:         OpTypeMigration,
			Description:  fmt.Sprintf("Run %s migrations for database %s", strategy, dbName),
			Dependencies: []string{},
			Metadata: map[string]interface{}{
				"database": dbName,
				"strategy": strategy,
				"engine":   dbCfg.Migrations.Engine,
				"path":     dbCfg.Migrations.Path,
				"conn_env": dbCfg.ConnectionEnv,
			},
		})
	}

	return opIDs
}

// addBuildOps adds build operations.
func (p *Planner) addBuildOps(plan *Plan) {
	if p.config.Backend != nil {
		opID := "build_backend"
		plan.Operations = append(plan.Operations, Operation{
			ID:           opID,
			Type:         OpTypeBuild,
			Description:  fmt.Sprintf("Build backend using provider %s", p.config.Backend.Provider),
			Dependencies: []string{},
			Metadata: map[string]interface{}{
				"provider": p.config.Backend.Provider,
			},
		})
	}
}

// addDeployOps adds deployment operations.
// preDeployMigrationIDs are the IDs of pre-deploy migration operations that must complete first.
// IDs are expected to be sorted for deterministic dependency ordering.
func (p *Planner) addDeployOps(plan *Plan, preDeployMigrationIDs []string) {
	opID := fmt.Sprintf("deploy_%s", plan.Environment)

	deps := []string{}
	if p.config.Backend != nil {
		deps = append(deps, "build_backend")
	}
	// Add all pre-deploy migration dependencies (already sorted)
	deps = append(deps, preDeployMigrationIDs...)

	plan.Operations = append(plan.Operations, Operation{
		ID:           opID,
		Type:         OpTypeDeploy,
		Description:  fmt.Sprintf("Deploy to environment %s", plan.Environment),
		Dependencies: deps,
		Metadata: map[string]interface{}{
			"environment": plan.Environment,
		},
	})
}

// addHealthCheckOps adds health check operations.
func (p *Planner) addHealthCheckOps(plan *Plan) {
	env := plan.Environment
	opID := fmt.Sprintf("health_check_%s", env)

	plan.Operations = append(plan.Operations, Operation{
		ID:           opID,
		Type:         OpTypeHealthCheck,
		Description:  fmt.Sprintf("Health check for environment %s", env),
		Dependencies: []string{fmt.Sprintf("deploy_%s", env)},
		Metadata: map[string]interface{}{
			"environment": env,
		},
	})
}
