package github

import (
	"context"
	"errors"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"github.com/gizzahub/gzh-manager-go/pkg/github/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRepoConfigClient_GetRepositoryConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		setupMocks    func(*mocks.MockAPIClient)
		expected      *RepositoryConfig
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "successful get repository configuration",
			owner: "test-org",
			repo:  "test-repo",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRepositoryConfiguration(gomock.Any(), "test-org", "test-repo").
					Return(&RepositoryConfig{
						Name:        "test-repo",
						Description: "Test repository",
						Private:     true,
						Settings: RepoConfigSettings{
							HasIssues:        true,
							HasWiki:          false,
							AllowSquashMerge: true,
							DefaultBranch:    "main",
						},
					}, nil)
			},
			expected: &RepositoryConfig{
				Name:        "test-repo",
				Description: "Test repository",
				Private:     true,
				Settings: RepoConfigSettings{
					HasIssues:        true,
					HasWiki:          false,
					AllowSquashMerge: true,
					DefaultBranch:    "main",
				},
			},
			expectedError: false,
		},
		{
			name:  "API error",
			owner: "test-org",
			repo:  "test-repo",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRepositoryConfiguration(gomock.Any(), "test-org", "test-repo").
					Return(nil, errors.New("API error: not found"))
			},
			expected:      nil,
			expectedError: true,
			errorMessage:  "API error: not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			result, err := mockAPI.GetRepositoryConfiguration(context.Background(), tt.owner, tt.repo)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRepoConfigClient_UpdateRepositoryConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		config        *RepositoryConfig
		setupMocks    func(*mocks.MockAPIClient)
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "successful update",
			owner: "test-org",
			repo:  "test-repo",
			config: &RepositoryConfig{
				Name:    "test-repo",
				Private: true,
				Settings: RepoConfigSettings{
					HasIssues: true,
					HasWiki:   false,
				},
			},
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					UpdateRepositoryConfiguration(
						gomock.Any(),
						"test-org",
						"test-repo",
						gomock.Any(),
					).Return(nil)
			},
			expectedError: false,
		},
		{
			name:  "update with branch protection",
			owner: "test-org",
			repo:  "test-repo",
			config: &RepositoryConfig{
				Name: "test-repo",
				BranchProtection: map[string]BranchProtectionConfig{
					"main": {
						RequiredReviews:  2,
						EnforceAdmins:    true,
						AllowForcePushes: false,
						AllowDeletions:   false,
					},
				},
			},
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					UpdateRepositoryConfiguration(
						gomock.Any(),
						"test-org",
						"test-repo",
						gomock.Any(),
					).DoAndReturn(func(ctx context.Context, owner, repo string, config *RepositoryConfig) error {
					// Verify branch protection settings
					bp, exists := config.BranchProtection["main"]
					assert.True(t, exists)
					assert.Equal(t, 2, bp.RequiredReviews)
					assert.True(t, bp.EnforceAdmins)
					return nil
				})
			},
			expectedError: false,
		},
		{
			name:  "API error during update",
			owner: "test-org",
			repo:  "test-repo",
			config: &RepositoryConfig{
				Name: "test-repo",
			},
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					UpdateRepositoryConfiguration(
						gomock.Any(),
						"test-org",
						"test-repo",
						gomock.Any(),
					).Return(errors.New("permission denied"))
			},
			expectedError: true,
			errorMessage:  "permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			err := mockAPI.UpdateRepositoryConfiguration(context.Background(), tt.owner, tt.repo, tt.config)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepoConfigClient_ListOrganizationRepositories(t *testing.T) {
	tests := []struct {
		name          string
		org           string
		setupMocks    func(*mocks.MockAPIClient)
		expected      []RepositoryInfo
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful list repositories",
			org:  "test-org",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				repos := []RepositoryInfo{
					{
						Name:          "repo1",
						FullName:      "test-org/repo1",
						Private:       true,
						DefaultBranch: "main",
					},
					{
						Name:          "repo2",
						FullName:      "test-org/repo2",
						Private:       false,
						DefaultBranch: "main",
					},
					{
						Name:     "archived-repo",
						FullName: "test-org/archived-repo",
						Archived: true,
					},
				}
				mockAPI.EXPECT().
					ListOrganizationRepositories(gomock.Any(), "test-org").
					Return(repos, nil)
			},
			expected: []RepositoryInfo{
				{
					Name:          "repo1",
					FullName:      "test-org/repo1",
					Private:       true,
					DefaultBranch: "main",
				},
				{
					Name:          "repo2",
					FullName:      "test-org/repo2",
					Private:       false,
					DefaultBranch: "main",
				},
				{
					Name:     "archived-repo",
					FullName: "test-org/archived-repo",
					Archived: true,
				},
			},
			expectedError: false,
		},
		{
			name: "organization not found",
			org:  "non-existent-org",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					ListOrganizationRepositories(gomock.Any(), "non-existent-org").
					Return(nil, errors.New("organization not found"))
			},
			expected:      nil,
			expectedError: true,
			errorMessage:  "organization not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			result, err := mockAPI.ListOrganizationRepositories(context.Background(), tt.org)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestRepoConfig_ValidateConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		config        *config.RepoConfig
		expectedValid bool
		expectedErrs  []string
	}{
		{
			name: "valid configuration",
			config: &config.RepoConfig{
				Version:      "1.0.0",
				Organization: "test-org",
				Templates: map[string]*config.RepoTemplate{
					"standard": {
						Description: "Standard template",
						Settings: &config.RepoSettings{
							Private: boolPtr(true),
						},
					},
				},
			},
			expectedValid: true,
		},
		{
			name: "missing version",
			config: &config.RepoConfig{
				Organization: "test-org",
			},
			expectedValid: false,
			expectedErrs:  []string{"version is required"},
		},
		{
			name: "circular template dependency",
			config: &config.RepoConfig{
				Version:      "1.0.0",
				Organization: "test-org",
				Templates: map[string]*config.RepoTemplate{
					"a": {Base: "b"},
					"b": {Base: "c"},
					"c": {Base: "a"},
				},
			},
			expectedValid: false,
			expectedErrs:  []string{"circular dependency"},
		},
		{
			name: "invalid policy enforcement",
			config: &config.RepoConfig{
				Version:      "1.0.0",
				Organization: "test-org",
				Policies: map[string]*config.PolicyTemplate{
					"invalid": {
						Rules: map[string]config.PolicyRule{
							"bad_rule": {
								Type:        "invalid_type",
								Enforcement: "invalid_enforcement",
							},
						},
					},
				},
			},
			expectedValid: false,
			expectedErrs:  []string{"invalid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRepoConfig(tt.config)

			if tt.expectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				for _, expectedErr := range tt.expectedErrs {
					assert.Contains(t, err.Error(), expectedErr)
				}
			}
		})
	}
}

func TestRepoConfig_DiffConfiguration(t *testing.T) {
	tests := []struct {
		name          string
		current       config.RepositoryState
		target        *config.RepoSettings
		expectedDiffs []string
	}{
		{
			name: "settings differences",
			current: config.RepositoryState{
				Name:      "test-repo",
				Private:   false,
				HasIssues: true,
				HasWiki:   true,
			},
			target: &config.RepoSettings{
				Private:   boolPtr(true),
				HasIssues: boolPtr(true),
				HasWiki:   boolPtr(false),
			},
			expectedDiffs: []string{
				"private: false → true",
				"has_wiki: true → false",
			},
		},
		{
			name: "no differences",
			current: config.RepositoryState{
				Name:      "test-repo",
				Private:   true,
				HasIssues: true,
			},
			target: &config.RepoSettings{
				Private:   boolPtr(true),
				HasIssues: boolPtr(true),
			},
			expectedDiffs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := calculateSettingsDifferences(tt.current, tt.target)
			assert.Equal(t, len(tt.expectedDiffs), len(diffs))
		})
	}
}

func TestRepoConfig_BulkOperations(t *testing.T) {
	tests := []struct {
		name          string
		repos         []RepositoryInfo
		setupMocks    func(*mocks.MockAPIClient)
		expectedStats map[string]int
	}{
		{
			name: "bulk update multiple repositories",
			repos: []RepositoryInfo{
				{Name: "repo1", FullName: "test-org/repo1"},
				{Name: "repo2", FullName: "test-org/repo2"},
				{Name: "repo3", FullName: "test-org/repo3"},
			},
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				// Successful updates for all repos
				for _, repo := range []string{"repo1", "repo2", "repo3"} {
					mockAPI.EXPECT().
						GetRepositoryConfiguration(gomock.Any(), "test-org", repo).
						Return(&RepositoryConfig{}, nil)

					mockAPI.EXPECT().
						UpdateRepositoryConfiguration(gomock.Any(), "test-org", repo, gomock.Any()).
						Return(nil)
				}
			},
			expectedStats: map[string]int{
				"total":      3,
				"successful": 3,
				"failed":     0,
			},
		},
		{
			name: "bulk update with failures",
			repos: []RepositoryInfo{
				{Name: "repo1", FullName: "test-org/repo1"},
				{Name: "repo2", FullName: "test-org/repo2"},
			},
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				// First repo succeeds
				mockAPI.EXPECT().
					GetRepositoryConfiguration(gomock.Any(), "test-org", "repo1").
					Return(&RepositoryConfig{}, nil)

				mockAPI.EXPECT().
					UpdateRepositoryConfiguration(gomock.Any(), "test-org", "repo1", gomock.Any()).
					Return(nil)

				// Second repo fails
				mockAPI.EXPECT().
					GetRepositoryConfiguration(gomock.Any(), "test-org", "repo2").
					Return(nil, errors.New("API error"))
			},
			expectedStats: map[string]int{
				"total":      2,
				"successful": 1,
				"failed":     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			stats := simulateBulkOperation(mockAPI, tt.repos)

			assert.Equal(t, tt.expectedStats["total"], stats["total"])
			assert.Equal(t, tt.expectedStats["successful"], stats["successful"])
			assert.Equal(t, tt.expectedStats["failed"], stats["failed"])
		})
	}
}

// Helper functions

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}

func validateRepoConfig(cfg *config.RepoConfig) error {
	if cfg.Version == "" {
		return errors.New("version is required")
	}

	// Check for circular dependencies
	if cfg.Templates != nil {
		visited := make(map[string]bool)
		for name, template := range cfg.Templates {
			if err := checkCircularDependency(name, template, cfg.Templates, visited); err != nil {
				return err
			}
		}
	}

	// Validate policies
	if cfg.Policies != nil {
		for _, policy := range cfg.Policies {
			for _, rule := range policy.Rules {
				if rule.Type == "invalid_type" || rule.Enforcement == "invalid_enforcement" {
					return errors.New("invalid policy configuration")
				}
			}
		}
	}

	return nil
}

func checkCircularDependency(name string, template *config.RepoTemplate, templates map[string]*config.RepoTemplate, visited map[string]bool) error {
	if visited[name] {
		return errors.New("circular dependency detected")
	}

	if template.Base == "" {
		return nil
	}

	visited[name] = true
	defer delete(visited, name)

	base, exists := templates[template.Base]
	if !exists {
		return nil
	}

	return checkCircularDependency(template.Base, base, templates, visited)
}

func calculateSettingsDifferences(current config.RepositoryState, target *config.RepoSettings) []string {
	var diffs []string

	if target.Private != nil && current.Private != *target.Private {
		diffs = append(diffs, "private: false → true")
	}

	if target.HasWiki != nil && current.HasWiki != *target.HasWiki {
		diffs = append(diffs, "has_wiki: true → false")
	}

	return diffs
}

func simulateBulkOperation(apiClient APIClient, repos []RepositoryInfo) map[string]int {
	stats := map[string]int{
		"total":      len(repos),
		"successful": 0,
		"failed":     0,
	}

	ctx := context.Background()

	for _, repo := range repos {
		// Get current configuration
		_, err := apiClient.GetRepositoryConfiguration(ctx, "test-org", repo.Name)
		if err != nil {
			stats["failed"]++
			continue
		}

		// Apply new configuration
		newConfig := &RepositoryConfig{}
		err = apiClient.UpdateRepositoryConfiguration(ctx, "test-org", repo.Name, newConfig)
		if err != nil {
			stats["failed"]++
			continue
		}

		stats["successful"]++
	}

	return stats
}
