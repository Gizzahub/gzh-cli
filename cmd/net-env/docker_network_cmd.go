package netenv

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// newDockerNetworkCmd creates the docker-network command.
func newDockerNetworkCmd(logger *zap.Logger, configDir string) *cobra.Command {
	dm := NewDockerNetworkManager(logger, configDir)

	cmd := &cobra.Command{
		Use:   "docker-network",
		Short: "Manage Docker network profiles",
		Long:  `Manage Docker network profiles for container-specific network configurations and Docker Compose integration.`,
	}

	// Add subcommands
	cmd.AddCommand(newDockerNetworkCreateCmd(dm))
	cmd.AddCommand(newDockerNetworkListCmd(dm))
	cmd.AddCommand(newDockerNetworkApplyCmd(dm))
	cmd.AddCommand(newDockerNetworkDeleteCmd(dm))
	cmd.AddCommand(newDockerNetworkStatusCmd(dm))
	cmd.AddCommand(newDockerNetworkImportCmd(dm))
	cmd.AddCommand(newDockerNetworkExportCmd(dm))
	cmd.AddCommand(newDockerNetworkDetectCmd(dm))
	cmd.AddCommand(newDockerNetworkContainerCmd(dm))
	cmd.AddCommand(newDockerNetworkCloneCmd(dm))

	return cmd
}

// newDockerNetworkCreateCmd creates the create subcommand.
func newDockerNetworkCreateCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [profile-name]",
		Short: "Create a new Docker network profile",
		Long:  `Create a new Docker network profile with custom network and container configurations.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Get flags
			description, _ := cmd.Flags().GetString("description")
			networkName, _ := cmd.Flags().GetString("network")
			driver, _ := cmd.Flags().GetString("driver")
			subnet, _ := cmd.Flags().GetString("subnet")
			gateway, _ := cmd.Flags().GetString("gateway")
			interactive, _ := cmd.Flags().GetBool("interactive")

			profile := &DockerNetworkProfile{
				Name:        profileName,
				Description: description,
				Networks:    make(map[string]*DockerNetwork),
				Containers:  make(map[string]*ContainerNetwork),
				Metadata:    make(map[string]string),
			}

			if interactive {
				return createProfileInteractively(dm, profile)
			}

			// Create a simple network if specified
			if networkName != "" {
				network := &DockerNetwork{
					Name:    networkName,
					Driver:  driver,
					Subnet:  subnet,
					Gateway: gateway,
				}
				profile.Networks[networkName] = network
			}

			if err := dm.CreateProfile(profile); err != nil {
				return fmt.Errorf("failed to create profile: %w", err)
			}

			fmt.Printf("✅ Created Docker network profile: %s\n", profileName)
			if description != "" {
				fmt.Printf("   Description: %s\n", description)
			}

			return nil
		},
	}

	cmd.Flags().String("description", "", "Profile description")
	cmd.Flags().String("network", "", "Network name to create")
	cmd.Flags().String("driver", "bridge", "Network driver")
	cmd.Flags().String("subnet", "", "Network subnet (e.g., 172.20.0.0/16)")
	cmd.Flags().String("gateway", "", "Network gateway (e.g., 172.20.0.1)")
	cmd.Flags().BoolP("interactive", "i", false, "Create profile interactively")

	return cmd
}

// newDockerNetworkListCmd creates the list subcommand.
func newDockerNetworkListCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List Docker network profiles",
		Long:    `List all available Docker network profiles with their status and basic information.`,
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			profiles, err := dm.ListProfiles()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				fmt.Println("No Docker network profiles found.")
				return nil
			}

			// Get output format
			output, _ := cmd.Flags().GetString("output")

			switch output {
			case "json":
				return printProfilesJSON(profiles)
			case "yaml":
				return printProfilesYAML(profiles)
			default:
				return printProfilesTable(profiles)
			}
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newDockerNetworkApplyCmd creates the apply subcommand.
func newDockerNetworkApplyCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply [profile-name]",
		Short: "Apply a Docker network profile",
		Long:  `Apply a Docker network profile to create networks and configure containers.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			dryRun, _ := cmd.Flags().GetBool("dry-run")
			if dryRun {
				fmt.Printf("🔍 Dry run mode - would apply profile: %s\n", profileName)
				profile, err := dm.LoadProfile(profileName)
				if err != nil {
					return fmt.Errorf("failed to load profile: %w", err)
				}
				return printProfileDetails(profile)
			}

			fmt.Printf("⏳ Applying Docker network profile: %s\n", profileName)

			if err := dm.ApplyProfile(profileName); err != nil {
				return fmt.Errorf("failed to apply profile: %w", err)
			}

			fmt.Printf("✅ Successfully applied Docker network profile: %s\n", profileName)
			return nil
		},
	}

	cmd.Flags().Bool("dry-run", false, "Show what would be applied without making changes")

	return cmd
}

// newDockerNetworkDeleteCmd creates the delete subcommand.
func newDockerNetworkDeleteCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete [profile-name]",
		Short:   "Delete a Docker network profile",
		Long:    `Delete a Docker network profile. This does not affect existing networks or containers.`,
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			force, _ := cmd.Flags().GetBool("force")
			if !force {
				fmt.Printf("⚠️  Are you sure you want to delete profile '%s'? (y/N): ", profileName)
				var response string
				fmt.Scanln(&response)
				if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
					fmt.Println("❌ Deletion cancelled.")
					return nil
				}
			}

			if err := dm.DeleteProfile(profileName); err != nil {
				return fmt.Errorf("failed to delete profile: %w", err)
			}

			fmt.Printf("✅ Deleted Docker network profile: %s\n", profileName)
			return nil
		},
	}

	cmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	return cmd
}

// newDockerNetworkStatusCmd creates the status subcommand.
func newDockerNetworkStatusCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show Docker network and container status",
		Long:  `Show the current status of Docker networks and running containers with their network configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			showContainers, _ := cmd.Flags().GetBool("containers")
			output, _ := cmd.Flags().GetString("output")

			// Get network status
			networks, err := dm.GetNetworkStatus()
			if err != nil {
				return fmt.Errorf("failed to get network status: %w", err)
			}

			if output == "json" {
				data := map[string]interface{}{
					"networks": networks,
				}

				if showContainers {
					containers, err := dm.GetContainerNetworkInfo()
					if err != nil {
						return fmt.Errorf("failed to get container info: %w", err)
					}
					data["containers"] = containers
				}

				return json.NewEncoder(os.Stdout).Encode(data)
			}

			// Print networks
			fmt.Println("🌐 Docker Networks:")
			if len(networks) == 0 {
				fmt.Println("  No networks found.")
			} else {
				printNetworksTable(networks)
			}

			// Print containers if requested
			if showContainers {
				fmt.Println("\n📦 Container Network Info:")
				containers, err := dm.GetContainerNetworkInfo()
				if err != nil {
					return fmt.Errorf("failed to get container info: %w", err)
				}

				if len(containers) == 0 {
					fmt.Println("  No running containers found.")
				} else {
					printContainersTable(containers)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolP("containers", "c", false, "Show container network information")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json)")

	return cmd
}

// newDockerNetworkImportCmd creates the import subcommand.
func newDockerNetworkImportCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import [compose-file] [profile-name]",
		Short: "Import Docker Compose file as network profile",
		Long:  `Import a Docker Compose file and create a network profile from its configuration.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			composePath := args[0]
			profileName := args[1]

			// Convert relative path to absolute
			if !filepath.IsAbs(composePath) {
				wd, _ := os.Getwd()
				composePath = filepath.Join(wd, composePath)
			}

			fmt.Printf("📥 Importing Docker Compose file: %s\n", composePath)

			if err := dm.CreateProfileFromCompose(composePath, profileName); err != nil {
				return fmt.Errorf("failed to import compose file: %w", err)
			}

			fmt.Printf("✅ Created profile '%s' from compose file\n", profileName)
			return nil
		},
	}

	return cmd
}

// newDockerNetworkExportCmd creates the export subcommand.
func newDockerNetworkExportCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export [profile-name] [output-file]",
		Short: "Export Docker network profile",
		Long:  `Export a Docker network profile to a file in YAML or JSON format.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			outputFile := args[1]

			profile, err := dm.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			format := strings.ToLower(filepath.Ext(outputFile))
			var data []byte

			switch format {
			case ".json":
				data, err = json.MarshalIndent(profile, "", "  ")
			case ".yaml", ".yml":
				data, err = yaml.Marshal(profile)
			default:
				return fmt.Errorf("unsupported output format: %s (use .json, .yaml, or .yml)", format)
			}

			if err != nil {
				return fmt.Errorf("failed to marshal profile: %w", err)
			}

			if err := os.WriteFile(outputFile, data, 0o644); err != nil {
				return fmt.Errorf("failed to write output file: %w", err)
			}

			fmt.Printf("✅ Exported profile '%s' to %s\n", profileName, outputFile)
			return nil
		},
	}

	return cmd
}

// newDockerNetworkDetectCmd creates the detect subcommand.
func newDockerNetworkDetectCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect running Docker Compose projects",
		Long:  `Detect and list running Docker Compose projects that can be imported as network profiles.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			projects, err := dm.DetectDockerComposeProjects()
			if err != nil {
				return fmt.Errorf("failed to detect compose projects: %w", err)
			}

			if len(projects) == 0 {
				fmt.Println("No running Docker Compose projects found.")
				return nil
			}

			fmt.Println("🔍 Detected Docker Compose Projects:")
			for i, project := range projects {
				fmt.Printf("%d. %s\n", i+1, project)
			}

			fmt.Printf("\n💡 Use 'gz net-env docker-network import <compose-file> <profile-name>' to create profiles from compose files.\n")
			return nil
		},
	}

	return cmd
}

// Helper functions

func createProfileInteractively(dm *DockerNetworkManager, profile *DockerNetworkProfile) error {
	fmt.Printf("Creating Docker network profile interactively...\n\n")

	// Get description
	if profile.Description == "" {
		fmt.Print("Enter description (optional): ")
		fmt.Scanln(&profile.Description)
	}

	// Ask if user wants to add networks
	fmt.Print("Add a network? (y/N): ")

	var addNetwork string
	fmt.Scanln(&addNetwork)

	if strings.ToLower(addNetwork) == "y" || strings.ToLower(addNetwork) == "yes" {
		for {
			var networkName, driver, subnet, gateway string

			fmt.Print("Network name: ")
			fmt.Scanln(&networkName)

			fmt.Print("Driver (bridge/overlay/macvlan) [bridge]: ")
			fmt.Scanln(&driver)

			if driver == "" {
				driver = "bridge"
			}

			fmt.Print("Subnet (optional, e.g., 172.20.0.0/16): ")
			fmt.Scanln(&subnet)

			if subnet != "" {
				fmt.Print("Gateway (optional, e.g., 172.20.0.1): ")
				fmt.Scanln(&gateway)
			}

			network := &DockerNetwork{
				Name:    networkName,
				Driver:  driver,
				Subnet:  subnet,
				Gateway: gateway,
			}

			profile.Networks[networkName] = network
			fmt.Printf("✅ Added network: %s\n", networkName)

			fmt.Print("Add another network? (y/N): ")

			var another string
			fmt.Scanln(&another)

			if strings.ToLower(another) != "y" && strings.ToLower(another) != "yes" {
				break
			}
		}
	}

	return dm.CreateProfile(profile)
}

func printProfilesJSON(profiles []*DockerNetworkProfile) error {
	return json.NewEncoder(os.Stdout).Encode(profiles)
}

func printProfilesYAML(profiles []*DockerNetworkProfile) error {
	data, err := yaml.Marshal(profiles)
	if err != nil {
		return err
	}

	fmt.Print(string(data))

	return nil
}

func printProfilesTable(profiles []*DockerNetworkProfile) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NAME\tDESCRIPTION\tNETWORKS\tCONTAINERS\tACTIVE\tCREATED")

	for _, profile := range profiles {
		active := "No"
		if profile.Active {
			active = "Yes"
		}

		created := profile.CreatedAt.Format("2006-01-02")
		if profile.CreatedAt.IsZero() {
			created = "-"
		}

		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%s\t%s\n",
			profile.Name,
			truncateString(profile.Description, 40),
			len(profile.Networks),
			len(profile.Containers),
			active,
			created)
	}

	return w.Flush()
}

func printProfileDetails(profile *DockerNetworkProfile) error {
	fmt.Printf("Profile: %s\n", profile.Name)

	if profile.Description != "" {
		fmt.Printf("Description: %s\n", profile.Description)
	}

	fmt.Printf("\nNetworks (%d):\n", len(profile.Networks))

	for name, network := range profile.Networks {
		fmt.Printf("  • %s (driver: %s", name, network.Driver)

		if network.Subnet != "" {
			fmt.Printf(", subnet: %s", network.Subnet)
		}

		fmt.Printf(")\n")
	}

	fmt.Printf("\nContainers (%d):\n", len(profile.Containers))

	for name, container := range profile.Containers {
		fmt.Printf("  • %s", name)

		if container.Image != "" {
			fmt.Printf(" (image: %s)", container.Image)
		}

		fmt.Printf("\n")
	}

	if profile.Compose != nil {
		fmt.Printf("\nDocker Compose:\n")
		fmt.Printf("  File: %s\n", profile.Compose.File)

		if profile.Compose.Project != "" {
			fmt.Printf("  Project: %s\n", profile.Compose.Project)
		}

		fmt.Printf("  Auto-apply: %t\n", profile.Compose.AutoApply)
	}

	return nil
}

func printNetworksTable(networks []*DockerNetworkStatus) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSCOPE\tCONTAINERS")

	for _, network := range networks {
		networkID := truncateString(network.NetworkID, 12)
		containerCount := strconv.Itoa(len(network.Containers))

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			networkID,
			network.Name,
			network.Driver,
			network.Scope,
			containerCount)
	}

	w.Flush()
}

func printContainersTable(containers []*ContainerNetworkInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprintln(w, "CONTAINER ID\tNAME\tIMAGE\tSTATE\tNETWORKS")

	for _, container := range containers {
		containerID := truncateString(container.ContainerID, 12)
		image := truncateString(container.Image, 30)
		networkCount := strconv.Itoa(len(container.Networks))

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			containerID,
			container.Name,
			image,
			container.State,
			networkCount)
	}

	w.Flush()
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	return s[:maxLen-3] + "..."
}

// newDockerNetworkContainerCmd creates the container subcommand.
func newDockerNetworkContainerCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Manage container-specific network configurations",
		Long:  `Manage container-specific network configurations within Docker network profiles.`,
	}

	// Add container subcommands
	cmd.AddCommand(newContainerAddCmd(dm))
	cmd.AddCommand(newContainerUpdateCmd(dm))
	cmd.AddCommand(newContainerRemoveCmd(dm))
	cmd.AddCommand(newContainerShowCmd(dm))
	cmd.AddCommand(newContainerConnectCmd(dm))
	cmd.AddCommand(newContainerDisconnectCmd(dm))

	return cmd
}

// newContainerAddCmd creates the container add subcommand.
func newContainerAddCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile-name] [container-name]",
		Short: "Add a container to a network profile",
		Long:  `Add a container with specific network configuration to a Docker network profile.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			containerName := args[1]

			// Get flags
			image, _ := cmd.Flags().GetString("image")
			networks, _ := cmd.Flags().GetStringSlice("network")
			ports, _ := cmd.Flags().GetStringSlice("port")
			envs, _ := cmd.Flags().GetStringSlice("env")
			dns, _ := cmd.Flags().GetStringSlice("dns")
			aliases, _ := cmd.Flags().GetStringSlice("alias")
			hostname, _ := cmd.Flags().GetString("hostname")

			// Build container configuration
			config := &ContainerNetwork{
				Image:        image,
				Networks:     networks,
				Ports:        ports,
				Environment:  make(map[string]string),
				DNSServers:   dns,
				NetworkAlias: aliases,
				Hostname:     hostname,
			}

			// Parse environment variables
			for _, env := range envs {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) == 2 {
					config.Environment[parts[0]] = parts[1]
				}
			}

			// Validate configuration
			if err := dm.ValidateContainerNetwork(config); err != nil {
				return fmt.Errorf("invalid container configuration: %w", err)
			}

			// Add container to profile
			if err := dm.UpdateContainerNetwork(profileName, containerName, config); err != nil {
				return fmt.Errorf("failed to add container: %w", err)
			}

			fmt.Printf("✅ Added container '%s' to profile '%s'\n", containerName, profileName)
			return nil
		},
	}

	cmd.Flags().StringP("image", "i", "", "Container image (required)")
	cmd.MarkFlagRequired("image")
	cmd.Flags().StringSliceP("network", "n", []string{}, "Networks to connect to")
	cmd.Flags().StringSliceP("port", "p", []string{}, "Port mappings (e.g., 80:80)")
	cmd.Flags().StringSliceP("env", "e", []string{}, "Environment variables (e.g., KEY=value)")
	cmd.Flags().StringSlice("dns", []string{}, "DNS servers")
	cmd.Flags().StringSlice("alias", []string{}, "Network aliases")
	cmd.Flags().String("hostname", "", "Container hostname")

	return cmd
}

// newContainerUpdateCmd creates the container update subcommand.
func newContainerUpdateCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [profile-name] [container-name]",
		Short: "Update container network configuration",
		Long:  `Update the network configuration for a container in a Docker network profile.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			containerName := args[1]

			// Load existing profile
			profile, err := dm.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			// Get existing container config
			config, exists := profile.Containers[containerName]
			if !exists {
				return fmt.Errorf("container '%s' not found in profile '%s'", containerName, profileName)
			}

			// Update with new values if provided
			if image, _ := cmd.Flags().GetString("image"); image != "" {
				config.Image = image
			}

			if networks, _ := cmd.Flags().GetStringSlice("network"); len(networks) > 0 {
				config.Networks = networks
			}

			if ports, _ := cmd.Flags().GetStringSlice("port"); len(ports) > 0 {
				config.Ports = ports
			}

			if envs, _ := cmd.Flags().GetStringSlice("env"); len(envs) > 0 {
				config.Environment = make(map[string]string)
				for _, env := range envs {
					parts := strings.SplitN(env, "=", 2)
					if len(parts) == 2 {
						config.Environment[parts[0]] = parts[1]
					}
				}
			}

			if dns, _ := cmd.Flags().GetStringSlice("dns"); len(dns) > 0 {
				config.DNSServers = dns
			}

			if aliases, _ := cmd.Flags().GetStringSlice("alias"); len(aliases) > 0 {
				config.NetworkAlias = aliases
			}

			if hostname, _ := cmd.Flags().GetString("hostname"); hostname != "" {
				config.Hostname = hostname
			}

			// Validate updated configuration
			if err := dm.ValidateContainerNetwork(config); err != nil {
				return fmt.Errorf("invalid container configuration: %w", err)
			}

			// Update container in profile
			if err := dm.UpdateContainerNetwork(profileName, containerName, config); err != nil {
				return fmt.Errorf("failed to update container: %w", err)
			}

			fmt.Printf("✅ Updated container '%s' in profile '%s'\n", containerName, profileName)
			return nil
		},
	}

	cmd.Flags().StringP("image", "i", "", "Container image")
	cmd.Flags().StringSliceP("network", "n", []string{}, "Networks to connect to")
	cmd.Flags().StringSliceP("port", "p", []string{}, "Port mappings (e.g., 80:80)")
	cmd.Flags().StringSliceP("env", "e", []string{}, "Environment variables (e.g., KEY=value)")
	cmd.Flags().StringSlice("dns", []string{}, "DNS servers")
	cmd.Flags().StringSlice("alias", []string{}, "Network aliases")
	cmd.Flags().String("hostname", "", "Container hostname")

	return cmd
}

// newContainerRemoveCmd creates the container remove subcommand.
func newContainerRemoveCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove [profile-name] [container-name]",
		Short:   "Remove a container from a network profile",
		Long:    `Remove a container from a Docker network profile.`,
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			containerName := args[1]

			if err := dm.RemoveContainerFromProfile(profileName, containerName); err != nil {
				return fmt.Errorf("failed to remove container: %w", err)
			}

			fmt.Printf("✅ Removed container '%s' from profile '%s'\n", containerName, profileName)
			return nil
		},
	}

	return cmd
}

// newContainerShowCmd creates the container show subcommand.
func newContainerShowCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [profile-name] [container-name]",
		Short: "Show container network configuration",
		Long:  `Show the network configuration for a container in a Docker network profile.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			containerName := args[1]

			profile, err := dm.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			config, exists := profile.Containers[containerName]
			if !exists {
				return fmt.Errorf("container '%s' not found in profile '%s'", containerName, profileName)
			}

			output, _ := cmd.Flags().GetString("output")
			switch output {
			case "json":
				data, err := json.MarshalIndent(config, "", "  ")
				if err != nil {
					return err
				}
				fmt.Println(string(data))
			case "yaml":
				data, err := yaml.Marshal(config)
				if err != nil {
					return err
				}
				fmt.Print(string(data))
			default:
				// Table format
				fmt.Printf("Container: %s\n", containerName)
				fmt.Printf("Image: %s\n", config.Image)
				if config.NetworkMode != "" {
					fmt.Printf("Network Mode: %s\n", config.NetworkMode)
				}
				if len(config.Networks) > 0 {
					fmt.Printf("Networks: %s\n", strings.Join(config.Networks, ", "))
				}
				if len(config.Ports) > 0 {
					fmt.Printf("Ports: %s\n", strings.Join(config.Ports, ", "))
				}
				if len(config.Environment) > 0 {
					fmt.Println("Environment:")
					for k, v := range config.Environment {
						fmt.Printf("  %s=%s\n", k, v)
					}
				}
				if len(config.DNSServers) > 0 {
					fmt.Printf("DNS Servers: %s\n", strings.Join(config.DNSServers, ", "))
				}
				if len(config.NetworkAlias) > 0 {
					fmt.Printf("Network Aliases: %s\n", strings.Join(config.NetworkAlias, ", "))
				}
				if config.Hostname != "" {
					fmt.Printf("Hostname: %s\n", config.Hostname)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newContainerConnectCmd creates the container connect subcommand.
func newContainerConnectCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect [container-name] [network-name]",
		Short: "Connect a running container to a network",
		Long:  `Connect a running Docker container to a network with optional aliases.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerName := args[0]
			networkName := args[1]

			aliases, _ := cmd.Flags().GetStringSlice("alias")

			// Build connect command
			connectCmd := fmt.Sprintf("docker network connect %s %s", networkName, containerName)
			for _, alias := range aliases {
				connectCmd += fmt.Sprintf(" --alias %s", alias)
			}

			result, err := dm.executor.ExecuteWithTimeout(context.Background(), connectCmd, 10*time.Second)
			if err != nil || result.ExitCode != 0 {
				return fmt.Errorf("failed to connect container to network: %w", err)
			}

			fmt.Printf("✅ Connected container '%s' to network '%s'\n", containerName, networkName)
			if len(aliases) > 0 {
				fmt.Printf("   Aliases: %s\n", strings.Join(aliases, ", "))
			}

			return nil
		},
	}

	cmd.Flags().StringSlice("alias", []string{}, "Network aliases for the container")

	return cmd
}

// newContainerDisconnectCmd creates the container disconnect subcommand.
func newContainerDisconnectCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disconnect [container-name] [network-name]",
		Short: "Disconnect a running container from a network",
		Long:  `Disconnect a running Docker container from a network.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			containerName := args[0]
			networkName := args[1]

			if err := dm.DisconnectContainerFromNetwork(containerName, networkName); err != nil {
				return fmt.Errorf("failed to disconnect container: %w", err)
			}

			fmt.Printf("✅ Disconnected container '%s' from network '%s'\n", containerName, networkName)
			return nil
		},
	}

	return cmd
}

// newDockerNetworkCloneCmd creates the clone subcommand.
func newDockerNetworkCloneCmd(dm *DockerNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clone [source-profile] [target-profile]",
		Short: "Clone an existing Docker network profile",
		Long:  `Create a copy of an existing Docker network profile with a new name.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			sourceName := args[0]
			targetName := args[1]

			if err := dm.CloneProfile(sourceName, targetName); err != nil {
				return fmt.Errorf("failed to clone profile: %w", err)
			}

			fmt.Printf("✅ Cloned profile '%s' to '%s'\n", sourceName, targetName)
			return nil
		},
	}

	return cmd
}
