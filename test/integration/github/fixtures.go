package github_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gizzahub/gzh-manager-go/pkg/config"
	"gopkg.in/yaml.v3"
)

// TestFixtures provides test data for integration tests
type TestFixtures struct {
	Organization string
	Token        string
	BaseDir      string
}

// NewTestFixtures creates fixtures from environment
func NewTestFixtures() (*TestFixtures, error) {
	org := os.Getenv(testOrgEnvVar)
	if org == "" {
		return nil, fmt.Errorf("%s environment variable not set", testOrgEnvVar)
	}

	token := os.Getenv(tokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("%s environment variable not set", tokenEnvVar)
	}

	tmpDir, err := os.MkdirTemp("", "github-integration-test-*")
	if err != nil {
		return nil, err
	}

	return &TestFixtures{
		Organization: org,
		Token:        token,
		BaseDir:      tmpDir,
	}, nil
}

// Cleanup removes temporary files
func (f *TestFixtures) Cleanup() {
	if f.BaseDir != "" {
		os.RemoveAll(f.BaseDir)
	}
}

// CreateConfigFile creates a test configuration file
func (f *TestFixtures) CreateConfigFile(filename string, config interface{}) (string, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return "", err
	}

	filepath := filepath.Join(f.BaseDir, filename)
	err = os.WriteFile(filepath, data, 0o644)
	if err != nil {
		return "", err
	}

	return filepath, nil
}

// SampleRepoConfigs provides various test configurations
var SampleRepoConfigs = struct {
	Basic          *config.RepoConfig
	SecurityFocus  *config.RepoConfig
	OpenSource     *config.RepoConfig
	Enterprise     *config.RepoConfig
	WithExceptions *config.RepoConfig
}{
	Basic: &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"basic": {
				Description: "Basic repository template",
				Settings: &config.RepoSettings{
					Private:   boolPtr(false),
					HasIssues: boolPtr(true),
					HasWiki:   boolPtr(true),
				},
			},
		},
		Repositories: []config.RepoSpecificConfig{
			{
				Name:     "*",
				Template: "basic",
			},
		},
	},

	SecurityFocus: &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"security-enhanced": {
				Description: "Security-focused template",
				Base:        "basic",
				Settings: &config.RepoSettings{
					Private:                  boolPtr(true),
					WebCommitSignoffRequired: boolPtr(true),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts:          boolPtr(true),
					AutomatedSecurityFixes:       boolPtr(true),
					SecretScanning:               boolPtr(true),
					SecretScanningPushProtection: boolPtr(true),
					BranchProtection: map[string]*config.BranchProtectionRule{
						"main": {
							RequiredReviews:               intPtr(2),
							DismissStaleReviews:           boolPtr(true),
							RequireCodeOwnerReviews:       boolPtr(true),
							RequiredStatusChecks:          []string{"ci/build", "ci/test", "security/scan"},
							StrictStatusChecks:            boolPtr(true),
							EnforceAdmins:                 boolPtr(true),
							RequireConversationResolution: boolPtr(true),
							AllowForcePushes:              boolPtr(false),
							AllowDeletions:                boolPtr(false),
						},
					},
				},
			},
		},
		Policies: map[string]*config.PolicyTemplate{
			"security-compliance": {
				Description: "Security compliance policy",
				Rules: map[string]config.PolicyRule{
					"private_required": {
						Type:        "visibility",
						Value:       "private",
						Enforcement: "required",
						Message:     "All repositories must be private",
					},
					"vulnerability_alerts": {
						Type:        "security_feature",
						Value:       true,
						Enforcement: "required",
						Message:     "Vulnerability alerts must be enabled",
					},
					"branch_protection": {
						Type:        "branch_protection",
						Value:       true,
						Enforcement: "required",
						Message:     "Main branch must be protected",
					},
				},
			},
		},
	},

	OpenSource: &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"opensource": {
				Description: "Open source project template",
				Settings: &config.RepoSettings{
					Private:     boolPtr(false),
					HasIssues:   boolPtr(true),
					HasWiki:     boolPtr(true),
					HasProjects: boolPtr(true),
					HasPages:    boolPtr(true),
				},
				Permissions: &config.RepoPermissions{
					Admin:    []string{"opensource-admins"},
					Maintain: []string{"opensource-maintainers"},
					Push:     []string{"opensource-contributors"},
				},
				RequiredFiles: []config.RequiredFile{
					{
						Path:    "LICENSE",
						Content: "MIT License\n\nCopyright (c) 2024 Test Organization\n",
					},
					{
						Path:    "README.md",
						Content: "# Project Name\n\n## Description\n\n## Installation\n\n## Usage\n\n## Contributing\n\n## License\n",
					},
					{
						Path:    "CONTRIBUTING.md",
						Content: "# Contributing Guidelines\n\n## Code of Conduct\n\n## How to Contribute\n\n## Pull Request Process\n",
					},
					{
						Path:    "CODE_OF_CONDUCT.md",
						Content: "# Code of Conduct\n\n## Our Pledge\n\n## Our Standards\n\n## Enforcement\n",
					},
				},
			},
		},
	},

	Enterprise: &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"enterprise-standard": {
				Description: "Enterprise standard template",
				Settings: &config.RepoSettings{
					Private:             boolPtr(true),
					HasIssues:           boolPtr(true),
					HasWiki:             boolPtr(false),
					HasProjects:         boolPtr(true),
					AllowSquashMerge:    boolPtr(true),
					AllowMergeCommit:    boolPtr(false),
					AllowRebaseMerge:    boolPtr(false),
					DeleteBranchOnMerge: boolPtr(true),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
					SecretScanning:      boolPtr(true),
					BranchProtection: map[string]*config.BranchProtectionRule{
						"main": {
							RequiredReviews:         intPtr(2),
							RequireCodeOwnerReviews: boolPtr(true),
							RequiredStatusChecks:    []string{"ci/build", "ci/test", "sonarqube"},
							EnforceAdmins:           boolPtr(false),
							RestrictPushAccess:      []string{"senior-developers"},
						},
						"develop": {
							RequiredReviews:      intPtr(1),
							RequiredStatusChecks: []string{"ci/build", "ci/test"},
						},
					},
				},
				Webhooks: []config.WebhookConfig{
					{
						Name:   "jenkins",
						URL:    "https://jenkins.internal/github-webhook/",
						Events: []string{"push", "pull_request"},
						Active: boolPtr(true),
					},
					{
						Name:   "sonarqube",
						URL:    "https://sonar.internal/webhook/github",
						Events: []string{"push"},
						Active: boolPtr(true),
					},
				},
			},
		},
	},

	WithExceptions: &config.RepoConfig{
		Version:      "1.0.0",
		Organization: "test-org",
		Templates: map[string]*config.RepoTemplate{
			"strict": {
				Description: "Strict security template",
				Settings: &config.RepoSettings{
					Private: boolPtr(true),
				},
				Security: &config.SecuritySettings{
					VulnerabilityAlerts: boolPtr(true),
					SecretScanning:      boolPtr(true),
				},
			},
		},
		Policies: map[string]*config.PolicyTemplate{
			"strict-policy": {
				Description: "Strict security policy",
				Rules: map[string]config.PolicyRule{
					"must_be_private": {
						Type:        "visibility",
						Value:       "private",
						Enforcement: "required",
						Message:     "All repositories must be private",
					},
				},
			},
		},
		Repositories: []config.RepoSpecificConfig{
			{
				Name:     "public-exception-repo",
				Template: "strict",
				Exceptions: []config.PolicyException{
					{
						PolicyName: "strict-policy",
						RuleName:   "must_be_private",
						Reason:     "This repository needs to be public for documentation",
						ApprovedBy: "security-team",
						ExpiresAt:  "2025-01-01",
					},
				},
			},
		},
	},
}

// TestScenarios provides predefined test scenarios
var TestScenarios = struct {
	BasicSetup          TestScenario
	SecurityCompliance  TestScenario
	BulkOperations      TestScenario
	PolicyEnforcement   TestScenario
	TemplateInheritance TestScenario
}{
	BasicSetup: TestScenario{
		Name:        "Basic Repository Setup",
		Description: "Test basic repository configuration management",
		Config:      SampleRepoConfigs.Basic,
		Steps: []TestStep{
			{
				Name:   "List repositories",
				Action: "list",
			},
			{
				Name:   "Apply basic configuration",
				Action: "apply",
			},
			{
				Name:   "Verify configuration",
				Action: "verify",
			},
		},
	},

	SecurityCompliance: TestScenario{
		Name:        "Security Compliance Check",
		Description: "Test security policy compliance and enforcement",
		Config:      SampleRepoConfigs.SecurityFocus,
		Steps: []TestStep{
			{
				Name:   "Define security policies",
				Action: "create-policy",
			},
			{
				Name:   "Run compliance audit",
				Action: "audit",
			},
			{
				Name:   "Apply security configuration",
				Action: "apply",
			},
			{
				Name:   "Verify compliance",
				Action: "verify-compliance",
			},
		},
	},

	BulkOperations: TestScenario{
		Name:        "Bulk Repository Updates",
		Description: "Test bulk operations on multiple repositories",
		Config:      SampleRepoConfigs.Enterprise,
		Steps: []TestStep{
			{
				Name:   "Select target repositories",
				Action: "filter",
			},
			{
				Name:   "Apply bulk configuration",
				Action: "bulk-apply",
			},
			{
				Name:   "Verify all updates",
				Action: "bulk-verify",
			},
		},
	},
}

// TestScenario represents a complete test scenario
type TestScenario struct {
	Name        string
	Description string
	Config      *config.RepoConfig
	Steps       []TestStep
}

// TestStep represents a single step in a test scenario
type TestStep struct {
	Name   string
	Action string
	Data   interface{}
}
