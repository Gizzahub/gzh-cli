package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateProviderCloner(t *testing.T) {
	tests := []struct {
		name         string
		providerName string
		token        string
		expectError  bool
		expectedType string
	}{
		{
			name:         "GitHub cloner",
			providerName: ProviderGitHub,
			token:        "test-token",
			expectError:  false,
			expectedType: ProviderGitHub,
		},
		{
			name:         "GitLab cloner",
			providerName: ProviderGitLab,
			token:        "test-token",
			expectError:  false,
			expectedType: ProviderGitLab,
		},
		{
			name:         "Gitea cloner",
			providerName: ProviderGitea,
			token:        "test-token",
			expectError:  false,
			expectedType: ProviderGitea,
		},
		{
			name:         "Gogs cloner",
			providerName: ProviderGogs,
			token:        "test-token",
			expectError:  false,
			expectedType: ProviderGogs,
		},
		{
			name:         "invalid provider",
			providerName: "invalid",
			token:        "test-token",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloner, err := CreateProviderCloner(tt.providerName, tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cloner)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cloner)
				assert.Equal(t, tt.expectedType, cloner.GetName())
			}
		})
	}
}

func TestGitHubCloner(t *testing.T) {
	cloner := NewGitHubCloner("test-token")
	
	assert.Equal(t, ProviderGitHub, cloner.GetName())
	assert.Equal(t, "test-token", cloner.token)

	// Test SetToken
	cloner.SetToken("new-token")
	assert.Equal(t, "new-token", cloner.token)
}

func TestGitLabCloner(t *testing.T) {
	cloner := NewGitLabCloner("test-token")
	
	assert.Equal(t, ProviderGitLab, cloner.GetName())
	assert.Equal(t, "test-token", cloner.token)

	// Test SetToken
	cloner.SetToken("new-token")
	assert.Equal(t, "new-token", cloner.token)
}

func TestBulkCloneResult_GetSummary(t *testing.T) {
	result := &BulkCloneResult{
		TotalTargets:      10,
		SuccessfulTargets: 7,
		FailedTargets:     2,
		SkippedTargets:    1,
	}

	expected := "Total: 10, Successful: 7, Failed: 2, Skipped: 1"
	assert.Equal(t, expected, result.GetSummary())
}

func TestNewBulkCloneExecutor(t *testing.T) {
	config := &Config{
		Version: "1.0.0",
		Providers: map[string]Provider{
			ProviderGitHub: {
				Token: "test-token",
				Orgs: []GitTarget{
					{Name: "test-org"},
				},
			},
		},
	}

	executor, err := NewBulkCloneExecutor(config)
	assert.NoError(t, err)
	assert.NotNil(t, executor)
	assert.Len(t, executor.cloners, 1)
	assert.Contains(t, executor.cloners, ProviderGitHub)
}

func TestTargetResult(t *testing.T) {
	result := TargetResult{
		Provider: ProviderGitHub,
		Name:     "test-org",
		CloneDir: "/path/to/clone",
		Strategy: StrategyReset,
		Success:  true,
	}

	assert.Equal(t, ProviderGitHub, result.Provider)
	assert.Equal(t, "test-org", result.Name)
	assert.True(t, result.Success)
	assert.Empty(t, result.Error)
}

func TestProviderClonerInterface(t *testing.T) {
	// Test that all cloners implement the interface correctly
	var cloner ProviderCloner

	cloner = NewGitHubCloner("token")
	assert.NotNil(t, cloner)
	assert.Equal(t, ProviderGitHub, cloner.GetName())

	cloner = NewGitLabCloner("token")
	assert.NotNil(t, cloner)
	assert.Equal(t, ProviderGitLab, cloner.GetName())

	cloner = NewGiteaCloner("token")
	assert.NotNil(t, cloner)
	assert.Equal(t, ProviderGitea, cloner.GetName())

	cloner = NewGogsCloner("token")
	assert.NotNil(t, cloner)
	assert.Equal(t, ProviderGogs, cloner.GetName())
}