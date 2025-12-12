// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"stagecraft/internal/core"
	"stagecraft/pkg/config"
	"stagecraft/pkg/logging"
	backendproviders "stagecraft/pkg/providers/backend"
)

// Feature: CLI_PLAN
// Spec: spec/commands/plan.md

// NewPlanCommand returns the `stagecraft plan` command.
func NewPlanCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Show the deployment plan without executing it",
		Long:  "Generates and displays a read-only deployment plan for the specified environment",
		RunE:  runPlan,
	}

	// Add subcommands
	cmd.AddCommand(NewPlanDeployCommand())
	cmd.AddCommand(NewPlanSliceCommand())

	cmd.Flags().StringP("env", "e", "", "Target environment (e.g. staging, prod)")
	cmd.Flags().StringP("version", "v", "", "Version to plan for (defaults to 'unknown' if omitted)")
	cmd.Flags().String("services", "", "Comma-separated list of services to include")
	cmd.Flags().String("format", "text", "Output format: text or json")
	cmd.Flags().BoolP("verbose", "V", false, "Show more detail")

	// Future extensions (v1 minimal, can be stubbed):
	// cmd.Flags().String("roles", "", "Comma-separated list of host roles")
	// cmd.Flags().String("hosts", "", "Comma-separated list of hostnames")
	// cmd.Flags().String("phases", "", "Comma-separated list of phase IDs/prefixes")

	_ = cmd.MarkFlagRequired("env")

	return cmd
}

// runPlan executes the plan command.
func runPlan(cmd *cobra.Command, args []string) error {
	// 1. Resolve global flags
	flags, err := ResolveFlags(cmd, nil)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// 2. Load config
	cfg, err := config.Load(flags.Config)
	if err != nil {
		if err == config.ErrConfigNotFound {
			return fmt.Errorf("stagecraft config not found at %s", flags.Config)
		}
		return fmt.Errorf("loading config: %w", err)
	}

	// 3. Re-resolve flags with config for environment validation
	flags, err = ResolveFlags(cmd, cfg)
	if err != nil {
		return fmt.Errorf("resolving flags: %w", err)
	}

	// 4. Validate environment is provided
	if flags.Env == "" {
		return fmt.Errorf("environment is required; use --env flag")
	}

	// 5. Initialize logger
	logger := logging.NewLogger(flags.Verbose)

	// 6. Parse plan-specific flags
	versionFlag, _ := cmd.Flags().GetString("version")
	servicesFlag, _ := cmd.Flags().GetString("services")
	formatFlag, _ := cmd.Flags().GetString("format")
	verboseFlag, _ := cmd.Flags().GetBool("verbose")

	// 7. Resolve version (plan-specific: no git, use "unknown" if omitted)
	version := resolvePlanVersion(versionFlag)

	// 8. Parse services list
	var services []string
	if servicesFlag != "" {
		services = parseServicesList(servicesFlag)
	}

	// 9. Generate plan
	planner := core.NewPlanner(cfg)
	plan, err := planner.PlanDeploy(flags.Env)
	if err != nil {
		return fmt.Errorf("generating deployment plan: %w", err)
	}

	// 10. Store version in plan metadata for rendering
	if plan.Metadata == nil {
		plan.Metadata = make(map[string]interface{})
	}
	plan.Metadata["version"] = version

	// 11. Get provider plans if backend is configured
	providerPlans := make(map[string]backendproviders.ProviderPlan)
	if cfg.Backend != nil {
		ctx := cmd.Context()
		if ctx == nil {
			ctx = context.Background()
		}

		// Get backend provider
		providerID := cfg.Backend.Provider
		provider, err := backendproviders.Get(providerID)
		if err != nil {
			// Log warning but don't fail - plan can still be generated without provider plan
			logger.Debug("Could not get backend provider for planning",
				logging.NewField("provider", providerID),
				logging.NewField("error", err.Error()),
			)
		} else {
			// Get provider config
			providerCfg, err := cfg.Backend.GetProviderConfig()
			if err != nil {
				logger.Debug("Could not get provider config for planning",
					logging.NewField("error", err.Error()),
				)
			} else {
				// Construct image tag (same logic as deploy)
				imageTag := fmt.Sprintf("%s:%s", cfg.Project.Name, version)

				// Get workdir
				workdir, err := os.Getwd()
				if err != nil {
					workdir = "."
				}

				// Call provider Plan()
				planOpts := backendproviders.PlanOptions{
					Config:   providerCfg,
					ImageTag: imageTag,
					WorkDir:  workdir,
				}

				providerPlan, err := provider.Plan(ctx, planOpts)
				if err != nil {
					logger.Debug("Provider plan generation failed",
						logging.NewField("provider", providerID),
						logging.NewField("error", err.Error()),
					)
				} else {
					providerPlans[providerID] = providerPlan
				}
			}
		}
	}

	// Store provider plans in metadata
	plan.Metadata["provider_plans"] = providerPlans

	// 12. Apply filters
	filteredPlan, err := applyFilters(plan, services, nil, nil, nil) // roles, hosts, phases stubbed for v1
	if err != nil {
		return fmt.Errorf("applying filters: %w", err)
	}

	// 13. Render output
	opts := PlanRenderOptions{
		Format:  formatFlag,
		Verbose: verboseFlag,
	}
	return renderPlan(cmd.OutOrStdout(), filteredPlan, flags.Env, version, opts, logger)
}

// resolvePlanVersion resolves the version for plan command.
// Unlike deploy/build, plan does NOT shell out to git.
// If --version is provided, use it. Otherwise, use "unknown".
func resolvePlanVersion(versionFlag string) string {
	if versionFlag != "" {
		return versionFlag
	}
	return "unknown"
}

// PlanRenderOptions contains options for rendering a plan.
type PlanRenderOptions struct {
	Format  string // "text" or "json"
	Verbose bool
}

// applyFilters applies service, role, host, and phase filters to a plan.
func applyFilters(plan *core.Plan, services, roles, hosts, phases []string) (*core.Plan, error) {
	// For v1, implement service filtering
	// Future: add role/host/phase filtering

	if len(services) == 0 && len(roles) == 0 && len(hosts) == 0 && len(phases) == 0 {
		return plan, nil
	}

	// Build set of services to include
	serviceSet := make(map[string]bool)
	for _, svc := range services {
		serviceSet[svc] = true
	}

	// Filter operations: keep if they touch at least one service
	filteredOps := []core.Operation{}
	for _, op := range plan.Operations {
		keep := true

		// Service filtering
		if len(serviceSet) > 0 {
			if !operationTouchesServices(op, serviceSet) {
				keep = false
			}
		}

		if keep {
			filteredOps = append(filteredOps, op)
		}
	}

	// Validate that all requested services are present
	if len(services) > 0 {
		foundServices := extractServicesFromPlan(filteredOps)
		for _, svc := range services {
			if !foundServices[svc] {
				return nil, fmt.Errorf("service %q not found in plan", svc)
			}
		}
	}

	return &core.Plan{
		Environment: plan.Environment,
		Operations:  filteredOps,
		Metadata:    plan.Metadata,
	}, nil
}

// operationTouchesServices checks if an operation touches any of the specified services.
// Returns true if the operation has no service metadata (to preserve dependencies like migrations).
func operationTouchesServices(op core.Operation, serviceSet map[string]bool) bool {
	// Check metadata for service information
	if services, ok := op.Metadata["services"].([]string); ok {
		for _, svc := range services {
			if serviceSet[svc] {
				return true
			}
		}
		// If services are specified but none match, return false
		return false
	}

	// Check if services is stored as []interface{} (from JSON unmarshaling)
	if services, ok := op.Metadata["services"].([]interface{}); ok {
		for _, svc := range services {
			if svcStr, ok := svc.(string); ok && serviceSet[svcStr] {
				return true
			}
		}
		// If services are specified but none match, return false
		return false
	}

	// If no service info, include by default (preserves dependencies like migrations)
	return true
}

// extractServicesFromPlan extracts all services mentioned in the plan operations.
func extractServicesFromPlan(ops []core.Operation) map[string]bool {
	services := make(map[string]bool)
	for _, op := range ops {
		if svcList, ok := op.Metadata["services"].([]string); ok {
			for _, svc := range svcList {
				services[svc] = true
			}
		}
		if svcList, ok := op.Metadata["services"].([]interface{}); ok {
			for _, svc := range svcList {
				if svcStr, ok := svc.(string); ok {
					services[svcStr] = true
				}
			}
		}
	}
	return services
}

// renderPlan renders the plan to the output writer.
func renderPlan(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions, logger logging.Logger) error {
	switch opts.Format {
	case "text":
		return renderPlanText(out, plan, env, version, opts, logger)
	case "json":
		return renderPlanJSON(out, plan, env, version, opts)
	default:
		return fmt.Errorf("invalid format: %s (must be 'text' or 'json')", opts.Format)
	}
}

// renderPlanText renders the plan in human-readable text format.
func renderPlanText(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions, logger logging.Logger) error {
	_ = opts.Verbose // Reserved for future verbose output enhancements
	_ = logger       // Reserved for future logging enhancements
	// Header
	_, _ = fmt.Fprintf(out, "Environment: %s\n", env)
	_, _ = fmt.Fprintf(out, "Version: %s\n", version)

	// Extract services and hosts from operations
	services := extractServicesFromPlan(plan.Operations)
	hosts := extractHostsFromPlan(plan.Operations)

	// Services
	if len(services) > 0 {
		serviceList := make([]string, 0, len(services))
		for svc := range services {
			serviceList = append(serviceList, svc)
		}
		sort.Strings(serviceList)
		_, _ = fmt.Fprintf(out, "Services: %s\n", strings.Join(serviceList, ", "))
	} else {
		_, _ = fmt.Fprintf(out, "Services: (all)\n")
	}

	// Hosts
	if len(hosts) > 0 {
		hostList := make([]string, 0, len(hosts))
		for host := range hosts {
			hostList = append(hostList, host)
		}
		sort.Strings(hostList)
		_, _ = fmt.Fprintf(out, "Hosts: %s\n", strings.Join(hostList, ", "))
	} else {
		_, _ = fmt.Fprintf(out, "Hosts: (all)\n")
	}

	_, _ = fmt.Fprintf(out, "\nPhases:\n")

	// Sort operations by ID (deterministic ordering)
	// For now, we'll use operation index as ID, but in future we can use actual IDs
	sortedOps := make([]core.Operation, len(plan.Operations))
	copy(sortedOps, plan.Operations)
	sort.Slice(sortedOps, func(i, j int) bool {
		// Sort by type first, then by description
		if sortedOps[i].Type != sortedOps[j].Type {
			return sortedOps[i].Type < sortedOps[j].Type
		}
		return sortedOps[i].Description < sortedOps[j].Description
	})

	// Render each phase
	for i, op := range sortedOps {
		_, _ = fmt.Fprintf(out, "  %d. %s\n", i+1, getOperationID(op, i))
		_, _ = fmt.Fprintf(out, "     - kind: %s\n", op.Type)

		// Services
		opServices := extractServicesFromOperation(op)
		if len(opServices) > 0 {
			sort.Strings(opServices)
			_, _ = fmt.Fprintf(out, "     - services: [%s]\n", strings.Join(opServices, ", "))
		} else {
			_, _ = fmt.Fprintf(out, "     - services: []\n")
		}

		// Hosts
		opHosts := extractHostsFromOperation(op)
		if len(opHosts) > 0 {
			sort.Strings(opHosts)
			_, _ = fmt.Fprintf(out, "     - hosts: [%s]\n", strings.Join(opHosts, ", "))
		} else {
			_, _ = fmt.Fprintf(out, "     - hosts: []\n")
		}

		// Description
		_, _ = fmt.Fprintf(out, "     - description: %s\n", op.Description)

		// Dependencies
		if len(op.Dependencies) > 0 {
			sort.Strings(op.Dependencies)
			_, _ = fmt.Fprintf(out, "     - depends_on: [%s]\n", strings.Join(op.Dependencies, ", "))
		} else {
			_, _ = fmt.Fprintf(out, "     - depends_on: []\n")
		}

		if i < len(sortedOps)-1 {
			_, _ = fmt.Fprintf(out, "\n")
		}
	}

	// Render provider plans if available
	if plan.Metadata != nil {
		if providerPlansRaw, ok := plan.Metadata["provider_plans"]; ok {
			if providerPlans, ok := providerPlansRaw.(map[string]backendproviders.ProviderPlan); ok && len(providerPlans) > 0 {
				_, _ = fmt.Fprintf(out, "\nPROVIDER PLANS:\n")

				// Sort provider IDs for deterministic output
				providerIDs := make([]string, 0, len(providerPlans))
				for id := range providerPlans {
					providerIDs = append(providerIDs, id)
				}
				sort.Strings(providerIDs)

				for _, providerID := range providerIDs {
					providerPlan := providerPlans[providerID]
					_, _ = fmt.Fprintf(out, "\nProvider: %s\n", providerID)
					for j, step := range providerPlan.Steps {
						_, _ = fmt.Fprintf(out, "  %d. %s\n", j+1, step.Name)
						_, _ = fmt.Fprintf(out, "     - %s\n", step.Description)
					}
				}
			}
		}
	}

	return nil
}

// renderPlanJSON renders the plan in JSON format.
func renderPlanJSON(out io.Writer, plan *core.Plan, env, version string, opts PlanRenderOptions) error {
	_ = opts.Verbose // Reserved for future verbose output enhancements
	// Build JSON structure
	jsonPlan := jsonPlan{
		Env:     env,
		Version: version,
		Phases:  []jsonPhase{},
	}

	// Sort operations for deterministic output
	sortedOps := make([]core.Operation, len(plan.Operations))
	copy(sortedOps, plan.Operations)
	sort.Slice(sortedOps, func(i, j int) bool {
		if sortedOps[i].Type != sortedOps[j].Type {
			return sortedOps[i].Type < sortedOps[j].Type
		}
		return sortedOps[i].Description < sortedOps[j].Description
	})

	// Convert operations to JSON phases
	for i, op := range sortedOps {
		phase := jsonPhase{
			ID:          getOperationID(op, i),
			Kind:        string(op.Type),
			Services:    extractServicesFromOperation(op),
			Hosts:       extractHostsFromOperation(op),
			Description: op.Description,
			DependsOn:   make([]string, len(op.Dependencies)),
			Metadata:    make(map[string]interface{}),
		}

		// Sort services and hosts
		sort.Strings(phase.Services)
		sort.Strings(phase.Hosts)

		// Copy dependencies
		copy(phase.DependsOn, op.Dependencies)
		sort.Strings(phase.DependsOn)

		// Copy metadata (excluding services and hosts which are already in top-level fields)
		for k, v := range op.Metadata {
			if k != "services" && k != "hosts" {
				phase.Metadata[k] = v
			}
		}

		jsonPlan.Phases = append(jsonPlan.Phases, phase)
	}

	// Add provider plans if available
	if plan.Metadata != nil {
		if providerPlansRaw, ok := plan.Metadata["provider_plans"]; ok {
			if providerPlans, ok := providerPlansRaw.(map[string]backendproviders.ProviderPlan); ok && len(providerPlans) > 0 {
				// Sort provider IDs for deterministic output
				providerIDs := make([]string, 0, len(providerPlans))
				for id := range providerPlans {
					providerIDs = append(providerIDs, id)
				}
				sort.Strings(providerIDs)

				// Use slice instead of map for true JSON determinism
				jsonPlan.ProviderPlans = make([]jsonProviderPlan, 0, len(providerPlans))

				for _, providerID := range providerIDs {
					providerPlan := providerPlans[providerID]
					jsonProviderPlan := jsonProviderPlan{
						Provider: providerPlan.Provider,
						Steps:    make([]jsonProviderStep, len(providerPlan.Steps)),
					}

					for i, step := range providerPlan.Steps {
						jsonProviderPlan.Steps[i] = jsonProviderStep{
							Name:        step.Name,
							Description: step.Description,
						}
					}

					jsonPlan.ProviderPlans = append(jsonPlan.ProviderPlans, jsonProviderPlan)
				}
			}
		}
	}

	// Marshal to JSON with indentation
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jsonPlan)
}

// jsonPlan is the JSON representation of a plan.
type jsonPlan struct {
	Env           string             `json:"env"`
	Version       string             `json:"version"`
	Phases        []jsonPhase        `json:"phases"`
	ProviderPlans []jsonProviderPlan `json:"provider_plans,omitempty"`
}

// jsonPhase is the JSON representation of a phase.
type jsonPhase struct {
	ID          string                 `json:"id"`
	Kind        string                 `json:"kind"`
	Services    []string               `json:"services"`
	Hosts       []string               `json:"hosts"`
	Description string                 `json:"description"`
	DependsOn   []string               `json:"depends_on"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// jsonProviderPlan is the JSON representation of a provider plan.
type jsonProviderPlan struct {
	Provider string             `json:"provider"`
	Steps    []jsonProviderStep `json:"steps"`
}

// jsonProviderStep is the JSON representation of a provider step.
type jsonProviderStep struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// getOperationID generates a deterministic ID for an operation.
// Prefers Operation.ID (canonical) over metadata fallback.
func getOperationID(op core.Operation, index int) string {
	// Use canonical Operation.ID if present (matches engine/agent view)
	if op.ID != "" {
		return op.ID
	}

	// Fallback: try to get ID from metadata (legacy)
	if id, ok := op.Metadata["id"].(string); ok && id != "" {
		return id
	}

	// Generate ID based on type and description
	// This ensures deterministic IDs
	prefix := strings.ToUpper(string(op.Type))
	parts := strings.Fields(op.Description)
	if len(parts) > 0 {
		// Use first word of description as suffix
		suffix := strings.ToUpper(parts[0])
		return fmt.Sprintf("%s_%s", prefix, suffix)
	}

	// Fallback to index-based ID
	return fmt.Sprintf("OP_%d", index)
}

// extractServicesFromOperation extracts services from an operation's metadata.
func extractServicesFromOperation(op core.Operation) []string {
	services := []string{}

	switch v := op.Metadata["services"].(type) {
	case []string:
		services = v
	case []interface{}:
		for _, svc := range v {
			if svcStr, ok := svc.(string); ok {
				services = append(services, svcStr)
			}
		}
	}

	return services
}

// extractHostsFromPlan extracts all hosts mentioned in the plan operations.
func extractHostsFromPlan(ops []core.Operation) map[string]bool {
	hosts := make(map[string]bool)
	for _, op := range ops {
		opHosts := extractHostsFromOperation(op)
		for _, host := range opHosts {
			hosts[host] = true
		}
	}
	return hosts
}

// extractHostsFromOperation extracts hosts from an operation's metadata.
func extractHostsFromOperation(op core.Operation) []string {
	hosts := []string{}

	switch v := op.Metadata["hosts"].(type) {
	case []string:
		hosts = v
	case []interface{}:
		for _, host := range v {
			if hostStr, ok := host.(string); ok {
				hosts = append(hosts, hostStr)
			}
		}
	}

	return hosts
}
