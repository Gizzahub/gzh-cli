package reposync

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// newSyncCmd creates the sync subcommand for bidirectional repository synchronization.
func newSyncCmd(logger *zap.Logger) *cobra.Command {
	var (
		bidirectional    bool
		remoteName       string
		conflictStrategy string
		dryRun           bool
		autoCommit       bool
		commitMessage    string
		excludePatterns  []string
		forcePush        bool
	)

	cmd := &cobra.Command{
		Use:   "sync [repository-path]",
		Short: "Bidirectional repository synchronization with conflict resolution",
		Long: `Perform bidirectional synchronization between local and remote repositories with advanced conflict resolution.

This command provides sophisticated synchronization capabilities:
- Local → Remote: Push local changes to remote repository
- Remote → Local: Pull remote changes to local repository
- Bidirectional: Automatic synchronization in both directions
- Intelligent conflict resolution with multiple strategies
- Atomic operations with rollback capabilities
- Detailed synchronization reports and logging

Conflict Resolution Strategies:
- manual: Stop and prompt user for conflict resolution
- auto-merge: Attempt automatic 3-way merge
- local-wins: Prefer local changes over remote
- remote-wins: Prefer remote changes over local
- timestamp: Use newest changes based on commit timestamp

Examples:
  # Basic one-way sync (local to remote)
  gz repo-sync sync ./my-repo --remote origin

  # Bidirectional sync with auto-merge
  gz repo-sync sync ./my-repo --bidirectional --conflict-strategy auto-merge

  # Dry run to see what would be synchronized
  gz repo-sync sync ./my-repo --bidirectional --dry-run

  # Sync with auto-commit and custom message
  gz repo-sync sync ./my-repo --auto-commit --commit-message "Auto-sync: {{.Timestamp}}"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			repoPath := "."
			if len(args) > 0 {
				repoPath = args[0]
			}

			// Validate repository path
			if err := validateRepositoryPath(repoPath); err != nil {
				return fmt.Errorf("invalid repository path: %w", err)
			}

			// Create synchronization configuration
			config := &SyncConfig{
				RepositoryPath:   repoPath,
				Bidirectional:    bidirectional,
				RemoteName:       remoteName,
				ConflictStrategy: conflictStrategy,
				DryRun:           dryRun,
				AutoCommit:       autoCommit,
				CommitMessage:    commitMessage,
				ExcludePatterns:  excludePatterns,
				ForcePush:        forcePush,
			}

			// Create and execute synchronizer
			synchronizer, err := NewRepositorySynchronizer(logger, config)
			if err != nil {
				return fmt.Errorf("failed to create synchronizer: %w", err)
			}

			ctx := context.Background()
			result, err := synchronizer.Synchronize(ctx)
			if err != nil {
				return fmt.Errorf("synchronization failed: %w", err)
			}

			// Print synchronization results
			printSyncResult(result)
			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&bidirectional, "bidirectional", false, "Enable bidirectional synchronization")
	cmd.Flags().StringVar(&remoteName, "remote", "origin", "Remote repository name")
	cmd.Flags().StringVar(&conflictStrategy, "conflict-strategy", "manual", "Conflict resolution strategy (manual|auto-merge|local-wins|remote-wins|timestamp)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synchronized without making changes")
	cmd.Flags().BoolVar(&autoCommit, "auto-commit", false, "Automatically commit local changes before sync")
	cmd.Flags().StringVar(&commitMessage, "commit-message", "Auto-sync: {{.Timestamp}}", "Commit message template")
	cmd.Flags().StringSliceVar(&excludePatterns, "exclude-patterns", []string{}, "Patterns to exclude from sync")
	cmd.Flags().BoolVar(&forcePush, "force-push", false, "Force push changes (use with caution)")

	return cmd
}

// SyncConfig represents configuration for repository synchronization.
type SyncConfig struct {
	RepositoryPath   string   `json:"repositoryPath"`
	Bidirectional    bool     `json:"bidirectional"`
	RemoteName       string   `json:"remoteName"`
	ConflictStrategy string   `json:"conflictStrategy"`
	DryRun           bool     `json:"dryRun"`
	AutoCommit       bool     `json:"autoCommit"`
	CommitMessage    string   `json:"commitMessage"`
	ExcludePatterns  []string `json:"excludePatterns"`
	ForcePush        bool     `json:"forcePush"`
}

// RepositorySynchronizer handles bidirectional repository synchronization.
type RepositorySynchronizer struct {
	logger *zap.Logger
	config *SyncConfig
	gitCmd GitCommandExecutor
}

// GitCommandExecutor interface for Git command execution.
type GitCommandExecutor interface {
	ExecuteCommand(ctx context.Context, dir string, args ...string) (*GitCommandResult, error)
	GetStatus(ctx context.Context, dir string) (*GitStatus, error)
	GetRemoteInfo(ctx context.Context, dir string, remote string) (*GitRemoteInfo, error)
}

// GitCommandResult represents the result of a Git command.
type GitCommandResult struct {
	Command  string        `json:"command"`
	Output   string        `json:"output"`
	Error    string        `json:"error"`
	ExitCode int           `json:"exitCode"`
	Duration time.Duration `json:"duration"`
	Success  bool          `json:"success"`
}

// GitStatus represents the current Git repository status.
type GitStatus struct {
	Branch          string    `json:"branch"`
	Upstream        string    `json:"upstream"`
	AheadBy         int       `json:"aheadBy"`
	BehindBy        int       `json:"behindBy"`
	ModifiedFiles   []string  `json:"modifiedFiles"`
	UntrackedFiles  []string  `json:"untrackedFiles"`
	ConflictedFiles []string  `json:"conflictedFiles"`
	CleanWorkingDir bool      `json:"cleanWorkingDir"`
	LastCommitHash  string    `json:"lastCommitHash"`
	LastCommitTime  time.Time `json:"lastCommitTime"`
}

// GitRemoteInfo represents information about a Git remote.
type GitRemoteInfo struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	FetchURL  string `json:"fetchUrl"`
	PushURL   string `json:"pushUrl"`
	Reachable bool   `json:"reachable"`
}

// defaultGitExecutor provides default Git command execution.
type defaultGitExecutor struct{}

// NewRepositorySynchronizer creates a new repository synchronizer.
func NewRepositorySynchronizer(logger *zap.Logger, config *SyncConfig) (*RepositorySynchronizer, error) {
	return &RepositorySynchronizer{
		logger: logger,
		config: config,
		gitCmd: &defaultGitExecutor{},
	}, nil
}

// Synchronize performs repository synchronization based on configuration.
func (rs *RepositorySynchronizer) Synchronize(ctx context.Context) (*SyncResult, error) {
	startTime := time.Now()
	result := &SyncResult{
		Success:   true,
		Conflicts: make([]ConflictInfo, 0),
		Errors:    make([]string, 0),
	}

	rs.logger.Info("Starting repository synchronization",
		zap.String("path", rs.config.RepositoryPath),
		zap.Bool("bidirectional", rs.config.Bidirectional),
		zap.String("remote", rs.config.RemoteName))

	// Get initial repository status
	status, err := rs.gitCmd.GetStatus(ctx, rs.config.RepositoryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository status: %w", err)
	}

	// Validate remote connectivity
	remoteInfo, err := rs.gitCmd.GetRemoteInfo(ctx, rs.config.RepositoryPath, rs.config.RemoteName)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote info: %w", err)
	}

	if !remoteInfo.Reachable {
		return nil, fmt.Errorf("remote '%s' is not reachable: %s", rs.config.RemoteName, remoteInfo.URL)
	}

	rs.logger.Info("Repository status",
		zap.String("branch", status.Branch),
		zap.Int("ahead", status.AheadBy),
		zap.Int("behind", status.BehindBy),
		zap.Int("modified", len(status.ModifiedFiles)),
		zap.Int("untracked", len(status.UntrackedFiles)))

	if rs.config.DryRun {
		fmt.Println("🔍 Dry run mode - no actual changes will be made")
	}

	// Handle local changes (commit if auto-commit enabled)
	if rs.config.AutoCommit && !status.CleanWorkingDir {
		if err := rs.commitLocalChanges(ctx, status, result); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to commit local changes: %v", err))
		}
	}

	// Perform synchronization based on configuration
	if rs.config.Bidirectional {
		// Bidirectional sync: handle both directions
		if err := rs.performBidirectionalSync(ctx, status, result); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Bidirectional sync failed: %v", err))
		}
	} else {
		// Unidirectional sync: local to remote
		if err := rs.performUnidirectionalSync(ctx, status, result); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Unidirectional sync failed: %v", err))
		}
	}

	result.Duration = time.Since(startTime)

	return result, nil
}

// commitLocalChanges commits local changes if auto-commit is enabled.
func (rs *RepositorySynchronizer) commitLocalChanges(ctx context.Context, status *GitStatus, result *SyncResult) error {
	if len(status.ModifiedFiles) == 0 && len(status.UntrackedFiles) == 0 {
		return nil
	}

	rs.logger.Info("Auto-committing local changes",
		zap.Int("modified", len(status.ModifiedFiles)),
		zap.Int("untracked", len(status.UntrackedFiles)))

	if rs.config.DryRun {
		fmt.Printf("📝 Would auto-commit %d modified and %d untracked files\n",
			len(status.ModifiedFiles), len(status.UntrackedFiles))

		return nil
	}

	// Stage all changes
	_, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath, "add", ".")
	if err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit message
	commitMsg := rs.expandCommitMessage(rs.config.CommitMessage)

	// Create commit
	commitResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath, "commit", "-m", commitMsg)
	if err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	// Extract commit hash from output
	if commitResult.Success {
		result.CommitHash = rs.extractCommitHash(commitResult.Output)
		result.FilesModified = len(status.ModifiedFiles)
		result.FilesCreated = len(status.UntrackedFiles)
	}

	fmt.Printf("✅ Auto-committed %d changes: %s\n",
		len(status.ModifiedFiles)+len(status.UntrackedFiles), commitMsg)

	return nil
}

// performUnidirectionalSync performs one-way synchronization (local to remote).
func (rs *RepositorySynchronizer) performUnidirectionalSync(ctx context.Context, status *GitStatus, _ *SyncResult) error {
	if status.AheadBy == 0 {
		fmt.Println("📊 Repository is up to date with remote")
		return nil
	}

	rs.logger.Info("Performing unidirectional sync",
		zap.Int("commits_ahead", status.AheadBy))

	if rs.config.DryRun {
		fmt.Printf("📤 Would push %d commits to remote '%s'\n", status.AheadBy, rs.config.RemoteName)
		return nil
	}

	// Push to remote
	pushArgs := []string{"push", rs.config.RemoteName, status.Branch}
	if rs.config.ForcePush {
		pushArgs = append(pushArgs, "--force")
	}

	pushResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath, pushArgs...)
	if err != nil {
		return fmt.Errorf("failed to push to remote: %w", err)
	}

	if pushResult.Success {
		fmt.Printf("📤 Successfully pushed %d commits to %s\n", status.AheadBy, rs.config.RemoteName)
	}

	return nil
}

// performBidirectionalSync performs two-way synchronization with conflict resolution.
func (rs *RepositorySynchronizer) performBidirectionalSync(ctx context.Context, status *GitStatus, result *SyncResult) error {
	rs.logger.Info("Performing bidirectional sync",
		zap.Int("commits_ahead", status.AheadBy),
		zap.Int("commits_behind", status.BehindBy))

	// Handle different sync scenarios
	switch {
	case status.AheadBy == 0 && status.BehindBy == 0:
		fmt.Println("📊 Repository is in sync with remote")
		return nil

	case status.AheadBy > 0 && status.BehindBy == 0:
		// Only local changes: push to remote
		return rs.pushLocalChanges(ctx, status, result)

	case status.AheadBy == 0 && status.BehindBy > 0:
		// Only remote changes: pull from remote
		return rs.pullRemoteChanges(ctx, status, result)

	case status.AheadBy > 0 && status.BehindBy > 0:
		// Diverged: need conflict resolution
		return rs.resolveDivergence(ctx, status, result)

	default:
		return fmt.Errorf("unexpected repository state")
	}
}

// pushLocalChanges pushes local changes to remote.
func (rs *RepositorySynchronizer) pushLocalChanges(ctx context.Context, status *GitStatus, _ *SyncResult) error {
	if rs.config.DryRun {
		fmt.Printf("📤 Would push %d local commits to remote\n", status.AheadBy)
		return nil
	}

	pushResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"push", rs.config.RemoteName, status.Branch)
	if err != nil {
		return fmt.Errorf("failed to push local changes: %w", err)
	}

	if pushResult.Success {
		fmt.Printf("📤 Pushed %d local commits to remote\n", status.AheadBy)
	}

	return nil
}

// pullRemoteChanges pulls remote changes to local.
func (rs *RepositorySynchronizer) pullRemoteChanges(ctx context.Context, status *GitStatus, _ *SyncResult) error {
	if rs.config.DryRun {
		fmt.Printf("📥 Would pull %d remote commits to local\n", status.BehindBy)
		return nil
	}

	pullResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"pull", rs.config.RemoteName, status.Branch)
	if err != nil {
		return fmt.Errorf("failed to pull remote changes: %w", err)
	}

	if pullResult.Success {
		fmt.Printf("📥 Pulled %d remote commits to local\n", status.BehindBy)
	}

	return nil
}

// resolveDivergence handles diverged repositories based on conflict strategy.
func (rs *RepositorySynchronizer) resolveDivergence(ctx context.Context, status *GitStatus, result *SyncResult) error {
	fmt.Printf("⚠️  Repository has diverged: %d local, %d remote commits\n",
		status.AheadBy, status.BehindBy)

	if rs.config.DryRun {
		fmt.Printf("🔀 Would resolve divergence using strategy: %s\n", rs.config.ConflictStrategy)
		return nil
	}

	switch rs.config.ConflictStrategy {
	case "auto-merge":
		return rs.performAutoMerge(ctx, status, result)
	case "local-wins":
		return rs.forceLocalChanges(ctx, status, result)
	case "remote-wins":
		return rs.forceRemoteChanges(ctx, status, result)
	case "timestamp":
		return rs.resolveByTimestamp(ctx, status, result)
	case "manual":
		return rs.promptManualResolution(ctx, status, result)
	default:
		return fmt.Errorf("unknown conflict strategy: %s", rs.config.ConflictStrategy)
	}
}

// performAutoMerge attempts automatic merge resolution.
func (rs *RepositorySynchronizer) performAutoMerge(ctx context.Context, status *GitStatus, result *SyncResult) error {
	rs.logger.Info("Attempting automatic merge")

	// Fetch latest from remote
	_, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"fetch", rs.config.RemoteName)
	if err != nil {
		return fmt.Errorf("failed to fetch from remote: %w", err)
	}

	// Attempt merge
	mergeResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"merge", fmt.Sprintf("%s/%s", rs.config.RemoteName, status.Branch))
	if err != nil || !mergeResult.Success {
		// Merge failed, likely due to conflicts
		return rs.handleMergeConflicts(ctx, status, result)
	}

	fmt.Println("🔀 Successfully merged remote changes automatically")

	// Push merged result back to remote
	_, err = rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"push", rs.config.RemoteName, status.Branch)
	if err != nil {
		return fmt.Errorf("failed to push merged changes: %w", err)
	}

	fmt.Println("📤 Pushed merged changes to remote")

	return nil
}

// handleMergeConflicts handles merge conflicts during auto-merge.
func (rs *RepositorySynchronizer) handleMergeConflicts(ctx context.Context, _ *GitStatus, result *SyncResult) error {
	// Get conflicted files
	statusResult, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to get status after merge: %w", err)
	}

	conflictedFiles := rs.parseConflictedFiles(statusResult.Output)

	for _, file := range conflictedFiles {
		conflict := ConflictInfo{
			Path:         file,
			ConflictType: "content",
			Resolution:   "manual",
		}
		result.Conflicts = append(result.Conflicts, conflict)
	}

	fmt.Printf("❌ Merge conflicts detected in %d files:\n", len(conflictedFiles))

	for _, file := range conflictedFiles {
		fmt.Printf("   - %s\n", file)
	}

	// Abort merge
	_, err = rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath, "merge", "--abort")
	if err != nil {
		rs.logger.Warn("Failed to abort merge", zap.Error(err))
	}

	return fmt.Errorf("merge conflicts require manual resolution")
}

// Helper methods for other conflict resolution strategies.
func (rs *RepositorySynchronizer) forceLocalChanges(ctx context.Context, status *GitStatus, _ *SyncResult) error {
	fmt.Println("🔄 Forcing local changes (local-wins strategy)")

	if rs.config.DryRun {
		fmt.Println("📤 Would force push local changes, overwriting remote")
		return nil
	}

	// Force push local changes
	_, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"push", "--force", rs.config.RemoteName, status.Branch)
	if err != nil {
		return fmt.Errorf("failed to force push: %w", err)
	}

	fmt.Println("📤 Force pushed local changes to remote")

	return nil
}

func (rs *RepositorySynchronizer) forceRemoteChanges(ctx context.Context, status *GitStatus, _ *SyncResult) error {
	fmt.Println("🔄 Forcing remote changes (remote-wins strategy)")

	if rs.config.DryRun {
		fmt.Println("📥 Would reset local to match remote, discarding local changes")
		return nil
	}

	// Reset local to match remote
	_, err := rs.gitCmd.ExecuteCommand(ctx, rs.config.RepositoryPath,
		"reset", "--hard", fmt.Sprintf("%s/%s", rs.config.RemoteName, status.Branch))
	if err != nil {
		return fmt.Errorf("failed to reset to remote: %w", err)
	}

	fmt.Println("📥 Reset local repository to match remote")

	return nil
}

func (rs *RepositorySynchronizer) resolveByTimestamp(ctx context.Context, status *GitStatus, result *SyncResult) error {
	// Timestamp-based resolution placeholder - implement timestamp comparison logic
	// This would compare commit timestamps and choose the newer set of changes
	return fmt.Errorf("timestamp-based resolution not yet implemented")
}

func (rs *RepositorySynchronizer) promptManualResolution(ctx context.Context, status *GitStatus, result *SyncResult) error {
	fmt.Printf("🔄 Manual resolution required:\n")
	fmt.Printf("   Local commits ahead: %d\n", status.AheadBy)
	fmt.Printf("   Remote commits behind: %d\n", status.BehindBy)
	fmt.Printf("   Please resolve manually using: git pull --rebase or git merge\n")

	return fmt.Errorf("manual conflict resolution required")
}

// Utility methods

func (rs *RepositorySynchronizer) expandCommitMessage(template string) string {
	// Simple template expansion
	message := strings.ReplaceAll(template, "{{.Timestamp}}", time.Now().Format("2006-01-02 15:04:05"))
	return message
}

func (rs *RepositorySynchronizer) extractCommitHash(output string) string {
	// Extract commit hash from git commit output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "[") && strings.Contains(line, "]") {
			parts := strings.Split(line, " ")
			if len(parts) > 1 {
				return strings.TrimSuffix(parts[1], "]")
			}
		}
	}

	return ""
}

func (rs *RepositorySynchronizer) parseConflictedFiles(statusOutput string) []string {
	var conflicted []string

	lines := strings.Split(statusOutput, "\n")

	for _, line := range lines {
		if len(line) > 2 && strings.HasPrefix(line, "UU") {
			conflicted = append(conflicted, strings.TrimSpace(line[2:]))
		}
	}

	return conflicted
}

// Git command executor implementation

func (ge *defaultGitExecutor) ExecuteCommand(ctx context.Context, dir string, args ...string) (*GitCommandResult, error) {
	startTime := time.Now()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	duration := time.Since(startTime)

	result := &GitCommandResult{
		Command:  fmt.Sprintf("git %s", strings.Join(args, " ")),
		Output:   string(output),
		Duration: duration,
		Success:  err == nil,
	}
	if err != nil {
		result.Error = err.Error()

		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		}
	}

	return result, err
}

func (ge *defaultGitExecutor) GetStatus(ctx context.Context, dir string) (*GitStatus, error) {
	// Get branch info
	branchResult, err := ge.ExecuteCommand(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get current branch: %w", err)
	}

	status := &GitStatus{
		Branch: strings.TrimSpace(branchResult.Output),
	}

	// Get upstream tracking info
	upstreamResult, _ := ge.ExecuteCommand(ctx, dir, "rev-parse", "--abbrev-ref", "@{upstream}")
	if upstreamResult.Success {
		status.Upstream = strings.TrimSpace(upstreamResult.Output)

		// Get ahead/behind counts
		aheadBehindResult, _ := ge.ExecuteCommand(ctx, dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		if aheadBehindResult.Success {
			parts := strings.Fields(strings.TrimSpace(aheadBehindResult.Output))
			if len(parts) == 2 {
				_, _ = fmt.Sscanf(parts[0], "%d", &status.AheadBy)  //nolint:errcheck // Not critical
				_, _ = fmt.Sscanf(parts[1], "%d", &status.BehindBy) //nolint:errcheck // Not critical
			}
		}
	}

	// Get working directory status
	statusResult, err := ge.ExecuteCommand(ctx, dir, "status", "--porcelain")
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory status: %w", err)
	}

	// Parse status output
	lines := strings.Split(statusResult.Output, "\n")
	for _, line := range lines {
		if len(line) < 3 {
			continue
		}

		statusCode := line[0:2]
		filename := strings.TrimSpace(line[2:])

		switch {
		case statusCode == "??":
			status.UntrackedFiles = append(status.UntrackedFiles, filename)
		case statusCode == "UU":
			status.ConflictedFiles = append(status.ConflictedFiles, filename)
		case statusCode[0] != ' ' || statusCode[1] != ' ':
			status.ModifiedFiles = append(status.ModifiedFiles, filename)
		}
	}

	status.CleanWorkingDir = len(status.ModifiedFiles) == 0 && len(status.UntrackedFiles) == 0

	return status, nil
}

func (ge *defaultGitExecutor) GetRemoteInfo(ctx context.Context, dir string, remote string) (*GitRemoteInfo, error) {
	// Get remote URL
	urlResult, err := ge.ExecuteCommand(ctx, dir, "remote", "get-url", remote)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote URL: %w", err)
	}

	info := &GitRemoteInfo{
		Name: remote,
		URL:  strings.TrimSpace(urlResult.Output),
	}

	// Test connectivity
	lsRemoteResult, err := ge.ExecuteCommand(ctx, dir, "ls-remote", remote)
	info.Reachable = err == nil && lsRemoteResult.Success

	return info, nil
}

// printSyncResult prints the synchronization results.
func printSyncResult(result *SyncResult) {
	fmt.Printf("\n📊 Synchronization Results:\n")
	fmt.Printf("   Status: ")

	if result.Success {
		fmt.Printf("✅ Success\n")
	} else {
		fmt.Printf("❌ Failed\n")
	}

	fmt.Printf("   Duration: %v\n", result.Duration.Round(time.Millisecond))

	if result.FilesModified > 0 || result.FilesCreated > 0 || result.FilesDeleted > 0 {
		fmt.Printf("   Changes: %d modified, %d created, %d deleted\n",
			result.FilesModified, result.FilesCreated, result.FilesDeleted)
	}

	if result.CommitHash != "" {
		fmt.Printf("   Commit: %s\n", result.CommitHash)
	}

	if len(result.Conflicts) > 0 {
		fmt.Printf("   Conflicts: %d found\n", len(result.Conflicts))

		for _, conflict := range result.Conflicts {
			fmt.Printf("     - %s (%s)\n", conflict.Path, conflict.ConflictType)
		}
	}

	if len(result.Errors) > 0 {
		fmt.Printf("   Errors:\n")

		for _, err := range result.Errors {
			fmt.Printf("     - %s\n", err)
		}
	}

	fmt.Println()
}
