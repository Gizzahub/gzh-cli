// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package bootstrap

import (
	"fmt"
	"sort"
)

// DependencyResolver handles dependency resolution for package manager installation order.
type DependencyResolver struct {
	dependencies map[string][]string // manager -> list of dependencies
}

// NewDependencyResolver creates a new dependency resolver.
func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		dependencies: make(map[string][]string),
	}
}

// AddDependency adds a dependency relationship.
func (dr *DependencyResolver) AddDependency(manager string, deps []string) {
	dr.dependencies[manager] = deps
}

// ResolveDependencies resolves the installation order for given managers using topological sort.
func (dr *DependencyResolver) ResolveDependencies(managers []string) ([]string, error) {
	if len(managers) == 0 {
		return []string{}, nil
	}

	// Build a graph of all managers and their dependencies
	graph := make(map[string][]string)
	inDegree := make(map[string]int)

	// Initialize all requested managers
	for _, manager := range managers {
		if _, exists := graph[manager]; !exists {
			graph[manager] = []string{}
			inDegree[manager] = 0
		}
	}

	// Add dependencies to the graph
	visited := make(map[string]bool)
	if err := dr.addDependenciesToGraph(managers, graph, inDegree, visited); err != nil {
		return nil, err
	}

	// Perform topological sort using Kahn's algorithm
	result := make([]string, 0)
	queue := make([]string, 0)

	// Find all nodes with no incoming edges
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	// Sort the initial queue for consistent ordering
	sort.Strings(queue)

	for len(queue) > 0 {
		// Remove node with no dependencies
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		// Remove this node from the graph
		for _, neighbor := range graph[current] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
				// Keep queue sorted for consistent ordering
				sort.Strings(queue)
			}
		}
	}

	// Check for circular dependencies
	if len(result) != len(inDegree) {
		return nil, fmt.Errorf("circular dependency detected in package managers")
	}

	// Filter result to only include originally requested managers
	requestedSet := make(map[string]bool)
	for _, manager := range managers {
		requestedSet[manager] = true
	}

	filteredResult := make([]string, 0)
	for _, manager := range result {
		if requestedSet[manager] {
			filteredResult = append(filteredResult, manager)
		}
	}

	return filteredResult, nil
}

// addDependenciesToGraph recursively adds dependencies to the graph.
func (dr *DependencyResolver) addDependenciesToGraph(managers []string, graph map[string][]string, inDegree map[string]int, visited map[string]bool) error {
	for _, manager := range managers {
		if visited[manager] {
			continue
		}
		visited[manager] = true

		deps := dr.dependencies[manager]
		for _, dep := range deps {
			// Initialize dependency node if not exists
			if _, exists := graph[dep]; !exists {
				graph[dep] = []string{}
				inDegree[dep] = 0
			}

			// Add edge from dependency to manager
			graph[dep] = append(graph[dep], manager)
			inDegree[manager]++

			// Recursively add dependencies of dependencies
			if err := dr.addDependenciesToGraph([]string{dep}, graph, inDegree, visited); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetDependencies returns the direct dependencies of a manager.
func (dr *DependencyResolver) GetDependencies(manager string) []string {
	deps, exists := dr.dependencies[manager]
	if !exists {
		return []string{}
	}
	// Return a copy to prevent modification
	result := make([]string, len(deps))
	copy(result, deps)
	return result
}

// HasDependency checks if a manager has a specific dependency.
func (dr *DependencyResolver) HasDependency(manager, dependency string) bool {
	deps := dr.dependencies[manager]
	for _, dep := range deps {
		if dep == dependency {
			return true
		}
	}
	return false
}

// GetAllDependencies returns all transitive dependencies of a manager.
func (dr *DependencyResolver) GetAllDependencies(manager string) ([]string, error) {
	visited := make(map[string]bool)
	result := make([]string, 0)

	if err := dr.collectDependencies(manager, visited, &result); err != nil {
		return nil, err
	}

	// Remove the manager itself from the result if present
	filtered := make([]string, 0)
	for _, dep := range result {
		if dep != manager {
			filtered = append(filtered, dep)
		}
	}

	return filtered, nil
}

// collectDependencies recursively collects all dependencies.
func (dr *DependencyResolver) collectDependencies(manager string, visited map[string]bool, result *[]string) error {
	if visited[manager] {
		return fmt.Errorf("circular dependency detected involving %s", manager)
	}

	visited[manager] = true
	*result = append(*result, manager)

	deps := dr.dependencies[manager]
	for _, dep := range deps {
		if err := dr.collectDependencies(dep, visited, result); err != nil {
			return err
		}
	}

	visited[manager] = false
	return nil
}

// ValidateNoCycles checks if the current dependency graph has any cycles.
func (dr *DependencyResolver) ValidateNoCycles() error {
	// Get all managers
	allManagers := make([]string, 0, len(dr.dependencies))
	for manager := range dr.dependencies {
		allManagers = append(allManagers, manager)
	}

	// Try to resolve all managers - this will detect cycles
	_, err := dr.ResolveDependencies(allManagers)
	return err
}
