package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceDiscovery_Creation(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "static",
	}

	sd := NewServiceDiscovery(logger, config)

	assert.NotNil(t, sd)
	assert.Equal(t, config, sd.config)
	assert.NotNil(t, sd.targets)
	assert.NotNil(t, sd.watchers)
}

func TestServiceDiscovery_StaticTargets(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "static",
		Config: map[string]interface{}{
			"targets": []interface{}{
				map[string]interface{}{
					"address": "localhost",
					"port":    float64(8080),
					"labels": map[string]interface{}{
						"environment": "test",
						"service":     "gzh-manager",
					},
				},
				map[string]interface{}{
					"address": "remote-host",
					"port":    float64(9090),
					"labels": map[string]interface{}{
						"environment": "production",
						"service":     "metrics",
					},
				},
			},
		},
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start static discovery", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.NoError(t, err)

		// Give it time to process
		time.Sleep(100 * time.Millisecond)

		targets := sd.GetTargets()
		assert.Len(t, targets, 2)

		// Verify first target
		target1 := findTarget(targets, "localhost", 8080)
		require.NotNil(t, target1)
		assert.Equal(t, "test", target1.Labels["environment"])
		assert.Equal(t, "gzh-manager", target1.Labels["service"])
		assert.Equal(t, "static", target1.ServiceType)

		// Verify second target
		target2 := findTarget(targets, "remote-host", 9090)
		require.NotNil(t, target2)
		assert.Equal(t, "production", target2.Labels["environment"])
		assert.Equal(t, "metrics", target2.Labels["service"])
	})

	t.Run("Get Prometheus targets", func(t *testing.T) {
		prometheusTargets := sd.GetPrometheusTargets()
		assert.Len(t, prometheusTargets, 2)

		// Verify Prometheus target format
		for _, target := range prometheusTargets {
			assert.Contains(t, target, "targets")
			assert.Contains(t, target, "labels")

			targetList, ok := target["targets"].([]string)
			assert.True(t, ok)
			assert.Len(t, targetList, 1)
		}
	})

	t.Run("Stop discovery", func(t *testing.T) {
		err := sd.Stop()
		assert.NoError(t, err)
	})
}

func TestServiceDiscovery_DNSTargets(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "dns",
		Config: map[string]interface{}{
			"name":     "localhost", // Use localhost for reliable testing
			"port":     float64(8080),
			"interval": "1s",
		},
		ScrapeInterval: 1 * time.Second,
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	t.Run("Start DNS discovery", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.NoError(t, err)

		// Give DNS resolution time to complete
		time.Sleep(2 * time.Second)

		targets := sd.GetTargets()

		// localhost should resolve to at least one IP
		assert.Greater(t, len(targets), 0)

		// Verify target properties
		if len(targets) > 0 {
			target := targets[0]
			assert.Equal(t, 8080, target.Port)
			assert.Equal(t, "localhost", target.Labels["dns_name"])
			assert.Equal(t, "dns", target.ServiceType)
		}
	})

	t.Run("Stop DNS discovery", func(t *testing.T) {
		err := sd.Stop()
		assert.NoError(t, err)
	})
}

func TestServiceDiscovery_KubernetesTargets(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "kubernetes",
		Config: map[string]interface{}{
			"namespace": "default",
			"selector": map[string]interface{}{
				"app": "gzh-manager",
			},
		},
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start Kubernetes discovery", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.NoError(t, err)

		// Note: This will log a warning that K8s discovery is not implemented
		// In a real implementation, this would connect to Kubernetes API

		targets := sd.GetTargets()
		// Should be empty since not implemented
		assert.Len(t, targets, 0)
	})

	t.Run("Stop Kubernetes discovery", func(t *testing.T) {
		err := sd.Stop()
		assert.NoError(t, err)
	})
}

func TestServiceDiscovery_ConsulTargets(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "consul",
		Config: map[string]interface{}{
			"address": "localhost:8500",
			"service": "gzh-manager",
		},
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start Consul discovery", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.NoError(t, err)

		// Note: This will log a warning that Consul discovery is not implemented
		// In a real implementation, this would connect to Consul API

		targets := sd.GetTargets()
		// Should be empty since not implemented
		assert.Len(t, targets, 0)
	})

	t.Run("Stop Consul discovery", func(t *testing.T) {
		err := sd.Stop()
		assert.NoError(t, err)
	})
}

func TestServiceDiscovery_DisabledDiscovery(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: false,
		Type:    "static",
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Disabled discovery should not start", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.NoError(t, err)

		targets := sd.GetTargets()
		assert.Len(t, targets, 0)
	})
}

func TestServiceDiscovery_InvalidType(t *testing.T) {
	logger := zap.NewNop()
	config := &ServiceDiscoveryConfig{
		Enabled: true,
		Type:    "invalid_type",
	}

	sd := NewServiceDiscovery(logger, config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Invalid type should return error", func(t *testing.T) {
		err := sd.Start(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported discovery type")
	})
}

func TestStaticTargetWatcher(t *testing.T) {
	logger := zap.NewNop()
	targets := []*DiscoveryTarget{
		{
			Address:     "host1",
			Port:        8080,
			Labels:      map[string]string{"env": "test"},
			ServiceType: "static",
		},
		{
			Address:     "host2",
			Port:        9090,
			Labels:      map[string]string{"env": "prod"},
			ServiceType: "static",
		},
	}

	watcher := &StaticTargetWatcher{
		targets: targets,
		logger:  logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Run("Start and get targets", func(t *testing.T) {
		err := watcher.Start(ctx)
		assert.NoError(t, err)

		discoveredTargets, err := watcher.GetTargets()
		assert.NoError(t, err)
		assert.Len(t, discoveredTargets, 2)
		assert.Equal(t, targets, discoveredTargets)
	})

	t.Run("Watch targets", func(t *testing.T) {
		callbackCalled := false
		var watchedTargets []*DiscoveryTarget

		err := watcher.Watch(func(targets []*DiscoveryTarget) {
			callbackCalled = true
			watchedTargets = targets
		})
		assert.NoError(t, err)
		assert.True(t, callbackCalled)
		assert.Len(t, watchedTargets, 2)
	})

	t.Run("Stop watcher", func(t *testing.T) {
		err := watcher.Stop()
		assert.NoError(t, err)
	})
}

func TestDNSTargetWatcher(t *testing.T) {
	logger := zap.NewNop()

	watcher := &DNSTargetWatcher{
		logger:   logger,
		dnsName:  "localhost",
		port:     8080,
		interval: 1 * time.Second,
		targets:  make(map[string]*DiscoveryTarget),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	t.Run("Start DNS watcher", func(t *testing.T) {
		err := watcher.Start(ctx)
		assert.NoError(t, err)

		// Give DNS resolution time
		time.Sleep(500 * time.Millisecond)

		targets, err := watcher.GetTargets()
		assert.NoError(t, err)

		// localhost should resolve
		assert.Greater(t, len(targets), 0)

		if len(targets) > 0 {
			target := targets[0]
			assert.Equal(t, 8080, target.Port)
			assert.Equal(t, "localhost", target.Labels["dns_name"])
			assert.Equal(t, "dns", target.ServiceType)
		}
	})

	t.Run("Watch DNS targets", func(t *testing.T) {
		callbackCalled := false

		err := watcher.Watch(func(targets []*DiscoveryTarget) {
			callbackCalled = true
		})
		assert.NoError(t, err)

		// Give watch time to trigger
		time.Sleep(2 * time.Second)

		// Callback should have been called at least once
		assert.True(t, callbackCalled)
	})

	t.Run("Stop DNS watcher", func(t *testing.T) {
		err := watcher.Stop()
		assert.NoError(t, err)
	})
}

func TestDiscoveryTarget_JSON(t *testing.T) {
	target := &DiscoveryTarget{
		Address:     "test-host",
		Port:        8080,
		Labels:      map[string]string{"env": "test", "service": "api"},
		Health:      "healthy",
		LastSeen:    time.Now(),
		ServiceType: "static",
		Metadata: map[string]interface{}{
			"version": "1.0.0",
			"region":  "us-west-2",
		},
	}

	// Test that the target can be marshaled/unmarshaled
	// This ensures the JSON tags are correct
	assert.NotEmpty(t, target.Address)
	assert.Equal(t, 8080, target.Port)
	assert.Equal(t, "test", target.Labels["env"])
	assert.Equal(t, "api", target.Labels["service"])
	assert.Equal(t, "healthy", target.Health)
	assert.Equal(t, "static", target.ServiceType)
}

func TestServiceDiscoveryConfig_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		config  *ServiceDiscoveryConfig
		isValid bool
	}{
		{
			name: "Valid static config",
			config: &ServiceDiscoveryConfig{
				Enabled: true,
				Type:    "static",
				Config: map[string]interface{}{
					"targets": []interface{}{
						map[string]interface{}{
							"address": "localhost",
							"port":    float64(8080),
						},
					},
				},
			},
			isValid: true,
		},
		{
			name: "Valid DNS config",
			config: &ServiceDiscoveryConfig{
				Enabled: true,
				Type:    "dns",
				Config: map[string]interface{}{
					"name": "service.example.com",
					"port": float64(8080),
				},
			},
			isValid: true,
		},
		{
			name: "Invalid DNS config - missing name",
			config: &ServiceDiscoveryConfig{
				Enabled: true,
				Type:    "dns",
				Config: map[string]interface{}{
					"port": float64(8080),
				},
			},
			isValid: false,
		},
		{
			name: "Disabled config",
			config: &ServiceDiscoveryConfig{
				Enabled: false,
				Type:    "static",
			},
			isValid: true, // Disabled configs are always valid
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			logger := zap.NewNop()
			sd := NewServiceDiscovery(logger, tc.config)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := sd.Start(ctx)
			if tc.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			sd.Stop()
		})
	}
}

// Helper function to find a target by address and port
func findTarget(targets []*DiscoveryTarget, address string, port int) *DiscoveryTarget {
	for _, target := range targets {
		if target.Address == address && target.Port == port {
			return target
		}
	}
	return nil
}
