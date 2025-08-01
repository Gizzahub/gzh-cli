// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package container

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock dependencies for testing.
type MockService struct {
	Name string
}

type MockDependentService struct {
	Service *MockService
}

func TestContainer_BasicOperations(t *testing.T) {
	container := NewContainer(nil)

	// Test registration and retrieval
	container.Register("mock", func(c *Container) (interface{}, error) {
		return &MockService{Name: "test"}, nil
	})

	// Test Get
	instance, err := container.Get("mock")
	require.NoError(t, err)

	mockService, ok := instance.(*MockService)
	require.True(t, ok)
	assert.Equal(t, "test", mockService.Name)

	// Test singleton behavior
	instance2, err := container.Get("mock")
	require.NoError(t, err)
	assert.Same(t, instance, instance2)

	// Test Has
	assert.True(t, container.Has("mock"))
	assert.False(t, container.Has("nonexistent"))

	// Test ListRegistered
	registered := container.ListRegistered()
	assert.Contains(t, registered, "mock")
}

func TestContainer_TypedGet(t *testing.T) {
	container := NewContainer(nil)

	container.Register("mock", func(c *Container) (interface{}, error) {
		return &MockService{Name: "typed"}, nil
	})

	// Test successful typed get
	service, err := GetTyped[*MockService](container, "mock")
	require.NoError(t, err)
	assert.Equal(t, "typed", service.Name)

	// Test failed type assertion
	_, err = GetTyped[string](container, "mock")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not of expected type")
}

func TestContainer_Dependencies(t *testing.T) {
	container := NewContainer(nil)

	// Register base service
	container.Register("base", func(c *Container) (interface{}, error) {
		return &MockService{Name: "base"}, nil
	})

	// Register dependent service
	container.Register("dependent", func(c *Container) (interface{}, error) {
		base, err := GetTyped[*MockService](c, "base")
		if err != nil {
			return nil, err
		}
		return &MockDependentService{Service: base}, nil
	})

	// Test dependency resolution
	dependent, err := GetTyped[*MockDependentService](container, "dependent")
	require.NoError(t, err)
	assert.NotNil(t, dependent.Service)
	assert.Equal(t, "base", dependent.Service.Name)
}

func TestContainer_ErrorHandling(t *testing.T) {
	container := NewContainer(nil)

	// Register factory that returns an error
	container.Register("error", func(c *Container) (interface{}, error) {
		return nil, errors.New("creation failed")
	})

	// Test error propagation
	_, err := container.Get("error")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "creation failed")

	// Test nonexistent dependency
	_, err = container.Get("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not registered")
}

func TestContainer_RegisterInstance(t *testing.T) {
	container := NewContainer(nil)

	service := &MockService{Name: "instance"}
	container.RegisterInstance("instance", service)

	retrieved, err := container.Get("instance")
	require.NoError(t, err)
	assert.Same(t, service, retrieved)
}

func TestContainer_Clear(t *testing.T) {
	container := NewContainer(nil)

	container.Register("test", func(c *Container) (interface{}, error) {
		return &MockService{Name: "test"}, nil
	})

	// Ensure dependency exists
	_, err := container.Get("test")
	require.NoError(t, err)

	// Clear container
	container.Clear()

	// Verify it's cleared
	assert.False(t, container.Has("test"))
	registered := container.ListRegistered()
	assert.Empty(t, registered)
}

func TestContainerBuilder(t *testing.T) {
	builder := NewContainerBuilder().
		WithHTTPTimeout(10*time.Second).
		WithMetrics(false).
		WithHealthChecks(true).
		Register("custom", func(c *Container) (interface{}, error) {
			return "custom value", nil
		}).
		RegisterInstance("instance", "instance value")

	container := builder.Build()

	// Test custom registration
	custom, err := container.Get("custom")
	require.NoError(t, err)
	assert.Equal(t, "custom value", custom)

	// Test custom instance
	instance, err := container.Get("instance")
	require.NoError(t, err)
	assert.Equal(t, "instance value", instance)

	// Test configuration was applied
	assert.Equal(t, 10*time.Second, container.configuration.HTTPTimeout)
	assert.False(t, container.configuration.EnableMetrics)
	assert.True(t, container.configuration.EnableHealthChecks)
}

func TestContextualContainer(t *testing.T) {
	container := NewContainer(nil)
	ctx := context.WithValue(context.Background(), "test", "value")

	contextualContainer := NewContextualContainer(ctx, container)

	// Test context retrieval
	retrievedCtx := contextualContainer.GetContext()
	assert.Equal(t, "value", retrievedCtx.Value("test"))
}

func TestDefaultContainer(t *testing.T) {
	// Test that default container has core dependencies
	dependencies := DefaultContainer.ListRegistered()

	expectedDeps := []string{"env", "logger", "httpClient", "configService", "providerFactory"}
	for _, dep := range expectedDeps {
		assert.Contains(t, dependencies, dep, "Default container should have %s dependency", dep)
	}

	// Test global convenience functions
	RegisterDefault("test", func(c *Container) (interface{}, error) {
		return "test value", nil
	})

	value, err := GetDefault("test")
	require.NoError(t, err)
	assert.Equal(t, "test value", value)

	typed, err := GetTypedDefault[string]("test")
	require.NoError(t, err)
	assert.Equal(t, "test value", typed)
}

func TestContainerModule(t *testing.T) {
	// Test GitHubModule
	githubModule := &GitHubModule{Token: "test-token"}

	builder := NewModuleBuilder()
	container, err := builder.AddModule(githubModule).Build()
	require.NoError(t, err)

	token, err := container.Get("githubToken")
	require.NoError(t, err)
	assert.Equal(t, "test-token", token)

	// Test GitLabModule
	gitlabModule := &GitLabModule{
		Token:   "gitlab-token",
		BaseURL: "https://gitlab.example.com",
	}

	container2, err := NewModuleBuilder().
		AddModule(gitlabModule).
		Build()
	require.NoError(t, err)

	token2, err := container2.Get("gitlabToken")
	require.NoError(t, err)
	assert.Equal(t, "gitlab-token", token2)

	baseURL, err := container2.Get("gitlabBaseURL")
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com", baseURL)
}

func TestContainer_ConcurrentAccess(t *testing.T) {
	container := NewContainer(nil)

	// Register a slow factory
	container.Register("slow", func(c *Container) (interface{}, error) {
		time.Sleep(10 * time.Millisecond)
		return "slow value", nil
	})

	// Test concurrent access
	const numGoroutines = 10
	results := make(chan interface{}, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			instance, err := container.Get("slow")
			if err != nil {
				errors <- err
			} else {
				results <- instance
			}
		}()
	}

	// Collect results
	var instances []interface{}
	for i := 0; i < numGoroutines; i++ {
		select {
		case instance := <-results:
			instances = append(instances, instance)
		case err := <-errors:
			t.Errorf("Unexpected error: %v", err)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for results")
		}
	}

	// All instances should be the same (singleton)
	assert.Len(t, instances, numGoroutines)
	for i := 1; i < len(instances); i++ {
		assert.Same(t, instances[0], instances[i])
	}
}
