// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package template

// GetBuiltinTemplates returns a map of built-in template configurations
func GetBuiltinTemplates() map[string]*TemplateConfig {
	return map[string]*TemplateConfig{
		"enterprise": getEnterpriseTemplate(),
		"minimal":    getMinimalTemplate(),
		"multi-org":  getMultiOrgTemplate(),
		"personal":   getPersonalTemplate(),
	}
}

// getEnterpriseTemplate returns the enterprise template configuration
func getEnterpriseTemplate() *TemplateConfig {
	return &TemplateConfig{
		Name:        "Enterprise Configuration",
		Description: "Multi-organization setup with security and compliance features",
		Template: map[string]interface{}{
			"version": "1.0.0",
			"global": map[string]interface{}{
				"clone_base_dir":   "${HOME}/enterprise-repos",
				"default_strategy": "reset",
				"concurrency": map[string]interface{}{
					"clone_workers":  5,
					"update_workers": 10,
				},
				"security": map[string]interface{}{
					"enforce_https":     true,
					"verify_signatures": true,
					"scan_for_secrets":  true,
				},
			},
			"providers": map[string]interface{}{
				"github": map[string]interface{}{
					"organizations": []map[string]interface{}{
						{
							"name":       "{{.CompanyOrg}}",
							"clone_dir":  "${HOME}/enterprise-repos/{{.CompanyOrg}}",
							"visibility": "private",
							"exclude": []string{
								".*-archive$",
								".*-deprecated$",
								".*-legacy$",
							},
							"auth": map[string]interface{}{
								"token": "${GITHUB_ENTERPRISE_TOKEN}",
							},
							"compliance": map[string]interface{}{
								"require_branch_protection": true,
								"scan_vulnerabilities":      true,
							},
						},
					},
				},
				"gitlab": map[string]interface{}{
					"groups": []map[string]interface{}{
						{
							"name":       "{{.CompanyGroup}}",
							"clone_dir":  "${HOME}/enterprise-repos/gitlab/{{.CompanyGroup}}",
							"visibility": "private",
							"auth": map[string]interface{}{
								"token": "${GITLAB_ENTERPRISE_TOKEN}",
							},
						},
					},
				},
			},
			"sync_mode": map[string]interface{}{
				"cleanup_orphans":       true,
				"conflict_resolution":   "remote-overwrite",
				"backup_before_cleanup": true,
			},
			"monitoring": map[string]interface{}{
				"enable_metrics": true,
				"log_level":      "info",
				"audit_trail":    true,
			},
		},
		Variables: []TemplateVariable{
			{
				Name:        "CompanyOrg",
				Description: "Your company's GitHub organization name",
				Required:    true,
				Type:        "string",
			},
			{
				Name:         "CompanyGroup",
				Description:  "Your company's GitLab group name",
				Required:     false,
				Type:         "string",
				DefaultValue: "{{.CompanyOrg}}",
			},
		},
	}
}

// getMinimalTemplate returns the minimal template configuration
func getMinimalTemplate() *TemplateConfig {
	return &TemplateConfig{
		Name:        "Minimal Configuration",
		Description: "Simple setup for personal or small team use",
		Template: map[string]interface{}{
			"version": "1.0.0",
			"global": map[string]interface{}{
				"clone_base_dir":   "${HOME}/repos",
				"default_strategy": "pull",
			},
			"providers": map[string]interface{}{
				"github": map[string]interface{}{
					"organizations": []map[string]interface{}{
						{
							"name":      "{{.GitHubOrg}}",
							"clone_dir": "${HOME}/repos/{{.GitHubOrg}}",
						},
					},
				},
			},
		},
		Variables: []TemplateVariable{
			{
				Name:        "GitHubOrg",
				Description: "GitHub organization or username",
				Required:    true,
				Type:        "string",
			},
		},
	}
}

// getMultiOrgTemplate returns the multi-organization template configuration
func getMultiOrgTemplate() *TemplateConfig {
	return &TemplateConfig{
		Name:        "Multi-Organization Configuration",
		Description: "Setup for managing multiple organizations across different platforms",
		Template: map[string]interface{}{
			"version": "1.0.0",
			"global": map[string]interface{}{
				"clone_base_dir":   "${HOME}/multi-org-repos",
				"default_strategy": "reset",
				"concurrency": map[string]interface{}{
					"clone_workers":  8,
					"update_workers": 12,
				},
			},
			"providers": map[string]interface{}{
				"github": map[string]interface{}{
					"organizations": []map[string]interface{}{
						{
							"name":      "{{.PrimaryGitHubOrg}}",
							"clone_dir": "${HOME}/multi-org-repos/github/{{.PrimaryGitHubOrg}}",
							"priority":  1,
						},
						{
							"name":      "{{.SecondaryGitHubOrg}}",
							"clone_dir": "${HOME}/multi-org-repos/github/{{.SecondaryGitHubOrg}}",
							"priority":  2,
						},
					},
				},
				"gitlab": map[string]interface{}{
					"groups": []map[string]interface{}{
						{
							"name":      "{{.GitLabGroup}}",
							"clone_dir": "${HOME}/multi-org-repos/gitlab/{{.GitLabGroup}}",
						},
					},
				},
				"gitea": map[string]interface{}{
					"organizations": []map[string]interface{}{
						{
							"name":      "{{.GiteaOrg}}",
							"clone_dir": "${HOME}/multi-org-repos/gitea/{{.GiteaOrg}}",
							"base_url":  "{{.GiteaURL}}",
						},
					},
				},
			},
			"sync_mode": map[string]interface{}{
				"cleanup_orphans":     true,
				"conflict_resolution": "interactive",
			},
		},
		Variables: []TemplateVariable{
			{
				Name:        "PrimaryGitHubOrg",
				Description: "Primary GitHub organization name",
				Required:    true,
				Type:        "string",
			},
			{
				Name:        "SecondaryGitHubOrg",
				Description: "Secondary GitHub organization name",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "GitLabGroup",
				Description: "GitLab group name",
				Required:    false,
				Type:        "string",
			},
			{
				Name:        "GiteaOrg",
				Description: "Gitea organization name",
				Required:    false,
				Type:        "string",
			},
			{
				Name:         "GiteaURL",
				Description:  "Gitea instance URL",
				Required:     false,
				Type:         "string",
				DefaultValue: "https://gitea.com",
			},
		},
	}
}

// getPersonalTemplate returns the personal template configuration
func getPersonalTemplate() *TemplateConfig {
	return &TemplateConfig{
		Name:        "Personal Configuration",
		Description: "Setup for personal repositories and projects",
		Template: map[string]interface{}{
			"version": "1.0.0",
			"global": map[string]interface{}{
				"clone_base_dir":   "${HOME}/personal-repos",
				"default_strategy": "pull",
				"concurrency": map[string]interface{}{
					"clone_workers":  3,
					"update_workers": 6,
				},
			},
			"providers": map[string]interface{}{
				"github": map[string]interface{}{
					"organizations": []map[string]interface{}{
						{
							"name":          "{{.GitHubUsername}}",
							"clone_dir":     "${HOME}/personal-repos/{{.GitHubUsername}}",
							"include_forks": false,
							"exclude": []string{
								".*-test$",
								".*-experiment$",
							},
						},
					},
				},
			},
			"sync_mode": map[string]interface{}{
				"cleanup_orphans":     false,
				"conflict_resolution": "local-keep",
			},
			"filters": map[string]interface{}{
				"min_stars":        0,
				"languages":        []string{"{{.PreferredLanguage}}"},
				"exclude_archived": true,
			},
		},
		Variables: []TemplateVariable{
			{
				Name:        "GitHubUsername",
				Description: "Your GitHub username",
				Required:    true,
				Type:        "string",
			},
			{
				Name:         "PreferredLanguage",
				Description:  "Your preferred programming language",
				Required:     false,
				Type:         "string",
				DefaultValue: "Go",
				Options:      []string{"Go", "Python", "JavaScript", "Java", "C++", "Rust"},
			},
		},
	}
}
