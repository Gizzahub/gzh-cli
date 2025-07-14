package testcontainers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// RedisContainer wraps the Redis testcontainer for integration tests
type RedisContainer struct {
	Container testcontainers.Container
	Host      string
	Port      string
	Address   string
}

// SetupRedisContainer creates and starts a Redis container for testing
func SetupRedisContainer(ctx context.Context, t *testing.T) *RedisContainer {
	t.Helper()

	req := testcontainers.ContainerRequest{
		Image:        "redis:7.2-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForAll(
			wait.ForLog("Ready to accept connections").WithStartupTimeout(30*time.Second),
			wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second),
		),
		Networks: []string{"bridge"},
		Cmd: []string{
			"redis-server",
			"--appendonly", "yes",
			"--maxmemory", "100mb",
			"--maxmemory-policy", "allkeys-lru",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	address := fmt.Sprintf("%s:%s", host, mappedPort.Port())

	return &RedisContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
		Address:   address,
	}
}

// Cleanup terminates the Redis container
func (r *RedisContainer) Cleanup(ctx context.Context) error {
	return r.Container.Terminate(ctx)
}

func TestRedisContainer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis container integration test in short mode")
	}

	ctx := context.Background()

	redis := SetupRedisContainer(ctx, t)
	defer func() {
		err := redis.Cleanup(ctx)
		assert.NoError(t, err)
	}()

	t.Logf("Redis container is ready at %s", redis.Address)

	// Test Redis connectivity by checking if container is running
	state, err := redis.Container.State(ctx)
	require.NoError(t, err)
	assert.True(t, state.Running)
}
