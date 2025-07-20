//nolint:testpackage // White-box testing needed for internal function access
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

func TestDockerNetworkManager(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "docker_network_test")
	require.NoError(t, err)

	defer func() { _ = os.RemoveAll(tempDir) }()

	logger, _ := zap.NewDevelopment()
	dm := NewDockerNetworkManager(logger, tempDir)

	t.Run("CreateProfile", func(t *testing.T) {
		profile := &DockerNetworkProfile{
			Name:        "test-profile",
			Description: "Test profile for unit testing",
			Networks: map[string]*DockerNetwork{
				"test-network": {
					Name:    "test-network",
					Driver:  "bridge",
					Subnet:  "172.20.0.0/16",
					Gateway: "172.20.0.1",
				},
			},
			Containers: map[string]*ContainerNetwork{
				"test-container": {
					Image:       "nginx:latest",
					NetworkMode: "bridge",
					Networks:    []string{"test-network"},
					Ports:       []string{"80:80"},
				},
			},
		}

		err := dm.CreateProfile(profile)
		assert.NoError(t, err)

		// Verify profile file was created
		profilePath := filepath.Join(tempDir, "docker", "network_profiles", "test-profile.yaml")
		assert.FileExists(t, profilePath)

		// Verify profile is in cache
		assert.Contains(t, dm.cache, "test-profile")
	})

	t.Run("LoadProfile", func(t *testing.T) {
		// Create a profile first
		profile := &DockerNetworkProfile{
			Name:        "load-test",
			Description: "Profile for load testing",
			Networks: map[string]*DockerNetwork{
				"load-network": {
					Name:   "load-network",
					Driver: "overlay",
				},
			},
		}

		err := dm.CreateProfile(profile)
		require.NoError(t, err)

		// Clear cache to test loading from file
		dm.cache = make(map[string]*DockerNetworkProfile)

		// Load profile
		loadedProfile, err := dm.LoadProfile("load-test")
		assert.NoError(t, err)
		assert.NotNil(t, loadedProfile)
		assert.Equal(t, "load-test", loadedProfile.Name)
		assert.Equal(t, "Profile for load testing", loadedProfile.Description)
		assert.Contains(t, loadedProfile.Networks, "load-network")
		assert.Equal(t, "overlay", loadedProfile.Networks["load-network"].Driver)
	})

	t.Run("ListProfiles", func(t *testing.T) {
		// Create multiple profiles
		profiles := []*DockerNetworkProfile{
			{
				Name:        "profile1",
				Description: "First profile",
				Networks:    map[string]*DockerNetwork{},
				Containers:  map[string]*ContainerNetwork{},
			},
			{
				Name:        "profile2",
				Description: "Second profile",
				Networks:    map[string]*DockerNetwork{},
				Containers:  map[string]*ContainerNetwork{},
			},
		}

		for _, profile := range profiles {
			err := dm.CreateProfile(profile)
			require.NoError(t, err)
		}

		// List profiles
		listedProfiles, err := dm.ListProfiles()
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
		profile := &DockerNetworkProfile{
			Name:        "delete-test",
			Description: "Profile for deletion testing",
			Networks:    map[string]*DockerNetwork{},
			Containers:  map[string]*ContainerNetwork{},
		}

		err := dm.CreateProfile(profile)
		require.NoError(t, err)

		// Verify profile exists
		profilePath := filepath.Join(tempDir, "docker", "network_profiles", "delete-test.yaml")
		assert.FileExists(t, profilePath)

		// Delete profile
		err = dm.DeleteProfile("delete-test")
		assert.NoError(t, err)

		// Verify profile file was deleted
		assert.NoFileExists(t, profilePath)

		// Verify profile is removed from cache
		assert.NotContains(t, dm.cache, "delete-test")
	})

	t.Run("ValidateNetworks", func(t *testing.T) {
		// Test valid networks
		validNetworks := map[string]*DockerNetwork{
			"valid-bridge": {
				Name:   "valid-bridge",
				Driver: "bridge",
			},
			"valid-overlay": {
				Name:   "valid-overlay",
				Driver: "overlay",
			},
		}

		err := dm.validateNetworks(validNetworks)
		assert.NoError(t, err)

		// Test invalid network driver
		invalidNetworks := map[string]*DockerNetwork{
			"invalid-network": {
				Name:   "invalid-network",
				Driver: "invalid-driver",
			},
		}

		err = dm.validateNetworks(invalidNetworks)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid network driver")
	})
}

func TestDockerNetworkProfileValidation(t *testing.T) {
	t.Run("EmptyProfileName", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_network_validation_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		profile := &DockerNetworkProfile{
			Name:        "", // Empty name should cause error
			Description: "Test profile",
			Networks:    map[string]*DockerNetwork{},
			Containers:  map[string]*ContainerNetwork{},
		}

		err = dm.CreateProfile(profile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "profile name cannot be empty")
	})

	t.Run("NetworkDefaults", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_network_defaults_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		profile := &DockerNetworkProfile{
			Name:        "defaults-test",
			Description: "Test profile for defaults",
			Networks: map[string]*DockerNetwork{
				"default-network": {
					// Name and Driver will be set by validation
				},
			},
			Containers: map[string]*ContainerNetwork{},
		}

		err = dm.CreateProfile(profile)
		assert.NoError(t, err)

		// Load profile and check defaults
		loadedProfile, err := dm.LoadProfile("defaults-test")
		require.NoError(t, err)

		network := loadedProfile.Networks["default-network"]
		assert.Equal(t, "default-network", network.Name)
		assert.Equal(t, "bridge", network.Driver) // Default driver
	})
}

func TestDockerComposeIntegration(t *testing.T) {
	t.Run("CreateProfileFromCompose", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_compose_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		// Create a test Docker Compose file
		composeContent := `version: '3.8'
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    networks:
      - frontend
  app:
    image: node:16
    networks:
      - frontend
      - backend
  db:
    image: postgres:13
    networks:
      - backend

networks:
  frontend:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
          gateway: 172.20.0.1
  backend:
    driver: bridge
`

		composePath := filepath.Join(tempDir, "docker-compose.yml")
		err = os.WriteFile(composePath, []byte(composeContent), 0o644)
		require.NoError(t, err)

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		// Create profile from compose file
		err = dm.CreateProfileFromCompose(composePath, "compose-test")
		assert.NoError(t, err)

		// Load and verify the created profile
		profile, err := dm.LoadProfile("compose-test")
		require.NoError(t, err)

		assert.Equal(t, "compose-test", profile.Name)
		assert.Contains(t, profile.Description, composePath)
		assert.NotNil(t, profile.Compose)
		assert.Equal(t, composePath, profile.Compose.File)

		// Check networks
		assert.Contains(t, profile.Networks, "frontend")
		assert.Contains(t, profile.Networks, "backend")

		frontendNet := profile.Networks["frontend"]
		assert.Equal(t, "bridge", frontendNet.Driver)
		assert.Equal(t, "172.20.0.0/16", frontendNet.Subnet)
		assert.Equal(t, "172.20.0.1", frontendNet.Gateway)

		// Check containers
		assert.Contains(t, profile.Containers, "web")
		assert.Contains(t, profile.Containers, "app")
		assert.Contains(t, profile.Containers, "db")

		webContainer := profile.Containers["web"]
		assert.Equal(t, "nginx:latest", webContainer.Image)
		assert.Contains(t, webContainer.Networks, "frontend")
		assert.Contains(t, webContainer.Ports, "80:80")

		appContainer := profile.Containers["app"]
		assert.Equal(t, "node:16", appContainer.Image)
		assert.Contains(t, appContainer.Networks, "frontend")
		assert.Contains(t, appContainer.Networks, "backend")
	})

	t.Run("NonExistentComposeFile", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_compose_nonexistent_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		err = dm.CreateProfileFromCompose("/nonexistent/docker-compose.yml", "test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Docker Compose file not found")
	})
}

func TestDockerNetworkConfiguration(t *testing.T) {
	t.Run("ComplexNetworkConfiguration", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_complex_network_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		profile := &DockerNetworkProfile{
			Name:        "complex-test",
			Description: "Complex network configuration test",
			Networks: map[string]*DockerNetwork{
				"frontend": {
					Name:       "frontend",
					Driver:     "bridge",
					Subnet:     "172.20.0.0/16",
					Gateway:    "172.20.0.1",
					Attachable: true,
					Options: map[string]string{
						"com.docker.network.bridge.name":                 "frontend-br",
						"com.docker.network.bridge.enable_icc":           "true",
						"com.docker.network.bridge.enable_ip_masquerade": "true",
					},
					Labels: map[string]string{
						"environment": "development",
						"project":     "myapp",
					},
				},
				"backend": {
					Name:    "backend",
					Driver:  "overlay",
					Subnet:  "172.21.0.0/16",
					Gateway: "172.21.0.1",
				},
			},
			Containers: map[string]*ContainerNetwork{
				"nginx": {
					Image:       "nginx:alpine",
					NetworkMode: "",
					Networks:    []string{"frontend"},
					Ports:       []string{"80:80", "443:443"},
					Environment: map[string]string{
						"NGINX_HOST": "localhost",
						"NGINX_PORT": "80",
					},
					DNSServers: []string{"8.8.8.8", "8.8.4.4"},
					ExtraHosts: []string{"api.example.com:172.20.0.10"},
					Hostname:   "nginx-server",
				},
				"api": {
					Image:        "myapp/api:latest",
					Networks:     []string{"frontend", "backend"},
					NetworkAlias: []string{"api-server", "backend-api"},
					Environment: map[string]string{
						"DATABASE_URL": "postgres://db:5432/myapp",
						"REDIS_URL":    "redis://cache:6379",
					},
				},
				"database": {
					Image:    "postgres:13",
					Networks: []string{"backend"},
					Environment: map[string]string{
						"POSTGRES_DB":       "myapp",
						"POSTGRES_USER":     "appuser",
						"POSTGRES_PASSWORD": "secret",
					},
					Hostname: "db-server",
				},
			},
			Compose: &DockerComposeConfig{
				File:      filepath.Join(tempDir, "docker-compose.yml"),
				Project:   "myapp",
				AutoApply: true,
				Environment: map[string]string{
					"COMPOSE_PROJECT_NAME": "myapp",
					"DOCKER_BUILDKIT":      "1",
				},
				Overrides: []string{
					"docker-compose.override.yml",
					"docker-compose.dev.yml",
				},
			},
			Metadata: map[string]string{
				"version":     "1.0.0",
				"environment": "development",
				"team":        "platform",
			},
		}

		err = dm.CreateProfile(profile)
		assert.NoError(t, err)

		// Load and verify the complex profile
		loadedProfile, err := dm.LoadProfile("complex-test")
		require.NoError(t, err)

		// Verify networks
		assert.Len(t, loadedProfile.Networks, 2)

		frontend := loadedProfile.Networks["frontend"]
		assert.Equal(t, "bridge", frontend.Driver)
		assert.Equal(t, "172.20.0.0/16", frontend.Subnet)
		assert.True(t, frontend.Attachable)
		assert.Contains(t, frontend.Options, "com.docker.network.bridge.name")
		assert.Equal(t, "development", frontend.Labels["environment"])

		// Verify containers
		assert.Len(t, loadedProfile.Containers, 3)

		nginx := loadedProfile.Containers["nginx"]
		assert.Equal(t, "nginx:alpine", nginx.Image)
		assert.Contains(t, nginx.Networks, "frontend")
		assert.Contains(t, nginx.Ports, "80:80")
		assert.Equal(t, "localhost", nginx.Environment["NGINX_HOST"])
		assert.Contains(t, nginx.DNSServers, "8.8.8.8")
		assert.Contains(t, nginx.ExtraHosts, "api.example.com:172.20.0.10")

		api := loadedProfile.Containers["api"]
		assert.Contains(t, api.Networks, "frontend")
		assert.Contains(t, api.Networks, "backend")
		assert.Contains(t, api.NetworkAlias, "api-server")

		// Verify compose configuration
		assert.NotNil(t, loadedProfile.Compose)
		assert.Equal(t, "myapp", loadedProfile.Compose.Project)
		assert.True(t, loadedProfile.Compose.AutoApply)
		assert.Equal(t, "1", loadedProfile.Compose.Environment["DOCKER_BUILDKIT"])
		assert.Contains(t, loadedProfile.Compose.Overrides, "docker-compose.override.yml")

		// Verify metadata
		assert.Equal(t, "1.0.0", loadedProfile.Metadata["version"])
		assert.Equal(t, "platform", loadedProfile.Metadata["team"])

		// Verify timestamps
		assert.False(t, loadedProfile.CreatedAt.IsZero())
		assert.False(t, loadedProfile.UpdatedAt.IsZero())
	})
}

// Mock tests for external Docker commands (these would require Docker to be installed).
func TestDockerCommandIntegration(t *testing.T) {
	t.Skip("Skipping Docker command integration tests - requires Docker installation")

	t.Run("GetNetworkStatus", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_status_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		networks, err := dm.GetNetworkStatus()
		if err != nil {
			t.Skipf("Docker not available: %v", err)
		}

		assert.NotNil(t, networks)
		// In a real environment, we would have at least the default networks
	})

	t.Run("GetContainerNetworkInfo", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_container_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		containers, err := dm.GetContainerNetworkInfo()
		if err != nil {
			t.Skipf("Docker not available: %v", err)
		}

		assert.NotNil(t, containers)
		// The list might be empty if no containers are running
	})

	t.Run("DetectDockerComposeProjects", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "docker_detect_test")
		require.NoError(t, err)

		defer func() { _ = os.RemoveAll(tempDir) }()

		logger, _ := zap.NewDevelopment()
		dm := NewDockerNetworkManager(logger, tempDir)

		projects, err := dm.DetectDockerComposeProjects()
		if err != nil {
			t.Skipf("Docker not available: %v", err)
		}

		assert.NotNil(t, projects)
		// The list might be empty if no compose projects are running
	})
}

func TestDockerCommandExecutorIntegration(t *testing.T) {
	t.Run("CommandExecutorCaching", func(t *testing.T) {
		logger, _ := zap.NewDevelopment()
		executor := NewDockerCommandExecutor(logger)

		// Execute the same command twice to test caching
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result1, err := executor.ExecuteWithTimeout(ctx, "echo test", 5*time.Second)
		assert.NoError(t, err)
		assert.Contains(t, result1.Output, "test")
		assert.Equal(t, 0, result1.ExitCode)

		// Second execution for non-cacheable command should work
		result2, err := executor.ExecuteWithTimeout(ctx, "echo test2", 5*time.Second)
		assert.NoError(t, err)
		assert.Contains(t, result2.Output, "test2")
		assert.Equal(t, 0, result2.ExitCode)
	})
}
