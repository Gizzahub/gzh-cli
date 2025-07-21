package netenv

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestKubernetesNetworkManager(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "k8s_network_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("CreateProfile", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:        "test-profile",
			Description: "Test profile for unit testing",
			Namespace:   "test-namespace",
			Policies: map[string]*NetworkPolicyConfig{
				"deny-all": {
					Name:        "deny-all",
					PodSelector: map[string]string{},
					PolicyTypes: []string{"Ingress", "Egress"},
				},
			},
			Services: map[string]*ServiceConfig{},
			Ingress:  map[string]*IngressConfig{},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Verify profile file was created
		profilePath := filepath.Join(tempDir, "kubernetes", "network_profiles", "test-profile.yaml")
		assert.FileExists(t, profilePath)

		// Verify profile is in cache
		assert.Contains(t, km.cache, "test-profile")
	})

	t.Run("LoadProfile", func(t *testing.T) {
		// Create a profile first
		profile := &KubernetesNetworkProfile{
			Name:        "load-test",
			Description: "Profile for load testing",
			Namespace:   "load-namespace",
			Policies: map[string]*NetworkPolicyConfig{
				"test-policy": {
					Name:        "test-policy",
					PodSelector: map[string]string{"app": "test"},
					PolicyTypes: []string{"Ingress"},
				},
			},
		}

		err := km.CreateProfile(profile)
		require.NoError(t, err)

		// Clear cache to test loading from file
		km.cache = make(map[string]*KubernetesNetworkProfile)

		// Load profile
		loadedProfile, err := km.LoadProfile("load-test")
		assert.NoError(t, err)
		assert.NotNil(t, loadedProfile)
		assert.Equal(t, "load-test", loadedProfile.Name)
		assert.Equal(t, "Profile for load testing", loadedProfile.Description)
		assert.Equal(t, "load-namespace", loadedProfile.Namespace)
		assert.Contains(t, loadedProfile.Policies, "test-policy")
	})

	t.Run("ListProfiles", func(t *testing.T) {
		// Create multiple profiles
		profiles := []*KubernetesNetworkProfile{
			{
				Name:        "profile1",
				Description: "First profile",
				Namespace:   "ns1",
				Policies:    map[string]*NetworkPolicyConfig{},
			},
			{
				Name:        "profile2",
				Description: "Second profile",
				Namespace:   "ns2",
				Policies:    map[string]*NetworkPolicyConfig{},
			},
		}

		for _, profile := range profiles {
			err := km.CreateProfile(profile)
			require.NoError(t, err)
		}

		// List profiles
		listedProfiles, err := km.ListProfiles()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(listedProfiles), 2)

		// Check that our profiles are in the list
		profileNames := make([]string, len(listedProfiles))
		for i, p := range listedProfiles {
			profileNames[i] = p.Name
		}
		assert.Contains(t, profileNames, "profile1")
		assert.Contains(t, profileNames, "profile2")
	})

	t.Run("DeleteProfile", func(t *testing.T) {
		// Create a profile to delete
		profile := &KubernetesNetworkProfile{
			Name:        "delete-test",
			Description: "Profile for deletion testing",
			Namespace:   "delete-ns",
			Policies:    map[string]*NetworkPolicyConfig{},
		}

		err := km.CreateProfile(profile)
		require.NoError(t, err)

		// Verify profile exists
		profilePath := filepath.Join(tempDir, "kubernetes", "network_profiles", "delete-test.yaml")
		assert.FileExists(t, profilePath)

		// Delete profile
		err = km.DeleteProfile("delete-test")
		assert.NoError(t, err)

		// Verify profile file was deleted
		assert.NoFileExists(t, profilePath)

		// Verify profile is removed from cache
		assert.NotContains(t, km.cache, "delete-test")
	})

	t.Run("ValidateNetworkPolicies", func(t *testing.T) {
		// Test valid policies
		validPolicies := map[string]*NetworkPolicyConfig{
			"valid-ingress": {
				Name:        "valid-ingress",
				PodSelector: map[string]string{"app": "web"},
				PolicyTypes: []string{"Ingress"},
			},
			"valid-egress": {
				Name:        "valid-egress",
				PodSelector: map[string]string{"app": "db"},
				PolicyTypes: []string{"Egress"},
			},
			"valid-both": {
				Name:        "valid-both",
				PodSelector: map[string]string{"app": "api"},
				PolicyTypes: []string{"Ingress", "Egress"},
			},
		}

		err := km.validateNetworkPolicies(validPolicies)
		assert.NoError(t, err)

		// Test invalid policy type
		invalidPolicies := map[string]*NetworkPolicyConfig{
			"invalid-policy": {
				Name:        "invalid-policy",
				PodSelector: map[string]string{"app": "test"},
				PolicyTypes: []string{"Invalid"},
			},
		}

		err = km.validateNetworkPolicies(invalidPolicies)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid policy type")
	})
}

func TestNetworkPolicyGeneration(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "k8s_policy_gen_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("GenerateBasicNetworkPolicy", func(t *testing.T) {
		port80 := int32(80)
		port443 := int32(443)

		config := &NetworkPolicyConfig{
			Name:        "web-policy",
			PodSelector: map[string]string{"app": "web", "tier": "frontend"},
			PolicyTypes: []string{"Ingress", "Egress"},
			Ingress: []NetworkPolicyIngressRule{
				{
					From: []NetworkPolicyPeer{
						{
							PodSelector: map[string]string{"app": "gateway"},
						},
						{
							NamespaceSelector: map[string]string{"name": "monitoring"},
						},
					},
					Ports: []NetworkPolicyPort{
						{
							Protocol: "TCP",
							Port:     &port80,
						},
						{
							Protocol: "TCP",
							Port:     &port443,
						},
					},
				},
			},
			Egress: []NetworkPolicyEgressRule{
				{
					To: []NetworkPolicyPeer{
						{
							PodSelector: map[string]string{"app": "api"},
						},
					},
					Ports: []NetworkPolicyPort{
						{
							Protocol: "TCP",
							Port:     &port80,
						},
					},
				},
			},
		}

		policy, err := km.GenerateNetworkPolicy("test-namespace", config)
		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, "web-policy", policy.Metadata.Name)
		assert.Equal(t, "test-namespace", policy.Metadata.Namespace)
		assert.Equal(t, 2, len(policy.Spec.PolicyTypes))
		assert.Equal(t, 1, len(policy.Spec.Ingress))
		assert.Equal(t, 1, len(policy.Spec.Egress))
	})

	t.Run("GenerateNetworkPolicyWithIPBlock", func(t *testing.T) {
		port22 := int32(22)

		config := &NetworkPolicyConfig{
			Name:        "ssh-policy",
			PodSelector: map[string]string{"app": "bastion"},
			PolicyTypes: []string{"Ingress"},
			Ingress: []NetworkPolicyIngressRule{
				{
					From: []NetworkPolicyPeer{
						{
							IPBlock: &IPBlock{
								CIDR:   "10.0.0.0/8",
								Except: []string{"10.1.0.0/16"},
							},
						},
					},
					Ports: []NetworkPolicyPort{
						{
							Protocol: "TCP",
							Port:     &port22,
						},
					},
				},
			},
		}

		policy, err := km.GenerateNetworkPolicy("secure-namespace", config)
		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, "ssh-policy", policy.Metadata.Name)
		assert.Equal(t, 1, len(policy.Spec.Ingress))
		assert.NotNil(t, policy.Spec.Ingress[0].From[0].IPBlock)
		assert.Equal(t, "10.0.0.0/8", policy.Spec.Ingress[0].From[0].IPBlock.CIDR)
		assert.Contains(t, policy.Spec.Ingress[0].From[0].IPBlock.Except, "10.1.0.0/16")
	})

	t.Run("GenerateNetworkPolicyWithPortRange", func(t *testing.T) {
		startPort := int32(8080)
		endPort := int32(8090)

		config := &NetworkPolicyConfig{
			Name:        "api-policy",
			PodSelector: map[string]string{"app": "api"},
			PolicyTypes: []string{"Ingress"},
			Ingress: []NetworkPolicyIngressRule{
				{
					From: []NetworkPolicyPeer{
						{
							PodSelector: map[string]string{},
						},
					},
					Ports: []NetworkPolicyPort{
						{
							Protocol: "TCP",
							Port:     &startPort,
							EndPort:  &endPort,
						},
					},
				},
			},
		}

		policy, err := km.GenerateNetworkPolicy("api-namespace", config)
		assert.NoError(t, err)
		assert.NotNil(t, policy)
		assert.Equal(t, "api-policy", policy.Metadata.Name)
		assert.Equal(t, int32(8080), policy.Spec.Ingress[0].Ports[0].Port.(int32))
		assert.Equal(t, int32(8090), *policy.Spec.Ingress[0].Ports[0].EndPort)
	})
}

func TestComplexNetworkProfile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "k8s_complex_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("MicroservicesNetworkProfile", func(t *testing.T) {
		// Create a complex microservices network profile
		port80 := int32(80)
		port443 := int32(443)
		port3306 := int32(3306)
		port6379 := int32(6379)
		port5432 := int32(5432)

		profile := &KubernetesNetworkProfile{
			Name:        "microservices",
			Description: "Microservices architecture network policies",
			Namespace:   "production",
			Policies: map[string]*NetworkPolicyConfig{
				"frontend-policy": {
					Name:        "frontend-policy",
					PodSelector: map[string]string{"tier": "frontend"},
					PolicyTypes: []string{"Ingress", "Egress"},
					Ingress: []NetworkPolicyIngressRule{
						{
							From: []NetworkPolicyPeer{
								{
									IPBlock: &IPBlock{
										CIDR: "0.0.0.0/0", // Allow from anywhere
									},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port80},
								{Protocol: "TCP", Port: &port443},
							},
						},
					},
					Egress: []NetworkPolicyEgressRule{
						{
							To: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "api"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port80},
							},
						},
						{
							To: []NetworkPolicyPeer{
								{
									// DNS resolution
									PodSelector:       map[string]string{},
									NamespaceSelector: map[string]string{"name": "kube-system"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "UDP", Port: func() *int32 { p := int32(53); return &p }()},
							},
						},
					},
				},
				"api-policy": {
					Name:        "api-policy",
					PodSelector: map[string]string{"tier": "api"},
					PolicyTypes: []string{"Ingress", "Egress"},
					Ingress: []NetworkPolicyIngressRule{
						{
							From: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "frontend"},
								},
								{
									NamespaceSelector: map[string]string{"name": "monitoring"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port80},
							},
						},
					},
					Egress: []NetworkPolicyEgressRule{
						{
							To: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "database"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port3306}, // MySQL
								{Protocol: "TCP", Port: &port5432}, // PostgreSQL
							},
						},
						{
							To: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "cache"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port6379}, // Redis
							},
						},
					},
				},
				"database-policy": {
					Name:        "database-policy",
					PodSelector: map[string]string{"tier": "database"},
					PolicyTypes: []string{"Ingress"},
					Ingress: []NetworkPolicyIngressRule{
						{
							From: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "api"},
								},
								{
									PodSelector: map[string]string{"app": "backup"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port3306},
								{Protocol: "TCP", Port: &port5432},
							},
						},
					},
				},
				"cache-policy": {
					Name:        "cache-policy",
					PodSelector: map[string]string{"tier": "cache"},
					PolicyTypes: []string{"Ingress"},
					Ingress: []NetworkPolicyIngressRule{
						{
							From: []NetworkPolicyPeer{
								{
									PodSelector: map[string]string{"tier": "api"},
								},
							},
							Ports: []NetworkPolicyPort{
								{Protocol: "TCP", Port: &port6379},
							},
						},
					},
				},
			},
			Services: map[string]*ServiceConfig{
				"frontend-service": {
					Name:     "frontend-service",
					Type:     "LoadBalancer",
					Selector: map[string]string{"tier": "frontend"},
					Ports: []ServicePort{
						{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
						{Name: "https", Protocol: "TCP", Port: 443, TargetPort: 8443},
					},
				},
				"api-service": {
					Name:     "api-service",
					Type:     "ClusterIP",
					Selector: map[string]string{"tier": "api"},
					Ports: []ServicePort{
						{Name: "http", Protocol: "TCP", Port: 80, TargetPort: 8080},
					},
				},
			},
			Metadata: map[string]string{
				"environment": "production",
				"team":        "platform",
				"version":     "1.0.0",
			},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Load and verify the profile
		loadedProfile, err := km.LoadProfile("microservices")
		require.NoError(t, err)

		assert.Equal(t, "microservices", loadedProfile.Name)
		assert.Equal(t, "production", loadedProfile.Namespace)
		assert.Len(t, loadedProfile.Policies, 4)
		assert.Contains(t, loadedProfile.Policies, "frontend-policy")
		assert.Contains(t, loadedProfile.Policies, "api-policy")
		assert.Contains(t, loadedProfile.Policies, "database-policy")
		assert.Contains(t, loadedProfile.Policies, "cache-policy")
		assert.Len(t, loadedProfile.Services, 2)
		assert.Equal(t, "production", loadedProfile.Metadata["environment"])
	})

	t.Run("ZeroTrustNetworkProfile", func(t *testing.T) {
		// Create a zero-trust network profile (deny all by default)
		profile := &KubernetesNetworkProfile{
			Name:        "zero-trust",
			Description: "Zero trust network security profile",
			Namespace:   "secure-namespace",
			Policies: map[string]*NetworkPolicyConfig{
				"default-deny-all": {
					Name:        "default-deny-all",
					PodSelector: map[string]string{}, // Empty selector = all pods
					PolicyTypes: []string{"Ingress", "Egress"},
					// No ingress or egress rules = deny all
				},
				"allow-dns": {
					Name:        "allow-dns",
					PodSelector: map[string]string{}, // All pods need DNS
					PolicyTypes: []string{"Egress"},
					Egress: []NetworkPolicyEgressRule{
						{
							To: []NetworkPolicyPeer{
								{
									NamespaceSelector: map[string]string{"name": "kube-system"},
									PodSelector:       map[string]string{"k8s-app": "kube-dns"},
								},
							},
							Ports: []NetworkPolicyPort{
								{
									Protocol: "UDP",
									Port:     func() *int32 { p := int32(53); return &p }(),
								},
							},
						},
					},
				},
			},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Verify the zero-trust profile
		loadedProfile, err := km.LoadProfile("zero-trust")
		require.NoError(t, err)

		assert.Equal(t, "zero-trust", loadedProfile.Name)
		assert.Contains(t, loadedProfile.Policies, "default-deny-all")
		assert.Contains(t, loadedProfile.Policies, "allow-dns")

		// Verify default-deny-all has no rules
		denyAll := loadedProfile.Policies["default-deny-all"]
		assert.Empty(t, denyAll.Ingress)
		assert.Empty(t, denyAll.Egress)
	})
}

func TestNetworkPolicyValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "k8s_validation_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logger, _ := zap.NewDevelopment()
	km := NewKubernetesNetworkManager(logger, tempDir)

	t.Run("EmptyProfileName", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:        "", // Empty name should cause error
			Description: "Test profile",
			Namespace:   "test",
			Policies:    map[string]*NetworkPolicyConfig{},
		}

		err := km.CreateProfile(profile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile name cannot be empty")
	})

	t.Run("DefaultNamespace", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:        "no-namespace-test",
			Description: "Test profile without namespace",
			// Namespace not specified
			Policies: map[string]*NetworkPolicyConfig{},
		}

		err := km.CreateProfile(profile)
		assert.NoError(t, err)

		// Load profile and check defaults
		loadedProfile, err := km.LoadProfile("no-namespace-test")
		require.NoError(t, err)

		assert.Equal(t, "default", loadedProfile.Namespace) // Should default to "default"
	})

	t.Run("InvalidPolicyType", func(t *testing.T) {
		profile := &KubernetesNetworkProfile{
			Name:      "invalid-policy-test",
			Namespace: "test",
			Policies: map[string]*NetworkPolicyConfig{
				"bad-policy": {
					Name:        "bad-policy",
					PodSelector: map[string]string{"app": "test"},
					PolicyTypes: []string{"InvalidType"}, // Invalid policy type
				},
			},
		}

		err := km.CreateProfile(profile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid policy type")
	})
}

func TestServiceMeshIntegration(t *testing.T) {
	t.Run("IstioServiceMeshConfig", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "k8s_istio_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		logger, _ := zap.NewDevelopment()
		km := NewKubernetesNetworkManager(logger, tempDir)

		// This test demonstrates how service mesh configs could be integrated
		profile := &KubernetesNetworkProfile{
			Name:        "istio-enabled",
			Description: "Profile with Istio service mesh integration",
			Namespace:   "istio-demo",
			Policies:    map[string]*NetworkPolicyConfig{},
			Metadata: map[string]string{
				"service-mesh":    "istio",
				"mtls-mode":       "STRICT",
				"traffic-policy":  "round-robin",
				"circuit-breaker": "enabled",
				"retry-attempts":  "3",
				"timeout":         "30s",
			},
		}

		err = km.CreateProfile(profile)
		assert.NoError(t, err)

		loadedProfile, err := km.LoadProfile("istio-enabled")
		require.NoError(t, err)

		assert.Equal(t, "istio", loadedProfile.Metadata["service-mesh"])
		assert.Equal(t, "STRICT", loadedProfile.Metadata["mtls-mode"])
	})

	t.Run("LinkerdServiceMeshConfig", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "k8s_linkerd_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		logger, _ := zap.NewDevelopment()
		km := NewKubernetesNetworkManager(logger, tempDir)

		profile := &KubernetesNetworkProfile{
			Name:        "linkerd-enabled",
			Description: "Profile with Linkerd service mesh integration",
			Namespace:   "linkerd-demo",
			Policies:    map[string]*NetworkPolicyConfig{},
			Metadata: map[string]string{
				"service-mesh":      "linkerd",
				"proxy-cpu-request": "100m",
				"proxy-cpu-limit":   "250m",
				"proxy-memory":      "64Mi",
				"tap-enabled":       "true",
			},
		}

		err = km.CreateProfile(profile)
		assert.NoError(t, err)

		loadedProfile, err := km.LoadProfile("linkerd-enabled")
		require.NoError(t, err)

		assert.Equal(t, "linkerd", loadedProfile.Metadata["service-mesh"])
		assert.Equal(t, "true", loadedProfile.Metadata["tap-enabled"])
	})
}

// TestKubernetesCommandExecutor tests the command executor.
func TestKubernetesCommandExecutor(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	executor := NewKubernetesCommandExecutor(logger)

	t.Run("CommandCaching", func(t *testing.T) {
		// Test that get commands are cached
		cmd := "kubectl get pods -n default"

		// First execution
		result1, err := executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, result1)

		// Check cache
		cached := executor.getCachedResult(cmd)
		assert.NotNil(t, cached)

		// Second execution should return cached result
		result2, err := executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)
		assert.NoError(t, err)
		assert.Equal(t, result1.Output, result2.Output)
	})

	t.Run("NonCacheableCommands", func(t *testing.T) {
		// Apply commands should not be cached
		cmd := "kubectl apply -f test.yaml"

		executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)

		// Should not be cached
		cached := executor.getCachedResult(cmd)
		assert.Nil(t, cached)
	})
}
