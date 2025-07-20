package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newContainerDetectionCmd creates the container detection command.
func newContainerDetectionCmd(logger *zap.Logger, _ string) *cobra.Command {
	cd := NewContainerDetector(logger)

	cmd := &cobra.Command{
		Use:   "container-detection",
		Short: "Detect running containers and environments",
		Long: `Detect and analyze running container environments including Docker, Podman, containerd, and nerdctl.
		
This command provides comprehensive container environment detection including:
- Available container runtimes (Docker, Podman, containerd, nerdctl)
- Running containers with network information
- Docker Compose projects
- Kubernetes cluster information
- Resource usage analysis
- Environment fingerprinting for change detection

Examples:
  # Detect current container environment
  gz net-env container-detection detect
  
  # Show detailed container information
  gz net-env container-detection status --detailed
  
  # List available container runtimes
  gz net-env container-detection runtimes
  
  # Monitor container changes
  gz net-env container-detection monitor --interval 30s`,
		SilenceUsage: true,
	}

	// Add subcommands
	cmd.AddCommand(newContainerDetectCmd(cd))
	cmd.AddCommand(newContainerStatusCmd(cd))
	cmd.AddCommand(newContainerRuntimesCmd(cd))
	cmd.AddCommand(newContainerMonitorCmd(cd))
	cmd.AddCommand(newContainerListCmd(cd))
	cmd.AddCommand(newContainerInspectCmd(cd))

	return cmd
}

// newContainerDetectCmd creates the detect subcommand.
func newContainerDetectCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect current container environment",
		Long:  `Perform a comprehensive detection of the current container environment including runtimes, containers, and orchestration platforms.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")
			detailed, _ := cmd.Flags().GetBool("detailed")

			fmt.Println("🔍 Detecting container environment...")

			env, err := cd.DetectContainerEnvironment(ctx)
			if err != nil {
				return fmt.Errorf("failed to detect container environment: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(env)
			case "yaml":
				// For now, just use structured output
				return printEnvironmentYAML(env)
			default:
				return printEnvironmentSummary(env, detailed)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")
	cmd.Flags().BoolP("detailed", "d", false, "Show detailed information")

	return cmd
}

// newContainerStatusCmd creates the status subcommand.
func newContainerStatusCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show container environment status",
		Long:  `Show the current status of container environments, including running containers and resource usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			showContainers, _ := cmd.Flags().GetBool("containers")
			showNetworks, _ := cmd.Flags().GetBool("networks")
			showResources, _ := cmd.Flags().GetBool("resources")
			output, _ := cmd.Flags().GetString("output")

			env, err := cd.DetectContainerEnvironment(ctx)
			if err != nil {
				return fmt.Errorf("failed to get container environment: %w", err)
			}

			if output == "json" {
				data := map[string]interface{}{
					"orchestration_platform": env.OrchestrationPlatform,
					"primary_runtime":        env.PrimaryRuntime,
					"detected_at":            env.DetectedAt,
				}

				if showContainers {
					data["containers"] = env.RunningContainers
				}
				if showNetworks {
					data["networks"] = env.Networks
				}
				if showResources {
					data["resource_usage"] = env.ResourceUsage
				}

				return json.NewEncoder(os.Stdout).Encode(data)
			}

			// Print status summary
			fmt.Printf("🐳 Container Environment Status\n\n")
			fmt.Printf("Primary Runtime: %s\n", env.PrimaryRuntime)
			fmt.Printf("Orchestration: %s\n", env.OrchestrationPlatform)
			fmt.Printf("Detected At: %s\n", env.DetectedAt.Format("2006-01-02 15:04:05"))

			if showContainers && len(env.RunningContainers) > 0 {
				fmt.Printf("\n📦 Running Containers (%d):\n", len(env.RunningContainers))
				_ = printDetectedContainersTable(env.RunningContainers) //nolint:errcheck // CLI display operations are non-critical
			}

			if showNetworks && len(env.Networks) > 0 {
				fmt.Printf("\n🌐 Networks (%d):\n", len(env.Networks))
				_ = printDetectedNetworksTable(env.Networks) //nolint:errcheck // CLI display operations are non-critical
			}

			if showResources && env.ResourceUsage != nil {
				fmt.Printf("\n📊 Resource Usage:\n")
				printResourceUsage(env.ResourceUsage)
			}

			return nil
		},
	}

	cmd.Flags().BoolP("containers", "c", false, "Show running containers")
	cmd.Flags().BoolP("networks", "n", false, "Show container networks")
	cmd.Flags().BoolP("resources", "r", false, "Show resource usage")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newContainerRuntimesCmd creates the runtimes subcommand.
func newContainerRuntimesCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "runtimes",
		Short: "List available container runtimes",
		Long:  `List all available container runtimes on the system with their versions and status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			output, _ := cmd.Flags().GetString("output")

			env, err := cd.DetectContainerEnvironment(ctx)
			if err != nil {
				return fmt.Errorf("failed to detect runtimes: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(env.AvailableRuntimes)
			default:
				return printRuntimesTable(env.AvailableRuntimes)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newContainerMonitorCmd creates the monitor subcommand.
func newContainerMonitorCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "Monitor container environment changes",
		Long:  `Monitor the container environment for changes and report when containers start, stop, or change configuration.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			interval, _ := cmd.Flags().GetDuration("interval")
			maxDuration, _ := cmd.Flags().GetDuration("duration")

			fmt.Printf("🔄 Monitoring container environment (interval: %s)\n", interval)
			fmt.Println("Press Ctrl+C to stop...")

			ctx := context.Background()
			if maxDuration > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, maxDuration)
				defer cancel()
			}

			var lastFingerprint string
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					fmt.Println("\n✅ Monitoring stopped")
					return nil
				case <-ticker.C:
					env, err := cd.DetectContainerEnvironment(ctx)
					if err != nil {
						fmt.Printf("❌ Detection failed: %v\n", err)
						continue
					}

					if lastFingerprint == "" {
						lastFingerprint = env.EnvironmentFingerprint
						fmt.Printf("📸 Initial environment fingerprint: %s\n", lastFingerprint[:12])
						continue
					}

					if env.EnvironmentFingerprint != lastFingerprint {
						fmt.Printf("\n🔄 Environment change detected at %s\n", time.Now().Format("15:04:05"))
						fmt.Printf("   Containers: %d running\n", len(env.RunningContainers))
						fmt.Printf("   Networks: %d active\n", len(env.Networks))
						fmt.Printf("   New fingerprint: %s\n", env.EnvironmentFingerprint[:12])
						lastFingerprint = env.EnvironmentFingerprint
					}
				}
			}
		},
	}

	cmd.Flags().DurationP("interval", "i", 30*time.Second, "Monitoring interval")
	cmd.Flags().DurationP("duration", "d", 0, "Maximum monitoring duration (0 = unlimited)")

	return cmd
}

// newContainerListCmd creates the list subcommand.
func newContainerListCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List running containers",
		Long:  `List all running containers across all detected container runtimes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			runtime, _ := cmd.Flags().GetString("runtime")
			output, _ := cmd.Flags().GetString("output")

			env, err := cd.DetectContainerEnvironment(ctx)
			if err != nil {
				return fmt.Errorf("failed to get containers: %w", err)
			}

			// Filter by runtime if specified
			containers := env.RunningContainers
			if runtime != "" {
				filtered := make([]DetectedContainer, 0)
				for _, container := range containers {
					if string(container.Runtime) == runtime {
						filtered = append(filtered, container)
					}
				}
				containers = filtered
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(containers)
			default:
				if len(containers) == 0 {
					fmt.Println("No running containers found.")
					return nil
				}
				fmt.Printf("📦 Running Containers (%d):\n", len(containers))
				_ = printDetectedContainersTable(containers) //nolint:errcheck // CLI display operations are non-critical
				return nil
			}
		},
	}

	cmd.Flags().StringP("runtime", "r", "", "Filter by runtime (docker|podman|containerd|nerdctl)")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newContainerInspectCmd creates the inspect subcommand.
func newContainerInspectCmd(cd *ContainerDetector) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect [container-id-or-name]",
		Short: "Inspect a specific container",
		Long:  `Get detailed information about a specific container including network configuration, mounts, and resource limits.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			containerID := args[0]
			output, _ := cmd.Flags().GetString("output")

			env, err := cd.DetectContainerEnvironment(ctx)
			if err != nil {
				return fmt.Errorf("failed to get container environment: %w", err)
			}

			// Find the container
			var container *DetectedContainer
			for _, c := range env.RunningContainers {
				if c.ID == containerID || c.Name == containerID || strings.HasPrefix(c.ID, containerID) {
					container = &c
					break
				}
			}

			if container == nil {
				return fmt.Errorf("container '%s' not found", containerID)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(container)
			default:
				return printContainerDetails(container)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// Helper functions for printing

func printEnvironmentSummary(env *ContainerEnvironment, detailed bool) error {
	fmt.Printf("🐳 Container Environment Summary\n\n")

	// Basic information
	fmt.Printf("Primary Runtime: %s\n", env.PrimaryRuntime)
	fmt.Printf("Orchestration Platform: %s\n", env.OrchestrationPlatform)
	fmt.Printf("Environment Fingerprint: %s\n", env.EnvironmentFingerprint[:12]+"...")
	fmt.Printf("Detected At: %s\n\n", env.DetectedAt.Format("2006-01-02 15:04:05"))

	// Available runtimes
	fmt.Printf("📋 Available Runtimes (%d):\n", len(env.AvailableRuntimes))
	_ = printRuntimesTable(env.AvailableRuntimes) //nolint:errcheck // CLI display operations are non-critical

	// Running containers
	if len(env.RunningContainers) > 0 {
		fmt.Printf("\n📦 Running Containers (%d):\n", len(env.RunningContainers))

		if detailed {
			for _, container := range env.RunningContainers {
				printContainerSummary(&container)
			}
		} else {
			_ = printDetectedContainersTable(env.RunningContainers) //nolint:errcheck // CLI display operations are non-critical
		}
	}

	// Networks
	if len(env.Networks) > 0 {
		fmt.Printf("\n🌐 Networks (%d):\n", len(env.Networks))
		_ = printDetectedNetworksTable(env.Networks) //nolint:errcheck // CLI display operations are non-critical
	}

	// Compose projects
	if len(env.ComposeProjects) > 0 {
		fmt.Printf("\n🐙 Docker Compose Projects (%d):\n", len(env.ComposeProjects))
		printComposeProjects(env.ComposeProjects)
	}

	// Kubernetes info
	if env.KubernetesInfo != nil {
		fmt.Printf("\n☸️  Kubernetes Cluster Info:\n")
		printKubernetesInfo(env.KubernetesInfo)
	}

	// Resource usage
	if env.ResourceUsage != nil {
		fmt.Printf("\n📊 Resource Usage:\n")
		printResourceUsage(env.ResourceUsage)
	}

	return nil
}

func printRuntimesTable(runtimes []RuntimeInfo) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "RUNTIME\tVERSION\tAVAILABLE\tEXECUTABLE") //nolint:errcheck // CLI table header output

	for _, runtime := range runtimes {
		available := "No"
		if runtime.Available {
			available = "Yes"
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", //nolint:errcheck // Table display errors are non-critical
			runtime.Runtime,
			runtime.Version,
			available,
			runtime.Executable)
	}

	return w.Flush()
}

func printDetectedContainersTable(containers []DetectedContainer) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "CONTAINER ID\tNAME\tIMAGE\tSTATUS\tRUNTIME\tNETWORKS") //nolint:errcheck // Table display errors are non-critical

	for _, container := range containers {
		containerID := truncateStringUtil(container.ID, 12)
		image := truncateStringUtil(container.Image, 30)
		networkCount := fmt.Sprintf("%d", len(container.Networks))

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", //nolint:errcheck // Table display errors are non-critical
			containerID,
			container.Name,
			image,
			container.Status,
			container.Runtime,
			networkCount)
	}

	return w.Flush()
}

func printDetectedNetworksTable(networks []DetectedNetwork) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSCOPE\tCREATED") //nolint:errcheck // Table display errors are non-critical

	for _, network := range networks {
		networkID := truncateStringUtil(network.ID, 12)

		created := "N/A"
		if !network.Created.IsZero() {
			created = network.Created.Format("2006-01-02")
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", //nolint:errcheck // Table display errors are non-critical
			networkID,
			network.Name,
			network.Driver,
			network.Scope,
			created)
	}

	return w.Flush()
}

func printContainerSummary(container *DetectedContainer) {
	fmt.Printf("  • %s (%s)\n", container.Name, container.ID[:12])
	fmt.Printf("    Image: %s\n", container.Image)
	fmt.Printf("    Status: %s\n", container.Status)
	fmt.Printf("    Runtime: %s\n", container.Runtime)

	if len(container.Ports) > 0 {
		fmt.Printf("    Ports: %d mapped\n", len(container.Ports))
	}

	fmt.Println()
}

func printContainerDetails(container *DetectedContainer) error {
	fmt.Printf("📦 Container Details\n\n")
	fmt.Printf("ID: %s\n", container.ID)
	fmt.Printf("Name: %s\n", container.Name)
	fmt.Printf("Image: %s\n", container.Image)
	fmt.Printf("Status: %s\n", container.Status)
	fmt.Printf("State: %s\n", container.State)
	fmt.Printf("Runtime: %s\n", container.Runtime)
	fmt.Printf("Created: %s\n", container.Created.Format("2006-01-02 15:04:05"))
	fmt.Printf("Started: %s\n", container.StartedAt.Format("2006-01-02 15:04:05"))

	if len(container.Ports) > 0 {
		fmt.Printf("\nPorts:\n")

		for _, port := range container.Ports {
			fmt.Printf("  %s:%d -> %d/%s\n", port.HostIP, port.HostPort, port.ContainerPort, port.Protocol)
		}
	}

	if len(container.Networks) > 0 {
		fmt.Printf("\nNetworks:\n")

		for _, network := range container.Networks {
			fmt.Printf("  • %s (IP: %s)\n", network.NetworkName, network.IPAddress)
		}
	}

	if len(container.Environment) > 0 {
		fmt.Printf("\nEnvironment Variables:\n")

		for _, env := range container.Environment {
			fmt.Printf("  %s\n", env)
		}
	}

	if len(container.Mounts) > 0 {
		fmt.Printf("\nMounts:\n")

		for _, mount := range container.Mounts {
			fmt.Printf("  %s -> %s (%s)\n", mount.Source, mount.Destination, mount.Type)
		}
	}

	return nil
}

func printComposeProjects(projects []ComposeProject) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "PROJECT\tFILE\tSERVICES\tCONTAINERS") //nolint:errcheck // Table display errors are non-critical

	for _, project := range projects {
		configFile := project.ConfigPath
		if configFile == "" {
			configFile = "N/A"
		}

		containerCount := 0
		for _, service := range project.Services {
			containerCount += len(service.Containers)
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%d\t%d\n", //nolint:errcheck // Table display errors are non-critical
			project.Name,
			configFile,
			len(project.Services),
			containerCount)
	}

	_ = w.Flush() //nolint:errcheck // Table display errors are non-critical
}

func printKubernetesInfo(info *KubernetesClusterInfo) {
	fmt.Printf("  Available: %t\n", info.Available)
	fmt.Printf("  Context: %s\n", info.Context)
	fmt.Printf("  Namespace: %s\n", info.Namespace)
	fmt.Printf("  Version: %s\n", info.Version)

	if len(info.Nodes) > 0 {
		fmt.Printf("  Nodes: %d\n", len(info.Nodes))
	}

	if info.ServiceMesh != nil {
		fmt.Printf("  Service Mesh: %s (%s)\n", info.ServiceMesh.Type, info.ServiceMesh.Version)
	}
}

func printResourceUsage(usage *ContainerResourceUsage) {
	fmt.Printf("  Total Containers: %d\n", usage.TotalContainers)
	fmt.Printf("  Running Containers: %d\n", usage.RunningContainers)
	fmt.Printf("  CPU Usage: %.2f%%\n", usage.ResourceSummary.CPUUsage)
	fmt.Printf("  Memory Usage: %d bytes\n", usage.ResourceSummary.MemoryUsage)
	fmt.Printf("  Network RX: %d bytes\n", usage.ResourceSummary.NetworkRx)
	fmt.Printf("  Network TX: %d bytes\n", usage.ResourceSummary.NetworkTx)
	fmt.Printf("  Block Read: %d bytes\n", usage.ResourceSummary.BlockRead)
	fmt.Printf("  Block Write: %d bytes\n", usage.ResourceSummary.BlockWrite)
}

func printEnvironmentYAML(env *ContainerEnvironment) error {
	// Simple YAML-like output for now
	fmt.Printf("container_environment:\n")
	fmt.Printf("  primary_runtime: %s\n", env.PrimaryRuntime)
	fmt.Printf("  orchestration_platform: %s\n", env.OrchestrationPlatform)
	fmt.Printf("  detected_at: %s\n", env.DetectedAt.Format(time.RFC3339))
	fmt.Printf("  environment_fingerprint: %s\n", env.EnvironmentFingerprint)
	fmt.Printf("  available_runtimes:\n")

	for _, runtime := range env.AvailableRuntimes {
		fmt.Printf("    - runtime: %s\n", runtime.Runtime)
		fmt.Printf("      version: %s\n", runtime.Version)
		fmt.Printf("      available: %t\n", runtime.Available)
	}

	fmt.Printf("  running_containers: %d\n", len(env.RunningContainers))
	fmt.Printf("  networks: %d\n", len(env.Networks))

	if len(env.ComposeProjects) > 0 {
		fmt.Printf("  compose_projects: %d\n", len(env.ComposeProjects))
	}

	return nil
}

func truncateStringUtil(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}
