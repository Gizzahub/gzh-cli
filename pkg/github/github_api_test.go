package github_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gizzahub/gzh-manager-go/pkg/github"
	"github.com/gizzahub/gzh-manager-go/pkg/github/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAPIClient_GetRepository(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		setupMocks    func(*mocks.MockAPIClient)
		expected      *github.RepositoryInfo
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "successful get repository",
			owner: "test-org",
			repo:  "test-repo",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRepository(gomock.Any(), "test-org", "test-repo").
					Return(&github.RepositoryInfo{
						Name:          "test-repo",
						FullName:      "test-org/test-repo",
						Description:   "Test repository",
						DefaultBranch: "main",
						CloneURL:      "https://github.com/test-org/test-repo.git",
						SSHURL:        "git@github.com:test-org/test-repo.git",
						HTMLURL:       "https://github.com/test-org/test-repo",
						Private:       true,
						Archived:      false,
						Language:      "Go",
						Size:          1024,
					}, nil)
			},
			expected: &github.RepositoryInfo{
				Name:          "test-repo",
				FullName:      "test-org/test-repo",
				Description:   "Test repository",
				DefaultBranch: "main",
				CloneURL:      "https://github.com/test-org/test-repo.git",
				SSHURL:        "git@github.com:test-org/test-repo.git",
				HTMLURL:       "https://github.com/test-org/test-repo",
				Private:       true,
				Archived:      false,
				Language:      "Go",
				Size:          1024,
			},
			expectedError: false,
		},
		{
			name:  "repository not found",
			owner: "test-org",
			repo:  "non-existent",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRepository(gomock.Any(), "test-org", "non-existent").
					Return(nil, errors.New("repository not found"))
			},
			expected:      nil,
			expectedError: true,
			errorMessage:  "repository not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			result, err := mockAPI.GetRepository(context.Background(), tt.owner, tt.repo)

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

func TestAPIClient_ListOrganizationRepositories(t *testing.T) {
	tests := []struct {
		name          string
		org           string
		setupMocks    func(*mocks.MockAPIClient)
		expected      []github.RepositoryInfo
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful list repositories",
			org:  "test-org",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				repos := []github.RepositoryInfo{
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
			expected: []github.RepositoryInfo{
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
			name: "empty organization",
			org:  "empty-org",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					ListOrganizationRepositories(gomock.Any(), "empty-org").
					Return([]github.RepositoryInfo{}, nil)
			},
			expected:      []github.RepositoryInfo{},
			expectedError: false,
		},
		{
			name: "API error",
			org:  "test-org",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					ListOrganizationRepositories(gomock.Any(), "test-org").
					Return(nil, errors.New("API rate limit exceeded"))
			},
			expected:      nil,
			expectedError: true,
			errorMessage:  "API rate limit exceeded",
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

func TestAPIClient_GetRateLimit(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.MockAPIClient)
		expected      *github.RateLimit
		expectedError bool
	}{
		{
			name: "successful get rate limit",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRateLimit(gomock.Any()).
					Return(&github.RateLimit{
						Limit:     5000,
						Remaining: 4500,
						Used:      500,
					}, nil)
			},
			expected: &github.RateLimit{
				Limit:     5000,
				Remaining: 4500,
				Used:      500,
			},
			expectedError: false,
		},
		{
			name: "API error",
			setupMocks: func(mockAPI *mocks.MockAPIClient) {
				mockAPI.EXPECT().
					GetRateLimit(gomock.Any()).
					Return(nil, errors.New("authentication failed"))
			},
			expected:      nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockAPI := mocks.NewMockAPIClient(ctrl)
			tt.setupMocks(mockAPI)

			result, err := mockAPI.GetRateLimit(context.Background())

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Limit, result.Limit)
				assert.Equal(t, tt.expected.Remaining, result.Remaining)
				assert.Equal(t, tt.expected.Used, result.Used)
			}
		})
	}
}

func TestCloneService_CloneRepository(t *testing.T) {
	tests := []struct {
		name          string
		repo          github.RepositoryInfo
		targetPath    string
		strategy      string
		setupMocks    func(*mocks.MockCloneService)
		expectedError bool
		errorMessage  string
	}{
		{
			name: "successful clone with reset strategy",
			repo: github.RepositoryInfo{
				Name:     "test-repo",
				CloneURL: "https://github.com/test-org/test-repo.git",
			},
			targetPath: "/tmp/repos",
			strategy:   "reset",
			setupMocks: func(mockClone *mocks.MockCloneService) {
				mockClone.EXPECT().
					CloneRepository(gomock.Any(), gomock.Any(), "/tmp/repos", "reset").
					Return(nil)
			},
			expectedError: false,
		},
		{
			name: "clone with invalid strategy",
			repo: github.RepositoryInfo{
				Name:     "test-repo",
				CloneURL: "https://github.com/test-org/test-repo.git",
			},
			targetPath: "/tmp/repos",
			strategy:   "invalid",
			setupMocks: func(mockClone *mocks.MockCloneService) {
				mockClone.EXPECT().
					CloneRepository(gomock.Any(), gomock.Any(), "/tmp/repos", "invalid").
					Return(errors.New("invalid strategy: invalid"))
			},
			expectedError: true,
			errorMessage:  "invalid strategy",
		},
		{
			name: "clone failure due to network error",
			repo: github.RepositoryInfo{
				Name:     "test-repo",
				CloneURL: "https://github.com/test-org/test-repo.git",
			},
			targetPath: "/tmp/repos",
			strategy:   "reset",
			setupMocks: func(mockClone *mocks.MockCloneService) {
				mockClone.EXPECT().
					CloneRepository(gomock.Any(), gomock.Any(), "/tmp/repos", "reset").
					Return(errors.New("network error: connection timeout"))
			},
			expectedError: true,
			errorMessage:  "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockClone := mocks.NewMockCloneService(ctrl)
			tt.setupMocks(mockClone)

			err := mockClone.CloneRepository(context.Background(), tt.repo, tt.targetPath, tt.strategy)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTokenValidator_ValidateToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		setupMocks    func(*mocks.MockTokenValidatorInterface)
		expected      *github.TokenInfoRecord
		expectedError bool
		errorMessage  string
	}{
		{
			name:  "valid token with all permissions",
			token: "ghp_validtoken123",
			setupMocks: func(mockValidator *mocks.MockTokenValidatorInterface) {
				mockValidator.EXPECT().
					ValidateToken(gomock.Any(), "ghp_validtoken123").
					Return(&github.TokenInfoRecord{
						Valid:  true,
						Scopes: []string{"repo", "user", "admin:org"},
						RateLimit: github.RateLimit{
							Limit:     5000,
							Remaining: 4999,
						},
						User:        "testuser",
						Permissions: []string{"read", "write", "admin"},
					}, nil)
			},
			expected: &github.TokenInfoRecord{
				Valid:  true,
				Scopes: []string{"repo", "user", "admin:org"},
				RateLimit: github.RateLimit{
					Limit:     5000,
					Remaining: 4999,
				},
				User:        "testuser",
				Permissions: []string{"read", "write", "admin"},
			},
			expectedError: false,
		},
		{
			name:  "invalid token",
			token: "invalid_token",
			setupMocks: func(mockValidator *mocks.MockTokenValidatorInterface) {
				mockValidator.EXPECT().
					ValidateToken(gomock.Any(), "invalid_token").
					Return(&github.TokenInfoRecord{
						Valid: false,
					}, errors.New("invalid authentication credentials"))
			},
			expected: &github.TokenInfoRecord{
				Valid: false,
			},
			expectedError: true,
			errorMessage:  "invalid authentication credentials",
		},
		{
			name:  "token with limited scopes",
			token: "ghp_limitedtoken456",
			setupMocks: func(mockValidator *mocks.MockTokenValidatorInterface) {
				mockValidator.EXPECT().
					ValidateToken(gomock.Any(), "ghp_limitedtoken456").
					Return(&github.TokenInfoRecord{
						Valid:  true,
						Scopes: []string{"repo:status", "public_repo"},
						RateLimit: github.RateLimit{
							Limit:     5000,
							Remaining: 3000,
						},
						User:        "limiteduser",
						Permissions: []string{"read"},
					}, nil)
			},
			expected: &github.TokenInfoRecord{
				Valid:  true,
				Scopes: []string{"repo:status", "public_repo"},
				RateLimit: github.RateLimit{
					Limit:     5000,
					Remaining: 3000,
				},
				User:        "limiteduser",
				Permissions: []string{"read"},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockValidator := mocks.NewMockTokenValidatorInterface(ctrl)
			tt.setupMocks(mockValidator)

			result, err := mockValidator.ValidateToken(context.Background(), tt.token)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected.Valid, result.Valid)
				assert.Equal(t, tt.expected.User, result.User)
				assert.Equal(t, tt.expected.Scopes, result.Scopes)
			}
		})
	}
}
