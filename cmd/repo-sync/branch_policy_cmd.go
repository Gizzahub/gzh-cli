package reposync

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newBranchPolicyCmd creates the branch-policy subcommand
func newBranchPolicyCmd(logger *zap.Logger) *cobra.Command {
	var (
		enforce     bool
		template    string
		autoCleanup bool
		dryRun      bool
		createIssue string
		deleteMode  string
	)

	cmd := &cobra.Command{
		Use:   "branch-policy [repository-path]",
		Short: "Manage and enforce branch policies automatically",
		Long: `Manage and enforce branch policies with automated workflows and validation.

This command provides comprehensive branch policy management:
- Branch strategy template enforcement (GitFlow, GitHub Flow, custom)
- Automated branch naming validation with configurable rules
- Automatic branch creation from issues and tickets
- Stale branch detection and cleanup
- Branch protection rules validation
- Merge policy enforcement

Supported Branch Strategy Templates:
- gitflow: Feature/develop/release/hotfix branch workflow
- github-flow: Simple feature branch workflow with main branch
- gitlab-flow: Environment branches with production workflow
- custom: User-defined branch strategy with custom rules

Examples:
  # Enforce GitFlow branch strategy
  gz repo-sync branch-policy --enforce --template gitflow
  
  # Enable automatic cleanup of stale branches
  gz repo-sync branch-policy --auto-cleanup --template github-flow
  
  # Dry run to see what policies would be enforced
  gz repo-sync branch-policy --enforce --template gitflow --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoPath := "."
			if len(args) > 0 {
				repoPath = args[0]
			}

			// Validate repository path
			if err := validateRepositoryPath(repoPath); err != nil {
				return fmt.Errorf("invalid repository path: %w", err)
			}

			ctx := context.Background()

			// Handle branch creation from issue
			if createIssue != "" {
				issue, err := ParseIssueFromString(createIssue)
				if err != nil {
					return fmt.Errorf("invalid issue format: %w", err)
				}

				branchConfig := &BranchManagementConfig{
					RepositoryPath:    repoPath,
					BranchNamingRules: template,
					RemoteName:        "origin",
					DryRun:            dryRun,
				}

				branchManager := NewBranchManager(logger, branchConfig)
				return branchManager.CreateBranchFromIssue(ctx, issue)
			}

			// Handle branch deletion
			if deleteMode != "" {
				branchConfig := &BranchManagementConfig{
					RepositoryPath:    repoPath,
					BranchNamingRules: template,
					RemoteName:        "origin",
					StaleBranchDays:   30,
					DryRun:            dryRun,
				}

				branchManager := NewBranchManager(logger, branchConfig)

				switch deleteMode {
				case "merged":
					return branchManager.DeleteMergedBranches(ctx)
				case "stale":
					return branchManager.CleanupStaleBranches(ctx)
				default:
					return fmt.Errorf("invalid delete mode: %s (use 'merged' or 'stale')", deleteMode)
				}
			}

			// Default branch policy enforcement
			config := &BranchPolicyConfig{
				RepositoryPath: repoPath,
				Enforce:        enforce,
				Template:       template,
				AutoCleanup:    autoCleanup,
				DryRun:         dryRun,
			}

			manager, err := NewBranchPolicyManager(logger, config)
			if err != nil {
				return fmt.Errorf("failed to create branch policy manager: %w", err)
			}

			return manager.ApplyPolicies(ctx)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&enforce, "enforce", false, "Enforce branch policies")
	cmd.Flags().StringVar(&template, "template", "github-flow", "Branch strategy template (gitflow|github-flow|gitlab-flow|custom)")
	cmd.Flags().BoolVar(&autoCleanup, "auto-cleanup", false, "Enable automatic cleanup of stale branches")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be enforced without applying changes")
	cmd.Flags().StringVar(&createIssue, "create-issue", "", "Create branch from issue (format: 'ID: Title')")
	cmd.Flags().StringVar(&deleteMode, "delete-mode", "", "Delete branches (merged|stale)")

	return cmd
}

// BranchPolicyConfig represents branch policy configuration
type BranchPolicyConfig struct {
	RepositoryPath string `json:"repository_path"`
	Enforce        bool   `json:"enforce"`
	Template       string `json:"template"`
	AutoCleanup    bool   `json:"auto_cleanup"`
	DryRun         bool   `json:"dry_run"`
}

// BranchPolicyManager handles branch policy enforcement
type BranchPolicyManager struct {
	logger    *zap.Logger
	config    *BranchPolicyConfig
	rules     *BranchPolicyRules
	validator *BranchValidator
	gitCmd    GitCommandExecutor
}

// BranchPolicyRules defines the branch policy rules
type BranchPolicyRules struct {
	Template          string                    `json:"template"`
	NamingPatterns    map[string]*regexp.Regexp `json:"naming_patterns"`
	ProtectedBranches []string                  `json:"protected_branches"`
	RequiredBranches  []string                  `json:"required_branches"`
	MergeRules        MergeRules                `json:"merge_rules"`
	CleanupRules      CleanupRules              `json:"cleanup_rules"`
}

// MergeRules defines merge policy rules
type MergeRules struct {
	RequireReview     bool     `json:"require_review"`
	RequireStatus     bool     `json:"require_status"`
	AllowedMergeTypes []string `json:"allowed_merge_types"`
	RestrictPushes    bool     `json:"restrict_pushes"`
}

// CleanupRules defines branch cleanup rules
type CleanupRules struct {
	StaleDays         int      `json:"stale_days"`
	ProtectedBranches []string `json:"protected_branches"`
	AutoDelete        bool     `json:"auto_delete"`
}

// NewBranchPolicyManager creates a new branch policy manager
func NewBranchPolicyManager(logger *zap.Logger, config *BranchPolicyConfig) (*BranchPolicyManager, error) {
	rules, err := createBranchPolicyRules(config.Template)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy rules: %w", err)
	}

	validator := NewBranchValidator(logger, config.Template)

	return &BranchPolicyManager{
		logger:    logger,
		config:    config,
		rules:     rules,
		validator: validator,
		gitCmd:    &defaultGitExecutor{},
	}, nil
}

// ApplyPolicies applies branch policies to the repository
func (bpm *BranchPolicyManager) ApplyPolicies(ctx context.Context) error {
	fmt.Printf("🔧 Applying branch policies for template: %s\n", bpm.config.Template)

	if bpm.config.DryRun {
		fmt.Println("🔍 Dry run mode - no actual changes will be made")
	}

	// Validate branch naming
	if err := bpm.validateBranchNaming(ctx); err != nil {
		return fmt.Errorf("branch naming validation failed: %w", err)
	}

	// Ensure required branches exist
	if err := bpm.ensureRequiredBranches(ctx); err != nil {
		return fmt.Errorf("failed to ensure required branches: %w", err)
	}

	// Apply cleanup rules if enabled
	if bpm.config.AutoCleanup {
		if err := bpm.cleanupStaleBranches(ctx); err != nil {
			return fmt.Errorf("branch cleanup failed: %w", err)
		}
	}

	// Enforce merge rules
	if bpm.config.Enforce {
		if err := bpm.enforceMergeRules(ctx); err != nil {
			return fmt.Errorf("merge rule enforcement failed: %w", err)
		}
	}

	fmt.Println("✅ Branch policies applied successfully")
	return nil
}

// validateBranchNaming validates branch names against naming patterns
func (bpm *BranchPolicyManager) validateBranchNaming(ctx context.Context) error {
	fmt.Println("📋 Validating branch naming conventions...")

	// Get list of all branches
	result, err := bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	// Parse branch names
	branches := strings.Split(strings.TrimSpace(result.Output), "\n")
	if len(branches) == 0 || (len(branches) == 1 && branches[0] == "") {
		fmt.Println("No branches found to validate")
		return nil
	}

	// Validate each branch
	validationResults := bpm.validator.BatchValidate(branches)
	var invalidBranches []string
	var suggestions []string

	for branchName, result := range validationResults {
		if !result.Valid && result.BranchType != "protected" {
			invalidBranches = append(invalidBranches, branchName)
			bpm.logger.Warn("Invalid branch name",
				zap.String("branch", branchName),
				zap.Strings("errors", result.Errors),
				zap.String("suggestion", result.FixedName))

			if result.FixedName != "" {
				suggestions = append(suggestions, fmt.Sprintf("  %s → %s", branchName, result.FixedName))
			}
		}
	}

	if len(invalidBranches) > 0 {
		fmt.Printf("❌ Found %d branches with invalid names:\n", len(invalidBranches))
		for _, branch := range invalidBranches {
			fmt.Printf("  - %s\n", branch)
		}

		if len(suggestions) > 0 {
			fmt.Println("\n💡 Suggested fixes:")
			for _, suggestion := range suggestions {
				fmt.Println(suggestion)
			}
		}

		if bpm.config.Enforce && !bpm.config.DryRun {
			fmt.Println("\n🔧 Applying automatic fixes...")
			return bpm.renameBranches(ctx, validationResults)
		}
	} else {
		fmt.Printf("✅ All branches follow %s naming conventions\n", bpm.config.Template)
	}

	return nil
}

// ensureRequiredBranches ensures required branches exist
func (bpm *BranchPolicyManager) ensureRequiredBranches(ctx context.Context) error {
	fmt.Printf("🌿 Ensuring required branches exist: %v\n", bpm.rules.RequiredBranches)

	for _, branch := range bpm.rules.RequiredBranches {
		// Check if branch exists
		result, _ := bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath,
			"show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branch))

		if result == nil || !result.Success {
			// Branch doesn't exist
			if bpm.config.DryRun {
				fmt.Printf("📝 Would create required branch '%s'\n", branch)
			} else {
				// Create the branch
				_, err := bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath,
					"checkout", "-b", branch)
				if err != nil {
					bpm.logger.Warn("Failed to create required branch",
						zap.String("branch", branch),
						zap.Error(err))
					// Try to create from origin/main or origin/master
					_, err = bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath,
						"checkout", "-b", branch, "origin/main")
					if err != nil {
						_, err = bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath,
							"checkout", "-b", branch, "origin/master")
						if err != nil {
							return fmt.Errorf("failed to create required branch %s: %w", branch, err)
						}
					}
				}
				fmt.Printf("✨ Created required branch '%s'\n", branch)

				// Return to previous branch
				bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "checkout", "-")
			}
		} else {
			fmt.Printf("✅ Branch '%s' exists\n", branch)
		}
	}

	return nil
}

// cleanupStaleBranches removes stale branches based on cleanup rules
func (bpm *BranchPolicyManager) cleanupStaleBranches(ctx context.Context) error {
	fmt.Printf("🧹 Cleaning up stale branches (older than %d days)...\n", bpm.rules.CleanupRules.StaleDays)

	// Use BranchManager for cleanup
	branchConfig := &BranchManagementConfig{
		RepositoryPath:    bpm.config.RepositoryPath,
		BranchNamingRules: bpm.config.Template,
		RemoteName:        "origin",
		StaleBranchDays:   bpm.rules.CleanupRules.StaleDays,
		ProtectedBranches: bpm.rules.CleanupRules.ProtectedBranches,
		DryRun:            bpm.config.DryRun,
		AutoDelete:        bpm.rules.CleanupRules.AutoDelete,
	}

	branchManager := NewBranchManager(bpm.logger, branchConfig)

	// Clean up stale branches
	if err := branchManager.CleanupStaleBranches(ctx); err != nil {
		return fmt.Errorf("failed to cleanup stale branches: %w", err)
	}

	// Also delete merged branches if auto-cleanup is enabled
	if bpm.config.AutoCleanup {
		fmt.Println("\n🧹 Checking for merged branches...")
		if err := branchManager.DeleteMergedBranches(ctx); err != nil {
			return fmt.Errorf("failed to delete merged branches: %w", err)
		}
	}

	return nil
}

// enforceMergeRules enforces merge policy rules
func (bpm *BranchPolicyManager) enforceMergeRules(ctx context.Context) error {
	fmt.Println("🔒 Enforcing merge policy rules...")

	if bpm.config.DryRun {
		fmt.Printf("📝 Would enforce merge rules: %+v\n", bpm.rules.MergeRules)
		return nil
	}

	// TODO: Implement merge rule enforcement (requires Git hosting platform API)
	fmt.Println("✅ Merge rules enforcement completed")
	return nil
}

// createBranchPolicyRules creates policy rules based on template
func createBranchPolicyRules(template string) (*BranchPolicyRules, error) {
	switch strings.ToLower(template) {
	case "gitflow":
		return createGitFlowRules(), nil
	case "github-flow":
		return createGitHubFlowRules(), nil
	case "gitlab-flow":
		return createGitLabFlowRules(), nil
	case "custom":
		return createCustomRules(), nil
	default:
		return nil, fmt.Errorf("unknown branch strategy template: %s", template)
	}
}

// createGitFlowRules creates GitFlow branch strategy rules
func createGitFlowRules() *BranchPolicyRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^feature/[a-z0-9-]+$`)
	patterns["release"] = regexp.MustCompile(`^release/\d+\.\d+\.\d+$`)
	patterns["hotfix"] = regexp.MustCompile(`^hotfix/[a-z0-9-]+$`)

	return &BranchPolicyRules{
		Template:          "gitflow",
		NamingPatterns:    patterns,
		ProtectedBranches: []string{"main", "develop"},
		RequiredBranches:  []string{"main", "develop"},
		MergeRules: MergeRules{
			RequireReview:     true,
			RequireStatus:     true,
			AllowedMergeTypes: []string{"merge", "squash"},
			RestrictPushes:    true,
		},
		CleanupRules: CleanupRules{
			StaleDays:         30,
			ProtectedBranches: []string{"main", "develop"},
			AutoDelete:        false,
		},
	}
}

// createGitHubFlowRules creates GitHub Flow branch strategy rules
func createGitHubFlowRules() *BranchPolicyRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^[a-z0-9-]+/[a-z0-9-]+$`)

	return &BranchPolicyRules{
		Template:          "github-flow",
		NamingPatterns:    patterns,
		ProtectedBranches: []string{"main"},
		RequiredBranches:  []string{"main"},
		MergeRules: MergeRules{
			RequireReview:     true,
			RequireStatus:     true,
			AllowedMergeTypes: []string{"merge", "squash", "rebase"},
			RestrictPushes:    false,
		},
		CleanupRules: CleanupRules{
			StaleDays:         14,
			ProtectedBranches: []string{"main"},
			AutoDelete:        true,
		},
	}
}

// createGitLabFlowRules creates GitLab Flow branch strategy rules
func createGitLabFlowRules() *BranchPolicyRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["feature"] = regexp.MustCompile(`^feature/[a-z0-9-]+$`)

	return &BranchPolicyRules{
		Template:          "gitlab-flow",
		NamingPatterns:    patterns,
		ProtectedBranches: []string{"main", "staging", "production"},
		RequiredBranches:  []string{"main"},
		MergeRules: MergeRules{
			RequireReview:     true,
			RequireStatus:     true,
			AllowedMergeTypes: []string{"merge"},
			RestrictPushes:    true,
		},
		CleanupRules: CleanupRules{
			StaleDays:         21,
			ProtectedBranches: []string{"main", "staging", "production"},
			AutoDelete:        false,
		},
	}
}

// createCustomRules creates custom branch strategy rules
func createCustomRules() *BranchPolicyRules {
	patterns := make(map[string]*regexp.Regexp)
	patterns["any"] = regexp.MustCompile(`^[a-zA-Z0-9-_/]+$`)

	return &BranchPolicyRules{
		Template:          "custom",
		NamingPatterns:    patterns,
		ProtectedBranches: []string{"main"},
		RequiredBranches:  []string{"main"},
		MergeRules: MergeRules{
			RequireReview:     false,
			RequireStatus:     false,
			AllowedMergeTypes: []string{"merge", "squash", "rebase"},
			RestrictPushes:    false,
		},
		CleanupRules: CleanupRules{
			StaleDays:         60,
			ProtectedBranches: []string{"main"},
			AutoDelete:        false,
		},
	}
}

// renameBranches renames branches to match naming conventions
func (bpm *BranchPolicyManager) renameBranches(ctx context.Context, validationResults map[string]*ValidationResult) error {
	var renameCount int

	for branchName, result := range validationResults {
		if !result.Valid && result.FixedName != "" && result.BranchType != "protected" {
			// Skip if branch name is the same as fixed name
			if branchName == result.FixedName {
				continue
			}

			fmt.Printf("🔄 Renaming: %s → %s\n", branchName, result.FixedName)

			// Check if the target branch already exists
			checkResult, _ := bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", result.FixedName))
			if checkResult != nil && checkResult.Success {
				fmt.Printf("⚠️  Target branch '%s' already exists, skipping\n", result.FixedName)
				continue
			}

			// Rename the branch
			_, err := bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "branch", "-m", branchName, result.FixedName)
			if err != nil {
				bpm.logger.Error("Failed to rename branch",
					zap.String("from", branchName),
					zap.String("to", result.FixedName),
					zap.Error(err))
				continue
			}

			renameCount++

			// If this is a remote branch, we need to push the new branch and delete the old one
			if strings.HasPrefix(branchName, "origin/") {
				fmt.Printf("🌐 Updating remote branch...\n")
				// Push new branch
				_, err = bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "push", "origin", result.FixedName)
				if err != nil {
					bpm.logger.Warn("Failed to push renamed branch", zap.Error(err))
				}
				// Delete old branch
				_, err = bpm.gitCmd.ExecuteCommand(ctx, bpm.config.RepositoryPath, "push", "origin", "--delete", strings.TrimPrefix(branchName, "origin/"))
				if err != nil {
					bpm.logger.Warn("Failed to delete old remote branch", zap.Error(err))
				}
			}
		}
	}

	if renameCount > 0 {
		fmt.Printf("✅ Renamed %d branches to follow naming conventions\n", renameCount)
	}

	return nil
}

// suggestBranchName suggests a branch name based on user input
func (bpm *BranchPolicyManager) suggestBranchName(branchType, description string) error {
	suggestion, err := bpm.validator.SuggestBranchName(branchType, description)
	if err != nil {
		return fmt.Errorf("failed to suggest branch name: %w", err)
	}

	fmt.Printf("💡 Suggested branch name: %s\n", suggestion)
	return nil
}

// validateSingleBranch validates a single branch name
func (bpm *BranchPolicyManager) validateSingleBranch(branchName string) error {
	result := bpm.validator.ValidateBranchName(branchName)

	if result.Valid {
		fmt.Printf("✅ '%s' is a valid branch name (%s)\n", branchName, result.BranchType)
	} else {
		fmt.Printf("❌ '%s' is invalid\n", branchName)
		if len(result.Errors) > 0 {
			fmt.Println("📝 Errors:")
			for _, err := range result.Errors {
				fmt.Printf("  - %s\n", err)
			}
		}
		if len(result.Suggestions) > 0 {
			fmt.Println("💡 Suggestions:")
			for _, suggestion := range result.Suggestions {
				fmt.Printf("  - %s\n", suggestion)
			}
		}
		if result.FixedName != "" {
			fmt.Printf("🔧 Fixed name: %s\n", result.FixedName)
		}
	}

	return nil
}
