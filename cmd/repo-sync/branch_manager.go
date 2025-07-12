package reposync

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// BranchManager handles automatic branch creation and deletion
type BranchManager struct {
	logger    *zap.Logger
	config    *BranchManagementConfig
	gitCmd    GitCommandExecutor
	validator *BranchValidator
}

// BranchManagementConfig represents branch management configuration
type BranchManagementConfig struct {
	RepositoryPath    string   `json:"repository_path"`
	AutoCreate        bool     `json:"auto_create"`
	AutoDelete        bool     `json:"auto_delete"`
	IssueTrackerType  string   `json:"issue_tracker_type"` // github, gitlab, jira
	IssuePattern      string   `json:"issue_pattern"`
	BranchTemplate    string   `json:"branch_template"`
	DeleteAfterMerge  bool     `json:"delete_after_merge"`
	StaleBranchDays   int      `json:"stale_branch_days"`
	ProtectedBranches []string `json:"protected_branches"`
	BranchNamingRules string   `json:"branch_naming_rules"` // gitflow, github-flow, etc.
	RemoteName        string   `json:"remote_name"`
	DryRun            bool     `json:"dry_run"`
}

// IssueInfo represents information about an issue
type IssueInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Type        string   `json:"type"` // feature, bug, enhancement, etc.
	Description string   `json:"description"`
	Labels      []string `json:"labels"`
}

// BranchInfo represents information about a branch
type BranchInfo struct {
	Name           string    `json:"name"`
	IsLocal        bool      `json:"is_local"`
	IsRemote       bool      `json:"is_remote"`
	LastCommitDate time.Time `json:"last_commit_date"`
	IsMerged       bool      `json:"is_merged"`
	MergedInto     string    `json:"merged_into,omitempty"`
	AuthorEmail    string    `json:"author_email"`
}

// NewBranchManager creates a new branch manager
func NewBranchManager(logger *zap.Logger, config *BranchManagementConfig) *BranchManager {
	validator := NewBranchValidator(logger, config.BranchNamingRules)

	return &BranchManager{
		logger:    logger,
		config:    config,
		gitCmd:    &defaultGitExecutor{},
		validator: validator,
	}
}

// CreateBranchFromIssue creates a new branch based on issue information
func (bm *BranchManager) CreateBranchFromIssue(ctx context.Context, issue *IssueInfo) error {
	bm.logger.Info("Creating branch from issue",
		zap.String("issue_id", issue.ID),
		zap.String("issue_title", issue.Title))

	// Generate branch name
	branchName, err := bm.generateBranchName(issue)
	if err != nil {
		return fmt.Errorf("failed to generate branch name: %w", err)
	}

	fmt.Printf("🌿 Creating branch for issue #%s: %s\n", issue.ID, branchName)

	// Validate branch name
	validationResult := bm.validator.ValidateBranchName(branchName)
	if !validationResult.Valid {
		if validationResult.FixedName != "" {
			fmt.Printf("📝 Using corrected branch name: %s\n", validationResult.FixedName)
			branchName = validationResult.FixedName
		} else {
			return fmt.Errorf("invalid branch name: %v", validationResult.Errors)
		}
	}

	// Check if branch already exists
	exists, err := bm.branchExists(ctx, branchName)
	if err != nil {
		return fmt.Errorf("failed to check branch existence: %w", err)
	}
	if exists {
		return fmt.Errorf("branch already exists: %s", branchName)
	}

	if bm.config.DryRun {
		fmt.Printf("🔍 [DRY RUN] Would create branch: %s\n", branchName)
		return nil
	}

	// Get current branch
	currentBranch, err := bm.getCurrentBranch(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Create branch
	baseBranch := bm.getBaseBranch(issue.Type)
	_, err = bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath, "checkout", "-b", branchName, baseBranch)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	fmt.Printf("✅ Created branch: %s (based on %s)\n", branchName, baseBranch)

	// Push to remote if configured
	if bm.config.RemoteName != "" {
		_, err = bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath, "push", "-u", bm.config.RemoteName, branchName)
		if err != nil {
			bm.logger.Warn("Failed to push branch to remote", zap.Error(err))
			// Restore original branch
			bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath, "checkout", currentBranch)
			return fmt.Errorf("failed to push branch: %w", err)
		}
		fmt.Printf("🌐 Pushed branch to remote: %s\n", bm.config.RemoteName)
	}

	return nil
}

// DeleteMergedBranches deletes branches that have been merged
func (bm *BranchManager) DeleteMergedBranches(ctx context.Context) error {
	fmt.Println("🧹 Checking for merged branches to delete...")

	branches, err := bm.listBranches(ctx)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	var deletedCount int
	for _, branch := range branches {
		if bm.isProtectedBranch(branch.Name) {
			continue
		}

		if branch.IsMerged && branch.MergedInto != "" {
			if bm.config.DryRun {
				fmt.Printf("🔍 [DRY RUN] Would delete merged branch: %s (merged into %s)\n", branch.Name, branch.MergedInto)
			} else {
				if err := bm.deleteBranch(ctx, branch); err != nil {
					bm.logger.Warn("Failed to delete branch",
						zap.String("branch", branch.Name),
						zap.Error(err))
					continue
				}
				deletedCount++
				fmt.Printf("🗑️  Deleted merged branch: %s\n", branch.Name)
			}
		}
	}

	if deletedCount > 0 {
		fmt.Printf("✅ Deleted %d merged branches\n", deletedCount)
	} else {
		fmt.Println("✅ No merged branches to delete")
	}

	return nil
}

// CleanupStaleBranches removes branches that haven't been updated recently
func (bm *BranchManager) CleanupStaleBranches(ctx context.Context) error {
	fmt.Printf("🧹 Checking for stale branches (older than %d days)...\n", bm.config.StaleBranchDays)

	branches, err := bm.listBranches(ctx)
	if err != nil {
		return fmt.Errorf("failed to list branches: %w", err)
	}

	staleDate := time.Now().AddDate(0, 0, -bm.config.StaleBranchDays)
	var cleanedCount int

	for _, branch := range branches {
		if bm.isProtectedBranch(branch.Name) {
			continue
		}

		if branch.LastCommitDate.Before(staleDate) {
			if bm.config.DryRun {
				fmt.Printf("🔍 [DRY RUN] Would delete stale branch: %s (last commit: %s)\n",
					branch.Name, branch.LastCommitDate.Format("2006-01-02"))
			} else {
				if err := bm.deleteBranch(ctx, branch); err != nil {
					bm.logger.Warn("Failed to delete stale branch",
						zap.String("branch", branch.Name),
						zap.Error(err))
					continue
				}
				cleanedCount++
				fmt.Printf("🗑️  Deleted stale branch: %s (last commit: %s)\n",
					branch.Name, branch.LastCommitDate.Format("2006-01-02"))
			}
		}
	}

	if cleanedCount > 0 {
		fmt.Printf("✅ Cleaned up %d stale branches\n", cleanedCount)
	} else {
		fmt.Println("✅ No stale branches to clean up")
	}

	return nil
}

// Helper methods

func (bm *BranchManager) generateBranchName(issue *IssueInfo) (string, error) {
	// Determine branch type from issue
	branchType := bm.determineBranchType(issue)

	// Clean issue title for branch name
	cleanTitle := bm.cleanIssueTitle(issue.Title)

	// Use template if provided
	if bm.config.BranchTemplate != "" {
		return bm.expandTemplate(bm.config.BranchTemplate, issue, branchType, cleanTitle), nil
	}

	// Generate using branch validator
	return bm.validator.SuggestBranchName(branchType, fmt.Sprintf("%s-%s", issue.ID, cleanTitle))
}

func (bm *BranchManager) determineBranchType(issue *IssueInfo) string {
	// Check issue type
	switch strings.ToLower(issue.Type) {
	case "bug", "bugfix", "fix":
		return "bugfix"
	case "feature", "enhancement", "story":
		return "feature"
	case "hotfix", "critical":
		return "hotfix"
	}

	// Check labels
	for _, label := range issue.Labels {
		labelLower := strings.ToLower(label)
		if strings.Contains(labelLower, "bug") {
			return "bugfix"
		}
		if strings.Contains(labelLower, "feature") || strings.Contains(labelLower, "enhancement") {
			return "feature"
		}
		if strings.Contains(labelLower, "hotfix") || strings.Contains(labelLower, "critical") {
			return "hotfix"
		}
	}

	// Default to feature
	return "feature"
}

func (bm *BranchManager) cleanIssueTitle(title string) string {
	// Remove special characters and convert to lowercase
	clean := strings.ToLower(title)
	clean = regexp.MustCompile(`[^a-z0-9\s-]+`).ReplaceAllString(clean, "")
	clean = regexp.MustCompile(`\s+`).ReplaceAllString(clean, "-")
	clean = strings.Trim(clean, "-")

	// Limit length
	if len(clean) > 40 {
		clean = clean[:40]
		if idx := strings.LastIndex(clean, "-"); idx > 30 {
			clean = clean[:idx]
		}
	}

	return clean
}

func (bm *BranchManager) expandTemplate(template string, issue *IssueInfo, branchType, cleanTitle string) string {
	result := template
	result = strings.ReplaceAll(result, "{{.IssueID}}", issue.ID)
	result = strings.ReplaceAll(result, "{{.IssueTitle}}", cleanTitle)
	result = strings.ReplaceAll(result, "{{.BranchType}}", branchType)
	result = strings.ReplaceAll(result, "{{.Timestamp}}", time.Now().Format("20060102"))
	return result
}

func (bm *BranchManager) getBaseBranch(issueType string) string {
	// Determine base branch based on issue type and branch strategy
	switch bm.config.BranchNamingRules {
	case "gitflow":
		if issueType == "hotfix" {
			return "main"
		}
		return "develop"
	default:
		return "main"
	}
}

func (bm *BranchManager) branchExists(ctx context.Context, branchName string) (bool, error) {
	result, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath,
		"show-ref", "--verify", "--quiet", fmt.Sprintf("refs/heads/%s", branchName))
	if err != nil {
		// Command returns non-zero exit code if branch doesn't exist
		return false, nil
	}

	return result.Success, nil
}

func (bm *BranchManager) getCurrentBranch(ctx context.Context) (string, error) {
	result, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(result.Output), nil
}

func (bm *BranchManager) listBranches(ctx context.Context) ([]*BranchInfo, error) {
	// Get all branches with their last commit dates
	result, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath,
		"for-each-ref", "--format=%(refname:short)|%(committerdate:iso8601)|%(authoremail)",
		"refs/heads/")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branches []*BranchInfo
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) != 3 {
			continue
		}

		branchName := parts[0]
		commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", parts[1])
		if err != nil {
			bm.logger.Warn("Failed to parse commit date",
				zap.String("branch", branchName),
				zap.String("date", parts[1]))
			continue
		}

		branch := &BranchInfo{
			Name:           branchName,
			LastCommitDate: commitDate,
			AuthorEmail:    parts[2],
			IsLocal:        true,
		}

		// Check if merged
		merged, mergedInto := bm.checkIfMerged(ctx, branchName)
		branch.IsMerged = merged
		branch.MergedInto = mergedInto

		branches = append(branches, branch)
	}

	return branches, nil
}

func (bm *BranchManager) checkIfMerged(ctx context.Context, branchName string) (bool, string) {
	// Check against main branches
	mainBranches := []string{"main", "master", "develop"}

	for _, mainBranch := range mainBranches {
		// Check if main branch exists
		if exists, _ := bm.branchExists(ctx, mainBranch); !exists {
			continue
		}

		// Check if branch is merged into main branch
		result, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath,
			"branch", "--merged", mainBranch)
		if err != nil {
			continue
		}

		if strings.Contains(result.Output, branchName) {
			return true, mainBranch
		}
	}

	return false, ""
}

func (bm *BranchManager) deleteBranch(ctx context.Context, branch *BranchInfo) error {
	// Delete local branch
	if branch.IsLocal {
		_, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath, "branch", "-D", branch.Name)
		if err != nil {
			return fmt.Errorf("failed to delete local branch: %w", err)
		}
	}

	// Delete remote branch if exists
	if branch.IsRemote && bm.config.RemoteName != "" {
		_, err := bm.gitCmd.ExecuteCommand(ctx, bm.config.RepositoryPath,
			"push", bm.config.RemoteName, "--delete", branch.Name)
		if err != nil {
			bm.logger.Warn("Failed to delete remote branch",
				zap.String("branch", branch.Name),
				zap.Error(err))
		}
	}

	return nil
}

func (bm *BranchManager) isProtectedBranch(branchName string) bool {
	// Default protected branches
	defaultProtected := []string{"main", "master", "develop", "staging", "production"}

	for _, protected := range defaultProtected {
		if branchName == protected {
			return true
		}
	}

	// Check configured protected branches
	for _, protected := range bm.config.ProtectedBranches {
		if branchName == protected {
			return true
		}
	}

	return false
}

// ParseIssueFromString parses issue information from a string (e.g., "123: Add user authentication")
func ParseIssueFromString(input string) (*IssueInfo, error) {
	// Try to parse issue ID and title
	pattern := regexp.MustCompile(`^(\w+[-]?\d+):\s*(.+)$`)
	matches := pattern.FindStringSubmatch(input)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid issue format, expected 'ID: Title'")
	}

	return &IssueInfo{
		ID:    matches[1],
		Title: matches[2],
		Type:  "feature", // Default type
	}, nil
}
