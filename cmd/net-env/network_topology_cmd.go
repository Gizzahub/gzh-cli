// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newNetworkTopologyCmd creates the network topology command.
func newNetworkTopologyCmd(logger *zap.Logger, _ string) *cobra.Command {
	containerDetector := NewContainerDetector(logger)
	analyzer := NewNetworkTopologyAnalyzer(logger, containerDetector)

	cmd := &cobra.Command{
		Use:   "network-topology",
		Short: "Analyze and visualize network topology",
		Long: `Analyze network topology including containers, services, connections, and dependencies.

This command provides comprehensive network topology analysis including:
- Container network mapping and relationships
- Service discovery and dependency analysis
- Network connection testing and status
- Logical cluster identification
- Topology complexity metrics
- Visual exports (DOT, Cytoscape, JSON)

Examples:
  # Analyze current network topology
  gz net-env network-topology analyze

  # Show topology summary
  gz net-env network-topology summary

  # List discovered services
  gz net-env network-topology services

  # Show network connections
  gz net-env network-topology connections

  # Export topology visualization
  gz net-env network-topology export --format dot --output topology.dot`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newTopologyAnalyzeCmd(analyzer))
	cmd.AddCommand(newTopologySummaryCmd(analyzer))
	cmd.AddCommand(newTopologyServicesCmd(analyzer))
	cmd.AddCommand(newTopologyConnectionsCmd(analyzer))
	cmd.AddCommand(newTopologyClustersCmd(analyzer))
	cmd.AddCommand(newTopologyExportCmd(analyzer))
	cmd.AddCommand(newTopologyMonitorCmd(analyzer))
	cmd.AddCommand(newTopologyValidateCmd(analyzer))

	return cmd
}

// newTopologyAnalyzeCmd creates the analyze subcommand.
func newTopologyAnalyzeCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Perform comprehensive network topology analysis",
		Long:  `Analyze the complete network topology including containers, services, connections, and clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")
			detailed, _ := cmd.Flags().GetBool("detailed")

			fmt.Println("🔍 Analyzing network topology...")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(topology)
			default:
				return printTopologyAnalysis(topology, detailed)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().BoolP("detailed", "d", false, "Show detailed analysis")

	return cmd
}

// newTopologySummaryCmd creates the summary subcommand.
func newTopologySummaryCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show network topology summary",
		Long:  `Display high-level summary statistics of the network topology.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(topology.Summary)
			default:
				return printTopologySummary(&topology.Summary)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newTopologyServicesCmd creates the services subcommand.
func newTopologyServicesCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "List discovered services",
		Long:  `List all services discovered in the network topology with their endpoints and dependencies.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")
			serviceType, _ := cmd.Flags().GetString("type")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			// Filter services by type if specified
			services := topology.Services
			if serviceType != "" {
				filtered := make([]TopologyService, 0)
				for _, service := range services {
					if string(service.Type) == serviceType {
						filtered = append(filtered, service)
					}
				}
				services = filtered
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(services)
			default:
				return printTopologyServices(services)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().StringP("type", "t", "", "Filter by service type (web|api|database|cache|queue|worker|proxy|other)")

	return cmd
}

// newTopologyConnectionsCmd creates the connections subcommand.
func newTopologyConnectionsCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connections",
		Short: "Show network connections",
		Long:  `Display network connections between containers and services with their status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")
			status, _ := cmd.Flags().GetString("status")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			// Filter connections by status if specified
			connections := topology.Connections
			if status != "" {
				filtered := make([]NetworkConnection, 0)
				for _, conn := range connections {
					if string(conn.Status) == status {
						filtered = append(filtered, conn)
					}
				}
				connections = filtered
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(connections)
			default:
				return printTopologyConnections(connections)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().StringP("status", "s", "", "Filter by connection status (active|idle|failed|unknown)")

	return cmd
}

// newTopologyClustersCmd creates the clusters subcommand.
func newTopologyClustersCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Show network clusters",
		Long:  `Display logical network clusters and their members.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")
			clusterType, _ := cmd.Flags().GetString("type")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			// Filter clusters by type if specified
			clusters := topology.Clusters
			if clusterType != "" {
				filtered := make([]NetworkCluster, 0)
				for _, cluster := range clusters {
					if string(cluster.Type) == clusterType {
						filtered = append(filtered, cluster)
					}
				}
				clusters = filtered
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(clusters)
			default:
				return printTopologyClusters(clusters)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")
	cmd.Flags().StringP("type", "t", "", "Filter by cluster type (namespace|project|environment|logical)")

	return cmd
}

// newTopologyExportCmd creates the export subcommand.
func newTopologyExportCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export topology visualization",
		Long:  `Export network topology to various visualization formats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			format, _ := cmd.Flags().GetString("format")
			outputFile, _ := cmd.Flags().GetString("output")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			data, err := analyzer.ExportTopology(topology, format)
			if err != nil {
				return fmt.Errorf("failed to export topology: %w", err)
			}

			if outputFile == "" {
				// Print to stdout
				fmt.Print(string(data))
			} else {
				// Write to file
				if err := os.WriteFile(outputFile, data, 0o644); err != nil {
					return fmt.Errorf("failed to write output file: %w", err)
				}
				fmt.Printf("✅ Topology exported to %s\n", outputFile)
			}

			return nil
		},
	}

	cmd.Flags().StringP("format", "f", "json", "Export format (json|dot|cytoscape)")
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")

	return cmd
}

// newTopologyMonitorCmd creates the monitor subcommand.
func newTopologyMonitorCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor topology changes",
		Long:  `Monitor network topology for changes and report when the topology structure changes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interval, _ := cmd.Flags().GetDuration("interval")
			maxDuration, _ := cmd.Flags().GetDuration("duration")

			fmt.Printf("🔄 Monitoring network topology changes (interval: %s)\n", interval)
			fmt.Println("Press Ctrl+C to stop...")

			ctx := context.Background()
			if maxDuration > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, maxDuration)
				defer cancel()
			}

			var lastTopologyHash string
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("\n✅ Monitoring stopped")
					return nil
				case <-ticker.C:
					topology, err := analyzer.AnalyzeNetworkTopology(ctx)
					if err != nil {
						fmt.Printf("❌ Analysis failed: %v\n", err)
						continue
					}

					currentHash := generateTopologyHash(topology)
					if lastTopologyHash == "" {
						lastTopologyHash = currentHash
						fmt.Printf("📸 Initial topology: %d networks, %d containers, %d services\n",
							len(topology.Networks), len(topology.Containers), len(topology.Services))
						continue
					}

					if currentHash != lastTopologyHash {
						fmt.Printf("\n🔄 Topology change detected at %s\n", time.Now().Format("15:04:05"))
						printTopologyChanges(topology)
						lastTopologyHash = currentHash
					}
				}
			}
		},
	}

	cmd.Flags().DurationP("interval", "i", 30*time.Second, "Monitoring interval")
	cmd.Flags().DurationP("duration", "d", 0, "Maximum monitoring duration (0 = unlimited)")

	return cmd
}

// newTopologyValidateCmd creates the validate subcommand.
func newTopologyValidateCmd(analyzer *NetworkTopologyAnalyzer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate network topology",
		Long:  `Validate network topology for common issues and misconfigurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()

			strict, _ := cmd.Flags().GetBool("strict")

			fmt.Println("🔍 Validating network topology...")

			topology, err := analyzer.AnalyzeNetworkTopology(ctx)
			if err != nil {
				return fmt.Errorf("failed to analyze network topology: %w", err)
			}

			issues := validateTopology(topology, strict)

			if len(issues) == 0 {
				fmt.Println("✅ Network topology validation passed - no issues found")
				return nil
			}

			fmt.Printf("⚠️  Found %d topology issues:\n\n", len(issues))
			for i, issue := range issues {
				fmt.Printf("%d. %s\n", i+1, issue)
			}

			if strict {
				return fmt.Errorf("topology validation failed with %d issues", len(issues))
			}

			return nil
		},
	}

	cmd.Flags().BoolP("strict", "s", false, "Fail on any validation issues")

	return cmd
}

// Helper functions for printing

func printTopologyAnalysis(topology *NetworkTopology, detailed bool) error {
	fmt.Printf("🌐 Network Topology Analysis\n\n")

	// Basic summary
	fmt.Printf("Generated: %s\n", topology.GeneratedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Analysis Duration: %s\n\n", topology.AnalysisMetrics.AnalysisDuration)

	// High-level metrics
	fmt.Printf("📊 Overview:\n")
	fmt.Printf("  Networks: %d\n", len(topology.Networks))
	fmt.Printf("  Containers: %d\n", len(topology.Containers))
	fmt.Printf("  Services: %d\n", len(topology.Services))
	fmt.Printf("  Connections: %d\n", len(topology.Connections))
	fmt.Printf("  Clusters: %d\n", len(topology.Clusters))
	fmt.Printf("  Dependencies: %d\n\n", len(topology.Dependencies))

	if detailed {
		// Detailed breakdown
		fmt.Printf("🔗 Network Details:\n")
		printTopologyNetworksTable(topology.Networks)

		fmt.Printf("\n📦 Container Details:\n")
		printTopologyContainersTable(topology.Containers)

		fmt.Printf("\n🚀 Service Details:\n")
		_ = printTopologyServices(topology.Services)

		fmt.Printf("\n🔌 Connection Details:\n")
		_ = printTopologyConnections(topology.Connections)

		fmt.Printf("\n🏷️  Cluster Details:\n")
		_ = printTopologyClusters(topology.Clusters)
	}

	// Complexity metrics
	fmt.Printf("📈 Complexity Metrics:\n")

	complexity := topology.Summary.TopologyComplexity
	fmt.Printf("  Network Complexity: %.2f\n", complexity.NetworkComplexity)
	fmt.Printf("  Service Complexity: %.2f\n", complexity.ServiceComplexity)
	fmt.Printf("  Connection Density: %.2f%%\n", complexity.ConnectionDensity*100)
	fmt.Printf("  Branching Factor: %.2f\n", complexity.BranchingFactor)
	fmt.Printf("  Cyclomatic Complexity: %d\n", complexity.CyclomaticComplexity)

	return nil
}

func printTopologySummary(summary *TopologySummary) error {
	fmt.Printf("📊 Network Topology Summary\n\n")

	fmt.Printf("Total Resources:\n")
	fmt.Printf("  Networks: %d\n", summary.TotalNetworks)
	fmt.Printf("  Containers: %d\n", summary.TotalContainers)
	fmt.Printf("  Services: %d\n", summary.TotalServices)
	fmt.Printf("  Connections: %d\n", summary.TotalConnections)
	fmt.Printf("  Clusters: %d\n\n", summary.TotalClusters)

	if len(summary.NetworksByDriver) > 0 {
		fmt.Printf("Networks by Driver:\n")

		for driver, count := range summary.NetworksByDriver {
			fmt.Printf("  %s: %d\n", driver, count)
		}

		fmt.Println()
	}

	if len(summary.ContainersByState) > 0 {
		fmt.Printf("Containers by State:\n")

		for state, count := range summary.ContainersByState {
			fmt.Printf("  %s: %d\n", state, count)
		}

		fmt.Println()
	}

	if len(summary.ServicesByType) > 0 {
		fmt.Printf("Services by Type:\n")

		for serviceType, count := range summary.ServicesByType {
			fmt.Printf("  %s: %d\n", serviceType, count)
		}

		fmt.Println()
	}

	complexity := summary.TopologyComplexity

	fmt.Printf("Complexity Metrics:\n")
	fmt.Printf("  Network Complexity: %.2f\n", complexity.NetworkComplexity)
	fmt.Printf("  Service Complexity: %.2f\n", complexity.ServiceComplexity)
	fmt.Printf("  Connection Density: %.2f%%\n", complexity.ConnectionDensity*100)
	fmt.Printf("  Branching Factor: %.2f\n", complexity.BranchingFactor)

	return nil
}

func printTopologyNetworksTable(networks []TopologyNetwork) {
	if len(networks) == 0 {
		fmt.Println("  No networks found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSUBNET\tCONTAINERS\tTYPE")

	for _, network := range networks {
		networkID := truncateStringUtil(network.ID, 12)

		subnet := network.Subnet
		if subnet == "" {
			subnet = "N/A"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
			networkID,
			network.Name,
			network.Driver,
			subnet,
			len(network.ConnectedContainers),
			network.NetworkType)
	}

	_ = w.Flush()
}

func printTopologyContainersTable(containers []TopologyContainer) {
	if len(containers) == 0 {
		fmt.Println("  No containers found.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "CONTAINER ID\tNAME\tIMAGE\tSTATE\tNETWORKS\tPORTS")

	for _, container := range containers {
		containerID := truncateStringUtil(container.ID, 12)
		image := truncateStringUtil(container.Image, 30)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%d\n",
			containerID,
			container.Name,
			image,
			container.State,
			len(container.NetworkInterfaces),
			len(container.ExposedPorts))
	}

	_ = w.Flush()
}

func printTopologyServices(services []TopologyService) error {
	if len(services) == 0 {
		fmt.Println("  No services found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "SERVICE\tTYPE\tCONTAINERS\tENDPOINTS\tHEALTH CHECKS")

	for _, service := range services {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n",
			service.Name,
			service.Type,
			len(service.Containers),
			len(service.Endpoints),
			len(service.HealthChecks))
	}

	return w.Flush()
}

func printTopologyConnections(connections []NetworkConnection) error {
	if len(connections) == 0 {
		fmt.Println("  No connections found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "SOURCE\tTARGET\tPROTOCOL\tPORT\tSTATUS\tLAST SEEN")

	for _, conn := range connections {
		sourceName := truncateStringUtil(conn.Source.Name, 20)
		targetName := truncateStringUtil(conn.Target.Name, 20)
		lastSeen := conn.LastSeen.Format("15:04:05")

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
			sourceName,
			targetName,
			conn.Protocol,
			conn.Port,
			conn.Status,
			lastSeen)
	}

	return w.Flush()
}

func printTopologyClusters(clusters []NetworkCluster) error {
	if len(clusters) == 0 {
		fmt.Println("  No clusters found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "CLUSTER ID\tNAME\tTYPE\tMEMBERS\tSUBNETS\tISOLATION")

	for _, cluster := range clusters {
		clusterID := truncateStringUtil(cluster.ID, 15)
		clusterName := truncateStringUtil(cluster.Name, 25)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%d\t%s\n",
			clusterID,
			clusterName,
			cluster.Type,
			len(cluster.Members),
			len(cluster.Subnets),
			cluster.Isolation)
	}

	return w.Flush()
}

func generateTopologyHash(topology *NetworkTopology) string {
	// Simple hash based on counts and key identifiers
	hash := fmt.Sprintf("nets:%d,containers:%d,services:%d,connections:%d",
		len(topology.Networks),
		len(topology.Containers),
		len(topology.Services),
		len(topology.Connections))

	// Add network IDs for more precise change detection
	networkIDs := make([]string, 0, len(topology.Networks))
	for _, network := range topology.Networks {
		networkIDs = append(networkIDs, network.ID[:8])
	}

	sort.Strings(networkIDs)
	hash += ",net_ids:" + strings.Join(networkIDs, ",")

	return hash
}

func printTopologyChanges(topology *NetworkTopology) {
	fmt.Printf("   Networks: %d\n", len(topology.Networks))
	fmt.Printf("   Containers: %d\n", len(topology.Containers))
	fmt.Printf("   Services: %d\n", len(topology.Services))
	fmt.Printf("   Connections: %d\n", len(topology.Connections))
	fmt.Printf("   Clusters: %d\n", len(topology.Clusters))
}

func validateTopology(topology *NetworkTopology, _ bool) []string {
	var issues []string

	// Check for isolated containers
	for _, container := range topology.Containers {
		if len(container.NetworkInterfaces) == 0 {
			issues = append(issues, fmt.Sprintf("Container %s (%s) has no network interfaces",
				container.Name, container.ID[:12]))
		}
	}

	// Check for services without containers
	for _, service := range topology.Services {
		if len(service.Containers) == 0 {
			issues = append(issues, fmt.Sprintf("Service %s has no containers", service.Name))
		}
	}

	// Check for failed connections
	failedConnections := 0

	for _, conn := range topology.Connections {
		if conn.Status == StatusFailed {
			failedConnections++
		}
	}

	if failedConnections > 0 {
		issues = append(issues, fmt.Sprintf("%d connections are in failed state", failedConnections))
	}

	// Check for high complexity
	complexity := topology.Summary.TopologyComplexity
	if complexity.ConnectionDensity > 0.8 {
		issues = append(issues, "High connection density (>80%) may indicate over-coupling")
	}

	if complexity.CyclomaticComplexity > 50 {
		issues = append(issues, "High cyclomatic complexity may indicate complex dependencies")
	}

	// Check for security issues
	for _, network := range topology.Networks {
		if !network.Internal && network.Driver == "bridge" && len(network.ConnectedContainers) > 10 {
			issues = append(issues, fmt.Sprintf("Network %s has many containers (%d) on external bridge",
				network.Name, len(network.ConnectedContainers)))
		}
	}

	return issues
}
