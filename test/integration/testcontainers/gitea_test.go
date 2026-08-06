//nolint:testpackage // White-box testing needed for internal function access
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

// GiteaTestContainer wraps the Gitea testcontainer for integration tests.
type GiteaTestContainer struct {
	Container testcontainers.Container
	BaseURL   string
	AdminUser string
	AdminPass string
}

// SetupGiteaTestContainer creates and starts a Gitea container for testing.
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

// Cleanup terminates the Gitea container.
func (g *GiteaTestContainer) Cleanup(ctx context.Context) error {
	return g.Container.Terminate(ctx)
}

// WaitForReady waits for Gitea to be fully ready for API calls.
func (g *GiteaTestContainer) WaitForReady(ctx context.Context) error {
	timeout := time.After(2 * time.Minute)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Gitea to be ready")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", g.BaseURL+"/api/v1/version", nil)
			if err != nil {
				continue
			}
			resp, err := http.DefaultClient.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				if err := resp.Body.Close(); err != nil {
					// Log but don't fail, this is a readiness check
				}
				return nil
			}

			if resp != nil {
				if err := resp.Body.Close(); err != nil {
					// Log but don't fail, this is a readiness check
				}
			}
		}
	}
}

// requireDocker는 도커가 없으면 시험을 건너뛴다.
//
// 없는 것과 고장난 것은 다르다. 도커가 아예 없는 기계에서 이 시험이
// 빨갛게 되면 진짜 고장과 구별할 수 없다. 건너뛴 것은 SKIP으로 남아
// 눈에 보이므로, 통과했다고 거짓말하는 것과도 다르다.
func requireDocker(t *testing.T) {
	t.Helper()

	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		t.Skipf("Docker를 쓸 수 없어 건너뛴다: %v", err)
	}

	defer func() { _ = provider.Close() }()

	if err := provider.Health(context.Background()); err != nil {
		t.Skipf("Docker 데몬에 닿지 않아 건너뛴다: %v", err)
	}
}

func TestGiteaContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gitea container integration test in short mode")
	}

	requireDocker(t)

	ctx := context.Background()

	// 이 파일 위쪽의 진짜 구현을 쓴다. 예전에는 gitea.go의
	// SetupGiteaContainer를 불렀는데 그것은 아무것도 띄우지 않고
	// http://localhost:3000을 그대로 돌려주는 껍데기다. WaitForReady도
	// 바로 nil을 돌려주니 "준비됐다"는 말만 하고 아무것도 기다리지 않았다.
	// 그래서 이 시험은 늘 그 주소로 요청을 보내고 connection refused로
	// 죽었다 -- 컨테이너가 뜨는지 보자는 시험이 컨테이너를 띄운 적이 없다.
	// 이름이 SetupGiteaContainer와 SetupGiteaTestContainer로 한 글자
	// 차이라 눈에 잘 안 띈다. 껍데기 쪽은 test/integration/docker가
	// 아직 쓰고 있어서 남겨 둔다.
	gitea := SetupGiteaTestContainer(ctx, t)

	defer func() {
		err := gitea.Cleanup(ctx)
		assert.NoError(t, err)
	}()

	// Wait for Gitea to be ready
	err := gitea.WaitForReady(ctx)
	require.NoError(t, err)

	// Test basic Gitea connectivity
	req, err := http.NewRequestWithContext(ctx, "GET", gitea.BaseURL+"/api/v1/version", nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Warning: failed to close response body: %v", err)
		}
	}()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	t.Logf("Gitea container is ready at %s", gitea.BaseURL)
}
