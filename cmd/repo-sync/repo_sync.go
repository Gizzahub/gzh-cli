package reposync

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// NewRepoSyncCmd creates the repository synchronization command
func NewRepoSyncCmd(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repo-sync",
		Short: "Advanced repository synchronization and management",
		Long: `Advanced repository synchronization and management with real-time file system monitoring.

This command provides sophisticated repository management capabilities:
- Real-time file system change detection using fsnotify
- Efficient batch processing of file changes
- Bidirectional synchronization (local ↔ remote)
- Automatic conflict resolution with configurable strategies
- Branch policy enforcement and automation
- Code quality metrics collection and monitoring

Examples:
  # Start real-time repository monitoring
  gz repo-sync watch ./my-repo
  
  # Enable bidirectional sync with remote
  gz repo-sync sync --bidirectional --remote origin
  
  # Monitor multiple repositories
  gz repo-sync watch-multi ./repo1 ./repo2 ./repo3
  
  # Setup branch policies
  gz repo-sync branch-policy --enforce gitflow --auto-cleanup
  
  # Analyze code quality metrics
  gz repo-sync quality-check --threshold 80`,
		SilenceUsage: true,
	}

	// Create logger
	logger, _ := zap.NewProduction()

	// Add subcommands
	cmd.AddCommand(newWatchCmd(logger))
	cmd.AddCommand(newSyncCmd(logger))
	cmd.AddCommand(newWatchMultiCmd(logger))
	cmd.AddCommand(newBranchPolicyCmd(logger))
	cmd.AddCommand(newQualityCheckCmd(logger))
	cmd.AddCommand(newTrendAnalyzeCmd(logger))

	return cmd
}

// getDefaultConfigDir returns the default configuration directory
func getDefaultConfigDir() string {
	if configDir := os.Getenv("GZH_CONFIG_DIR"); configDir != "" {
		return configDir
	}

	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".config", "gzh-manager")
}

// validateRepositoryPath validates that the path is a valid Git repository
func validateRepositoryPath(path string) error {
	// Check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}

	// Check if it's a Git repository
	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return fmt.Errorf("not a Git repository: %s (no .git directory found)", path)
	}

	return nil
}

// RepoSyncConfig represents the configuration for repository synchronization
type RepoSyncConfig struct {
	RepositoryPath   string        `json:"repository_path"`
	WatchPatterns    []string      `json:"watch_patterns"`
	IgnorePatterns   []string      `json:"ignore_patterns"`
	BatchSize        int           `json:"batch_size"`
	BatchTimeout     time.Duration `json:"batch_timeout"`
	Bidirectional    bool          `json:"bidirectional"`
	ConflictStrategy string        `json:"conflict_strategy"`
	RemoteName       string        `json:"remote_name"`
	AutoCommit       bool          `json:"auto_commit"`
	CommitMessage    string        `json:"commit_message"`
}

// DefaultRepoSyncConfig returns the default configuration
func DefaultRepoSyncConfig() *RepoSyncConfig {
	return &RepoSyncConfig{
		WatchPatterns:    []string{"**/*.go", "**/*.md", "**/*.yaml", "**/*.yml", "**/*.json"},
		IgnorePatterns:   []string{".git/**", "vendor/**", "node_modules/**", "*.tmp", "*.log"},
		BatchSize:        100,
		BatchTimeout:     5 * time.Second,
		Bidirectional:    false,
		ConflictStrategy: "manual",
		RemoteName:       "origin",
		AutoCommit:       false,
		CommitMessage:    "Auto-sync: {{.Timestamp}}",
	}
}

// FileChangeEvent represents a file system change event
type FileChangeEvent struct {
	Path        string    `json:"path"`
	Operation   string    `json:"operation"` // create, write, remove, rename, chmod
	IsDirectory bool      `json:"is_directory"`
	Timestamp   time.Time `json:"timestamp"`
	Size        int64     `json:"size"`
	Checksum    string    `json:"checksum,omitempty"`
}

// FileChangeBatch represents a batch of file changes
type FileChangeBatch struct {
	Events      []FileChangeEvent `json:"events"`
	BatchID     string            `json:"batch_id"`
	StartTime   time.Time         `json:"start_time"`
	EndTime     time.Time         `json:"end_time"`
	TotalEvents int               `json:"total_events"`
}

// SyncResult represents the result of a synchronization operation
type SyncResult struct {
	Success       bool           `json:"success"`
	FilesModified int            `json:"files_modified"`
	FilesCreated  int            `json:"files_created"`
	FilesDeleted  int            `json:"files_deleted"`
	Conflicts     []ConflictInfo `json:"conflicts"`
	Errors        []string       `json:"errors"`
	Duration      time.Duration  `json:"duration"`
	CommitHash    string         `json:"commit_hash,omitempty"`
}

// ConflictInfo represents information about a merge conflict
type ConflictInfo struct {
	Path         string    `json:"path"`
	ConflictType string    `json:"conflict_type"` // content, rename, delete
	LocalHash    string    `json:"local_hash"`
	RemoteHash   string    `json:"remote_hash"`
	Resolution   string    `json:"resolution"` // manual, auto, skip
	ResolvedAt   time.Time `json:"resolved_at,omitempty"`
}
