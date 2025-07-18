package config

// GetPredefinedPolicyTemplates returns predefined policy templates for common compliance frameworks.
func GetPredefinedPolicyTemplates() map[string]*PolicyTemplate {
	return map[string]*PolicyTemplate{
		// Security Policies
		"branch_protection": {
			Description: "Enforce branch protection rules on critical branches",
			Group:       "security",
			Severity:    "critical",
			Rules: map[string]PolicyRule{
				"main_protected": {
					Type:        "branch_protection",
					Value:       true,
					Enforcement: "required",
					Message:     "Main branch must be protected",
				},
				"require_reviews": {
					Type:        "min_reviews",
					Value:       2,
					Enforcement: "required",
					Message:     "Minimum 2 reviews required for pull requests",
				},
				"dismiss_stale": {
					Type:        "dismiss_stale_reviews",
					Value:       true,
					Enforcement: "recommended",
					Message:     "Stale reviews should be dismissed when new commits are pushed",
				},
				"enforce_admins": {
					Type:        "enforce_admins",
					Value:       true,
					Enforcement: "required",
					Message:     "Branch protection must apply to administrators",
				},
			},
			Tags: []string{"security", "code-review", "branch-management"},
		},

		"vulnerability_management": {
			Description: "Enable security vulnerability detection and management",
			Group:       "security",
			Severity:    "critical",
			Rules: map[string]PolicyRule{
				"vulnerability_alerts": {
					Type:        "security_feature",
					Value:       "vulnerability_alerts",
					Enforcement: "required",
					Message:     "Vulnerability alerts must be enabled",
				},
				"security_advisories": {
					Type:        "security_feature",
					Value:       "security_advisories",
					Enforcement: "required",
					Message:     "Security advisories must be enabled",
				},
				"dependabot_alerts": {
					Type:        "security_feature",
					Value:       "dependabot_security_updates",
					Enforcement: "recommended",
					Message:     "Dependabot security updates should be enabled",
				},
			},
			Tags: []string{"security", "vulnerabilities", "dependencies"},
		},

		"access_control": {
			Description: "Enforce proper access control and permissions",
			Group:       "security",
			Severity:    "high",
			Rules: map[string]PolicyRule{
				"restrict_push": {
					Type:        "restrict_push_access",
					Value:       true,
					Enforcement: "recommended",
					Message:     "Direct push access should be restricted",
				},
				"team_permissions": {
					Type:        "team_permission_model",
					Value:       "least_privilege",
					Enforcement: "required",
					Message:     "Teams should have least privilege access",
				},
			},
			Tags: []string{"security", "access-control", "permissions"},
		},

		// Compliance Policies
		"required_documentation": {
			Description: "Ensure required documentation files are present",
			Group:       "compliance",
			Severity:    "medium",
			Rules: map[string]PolicyRule{
				"readme": {
					Type:        "file_exists",
					Value:       "README.md",
					Enforcement: "required",
					Message:     "README.md file is required",
				},
				"license": {
					Type:        "file_exists",
					Value:       "LICENSE",
					Enforcement: "required",
					Message:     "LICENSE file is required",
				},
				"security_policy": {
					Type:        "file_exists",
					Value:       "SECURITY.md",
					Enforcement: "recommended",
					Message:     "SECURITY.md file should be present",
				},
			},
			Tags: []string{"compliance", "documentation"},
		},

		"audit_logging": {
			Description: "Ensure audit logging and monitoring is configured",
			Group:       "compliance",
			Severity:    "high",
			Rules: map[string]PolicyRule{
				"audit_webhook": {
					Type:        "webhook_exists",
					Value:       "audit",
					Enforcement: "required",
					Message:     "Audit webhook must be configured",
				},
				"event_logging": {
					Type:        "feature_enabled",
					Value:       "audit_log",
					Enforcement: "required",
					Message:     "Audit logging must be enabled",
				},
			},
			Tags: []string{"compliance", "audit", "monitoring"},
		},

		// Best Practice Policies
		"ci_cd_pipeline": {
			Description: "Ensure CI/CD pipelines are properly configured",
			Group:       "best-practice",
			Severity:    "medium",
			Rules: map[string]PolicyRule{
				"ci_workflow": {
					Type:        "workflow_exists",
					Value:       ".github/workflows/ci.yml",
					Enforcement: "recommended",
					Message:     "CI workflow should be configured",
				},
				"status_checks": {
					Type:        "required_status_checks",
					Value:       []string{"build", "test"},
					Enforcement: "recommended",
					Message:     "Status checks should be required",
				},
			},
			Tags: []string{"best-practice", "ci-cd", "automation"},
		},

		"code_quality": {
			Description: "Enforce code quality standards",
			Group:       "best-practice",
			Severity:    "low",
			Rules: map[string]PolicyRule{
				"linter_config": {
					Type:        "file_exists",
					Value:       ".eslintrc.json",
					Enforcement: "optional",
					Message:     "Linter configuration should be present",
				},
				"code_owners": {
					Type:        "file_exists",
					Value:       "CODEOWNERS",
					Enforcement: "recommended",
					Message:     "CODEOWNERS file should be defined",
				},
			},
			Tags: []string{"best-practice", "code-quality"},
		},

		// Repository Management
		"repository_hygiene": {
			Description: "Maintain clean and organized repositories",
			Group:       "best-practice",
			Severity:    "low",
			Rules: map[string]PolicyRule{
				"delete_merged_branches": {
					Type:        "delete_branch_on_merge",
					Value:       true,
					Enforcement: "recommended",
					Message:     "Merged branches should be automatically deleted",
				},
				"issue_templates": {
					Type:        "directory_exists",
					Value:       ".github/ISSUE_TEMPLATE",
					Enforcement: "optional",
					Message:     "Issue templates improve issue quality",
				},
			},
			Tags: []string{"best-practice", "repository-management"},
		},
	}
}

// GetPolicyPresets returns predefined policy presets for compliance frameworks.
func GetPolicyPresets() map[string]*PolicyPreset {
	return map[string]*PolicyPreset{
		"soc2": {
			Name:        "SOC 2 Type II",
			Description: "Service Organization Control 2 compliance requirements",
			Framework:   "SOC2",
			Version:     "2017",
			Groups:      []string{"security", "compliance"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"required_documentation",
				"audit_logging",
			},
			Overrides: map[string]PolicyOverride{
				"branch_protection": {
					Rules: map[string]RuleOverride{
						"require_reviews": {
							Value: 2, // SOC2 requires minimum 2 reviewers
						},
					},
				},
			},
		},

		"iso27001": {
			Name:        "ISO 27001:2022",
			Description: "Information Security Management System requirements",
			Framework:   "ISO27001",
			Version:     "2022",
			Groups:      []string{"security", "compliance"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"required_documentation",
				"audit_logging",
				"repository_hygiene",
			},
			Overrides: map[string]PolicyOverride{
				"access_control": {
					Enforcement: "required", // ISO 27001 makes this mandatory
				},
				"audit_logging": {
					Rules: map[string]RuleOverride{
						"event_logging": {
							Enforcement: "required",
						},
					},
				},
			},
		},

		"nist": {
			Name:        "NIST Cybersecurity Framework",
			Description: "NIST CSF security controls",
			Framework:   "NIST-CSF",
			Version:     "1.1",
			Groups:      []string{"security", "compliance", "best-practice"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"required_documentation",
				"audit_logging",
				"ci_cd_pipeline",
			},
			Overrides: map[string]PolicyOverride{
				"vulnerability_management": {
					Rules: map[string]RuleOverride{
						"dependabot_alerts": {
							Enforcement: "required", // NIST requires automated vulnerability scanning
						},
					},
				},
			},
		},

		"pci-dss": {
			Name:        "PCI DSS v4.0",
			Description: "Payment Card Industry Data Security Standard",
			Framework:   "PCI-DSS",
			Version:     "4.0",
			Groups:      []string{"security", "compliance"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"audit_logging",
			},
			Overrides: map[string]PolicyOverride{
				"branch_protection": {
					Rules: map[string]RuleOverride{
						"require_reviews": {
							Value:       2,
							Enforcement: "required",
						},
						"enforce_admins": {
							Value:       true,
							Enforcement: "required",
						},
					},
				},
				"access_control": {
					Enforcement: "required",
					Rules: map[string]RuleOverride{
						"restrict_push": {
							Value:       true,
							Enforcement: "required",
						},
					},
				},
			},
		},

		"hipaa": {
			Name:        "HIPAA Security Rule",
			Description: "Health Insurance Portability and Accountability Act requirements",
			Framework:   "HIPAA",
			Version:     "2013",
			Groups:      []string{"security", "compliance"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"audit_logging",
				"required_documentation",
			},
			Overrides: map[string]PolicyOverride{
				"audit_logging": {
					Enforcement: "required",
					Rules: map[string]RuleOverride{
						"audit_webhook": {
							Enforcement: "required",
						},
						"event_logging": {
							Enforcement: "required",
						},
					},
				},
				"required_documentation": {
					Rules: map[string]RuleOverride{
						"security_policy": {
							Enforcement: "required", // HIPAA requires security documentation
						},
					},
				},
			},
		},

		"gdpr": {
			Name:        "GDPR Compliance",
			Description: "General Data Protection Regulation requirements",
			Framework:   "GDPR",
			Version:     "2018",
			Groups:      []string{"compliance", "security"},
			Policies: []string{
				"access_control",
				"audit_logging",
				"required_documentation",
				"vulnerability_management",
			},
			Overrides: map[string]PolicyOverride{
				"required_documentation": {
					Rules: map[string]RuleOverride{
						"privacy_policy": {
							Value:       "PRIVACY.md",
							Enforcement: "required",
						},
					},
				},
			},
		},

		"minimal": {
			Name:        "Minimal Security",
			Description: "Basic security requirements for all repositories",
			Framework:   "Custom",
			Version:     "1.0",
			Groups:      []string{"security"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
			},
			Overrides: map[string]PolicyOverride{
				"branch_protection": {
					Rules: map[string]RuleOverride{
						"require_reviews": {
							Value: 1, // Minimal requires at least 1 reviewer
						},
						"enforce_admins": {
							Enforcement: "recommended", // Less strict for minimal
						},
					},
				},
			},
		},

		"enterprise": {
			Name:        "Enterprise Standard",
			Description: "Comprehensive enterprise security and compliance",
			Framework:   "Custom",
			Version:     "2.0",
			Groups:      []string{"security", "compliance", "best-practice"},
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
				"required_documentation",
				"audit_logging",
				"ci_cd_pipeline",
				"code_quality",
				"repository_hygiene",
			},
			Overrides: map[string]PolicyOverride{
				"branch_protection": {
					Rules: map[string]RuleOverride{
						"require_reviews": {
							Value: 3, // Enterprise requires more reviewers
						},
					},
				},
				"ci_cd_pipeline": {
					Enforcement: "required", // CI/CD is mandatory for enterprise
				},
			},
		},
	}
}

// GetPolicyGroups returns predefined policy groups.
func GetPolicyGroups() map[string]*PolicyGroup {
	return map[string]*PolicyGroup{
		"security": {
			Name:        "Security Policies",
			Description: "Policies related to security controls and vulnerability management",
			Weight:      0.4, // 40% of total score
			Policies: []string{
				"branch_protection",
				"vulnerability_management",
				"access_control",
			},
			Required: true, // All security policies must pass
			Tags:     []string{"critical", "security"},
		},
		"compliance": {
			Name:        "Compliance Policies",
			Description: "Policies ensuring regulatory and organizational compliance",
			Weight:      0.35, // 35% of total score
			Policies: []string{
				"required_documentation",
				"audit_logging",
			},
			Required: false,
			Tags:     []string{"compliance", "audit"},
		},
		"best-practice": {
			Name:        "Best Practice Policies",
			Description: "Recommended practices for maintainable and efficient repositories",
			Weight:      0.25, // 25% of total score
			Policies: []string{
				"ci_cd_pipeline",
				"code_quality",
				"repository_hygiene",
			},
			Required: false,
			Tags:     []string{"quality", "efficiency"},
		},
	}
}
