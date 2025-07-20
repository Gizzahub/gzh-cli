package netenv

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Service mesh constants.
const (
	serviceMeshIstio   = "istio"
	serviceMeshLinkerd = "linkerd"
)

// newServiceMeshCmd creates the service-mesh subcommand.
func newServiceMeshCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-mesh",
		Short:   "Manage service mesh integration (Istio/Linkerd)",
		Long:    `Configure and manage service mesh integration for Kubernetes network profiles, including Istio and Linkerd configurations.`,
		Aliases: []string{"sm", "mesh"},
	}

	// Add subcommands
	cmd.AddCommand(newServiceMeshDetectCmd(km))
	cmd.AddCommand(newServiceMeshEnableCmd(km))
	cmd.AddCommand(newServiceMeshDisableCmd(km))
	cmd.AddCommand(newServiceMeshConfigureCmd(km))
	cmd.AddCommand(newServiceMeshStatusCmd(km))
	cmd.AddCommand(newServiceMeshIstioCmd(km))
	cmd.AddCommand(newServiceMeshLinkerdCmd(km))

	return cmd
}

// newServiceMeshDetectCmd creates the detect subcommand.
func newServiceMeshDetectCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detect",
		Short: "Detect installed service mesh",
		Long:  `Detect which service mesh (Istio or Linkerd) is installed in the cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			meshType, err := km.DetectServiceMesh(cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to detect service mesh: %w", err)
			}

			if meshType == "" {
				fmt.Println("❌ No service mesh detected in the cluster")
				fmt.Println("\n💡 To install a service mesh:")
				fmt.Println("   - Istio: https://istio.io/latest/docs/setup/getting-started/")
				fmt.Println("   - Linkerd: https://linkerd.io/2/getting-started/")
			} else {
				fmt.Printf("✅ Detected service mesh: %s\n", meshType)
			}

			return nil
		},
	}

	return cmd
}

// newServiceMeshEnableCmd creates the enable subcommand.
func newServiceMeshEnableCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable [profile-name]",
		Short: "Enable service mesh for a profile",
		Long:  `Enable service mesh integration for a Kubernetes network profile.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Load profile
			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			// Get mesh type
			meshType, _ := cmd.Flags().GetString("type")
			if meshType == "" {
				// Auto-detect
				detectedType, err := km.DetectServiceMesh(cmd.Context())
				if err != nil {
					return fmt.Errorf("failed to detect service mesh: %w", err)
				}
				if detectedType == "" {
					return fmt.Errorf("no service mesh detected. Please install Istio or Linkerd first")
				}
				meshType = detectedType
			}

			// Enable service mesh
			if profile.ServiceMesh == nil {
				profile.ServiceMesh = &ServiceMeshConfig{
					Type:          meshType,
					Enabled:       true,
					Namespace:     profile.Namespace,
					TrafficPolicy: make(map[string]interface{}),
				}
			} else {
				profile.ServiceMesh.Type = meshType
				profile.ServiceMesh.Enabled = true
			}

			// Save profile
			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Enabled %s service mesh for profile '%s'\n", meshType, profileName)

			// Apply if requested
			apply, _ := cmd.Flags().GetBool("apply")
			if apply {
				fmt.Printf("⏳ Applying service mesh configuration...\n")
				if err := km.ApplyServiceMeshConfig(profile); err != nil {
					return fmt.Errorf("failed to apply service mesh config: %w", err)
				}
				fmt.Printf("✅ Service mesh configuration applied\n")
			}

			return nil
		},
	}

	cmd.Flags().String("type", "", "Service mesh type (istio or linkerd)")
	cmd.Flags().Bool("apply", false, "Apply configuration immediately")

	return cmd
}

// newServiceMeshDisableCmd creates the disable subcommand.
func newServiceMeshDisableCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable [profile-name]",
		Short: "Disable service mesh for a profile",
		Long:  `Disable service mesh integration for a Kubernetes network profile.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Load profile
			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || !profile.ServiceMesh.Enabled {
				fmt.Printf("ℹ️  Service mesh is already disabled for profile '%s'\n", profileName)
				return nil
			}

			// Disable service mesh
			profile.ServiceMesh.Enabled = false

			// Save profile
			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Disabled service mesh for profile '%s'\n", profileName)
			fmt.Println("⚠️  Note: This does not remove existing service mesh resources from the cluster")

			return nil
		},
	}

	return cmd
}

// newServiceMeshConfigureCmd creates the configure subcommand.
func newServiceMeshConfigureCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure [profile-name]",
		Short: "Configure service mesh settings",
		Long:  `Configure service mesh settings for a profile, including traffic policies, circuit breakers, and retry policies.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			// Load profile
			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil {
				return fmt.Errorf("service mesh is not enabled for profile '%s'", profileName)
			}

			// Configure based on mesh type
			switch profile.ServiceMesh.Type {
			case serviceMeshIstio:
				return configureIstioInteractive(km, profile)
			case serviceMeshLinkerd:
				return configureLinkerdInteractive(km, profile)
			default:
				return fmt.Errorf("unknown service mesh type: %s", profile.ServiceMesh.Type)
			}
		},
	}

	return cmd
}

// newServiceMeshStatusCmd creates the status subcommand.
func newServiceMeshStatusCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show service mesh status",
		Long:  `Show the current status of service mesh configuration in a namespace.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			namespace, _ := cmd.Flags().GetString("namespace")
			output, _ := cmd.Flags().GetString("output")

			status, err := km.GetServiceMeshStatus(namespace)
			if err != nil {
				return fmt.Errorf("failed to get service mesh status: %w", err)
			}

			switch output {
			case "json":
				return json.NewEncoder(os.Stdout).Encode(status)
			case "yaml":
				return yaml.NewEncoder(os.Stdout).Encode(status)
			default:
				return printServiceMeshStatus(status)
			}
		},
	}

	cmd.Flags().StringP("namespace", "n", "default", "Target namespace")
	cmd.Flags().StringP("output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newServiceMeshIstioCmd creates the istio subcommand.
func newServiceMeshIstioCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "istio",
		Short: "Manage Istio-specific configurations",
		Long:  `Manage Istio-specific configurations including VirtualServices, DestinationRules, ServiceEntries, and Gateways.`,
	}

	// Add Istio subcommands
	cmd.AddCommand(newIstioVirtualServiceCmd(km))
	cmd.AddCommand(newIstioDestinationRuleCmd(km))
	cmd.AddCommand(newIstioServiceEntryCmd(km))
	cmd.AddCommand(newIstioGatewayCmd(km))

	return cmd
}

// newServiceMeshLinkerdCmd creates the linkerd subcommand.
func newServiceMeshLinkerdCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "linkerd",
		Short: "Manage Linkerd-specific configurations",
		Long:  `Manage Linkerd-specific configurations including ServiceProfiles and TrafficSplits.`,
	}

	// Add Linkerd subcommands
	cmd.AddCommand(newLinkerdServiceProfileCmd(km))
	cmd.AddCommand(newLinkerdTrafficSplitCmd(km))

	return cmd
}

// Istio-specific commands

func newIstioVirtualServiceCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "virtual-service",
		Short:   "Manage Istio VirtualServices",
		Aliases: []string{"vs"},
	}

	// Add VirtualService
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [service-name]",
		Short: "Add a VirtualService to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			serviceName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "istio" {
				return fmt.Errorf("profile '%s' is not configured for Istio", profileName)
			}

			// Get flags
			hosts, _ := cmd.Flags().GetStringSlice("hosts")
			gateways, _ := cmd.Flags().GetStringSlice("gateways")
			destination, _ := cmd.Flags().GetString("destination")
			subset, _ := cmd.Flags().GetString("subset")
			weight, _ := cmd.Flags().GetInt32("weight")
			timeout, _ := cmd.Flags().GetString("timeout")

			// Create VirtualService
			vs := &IstioVirtualService{
				Name:     serviceName,
				Hosts:    hosts,
				Gateways: gateways,
			}

			// Add basic HTTP route
			if destination != "" {
				route := IstioHTTPRoute{
					Route: []IstioHTTPRouteDestination{
						{
							Destination: &IstioDestination{
								Host:   destination,
								Subset: subset,
							},
							Weight: weight,
						},
					},
				}
				if timeout != "" {
					route.Timeout = timeout
				}
				vs.HTTP = []IstioHTTPRoute{route}
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig)
			if !ok {
				istioConfig = &IstioConfig{
					VirtualServices:  make(map[string]*IstioVirtualService),
					DestinationRules: make(map[string]*IstioDestinationRule),
					ServiceEntries:   make(map[string]*IstioServiceEntry),
					Gateways:         make(map[string]*IstioGateway),
				}
				profile.ServiceMesh.TrafficPolicy["istio"] = istioConfig
			}

			istioConfig.VirtualServices[serviceName] = vs

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added VirtualService '%s' to profile '%s'\n", serviceName, profileName)
			return nil
		},
	}

	addCmd.Flags().StringSlice("hosts", []string{}, "Destination hosts")
	addCmd.Flags().StringSlice("gateways", []string{}, "Gateways (leave empty for mesh)")
	addCmd.Flags().String("destination", "", "Destination service")
	addCmd.Flags().String("subset", "", "Destination subset")
	addCmd.Flags().Int32("weight", 100, "Route weight")
	addCmd.Flags().String("timeout", "", "Request timeout (e.g., 30s)")

	// List VirtualServices
	listCmd := &cobra.Command{
		Use:   "list [profile-name]",
		Short: "List VirtualServices in a profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.TrafficPolicy == nil {
				fmt.Println("No VirtualServices configured")
				return nil
			}

			if istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig); ok {
				if len(istioConfig.VirtualServices) == 0 {
					fmt.Println("No VirtualServices configured")
					return nil
				}

				w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
				fmt.Fprintln(w, "NAME\tHOSTS\tGATEWAYS\tROUTES")
				for name, vs := range istioConfig.VirtualServices {
					hosts := strings.Join(vs.Hosts, ", ")
					gateways := strings.Join(vs.Gateways, ", ")
					if gateways == "" {
						gateways = "mesh"
					}
					routes := fmt.Sprintf("%d HTTP", len(vs.HTTP))
					if len(vs.TCP) > 0 {
						routes += fmt.Sprintf(", %d TCP", len(vs.TCP))
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", name, hosts, gateways, routes)
				}
				w.Flush()
			}

			return nil
		},
	}

	cmd.AddCommand(addCmd)
	cmd.AddCommand(listCmd)

	return cmd
}

func newIstioDestinationRuleCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "destination-rule",
		Short:   "Manage Istio DestinationRules",
		Aliases: []string{"dr"},
	}

	// Add DestinationRule
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [service-name]",
		Short: "Add a DestinationRule to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			serviceName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "istio" {
				return fmt.Errorf("profile '%s' is not configured for Istio", profileName)
			}

			// Get flags
			host, _ := cmd.Flags().GetString("host")
			loadBalancer, _ := cmd.Flags().GetString("load-balancer")
			consecutiveErrors, _ := cmd.Flags().GetInt32("consecutive-errors")
			interval, _ := cmd.Flags().GetString("interval")
			baseEjectionTime, _ := cmd.Flags().GetString("base-ejection-time")
			tlsMode, _ := cmd.Flags().GetString("tls-mode")

			// Create DestinationRule
			dr := &IstioDestinationRule{
				Name:          serviceName,
				Host:          host,
				TrafficPolicy: &IstioTrafficPolicy{},
			}

			// Configure load balancer
			if loadBalancer != "" {
				dr.TrafficPolicy.LoadBalancer = &LoadBalancerSettings{
					Simple: loadBalancer,
				}
			}

			// Configure outlier detection
			if consecutiveErrors > 0 {
				dr.TrafficPolicy.OutlierDetection = &OutlierDetection{
					ConsecutiveErrors: consecutiveErrors,
					Interval:          interval,
					BaseEjectionTime:  baseEjectionTime,
				}
			}

			// Configure TLS
			if tlsMode != "" {
				dr.TrafficPolicy.TLS = &ClientTLSSettings{
					Mode: tlsMode,
				}
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig)
			if !ok {
				istioConfig = &IstioConfig{
					VirtualServices:  make(map[string]*IstioVirtualService),
					DestinationRules: make(map[string]*IstioDestinationRule),
					ServiceEntries:   make(map[string]*IstioServiceEntry),
					Gateways:         make(map[string]*IstioGateway),
				}
				profile.ServiceMesh.TrafficPolicy["istio"] = istioConfig
			}

			istioConfig.DestinationRules[serviceName] = dr

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added DestinationRule '%s' to profile '%s'\n", serviceName, profileName)
			return nil
		},
	}

	addCmd.Flags().String("host", "", "Destination host")
	addCmd.Flags().String("load-balancer", "ROUND_ROBIN", "Load balancer policy (ROUND_ROBIN, LEAST_CONN, RANDOM, PASSTHROUGH)")
	addCmd.Flags().Int32("consecutive-errors", 5, "Consecutive errors before ejection")
	addCmd.Flags().String("interval", "30s", "Analysis interval")
	addCmd.Flags().String("base-ejection-time", "30s", "Minimum ejection duration")
	addCmd.Flags().String("tls-mode", "", "TLS mode (DISABLE, SIMPLE, MUTUAL, ISTIO_MUTUAL)")

	cmd.AddCommand(addCmd)

	return cmd
}

func newIstioServiceEntryCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-entry",
		Short:   "Manage Istio ServiceEntries",
		Aliases: []string{"se"},
	}

	// Add ServiceEntry
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [service-name]",
		Short: "Add a ServiceEntry to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			serviceName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "istio" {
				return fmt.Errorf("profile '%s' is not configured for Istio", profileName)
			}

			// Get flags
			hosts, _ := cmd.Flags().GetStringSlice("hosts")
			location, _ := cmd.Flags().GetString("location")
			resolution, _ := cmd.Flags().GetString("resolution")
			ports, _ := cmd.Flags().GetStringSlice("ports")

			// Create ServiceEntry
			se := &IstioServiceEntry{
				Name:       serviceName,
				Hosts:      hosts,
				Location:   location,
				Resolution: resolution,
				Ports:      []ServicePort{},
			}

			// Parse ports
			for _, portSpec := range ports {
				parts := strings.Split(portSpec, ":")
				if len(parts) == 3 {
					var port int32
					fmt.Sscanf(parts[1], "%d", &port)
					se.Ports = append(se.Ports, ServicePort{
						Name:     parts[0],
						Port:     port,
						Protocol: parts[2],
					})
				}
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig)
			if !ok {
				istioConfig = &IstioConfig{
					VirtualServices:  make(map[string]*IstioVirtualService),
					DestinationRules: make(map[string]*IstioDestinationRule),
					ServiceEntries:   make(map[string]*IstioServiceEntry),
					Gateways:         make(map[string]*IstioGateway),
				}
				profile.ServiceMesh.TrafficPolicy["istio"] = istioConfig
			}

			istioConfig.ServiceEntries[serviceName] = se

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added ServiceEntry '%s' to profile '%s'\n", serviceName, profileName)
			return nil
		},
	}

	addCmd.Flags().StringSlice("hosts", []string{}, "Service hosts")
	addCmd.Flags().String("location", "MESH_EXTERNAL", "Service location (MESH_EXTERNAL or MESH_INTERNAL)")
	addCmd.Flags().String("resolution", "DNS", "Service resolution (NONE, STATIC, DNS)")
	addCmd.Flags().StringSlice("ports", []string{}, "Service ports (format: name:number:protocol)")

	cmd.AddCommand(addCmd)

	return cmd
}

func newIstioGatewayCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gateway",
		Short:   "Manage Istio Gateways",
		Aliases: []string{"gw"},
	}

	// Add Gateway
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [gateway-name]",
		Short: "Add a Gateway to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			gatewayName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "istio" {
				return fmt.Errorf("profile '%s' is not configured for Istio", profileName)
			}

			// Get flags
			selector, _ := cmd.Flags().GetStringToString("selector")
			hosts, _ := cmd.Flags().GetStringSlice("hosts")
			port, _ := cmd.Flags().GetInt32("port")
			protocol, _ := cmd.Flags().GetString("protocol")
			tlsMode, _ := cmd.Flags().GetString("tls-mode")

			// Create Gateway
			gw := &IstioGateway{
				Name:     gatewayName,
				Selector: selector,
				Servers: []IstioServer{
					{
						Port: &GatewayPort{
							Number:   port,
							Name:     fmt.Sprintf("%s-%d", protocol, port),
							Protocol: protocol,
						},
						Hosts: hosts,
					},
				},
			}

			// Configure TLS if needed
			if tlsMode != "" {
				gw.Servers[0].TLS = &ServerTLSSettings{
					Mode: tlsMode,
				}
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			istioConfig, ok := profile.ServiceMesh.TrafficPolicy["istio"].(*IstioConfig)
			if !ok {
				istioConfig = &IstioConfig{
					VirtualServices:  make(map[string]*IstioVirtualService),
					DestinationRules: make(map[string]*IstioDestinationRule),
					ServiceEntries:   make(map[string]*IstioServiceEntry),
					Gateways:         make(map[string]*IstioGateway),
				}
				profile.ServiceMesh.TrafficPolicy["istio"] = istioConfig
			}

			istioConfig.Gateways[gatewayName] = gw

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added Gateway '%s' to profile '%s'\n", gatewayName, profileName)
			return nil
		},
	}

	addCmd.Flags().StringToString("selector", map[string]string{"istio": "ingressgateway"}, "Gateway selector labels")
	addCmd.Flags().StringSlice("hosts", []string{"*"}, "Gateway hosts")
	addCmd.Flags().Int32("port", 80, "Gateway port")
	addCmd.Flags().String("protocol", "HTTP", "Gateway protocol (HTTP, HTTPS, GRPC, HTTP2, MONGO, TCP, TLS)")
	addCmd.Flags().String("tls-mode", "", "TLS mode (PASSTHROUGH, SIMPLE, MUTUAL)")

	cmd.AddCommand(addCmd)

	return cmd
}

// Linkerd-specific commands

func newLinkerdServiceProfileCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "service-profile",
		Short:   "Manage Linkerd ServiceProfiles",
		Aliases: []string{"sp"},
	}

	// Add ServiceProfile
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [service-name]",
		Short: "Add a ServiceProfile to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			serviceName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "linkerd" {
				return fmt.Errorf("profile '%s' is not configured for Linkerd", profileName)
			}

			// Get flags
			routeName, _ := cmd.Flags().GetString("route-name")
			method, _ := cmd.Flags().GetString("method")
			pathRegex, _ := cmd.Flags().GetString("path-regex")
			timeout, _ := cmd.Flags().GetString("timeout")
			isRetryable, _ := cmd.Flags().GetBool("retryable")
			retryRatio, _ := cmd.Flags().GetFloat32("retry-ratio")
			minRetries, _ := cmd.Flags().GetInt32("min-retries")

			// Create ServiceProfile
			sp := &LinkerdServiceProfile{
				Name: serviceName,
				Routes: []LinkerdRoute{
					{
						Name: routeName,
						Condition: &LinkerdCondition{
							Method:    method,
							PathRegex: pathRegex,
						},
						Timeout:     timeout,
						IsRetryable: isRetryable,
					},
				},
			}

			// Configure retry budget
			if retryRatio > 0 {
				sp.RetryBudget = &RetryBudgetConfig{
					RetryRatio:          retryRatio,
					MinRetriesPerSecond: minRetries,
					TTL:                 "10s",
				}
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			linkerdConfig, ok := profile.ServiceMesh.TrafficPolicy["linkerd"].(*LinkerdConfig)
			if !ok {
				linkerdConfig = &LinkerdConfig{
					ServiceProfiles: make(map[string]*LinkerdServiceProfile),
					TrafficSplits:   make(map[string]*LinkerdTrafficSplit),
				}
				profile.ServiceMesh.TrafficPolicy["linkerd"] = linkerdConfig
			}

			linkerdConfig.ServiceProfiles[serviceName] = sp

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added ServiceProfile '%s' to profile '%s'\n", serviceName, profileName)
			return nil
		},
	}

	addCmd.Flags().String("route-name", "default", "Route name")
	addCmd.Flags().String("method", "", "HTTP method (GET, POST, etc.)")
	addCmd.Flags().String("path-regex", ".*", "Path regex pattern")
	addCmd.Flags().String("timeout", "30s", "Request timeout")
	addCmd.Flags().Bool("retryable", true, "Whether the route is retryable")
	addCmd.Flags().Float32("retry-ratio", 0.2, "Retry ratio")
	addCmd.Flags().Int32("min-retries", 10, "Minimum retries per second")

	cmd.AddCommand(addCmd)

	return cmd
}

func newLinkerdTrafficSplitCmd(km *KubernetesNetworkManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "traffic-split",
		Short:   "Manage Linkerd TrafficSplits",
		Aliases: []string{"ts"},
	}

	// Add TrafficSplit
	addCmd := &cobra.Command{
		Use:   "add [profile-name] [split-name]",
		Short: "Add a TrafficSplit to a profile",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			profileName := args[0]
			splitName := args[1]

			profile, err := km.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if profile.ServiceMesh == nil || profile.ServiceMesh.Type != "linkerd" {
				return fmt.Errorf("profile '%s' is not configured for Linkerd", profileName)
			}

			// Get flags
			service, _ := cmd.Flags().GetString("service")
			backends, _ := cmd.Flags().GetStringSlice("backends")

			// Create TrafficSplit
			ts := &LinkerdTrafficSplit{
				Name:     splitName,
				Service:  service,
				Backends: []LinkerdBackend{},
			}

			// Parse backends (format: service:weight)
			totalWeight := int32(0)
			for _, backend := range backends {
				parts := strings.Split(backend, ":")
				if len(parts) == 2 {
					var weight int32
					fmt.Sscanf(parts[1], "%d", &weight)
					ts.Backends = append(ts.Backends, LinkerdBackend{
						Service: parts[0],
						Weight:  weight,
					})
					totalWeight += weight
				}
			}

			// Validate weights
			if totalWeight != 100 {
				return fmt.Errorf("backend weights must sum to 100, got %d", totalWeight)
			}

			// Add to profile
			if profile.ServiceMesh.TrafficPolicy == nil {
				profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
			}

			linkerdConfig, ok := profile.ServiceMesh.TrafficPolicy["linkerd"].(*LinkerdConfig)
			if !ok {
				linkerdConfig = &LinkerdConfig{
					ServiceProfiles: make(map[string]*LinkerdServiceProfile),
					TrafficSplits:   make(map[string]*LinkerdTrafficSplit),
				}
				profile.ServiceMesh.TrafficPolicy["linkerd"] = linkerdConfig
			}

			linkerdConfig.TrafficSplits[splitName] = ts

			if err := km.saveProfile(profile); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			fmt.Printf("✅ Added TrafficSplit '%s' to profile '%s'\n", splitName, profileName)
			return nil
		},
	}

	addCmd.Flags().String("service", "", "Root service name")
	addCmd.Flags().StringSlice("backends", []string{}, "Backend services (format: service:weight)")

	cmd.AddCommand(addCmd)

	return cmd
}

// Helper functions

func configureIstioInteractive(km *KubernetesNetworkManager, profile *KubernetesNetworkProfile) error {
	fmt.Println("🔧 Configuring Istio settings")

	// Simple configuration for now
	fmt.Print("Enable mTLS mode? (DISABLE/SIMPLE/MUTUAL/ISTIO_MUTUAL) [ISTIO_MUTUAL]: ")

	var mtlsMode string
	fmt.Scanln(&mtlsMode)

	if mtlsMode == "" {
		mtlsMode = "ISTIO_MUTUAL"
	}

	// Initialize Istio config
	if profile.ServiceMesh.TrafficPolicy == nil {
		profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
	}

	istioConfig := &IstioConfig{
		MTLSMode:         mtlsMode,
		SidecarInjection: true,
		VirtualServices:  make(map[string]*IstioVirtualService),
		DestinationRules: make(map[string]*IstioDestinationRule),
		ServiceEntries:   make(map[string]*IstioServiceEntry),
		Gateways:         make(map[string]*IstioGateway),
	}

	profile.ServiceMesh.TrafficPolicy["istio"] = istioConfig

	if err := km.saveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Println("✅ Istio configuration saved")
	fmt.Println("💡 Use 'gz net-env kubernetes-network service-mesh istio' commands to manage specific resources")

	return nil
}

func configureLinkerdInteractive(km *KubernetesNetworkManager, profile *KubernetesNetworkProfile) error {
	fmt.Println("🔧 Configuring Linkerd settings")

	// Simple configuration for now
	fmt.Print("Enable proxy injection? (y/N): ")

	var enableProxy string
	fmt.Scanln(&enableProxy)

	// Initialize Linkerd config
	if profile.ServiceMesh.TrafficPolicy == nil {
		profile.ServiceMesh.TrafficPolicy = make(map[string]interface{})
	}

	linkerdConfig := &LinkerdConfig{
		ProxyInjection:  strings.ToLower(enableProxy) == "y",
		ServiceProfiles: make(map[string]*LinkerdServiceProfile),
		TrafficSplits:   make(map[string]*LinkerdTrafficSplit),
	}

	// Set default proxy resources
	linkerdConfig.ProxyResources = &ProxyResourceConfig{
		CPURequest:    "100m",
		CPULimit:      "250m",
		MemoryRequest: "64Mi",
		MemoryLimit:   "256Mi",
	}

	profile.ServiceMesh.TrafficPolicy["linkerd"] = linkerdConfig

	if err := km.saveProfile(profile); err != nil {
		return fmt.Errorf("failed to save profile: %w", err)
	}

	fmt.Println("✅ Linkerd configuration saved")
	fmt.Println("💡 Use 'gz net-env kubernetes-network service-mesh linkerd' commands to manage specific resources")

	return nil
}

func printServiceMeshStatus(status map[string]interface{}) error {
	fmt.Printf("🌐 Service Mesh Status\n")
	fmt.Printf("   Namespace: %s\n", status["namespace"])

	meshType, _ := status["mesh_type"].(string)
	if meshType == "" {
		fmt.Println("   Status: No service mesh detected")
		return nil
	}

	fmt.Printf("   Type: %s\n", meshType)

	switch meshType {
	case "istio":
		if injection, ok := status["sidecar_injection"].(bool); ok {
			fmt.Printf("   Sidecar Injection: %v\n", injection)
		}
	case "linkerd":
		if injection, ok := status["proxy_injection"].(bool); ok {
			fmt.Printf("   Proxy Injection: %v\n", injection)
		}
	}

	if resources, ok := status["resources"].(map[string]int); ok {
		fmt.Println("\n📊 Resources:")

		for resource, count := range resources {
			fmt.Printf("   - %s: %d\n", resource, count)
		}
	}

	return nil
}
