package devenv

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAzureSubscriptionManager(t *testing.T) {
	// Create temporary directory for testing
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	ctx := context.Background()
	manager, err := NewAzureSubscriptionManager(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, tmpDir+"/.azure", manager.configPath)
	assert.NotNil(t, manager.subscriptions)
	assert.NotNil(t, manager.tenants)
}

func TestAzureSubscription_JSONSerialization(t *testing.T) {
	now := time.Now()
	subscription := &AzureSubscription{
		ID:                "12345678-1234-1234-1234-123456789012",
		DisplayName:       "Test Subscription",
		Name:              "Test Subscription",
		State:             "Enabled",
		TenantID:          "87654321-4321-4321-4321-210987654321",
		TenantDisplayName: "Test Tenant",
		User:              "test@example.com",
		IsDefault:         true,
		IsActive:          true,
		LastUsed:          &now,
		Tags: map[string]string{
			"environment": "test",
			"team":        "dev",
		},
		ResourceGroups: []string{
			"rg-test-1",
			"rg-test-2",
		},
		Regions: []string{
			"eastus",
			"westus2",
			"westeurope",
		},
		EnvironmentName:  "AzureCloud",
		HomeTenantID:     "87654321-4321-4321-4321-210987654321",
		ManagedByTenants: []string{"11111111-1111-1111-1111-111111111111"},
	}

	// Test JSON marshaling
	data, err := json.Marshal(subscription)
	require.NoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled AzureSubscription
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, subscription.ID, unmarshaled.ID)
	assert.Equal(t, subscription.DisplayName, unmarshaled.DisplayName)
	assert.Equal(t, subscription.Name, unmarshaled.Name)
	assert.Equal(t, subscription.State, unmarshaled.State)
	assert.Equal(t, subscription.TenantID, unmarshaled.TenantID)
	assert.Equal(t, subscription.TenantDisplayName, unmarshaled.TenantDisplayName)
	assert.Equal(t, subscription.User, unmarshaled.User)
	assert.Equal(t, subscription.IsDefault, unmarshaled.IsDefault)
	assert.Equal(t, subscription.IsActive, unmarshaled.IsActive)
	assert.True(t, subscription.LastUsed.Equal(*unmarshaled.LastUsed))
	assert.Equal(t, subscription.Tags, unmarshaled.Tags)
	assert.Equal(t, subscription.ResourceGroups, unmarshaled.ResourceGroups)
	assert.Equal(t, subscription.Regions, unmarshaled.Regions)
	assert.Equal(t, subscription.EnvironmentName, unmarshaled.EnvironmentName)
	assert.Equal(t, subscription.HomeTenantID, unmarshaled.HomeTenantID)
	assert.Equal(t, subscription.ManagedByTenants, unmarshaled.ManagedByTenants)
}

func TestAzureSubscriptionManager_EmptyState(t *testing.T) {
	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	manager := &AzureSubscriptionManager{
		configPath:    tmpDir + "/.azure",
		subscriptions: make(map[string]*AzureSubscription),
		tenants:       make(map[string]string),
		ctx:           context.Background(),
	}

	// Test getCurrentSubscription with no active subscription
	currentSub := manager.getCurrentSubscription()
	assert.Empty(t, currentSub)

	// Test listSubscriptions with no subscriptions
	err := manager.listSubscriptions("table", "")
	assert.NoError(t, err) // Should handle empty state gracefully

	err = manager.listSubscriptions("json", "")
	assert.NoError(t, err) // Should handle empty state gracefully
}

func TestAzureSubscriptionManager_SubscriptionFiltering(t *testing.T) {
	manager := &AzureSubscriptionManager{
		subscriptions: map[string]*AzureSubscription{
			"sub1": {
				ID:                "sub1",
				DisplayName:       "Production Subscription",
				TenantID:          "tenant1",
				TenantDisplayName: "Production Tenant",
				State:             "Enabled",
				IsActive:          true,
			},
			"sub2": {
				ID:                "sub2",
				DisplayName:       "Development Subscription",
				TenantID:          "tenant2",
				TenantDisplayName: "Development Tenant",
				State:             "Enabled",
				IsActive:          false,
			},
			"sub3": {
				ID:                "sub3",
				DisplayName:       "Staging Subscription",
				TenantID:          "tenant1",
				TenantDisplayName: "Production Tenant",
				State:             "Disabled",
				IsActive:          false,
			},
		},
		tenants: map[string]string{
			"tenant1": "Production Tenant",
			"tenant2": "Development Tenant",
		},
		ctx: context.Background(),
	}

	t.Run("ListAllSubscriptions", func(t *testing.T) {
		err := manager.listSubscriptions("json", "")
		assert.NoError(t, err)
	})

	t.Run("FilterByTenant", func(t *testing.T) {
		err := manager.listSubscriptions("json", "tenant1")
		assert.NoError(t, err)
	})

	t.Run("FilterByNonExistentTenant", func(t *testing.T) {
		err := manager.listSubscriptions("json", "nonexistent")
		assert.NoError(t, err) // Should handle gracefully
	})
}

func TestAzureSubscriptionManager_EnrichSubscriptionDetails(t *testing.T) {
	manager := &AzureSubscriptionManager{
		ctx: context.Background(),
	}

	subscription := &AzureSubscription{
		ID:          "test-sub-id",
		DisplayName: "Test Subscription",
		TenantID:    "test-tenant-id",
	}

	// This will fail in test environment since Azure CLI is not available,
	// but we test that it doesn't panic
	assert.NotPanics(t, func() {
		manager.enrichSubscriptionDetails(subscription)
	})

	// Verify subscription structure is intact
	assert.Equal(t, "test-sub-id", subscription.ID)
	assert.Equal(t, "Test Subscription", subscription.DisplayName)
	assert.Equal(t, "test-tenant-id", subscription.TenantID)
}

func TestAzureSubscriptionManager_SelectSubscriptionInteractively(t *testing.T) {
	manager := &AzureSubscriptionManager{
		subscriptions: map[string]*AzureSubscription{
			"sub1": {
				ID:          "sub1",
				DisplayName: "Production Subscription",
				TenantID:    "tenant1",
				IsActive:    true,
			},
			"sub2": {
				ID:          "sub2",
				DisplayName: "Development Subscription",
				TenantID:    "tenant2",
				IsActive:    false,
			},
		},
		ctx: context.Background(),
	}

	t.Run("NoSubscriptionsAvailable", func(t *testing.T) {
		emptyManager := &AzureSubscriptionManager{
			subscriptions: make(map[string]*AzureSubscription),
			ctx:           context.Background(),
		}

		_, err := emptyManager.selectSubscriptionInteractively("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subscriptions available")
	})

	t.Run("NoSubscriptionsForTenant", func(t *testing.T) {
		_, err := manager.selectSubscriptionInteractively("nonexistent-tenant")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no subscriptions available for tenant")
	})

	// Note: Testing actual interactive selection would require mocking promptui,
	// which is complex in this context. The logic is tested indirectly through
	// the filter and subscription availability checks.
}

func TestAzureSubscriptionManager_TenantManagement(t *testing.T) {
	manager := &AzureSubscriptionManager{
		tenants: map[string]string{
			"tenant1": "Production Tenant",
			"tenant2": "Development Tenant",
			"tenant3": "Testing Tenant",
		},
		subscriptions: map[string]*AzureSubscription{
			"sub1": {TenantID: "tenant1"},
			"sub2": {TenantID: "tenant1"},
			"sub3": {TenantID: "tenant2"},
		},
		ctx: context.Background(),
	}

	t.Run("ListTenants", func(t *testing.T) {
		err := manager.listTenants("table")
		assert.NoError(t, err)

		err = manager.listTenants("json")
		assert.NoError(t, err)
	})

	t.Run("EmptyTenants", func(t *testing.T) {
		emptyManager := &AzureSubscriptionManager{
			tenants: make(map[string]string),
			ctx:     context.Background(),
		}

		err := emptyManager.listTenants("table")
		assert.NoError(t, err) // Should handle empty state gracefully
	})
}

func TestAzureSubscription_DefaultValues(t *testing.T) {
	subscription := &AzureSubscription{
		ID:          "test-subscription",
		DisplayName: "Test Subscription",
	}

	// Verify default values for optional fields
	assert.Empty(t, subscription.State)
	assert.Empty(t, subscription.TenantID)
	assert.Empty(t, subscription.User)
	assert.False(t, subscription.IsDefault)
	assert.False(t, subscription.IsActive)
	assert.Nil(t, subscription.LastUsed)
	assert.Empty(t, subscription.Tags)
	assert.Empty(t, subscription.ResourceGroups)
	assert.Empty(t, subscription.Regions)
	assert.Empty(t, subscription.ManagedByTenants)
}

func TestAzureSubscriptionManager_ShowSubscription(t *testing.T) {
	now := time.Now()
	manager := &AzureSubscriptionManager{
		subscriptions: map[string]*AzureSubscription{
			"test-sub": {
				ID:                "test-sub",
				DisplayName:       "Test Subscription",
				State:             "Enabled",
				TenantID:          "test-tenant",
				TenantDisplayName: "Test Tenant",
				User:              "test@example.com",
				IsActive:          true,
				LastUsed:          &now,
				ResourceGroups:    []string{"rg1", "rg2"},
				Regions:           []string{"eastus", "westus2"},
			},
		},
		ctx: context.Background(),
	}

	// Note: In a real test, we would mock getCurrentSubscription method
	// For this test, we'll test the actual functionality

	t.Run("ShowExistingSubscription", func(t *testing.T) {
		err := manager.showSubscription("test-sub", "table")
		assert.NoError(t, err)

		err = manager.showSubscription("test-sub", "json")
		assert.NoError(t, err)
	})

	t.Run("ShowNonExistentSubscription", func(t *testing.T) {
		err := manager.showSubscription("nonexistent", "table")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ShowNoActiveSubscription", func(t *testing.T) {
		err := manager.showSubscription("", "table")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no active subscription found")
	})
}

func TestAzureSubscriptionManager_Integration(t *testing.T) {
	// This test can be run with real Azure CLI for integration testing
	// Skip if Azure CLI is not available or user is not authenticated
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	ctx := context.Background()
	manager, err := NewAzureSubscriptionManager(ctx)

	// Should not fail even without real Azure CLI
	assert.NoError(t, err)
	assert.NotNil(t, manager)

	// Test that empty manager can handle basic operations
	err = manager.listSubscriptions("table", "")
	assert.NoError(t, err) // Should handle empty subscription list gracefully

	err = manager.listTenants("table")
	assert.NoError(t, err) // Should handle empty tenant list gracefully
}

func TestAzureSubscriptionManager_InvalidFormat(t *testing.T) {
	manager := &AzureSubscriptionManager{
		subscriptions: map[string]*AzureSubscription{
			"sub1": {
				ID:          "sub1",
				DisplayName: "Test Subscription",
			},
		},
		ctx: context.Background(),
	}

	t.Run("InvalidListFormat", func(t *testing.T) {
		err := manager.listSubscriptions("xml", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported output format")
	})

	t.Run("InvalidShowFormat", func(t *testing.T) {
		err := manager.showSubscription("sub1", "yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported output format")
	})

	t.Run("InvalidTenantFormat", func(t *testing.T) {
		err := manager.listTenants("csv")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported output format")
	})
}

// Benchmark tests for performance
func BenchmarkAzureSubscriptionManager_LoadSubscriptions(b *testing.B) {
	tmpDir := b.TempDir()
	originalHome := os.Getenv("HOME")
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tmpDir)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager := &AzureSubscriptionManager{
			configPath:    tmpDir + "/.azure",
			subscriptions: make(map[string]*AzureSubscription),
			tenants:       make(map[string]string),
			ctx:           context.Background(),
		}
		_ = manager.loadSubscriptions() // Will fail but tests the path
	}
}

func BenchmarkAzureSubscription_JSONSerialization(b *testing.B) {
	now := time.Now()
	subscription := &AzureSubscription{
		ID:                "benchmark-subscription",
		DisplayName:       "Benchmark Subscription",
		State:             "Enabled",
		TenantID:          "benchmark-tenant",
		TenantDisplayName: "Benchmark Tenant",
		User:              "bench@example.com",
		IsActive:          true,
		LastUsed:          &now,
		Tags: map[string]string{
			"environment": "benchmark",
			"team":        "performance",
		},
		ResourceGroups: []string{
			"rg-benchmark-1",
			"rg-benchmark-2",
			"rg-benchmark-3",
		},
		Regions: []string{
			"eastus",
			"westus2",
			"westeurope",
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(subscription)
		var unmarshaled AzureSubscription
		_ = json.Unmarshal(data, &unmarshaled)
	}
}
