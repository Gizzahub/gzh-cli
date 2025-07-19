package netenv

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceMeshIntegrationAdvanced(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "k8s_service_mesh_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("ServiceMeshDetection", func(t *testing.T) {
		// This test would normally detect real service mesh installations
		// For unit testing, we just verify the function doesn't error
		meshType, err := km.DetectServiceMesh(context.Background())
		assert.NoError(t, err)
		// In a real cluster, meshType would be "istio", "linkerd", or ""
		assert.NotNil(t, meshType) // Can be empty string
	})

	t.Run("ServiceMeshConfiguration", func(t *testing.T) {
		// Create a profile with service mesh enabled
		profile := &KubernetesNetworkProfile{
			Name:        "mesh-test",
			Description: "Service mesh test profile",
			Namespace:   "mesh-namespace",
			Policies:    make(map[string]*NetworkPolicyConfig),
			ServiceMesh: &ServiceMeshConfig{
				Type:          "istio",
				Enabled:       true,
				Namespace:     "mesh-namespace",
				TrafficPolicy: make(map[string]interface{}),
			},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Load and verify
		loadedProfile, err := km.LoadProfile("mesh-test")
		require.NoError(t, err)
		assert.NotNil(t, loadedProfile.ServiceMesh)
		assert.Equal(t, "istio", loadedProfile.ServiceMesh.Type)
		assert.True(t, loadedProfile.ServiceMesh.Enabled)
	})

	t.Run("ValidateServiceMeshConfig", func(t *testing.T) {
		// Test valid Istio config
		istioConfig := &ServiceMeshConfig{
			Type:      "istio",
			Enabled:   true,
			Namespace: "test",
		}
		err := km.ValidateServiceMeshConfig(istioConfig)
		assert.NoError(t, err)

		// Test valid Linkerd config
		linkerdConfig := &ServiceMeshConfig{
			Type:      "linkerd",
			Enabled:   true,
			Namespace: "test",
		}
		err = km.ValidateServiceMeshConfig(linkerdConfig)
		assert.NoError(t, err)

		// Test invalid mesh type
		invalidConfig := &ServiceMeshConfig{
			Type:    "invalid-mesh",
			Enabled: true,
		}
		err = km.ValidateServiceMeshConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported service mesh type")
	})
}

func TestIstioConfiguration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "istio_config_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("IstioVirtualService", func(t *testing.T) {
		vs := &IstioVirtualService{
			Name:  "test-vs",
			Hosts: []string{"test-service"},
			HTTP: []IstioHTTPRoute{
				{
					Name: "test-route",
					Match: []IstioHTTPMatchRequest{
						{
							URI: &StringMatch{Prefix: "/api"},
						},
					},
					Route: []IstioHTTPRouteDestination{
						{
							Destination: &IstioDestination{
								Host:   "test-service",
								Subset: "v1",
							},
							Weight: 80,
						},
						{
							Destination: &IstioDestination{
								Host:   "test-service",
								Subset: "v2",
							},
							Weight: 20,
						},
					},
					Timeout: "30s",
				},
			},
		}

		manifest, err := km.GenerateIstioVirtualService("test-namespace", "test-vs", vs)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "networking.istio.io/v1beta1", manifest["apiVersion"])
		assert.Equal(t, "VirtualService", manifest["kind"])

		metadata, ok := manifest["metadata"].(map[string]interface{})
		assert.True(t, ok, "metadata should be a map")
		assert.Equal(t, "test-vs", metadata["name"])
		assert.Equal(t, "test-namespace", metadata["namespace"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		assert.Equal(t, []string{"test-service"}, spec["hosts"])
		assert.NotNil(t, spec["http"])
	})

	t.Run("IstioDestinationRule", func(t *testing.T) {
		dr := &IstioDestinationRule{
			Name: "test-dr",
			Host: "test-service",
			TrafficPolicy: &IstioTrafficPolicy{
				LoadBalancer: &LoadBalancerSettings{
					Simple: "ROUND_ROBIN",
				},
				ConnectionPool: &ConnectionPoolSettings{
					TCP: &TCPSettings{
						MaxConnections: 100,
						ConnectTimeout: "30s",
					},
					HTTP: &HTTPSettings{
						HTTP1MaxPendingRequests: 100,
						HTTP2MaxRequests:        1000,
					},
				},
				OutlierDetection: &OutlierDetection{
					ConsecutiveErrors:  5,
					Interval:           "30s",
					BaseEjectionTime:   "30s",
					MaxEjectionPercent: 50,
				},
				TLS: &ClientTLSSettings{
					Mode: "ISTIO_MUTUAL",
				},
			},
			Subsets: []IstioSubset{
				{
					Name:   "v1",
					Labels: map[string]string{"version": "v1"},
				},
				{
					Name:   "v2",
					Labels: map[string]string{"version": "v2"},
				},
			},
		}

		manifest, err := km.GenerateIstioDestinationRule("test-namespace", "test-dr", dr)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "networking.istio.io/v1beta1", manifest["apiVersion"])
		assert.Equal(t, "DestinationRule", manifest["kind"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		assert.Equal(t, "test-service", spec["host"])
		assert.NotNil(t, spec["trafficPolicy"])
		assert.NotNil(t, spec["subsets"])
	})

	t.Run("IstioServiceEntry", func(t *testing.T) {
		se := &IstioServiceEntry{
			Name:  "external-service",
			Hosts: []string{"external.example.com"},
			Ports: []ServicePort{
				{
					Name:     "https",
					Port:     443,
					Protocol: "HTTPS",
				},
			},
			Location:   "MESH_EXTERNAL",
			Resolution: "DNS",
		}

		manifest, err := km.GenerateIstioServiceEntry("test-namespace", "external-service", se)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "networking.istio.io/v1beta1", manifest["apiVersion"])
		assert.Equal(t, "ServiceEntry", manifest["kind"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		assert.Equal(t, []string{"external.example.com"}, spec["hosts"])
		assert.Equal(t, "MESH_EXTERNAL", spec["location"])
		assert.Equal(t, "DNS", spec["resolution"])
	})

	t.Run("IstioGateway", func(t *testing.T) {
		gw := &IstioGateway{
			Name:     "test-gateway",
			Selector: map[string]string{"istio": "ingressgateway"},
			Servers: []IstioServer{
				{
					Port: &GatewayPort{
						Number:   80,
						Name:     "http",
						Protocol: "HTTP",
					},
					Hosts: []string{"*.example.com"},
				},
				{
					Port: &GatewayPort{
						Number:   443,
						Name:     "https",
						Protocol: "HTTPS",
					},
					Hosts: []string{"*.example.com"},
					TLS: &ServerTLSSettings{
						Mode: "SIMPLE",
					},
				},
			},
		}

		manifest, err := km.GenerateIstioGateway("test-namespace", "test-gateway", gw)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "networking.istio.io/v1beta1", manifest["apiVersion"])
		assert.Equal(t, "Gateway", manifest["kind"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		assert.Equal(t, map[string]string{"istio": "ingressgateway"}, spec["selector"])
		servers, ok := spec["servers"].([]map[string]interface{})
		assert.True(t, ok, "servers should be a slice of maps")
		assert.Len(t, servers, 2)
	})
}

func TestLinkerdConfiguration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "linkerd_config_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("LinkerdServiceProfile", func(t *testing.T) {
		sp := &LinkerdServiceProfile{
			Name: "test-service",
			Routes: []LinkerdRoute{
				{
					Name: "get-users",
					Condition: &LinkerdCondition{
						Method:    "GET",
						PathRegex: "/api/users/[^/]*",
					},
					Timeout:     "30s",
					IsRetryable: true,
				},
				{
					Name: "post-users",
					Condition: &LinkerdCondition{
						Method:    "POST",
						PathRegex: "/api/users",
					},
					Timeout:     "60s",
					IsRetryable: false,
				},
			},
			RetryBudget: &RetryBudgetConfig{
				RetryRatio:          0.2,
				MinRetriesPerSecond: 10,
				TTL:                 "10s",
			},
		}

		manifest, err := km.GenerateLinkerdServiceProfile("test-namespace", "test-service", sp)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "linkerd.io/v1alpha2", manifest["apiVersion"])
		assert.Equal(t, "ServiceProfile", manifest["kind"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		routes, ok := spec["routes"].([]map[string]interface{})
		assert.True(t, ok, "routes should be a slice of maps")
		assert.Len(t, routes, 2)
		assert.NotNil(t, spec["retryBudget"])
	})

	t.Run("LinkerdTrafficSplit", func(t *testing.T) {
		ts := &LinkerdTrafficSplit{
			Name:    "test-split",
			Service: "test-service",
			Backends: []LinkerdBackend{
				{
					Service: "test-service-v1",
					Weight:  80,
				},
				{
					Service: "test-service-v2",
					Weight:  20,
				},
			},
		}

		manifest, err := km.GenerateLinkerdTrafficSplit("test-namespace", "test-split", ts)
		assert.NoError(t, err)
		assert.NotNil(t, manifest)
		assert.Equal(t, "split.smi-spec.io/v1alpha1", manifest["apiVersion"])
		assert.Equal(t, "TrafficSplit", manifest["kind"])

		spec, ok := manifest["spec"].(map[string]interface{})
		assert.True(t, ok, "spec should be a map")
		assert.Equal(t, "test-service", spec["service"])
		backends, ok := spec["backends"].([]map[string]interface{})
		assert.True(t, ok, "backends should be a slice of maps")
		assert.Len(t, backends, 2)

		// Verify weights
		assert.Equal(t, int32(80), backends[0]["weight"])
		assert.Equal(t, int32(20), backends[1]["weight"])
	})
}

func TestServiceMeshProfileIntegration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mesh_profile_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("IstioProfileWithCompleteConfig", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:        "istio-complete",
			Description: "Complete Istio configuration",
			Namespace:   "istio-test",
			Policies:    make(map[string]*NetworkPolicyConfig),
			ServiceMesh: &ServiceMeshConfig{
				Type:      "istio",
				Enabled:   true,
				Namespace: "istio-test",
				TrafficPolicy: map[string]interface{}{
					"istio": &IstioConfig{
						MTLSMode:         "ISTIO_MUTUAL",
						SidecarInjection: true,
						VirtualServices: map[string]*IstioVirtualService{
							"frontend": {
								Name:  "frontend",
								Hosts: []string{"frontend"},
								HTTP: []IstioHTTPRoute{
									{
										Route: []IstioHTTPRouteDestination{
											{
												Destination: &IstioDestination{
													Host: "frontend",
												},
											},
										},
									},
								},
							},
						},
						DestinationRules: map[string]*IstioDestinationRule{
							"frontend": {
								Name: "frontend",
								Host: "frontend",
								TrafficPolicy: &IstioTrafficPolicy{
									TLS: &ClientTLSSettings{
										Mode: "ISTIO_MUTUAL",
									},
								},
							},
						},
						CircuitBreaker: &CircuitBreakerConfig{
							ConsecutiveErrors:  5,
							Interval:           "30s",
							BaseEjectionTime:   "30s",
							MaxEjectionPercent: 50,
						},
						RetryPolicy: &RetryPolicyConfig{
							Attempts:      3,
							PerTryTimeout: "30s",
							BackoffBase:   "1s",
							BackoffMax:    "10s",
							RetryOn:       []string{"5xx", "reset", "connect-failure"},
						},
					},
				},
			},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Verify profile was saved correctly
		profilePath := filepath.Join(tempDir, "kubernetes", "network_profiles", "istio-complete.yaml")
		assert.FileExists(t, profilePath)

		// Load and verify
		loadedProfile, err := km.LoadProfile("istio-complete")
		require.NoError(t, err)
		assert.NotNil(t, loadedProfile.ServiceMesh)
		assert.Equal(t, "istio", loadedProfile.ServiceMesh.Type)
		assert.True(t, loadedProfile.ServiceMesh.Enabled)

		// Verify Istio config was preserved
		istioConfig, ok := loadedProfile.ServiceMesh.TrafficPolicy["istio"]
		assert.True(t, ok)
		assert.NotNil(t, istioConfig)
	})

	t.Run("LinkerdProfileWithTrafficSplit", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:        "linkerd-canary",
			Description: "Linkerd canary deployment",
			Namespace:   "linkerd-test",
			Policies:    make(map[string]*NetworkPolicyConfig),
			ServiceMesh: &ServiceMeshConfig{
				Type:      "linkerd",
				Enabled:   true,
				Namespace: "linkerd-test",
				TrafficPolicy: map[string]interface{}{
					"linkerd": &LinkerdConfig{
						ProxyInjection: true,
						ServiceProfiles: map[string]*LinkerdServiceProfile{
							"api-service": {
								Name: "api-service",
								Routes: []LinkerdRoute{
									{
										Name: "api-route",
										Condition: &LinkerdCondition{
											Method:    "GET",
											PathRegex: "/api/.*",
										},
										Timeout:     "30s",
										IsRetryable: true,
									},
								},
								RetryBudget: &RetryBudgetConfig{
									RetryRatio:          0.2,
									MinRetriesPerSecond: 10,
									TTL:                 "10s",
								},
							},
						},
						TrafficSplits: map[string]*LinkerdTrafficSplit{
							"api-canary": {
								Name:    "api-canary",
								Service: "api-service",
								Backends: []LinkerdBackend{
									{
										Service: "api-service-stable",
										Weight:  90,
									},
									{
										Service: "api-service-canary",
										Weight:  10,
									},
								},
							},
						},
						ProxyResources: &ProxyResourceConfig{
							CPURequest:    "100m",
							CPULimit:      "250m",
							MemoryRequest: "64Mi",
							MemoryLimit:   "256Mi",
						},
					},
				},
			},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Load and verify
		loadedProfile, err := km.LoadProfile("linkerd-canary")
		require.NoError(t, err)
		assert.NotNil(t, loadedProfile.ServiceMesh)
		assert.Equal(t, "linkerd", loadedProfile.ServiceMesh.Type)
		assert.True(t, loadedProfile.ServiceMesh.Enabled)
	})
}

func TestServiceMeshValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mesh_validation_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("ValidateIstioMTLSModes", func(t *testing.T) {
		validModes := []string{"DISABLE", "SIMPLE", "MUTUAL", "ISTIO_MUTUAL"}
		for _, mode := range validModes {
			config := map[string]interface{}{
				"istio": &IstioConfig{
					MTLSMode: mode,
				},
			}
			err := km.validateIstioConfig(config["istio"])
			assert.NoError(t, err, "Mode %s should be valid", mode)
		}

		// Test invalid mode
		invalidConfig := map[string]interface{}{
			"istio": &IstioConfig{
				MTLSMode: "INVALID_MODE",
			},
		}
		err := km.validateIstioConfig(invalidConfig["istio"])
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid mTLS mode")
	})

	t.Run("ValidateLinkerdTrafficSplitWeights", func(t *testing.T) {
		// Valid weights (sum to 100)
		validConfig := &LinkerdConfig{
			TrafficSplits: map[string]*LinkerdTrafficSplit{
				"valid-split": {
					Name:    "valid-split",
					Service: "test-service",
					Backends: []LinkerdBackend{
						{Service: "v1", Weight: 70},
						{Service: "v2", Weight: 30},
					},
				},
			},
		}
		err := km.validateLinkerdConfig(validConfig)
		assert.NoError(t, err)

		// Invalid weights (don't sum to 100)
		invalidConfig := &LinkerdConfig{
			TrafficSplits: map[string]*LinkerdTrafficSplit{
				"invalid-split": {
					Name:    "invalid-split",
					Service: "test-service",
					Backends: []LinkerdBackend{
						{Service: "v1", Weight: 60},
						{Service: "v2", Weight: 30},
					},
				},
			},
		}
		err = km.validateLinkerdConfig(invalidConfig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "backend weights must sum to 100")
	})
}

func TestServiceMeshStatus(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mesh_status_test")
	require.NoError(t, err)

	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("GetServiceMeshStatus", func(t *testing.T) {
		// This test would normally interact with a real cluster
		// For unit testing, we just verify the function structure
		status, err := km.GetServiceMeshStatus("default")
		assert.NoError(t, err)
		assert.NotNil(t, status)

		// Verify expected fields
		assert.Contains(t, status, "mesh_type")
		assert.Contains(t, status, "namespace")
		assert.Equal(t, "default", status["namespace"])
	})
}
