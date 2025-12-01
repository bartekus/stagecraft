// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - A Go-based CLI for orchestrating local-first multi-service deployments using Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package core

import (
	"fmt"

	"stagecraft/pkg/config"
)

// Feature: CORE_PLAN
// Spec: spec/core/plan.md

// Plan represents a deployment plan with operations to execute.
type Plan struct {
	Environment string
	Operations  []Operation
}

// Operation represents a single step in a deployment plan.
type Operation struct {
	Type        OperationType
	Description string
	Dependencies []string // IDs of operations that must complete first
	Metadata    map[string]interface{}
}

// OperationType represents the kind of operation.
type OperationType string

const (
	OpTypeInfraProvision OperationType = "infra_provision"
	OpTypeMigration      OperationType = "migration"
	OpTypeBuild          OperationType = "build"
	OpTypeDeploy         OperationType = "deploy"
	OpTypeHealthCheck    OperationType = "health_check"
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

	// Add migration operations (pre-deploy)
	if err := p.addMigrationOps(plan, "pre_deploy"); err != nil {
		return nil, fmt.Errorf("planning pre-deploy migrations: %w", err)
	}

	// Add build operations
	if err := p.addBuildOps(plan); err != nil {
		return nil, fmt.Errorf("planning builds: %w", err)
	}

	// Add deploy operations
	if err := p.addDeployOps(plan); err != nil {
		return nil, fmt.Errorf("planning deployment: %w", err)
	}

	// Add migration operations (post-deploy)
	if err := p.addMigrationOps(plan, "post_deploy"); err != nil {
		return nil, fmt.Errorf("planning post-deploy migrations: %w", err)
	}

	// Add health check operations
	if err := p.addHealthCheckOps(plan); err != nil {
		return nil, fmt.Errorf("planning health checks: %w", err)
	}

	return plan, nil
}

// addMigrationOps adds migration operations for the given strategy.
func (p *Planner) addMigrationOps(plan *Plan, strategy string) error {
	for dbName, dbCfg := range p.config.Databases {
		if dbCfg.Migrations == nil {
			continue
		}

		if dbCfg.Migrations.Strategy != strategy {
			continue
		}

		opID := fmt.Sprintf("migration_%s_%s", dbName, strategy)
		plan.Operations = append(plan.Operations, Operation{
			Type:        OpTypeMigration,
			Description: fmt.Sprintf("Run %s migrations for database %s", strategy, dbName),
			Dependencies: []string{},
			Metadata: map[string]interface{}{
				"database":  dbName,
				"strategy": strategy,
				"engine":   dbCfg.Migrations.Engine,
				"path":     dbCfg.Migrations.Path,
				"conn_env": dbCfg.ConnectionEnv,
			},
		})
		_ = opID // For future dependency tracking
	}

	return nil
}

// addBuildOps adds build operations.
func (p *Planner) addBuildOps(plan *Plan) error {
	if p.config.Backend != nil {
		opID := "build_backend"
		plan.Operations = append(plan.Operations, Operation{
			Type:        OpTypeBuild,
			Description: fmt.Sprintf("Build backend using provider %s", p.config.Backend.Provider),
			Dependencies: []string{},
			Metadata: map[string]interface{}{
				"provider": p.config.Backend.Provider,
			},
		})
		_ = opID // For future dependency tracking
	}

	return nil
}

// addDeployOps adds deployment operations.
func (p *Planner) addDeployOps(plan *Plan) error {
	plan.Operations = append(plan.Operations, Operation{
		Type:        OpTypeDeploy,
		Description: fmt.Sprintf("Deploy to environment %s", plan.Environment),
		Dependencies: []string{}, // Will depend on builds and pre-deploy migrations
		Metadata: map[string]interface{}{
			"environment": plan.Environment,
		},
	})

	return nil
}

// addHealthCheckOps adds health check operations.
func (p *Planner) addHealthCheckOps(plan *Plan) error {
	plan.Operations = append(plan.Operations, Operation{
		Type:        OpTypeHealthCheck,
		Description: fmt.Sprintf("Health check for environment %s", plan.Environment),
		Dependencies: []string{}, // Will depend on deployment
		Metadata: map[string]interface{}{
			"environment": plan.Environment,
		},
	})

	return nil
}

