//nolint:testpackage // White-box testing needed for internal function access
package netenv

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestContainerNetworkManagement(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "docker_container_network_test")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	logger, _ := zap.NewDevelopment()
	dm := NewDockerNetworkManager(logger, tempDir)

	// Create a test profile
	profile := &DockerNetworkProfile{
		Name:        "test-container-profile",
		Description: "Test profile for container network management",
		Networks: map[string]*DockerNetwork{
			"frontend": {
				Name:    "frontend",
				Driver:  "bridge",
				Subnet:  "172.20.0.0/16",
				Gateway: "172.20.0.1",
			},
			"backend": {
				Name:   "backend",
				Driver: "bridge",
			},
		},
		Containers: map[string]*ContainerNetwork{},
	}

	err = dm.CreateProfile(profile)
	require.NoError(t, err)

	t.Run("UpdateContainerNetwork", func(t *testing.T) {
		containerConfig := &ContainerNetwork{
			Image:    "nginx:alpine",
			Networks: []string{"frontend"},
			Ports:    []string{"80:80", "443:443"},
			Environment: map[string]string{
				"NGINX_HOST": "example.com",
				"NGINX_PORT": "80",
			},
			DNSServers:   []string{"8.8.8.8", "8.8.4.4"},
			NetworkAlias: []string{"web", "nginx"},
			Hostname:     "web-server",
		}

		err := dm.UpdateContainerNetwork("test-container-profile", "web", containerConfig)
		assert.NoError(t, err)

		// Verify the container was added
		updatedProfile, err := dm.LoadProfile("test-container-profile")
		require.NoError(t, err)
		assert.Contains(t, updatedProfile.Containers, "web")
		assert.Equal(t, "nginx:alpine", updatedProfile.Containers["web"].Image)
		assert.Contains(t, updatedProfile.Containers["web"].Networks, "frontend")
		assert.Contains(t, updatedProfile.Containers["web"].Ports, "80:80")
		assert.Equal(t, "example.com", updatedProfile.Containers["web"].Environment["NGINX_HOST"])
	})

	t.Run("ValidateContainerNetwork", func(t *testing.T) {
		// Test valid configuration
		validConfig := &ContainerNetwork{
			Image:    "postgres:13",
			Networks: []string{"backend"},
			Ports:    []string{"5432:5432"},
			Environment: map[string]string{
				"POSTGRES_DB":       "testdb",
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
			},
			DNSServers: []string{"8.8.8.8"},
			ExtraHosts: []string{"api.local:172.20.0.10"},
		}

		err := dm.ValidateContainerNetwork(validConfig)
		assert.NoError(t, err)

		// Test invalid configurations
		invalidConfigs := []struct {
			name   string
			config *ContainerNetwork
			errMsg string
		}{
			{
				name: "empty image",
				config: &ContainerNetwork{
					Networks: []string{"backend"},
				},
				errMsg: "container image cannot be empty",
			},
			{
				name: "invalid network mode",
				config: &ContainerNetwork{
					Image:       "nginx:latest",
					NetworkMode: "invalid-mode",
				},
				errMsg: "invalid network mode",
			},
			{
				name: "invalid port mapping",
				config: &ContainerNetwork{
					Image: "nginx:latest",
					Ports: []string{"80:80:80:80"}, // Too many parts
				},
				errMsg: "invalid port mapping",
			},
			{
				name: "invalid DNS server",
				config: &ContainerNetwork{
					Image:      "nginx:latest",
					DNSServers: []string{"invalid-ip"},
				},
				errMsg: "invalid DNS server IP",
			},
			{
				name: "invalid extra host",
				config: &ContainerNetwork{
					Image:      "nginx:latest",
					ExtraHosts: []string{"hostname-without-ip"},
				},
				errMsg: "invalid extra host entry",
			},
		}

		for _, tc := range invalidConfigs {
			t.Run(tc.name, func(t *testing.T) {
				err := dm.ValidateContainerNetwork(tc.config)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			})
		}
	})

	t.Run("RemoveContainerFromProfile", func(t *testing.T) {
		// Add a container first
		containerConfig := &ContainerNetwork{
			Image:    "redis:6",
			Networks: []string{"backend"},
			Ports:    []string{"6379:6379"},
		}

		err := dm.UpdateContainerNetwork("test-container-profile", "cache", containerConfig)
		require.NoError(t, err)

		// Verify it was added
		profile, err := dm.LoadProfile("test-container-profile")
		require.NoError(t, err)
		assert.Contains(t, profile.Containers, "cache")

		// Remove the container
		err = dm.RemoveContainerFromProfile("test-container-profile", "cache")
		assert.NoError(t, err)

		// Verify it was removed
		profile, err = dm.LoadProfile("test-container-profile")
		require.NoError(t, err)
		assert.NotContains(t, profile.Containers, "cache")
	})

	t.Run("RemoveNonExistentContainer", func(t *testing.T) {
		err := dm.RemoveContainerFromProfile("test-container-profile", "non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container non-existent not found")
	})

	t.Run("CloneProfile", func(t *testing.T) {
		// Create a profile with containers
		sourceProfile := &DockerNetworkProfile{
			Name:        "source-profile",
			Description: "Source profile for cloning",
			Networks: map[string]*DockerNetwork{
				"app-network": {
					Name:   "app-network",
					Driver: "overlay",
					Subnet: "10.0.0.0/16",
					Options: map[string]string{
						"encrypted": "true",
					},
					Labels: map[string]string{
						"app": "myapp",
					},
				},
			},
			Containers: map[string]*ContainerNetwork{
				"app": {
					Image:    "myapp:latest",
					Networks: []string{"app-network"},
					Ports:    []string{"8080:8080"},
					Environment: map[string]string{
						"APP_ENV":  "production",
						"APP_PORT": "8080",
					},
					NetworkAlias: []string{"myapp", "app-server"},
				},
			},
			Metadata: map[string]string{
				"version": "1.0",
				"team":    "devops",
			},
		}

		err := dm.CreateProfile(sourceProfile)
		require.NoError(t, err)

		// Clone the profile
		err = dm.CloneProfile("source-profile", "cloned-profile")
		assert.NoError(t, err)

		// Verify the cloned profile
		clonedProfile, err := dm.LoadProfile("cloned-profile")
		require.NoError(t, err)

		assert.Equal(t, "cloned-profile", clonedProfile.Name)
		assert.Contains(t, clonedProfile.Description, "Cloned from source-profile")

		// Verify networks were cloned
		assert.Contains(t, clonedProfile.Networks, "app-network")
		assert.Equal(t, "overlay", clonedProfile.Networks["app-network"].Driver)
		assert.Equal(t, "10.0.0.0/16", clonedProfile.Networks["app-network"].Subnet)
		assert.Equal(t, "true", clonedProfile.Networks["app-network"].Options["encrypted"])
		assert.Equal(t, "myapp", clonedProfile.Networks["app-network"].Labels["app"])

		// Verify containers were cloned
		assert.Contains(t, clonedProfile.Containers, "app")
		assert.Equal(t, "myapp:latest", clonedProfile.Containers["app"].Image)
		assert.Contains(t, clonedProfile.Containers["app"].Networks, "app-network")
		assert.Contains(t, clonedProfile.Containers["app"].Ports, "8080:8080")
		assert.Equal(t, "production", clonedProfile.Containers["app"].Environment["APP_ENV"])
		assert.Contains(t, clonedProfile.Containers["app"].NetworkAlias, "myapp")

		// Verify metadata was cloned
		assert.Equal(t, "1.0", clonedProfile.Metadata["version"])
		assert.Equal(t, "devops", clonedProfile.Metadata["team"])
		assert.Equal(t, "source-profile", clonedProfile.Metadata["cloned_from"])
		assert.Contains(t, clonedProfile.Metadata, "cloned_at")
	})

	t.Run("ComplexContainerNetworkScenario", func(t *testing.T) {
		// Create a complex profile with multiple containers
		complexProfile := &DockerNetworkProfile{
			Name:        "microservices",
			Description: "Microservices architecture profile",
			Networks: map[string]*DockerNetwork{
				"public": {
					Name:   "public",
					Driver: "bridge",
					Subnet: "172.30.0.0/16",
				},
				"internal": {
					Name:   "internal",
					Driver: "bridge",
					Subnet: "172.31.0.0/16",
				},
				"database": {
					Name:   "database",
					Driver: "bridge",
					Subnet: "172.32.0.0/16",
				},
			},
			Containers: map[string]*ContainerNetwork{},
		}

		err := dm.CreateProfile(complexProfile)
		require.NoError(t, err)

		// Add API Gateway
		apiGatewayConfig := &ContainerNetwork{
			Image:    "nginx:alpine",
			Networks: []string{"public", "internal"},
			Ports:    []string{"80:80", "443:443"},
			Environment: map[string]string{
				"UPSTREAM_SERVERS": "api1,api2,api3",
			},
			NetworkAlias: []string{"gateway", "api-gateway"},
			Hostname:     "gateway",
		}
		err = dm.UpdateContainerNetwork("microservices", "api-gateway", apiGatewayConfig)
		assert.NoError(t, err)

		// Add API Services
		for i := 1; i <= 3; i++ {
			apiConfig := &ContainerNetwork{
				Image:    "myapp/api:latest",
				Networks: []string{"internal", "database"},
				Ports:    []string{},
				Environment: map[string]string{
					"API_ID":       string(rune('0' + i)),
					"DATABASE_URL": "postgres://db:5432/myapp",
					"REDIS_URL":    "redis://cache:6379",
				},
				NetworkAlias: []string{
					"api" + string(rune('0'+i)),
					"api-service-" + string(rune('0'+i)),
				},
			}
			err = dm.UpdateContainerNetwork("microservices", "api"+string(rune('0'+i)), apiConfig)
			assert.NoError(t, err)
		}

		// Add Database
		dbConfig := &ContainerNetwork{
			Image:    "postgres:13",
			Networks: []string{"database"},
			Ports:    []string{"5432:5432"},
			Environment: map[string]string{
				"POSTGRES_DB":       "myapp",
				"POSTGRES_USER":     "appuser",
				"POSTGRES_PASSWORD": "secret",
			},
			NetworkAlias: []string{"db", "postgres"},
			Hostname:     "database",
		}
		err = dm.UpdateContainerNetwork("microservices", "database", dbConfig)
		assert.NoError(t, err)

		// Add Cache
		cacheConfig := &ContainerNetwork{
			Image:        "redis:6-alpine",
			Networks:     []string{"database"},
			Ports:        []string{"6379:6379"},
			NetworkAlias: []string{"cache", "redis"},
		}
		err = dm.UpdateContainerNetwork("microservices", "cache", cacheConfig)
		assert.NoError(t, err)

		// Verify the complete profile
		finalProfile, err := dm.LoadProfile("microservices")
		require.NoError(t, err)

		assert.Len(t, finalProfile.Containers, 6) // gateway + 3 APIs + db + cache
		assert.Contains(t, finalProfile.Containers, "api-gateway")
		assert.Contains(t, finalProfile.Containers, "api1")
		assert.Contains(t, finalProfile.Containers, "api2")
		assert.Contains(t, finalProfile.Containers, "api3")
		assert.Contains(t, finalProfile.Containers, "database")
		assert.Contains(t, finalProfile.Containers, "cache")

		// Verify network connections
		assert.Contains(t, finalProfile.Containers["api-gateway"].Networks, "public")
		assert.Contains(t, finalProfile.Containers["api-gateway"].Networks, "internal")
		assert.Contains(t, finalProfile.Containers["api1"].Networks, "internal")
		assert.Contains(t, finalProfile.Containers["api1"].Networks, "database")
		assert.Contains(t, finalProfile.Containers["database"].Networks, "database")
	})

	t.Run("UpdateExistingContainer", func(t *testing.T) {
		// Create profile with initial container
		profile := &DockerNetworkProfile{
			Name: "update-test",
			Networks: map[string]*DockerNetwork{
				"net1": {Name: "net1", Driver: "bridge"},
				"net2": {Name: "net2", Driver: "bridge"},
			},
			Containers: map[string]*ContainerNetwork{
				"test-container": {
					Image:    "nginx:1.19",
					Networks: []string{"net1"},
					Ports:    []string{"80:80"},
					Environment: map[string]string{
						"VERSION": "1.19",
					},
				},
			},
		}

		err := dm.CreateProfile(profile)
		require.NoError(t, err)

		// Update the container configuration
		updatedConfig := &ContainerNetwork{
			Image:    "nginx:1.21",                 // Updated version
			Networks: []string{"net1", "net2"},     // Added network
			Ports:    []string{"80:80", "443:443"}, // Added port
			Environment: map[string]string{
				"VERSION": "1.21",
				"TLS":     "enabled", // New env var
			},
			DNSServers:   []string{"1.1.1.1"}, // Added DNS
			NetworkAlias: []string{"web"},     // Added alias
		}

		err = dm.UpdateContainerNetwork("update-test", "test-container", updatedConfig)
		assert.NoError(t, err)

		// Verify updates
		updatedProfile, err := dm.LoadProfile("update-test")
		require.NoError(t, err)

		container := updatedProfile.Containers["test-container"]
		assert.Equal(t, "nginx:1.21", container.Image)
		assert.Len(t, container.Networks, 2)
		assert.Contains(t, container.Networks, "net2")
		assert.Len(t, container.Ports, 2)
		assert.Contains(t, container.Ports, "443:443")
		assert.Equal(t, "1.21", container.Environment["VERSION"])
		assert.Equal(t, "enabled", container.Environment["TLS"])
		assert.Contains(t, container.DNSServers, "1.1.1.1")
		assert.Contains(t, container.NetworkAlias, "web")
	})
}

// TestContainerNetworkCommandValidation tests command-level validations.
func TestContainerNetworkCommandValidation(t *testing.T) {
	t.Run("ValidateNetworkMode", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_network_mode_test")
		require.NoError(t, err)

		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		// Valid network modes
		validModes := []string{
			"bridge",
			"host",
			"none",
			"container:other-container",
		}

		for _, mode := range validModes {
			config := &ContainerNetwork{
				Image:       "nginx:latest",
				NetworkMode: mode,
			}
			err := dm.ValidateContainerNetwork(config)
			assert.NoError(t, err, "Network mode %s should be valid", mode)
		}
	})

	t.Run("ContainerNetworkIsolation", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_isolation_test")
		require.NoError(t, err)

		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		// Create profile with isolated networks
		profile := &DockerNetworkProfile{
			Name: "isolation-test",
			Networks: map[string]*DockerNetwork{
				"dmz": {
					Name:   "dmz",
					Driver: "bridge",
					Subnet: "10.0.1.0/24",
				},
				"secure": {
					Name:   "secure",
					Driver: "bridge",
					Subnet: "10.0.2.0/24",
					Options: map[string]string{
						"com.docker.network.bridge.enable_icc": "false",
					},
				},
			},
			Containers: map[string]*ContainerNetwork{
				"public-web": {
					Image:    "nginx:latest",
					Networks: []string{"dmz"}, // Only in DMZ
					Ports:    []string{"80:80"},
				},
				"secure-db": {
					Image:    "postgres:13",
					Networks: []string{"secure"}, // Only in secure network
					Environment: map[string]string{
						"POSTGRES_PASSWORD": "secret",
					},
				},
				"api": {
					Image:    "myapp/api:latest",
					Networks: []string{"dmz", "secure"}, // Bridge between networks
					Environment: map[string]string{
						"DB_HOST": "secure-db",
					},
				},
			},
		}

		err = dm.CreateProfile(profile)
		assert.NoError(t, err)

		// Verify network isolation
		loadedProfile, err := dm.LoadProfile("isolation-test")
		require.NoError(t, err)

		// Public web should only be in DMZ
		assert.Len(t, loadedProfile.Containers["public-web"].Networks, 1)
		assert.Contains(t, loadedProfile.Containers["public-web"].Networks, "dmz")

		// Secure DB should only be in secure network
		assert.Len(t, loadedProfile.Containers["secure-db"].Networks, 1)
		assert.Contains(t, loadedProfile.Containers["secure-db"].Networks, "secure")

		// API should be in both networks
		assert.Len(t, loadedProfile.Containers["api"].Networks, 2)
		assert.Contains(t, loadedProfile.Containers["api"].Networks, "dmz")
		assert.Contains(t, loadedProfile.Containers["api"].Networks, "secure")
	})
}

// TestDockerCommandExecutorCaching tests the caching mechanism.
func TestDockerCommandExecutorCaching(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	executor := NewDockerCommandExecutor(logger)

	// Test that read-only commands are cached
	cacheableCommands := []string{
		"docker inspect test-container",
		"docker network ls",
		"docker ps",
	}

	for _, cmd := range cacheableCommands {
		t.Run("Cache_"+cmd, func(t *testing.T) {
			// First execution - should not be cached
			result1, err := executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)
			if err != nil {
				t.Skip("Docker command failed, skipping cache test")
			}

			// Get cached result immediately
			cachedResult := executor.getCachedResult(cmd)
			assert.NotNil(t, cachedResult)
			assert.Equal(t, result1.Output, cachedResult.Output)

			// Wait for cache to be valid
			time.Sleep(100 * time.Millisecond)

			// Second execution - should return cached result
			result2, err := executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)
			if err != nil {
				t.Skip("Docker command failed, skipping cache test")
			}

			// Results should be the same (from cache)
			assert.Equal(t, result1.Output, result2.Output)
		})
	}

	// Test that write commands are not cached
	nonCacheableCommands := []string{
		"docker run -d nginx",
		"docker network create test",
		"docker stop test-container",
	}

	for _, cmd := range nonCacheableCommands {
		t.Run("NoCache_"+cmd, func(t *testing.T) {
			// Execute command
			_, _ = executor.ExecuteWithTimeout(context.Background(), cmd, 5*time.Second)

			// Should not be cached
			cachedResult := executor.getCachedResult(cmd)
			assert.Nil(t, cachedResult)
		})
	}
}
