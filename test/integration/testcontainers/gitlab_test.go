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

// GitLabContainer wraps the GitLab testcontainer for integration tests
type GitLabContainer struct {
	Container testcontainers.Container
	BaseURL   string
}

// SetupGitLabContainer creates and starts a GitLab container for testing
func SetupGitLabContainer(ctx context.Context, t *testing.T) *GitLabContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "gitlab/gitlab-ce:16.11.0-ce.0",
		ExposedPorts: []string{"80/tcp", "22/tcp"},
		Env: map[string]string{
			"GITLAB_OMNIBUS_CONFIG": `
				external_url 'http://localhost'
				gitlab_rails['initial_root_password'] = 'testpassword123'
				gitlab_rails['gitlab_shell_ssh_port'] = 22
				puma['worker_processes'] = 0
				sidekiq['max_concurrency'] = 10
				prometheus_monitoring['enable'] = false
				alertmanager['enable'] = false
				grafana['enable'] = false
				gitlab_exporter['enable'] = false
				node_exporter['enable'] = false
				redis_exporter['enable'] = false
				postgres_exporter['enable'] = false
			`,
		},
		WaitingFor: wait.ForAll(
			wait.ForLog("Gitlab Reconfigure Complete").WithStartupTimeout(10*time.Minute),
			wait.ForHTTP("/").WithPort("80/tcp").WithStartupTimeout(10*time.Minute),
		),
		Networks: []string{"bridge"},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "80")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	baseURL := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	return &GitLabContainer{
		Container: container,
		BaseURL:   baseURL,
	}
}

// Cleanup terminates the GitLab container
func (g *GitLabContainer) Cleanup(ctx context.Context) error {
	return g.Container.Terminate(ctx)
}

// WaitForReady waits for GitLab to be fully ready for API calls
func (g *GitLabContainer) WaitForReady(ctx context.Context) error {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for GitLab to be ready")
		case <-ticker.C:
			resp, err := http.Get(g.BaseURL + "/api/v4/version")
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

func TestGitLabContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping GitLab container integration test in short mode")
	}

	ctx := context.Background()

	gitlab := SetupGitLabContainer(ctx, t)
	defer func() {
		err := gitlab.Cleanup(ctx)
		assert.NoError(t, err)
	}()

	// Wait for GitLab to be ready
	err := gitlab.WaitForReady(ctx)
	require.NoError(t, err)

	// Test basic GitLab connectivity
	resp, err := http.Get(gitlab.BaseURL + "/api/v4/version")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Logf("GitLab container is ready at %s", gitlab.BaseURL)
}
