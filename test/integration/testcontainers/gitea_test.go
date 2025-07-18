package testcontainers

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// GiteaTestContainer wraps the Gitea testcontainer for integration tests
type GiteaTestContainer struct {
	Container testcontainers.Container
	BaseURL   string
	AdminUser string
	AdminPass string
}

// SetupGiteaTestContainer creates and starts a Gitea container for testing
func SetupGiteaTestContainer(ctx context.Context, t *testing.T) *GiteaTestContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "gitea/gitea:1.21.10",
		ExposedPorts: []string{"3000/tcp", "22/tcp"},
		Env: map[string]string{
			"USER_UID":                             "1000",
			"USER_GID":                             "1000",
			"GITEA__database__DB_TYPE":             "sqlite3",
			"GITEA__database__PATH":                "/data/gitea/gitea.db",
			"GITEA__security__INSTALL_LOCK":        "true",
			"GITEA__security__SECRET_KEY":          "test-secret-key-for-integration-tests-only",
			"GITEA__security__INTERNAL_TOKEN":      "test-internal-token-for-integration-tests",
			"GITEA__service__DISABLE_REGISTRATION": "false",
			"GITEA__service__REQUIRE_SIGNIN_VIEW":  "false",
			"GITEA__server__ROOT_URL":              "http://localhost:3000/",
			"GITEA__server__SSH_DOMAIN":            "localhost",
			"GITEA__server__SSH_PORT":              "22",
			"GITEA__repository__DEFAULT_BRANCH":    "main",
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("Starting new Web server: tcp:0.0.0.0:3000").WithStartupTimeout(2*time.Minute),
			wait.ForHTTP("/").WithPort("3000/tcp").WithStartupTimeout(2*time.Minute),
		),
		Networks: []string{"bridge"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "3000")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	return &GiteaTestContainer{
		Container: container,
		BaseURL:   baseURL,
		AdminUser: "gitea_admin",
		AdminPass: "admin123",
	}
}

// Cleanup terminates the Gitea container
func (g *GiteaTestContainer) Cleanup(ctx context.Context) error {
	return g.Container.Terminate(ctx)
}

// WaitForReady waits for Gitea to be fully ready for API calls
func (g *GiteaTestContainer) WaitForReady(ctx context.Context) error {
	timeout := time.After(2 * time.Minute)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Gitea to be ready")
		case <-ticker.C:
			resp, err := http.Get(g.BaseURL + "/api/v1/version")
			if err == nil && resp.StatusCode == http.StatusOK {
				resp.Body.Close()
				return nil
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
	}
}

func TestGiteaContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gitea container integration test in short mode")
	}

	ctx := context.Background()

	gitea := SetupGiteaContainer(ctx, t)
	defer func() {
		err := gitea.Cleanup(ctx)
		assert.NoError(t, err)
	}()

	// Wait for Gitea to be ready
	err := gitea.WaitForReady(ctx)
	require.NoError(t, err)

	// Test basic Gitea connectivity
	resp, err := http.Get(gitea.BaseURL + "/api/v1/version")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Logf("Gitea container is ready at %s", gitea.BaseURL)
}
