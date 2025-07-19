package git

import (
	"context"
	"time"
)

// Repository represents a Git repository.
type Repository struct {
	Path          string            `json:"path"`
	RemoteURL     string            `json:"remoteUrl"`
	CurrentBranch string            `json:"currentBranch"`
	DefaultBranch string            `json:"defaultBranch"`
	Remotes       map[string]string `json:"remotes"`
	IsDirty       bool              `json:"isDirty"`
	IsDetached    bool              `json:"isDetached"`
	LastCommit    *Commit           `json:"lastCommit,omitempty"`
}

// Commit represents a Git commit.
type Commit struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	Date      time.Time `json:"date"`
}

// CloneOptions represents options for cloning a repository.
type CloneOptions struct {
	URL          string `json:"url"`
	Path         string `json:"path"`
	Branch       string `json:"branch,omitempty"`
	Depth        int    `json:"depth,omitempty"`
	SingleBranch bool   `json:"singleBranch"`
	Bare         bool   `json:"bare"`
	Mirror       bool   `json:"mirror"`
	Recursive    bool   `json:"recursive"`
	SSHKeyPath   string `json:"sshKeyPath,omitempty"`
	Token        string `json:"token,omitempty"`
}

// PullOptions represents options for pulling changes.
type PullOptions struct {
	Remote     string `json:"remote"`
	Branch     string `json:"branch,omitempty"`
	Strategy   string `json:"strategy"` // merge, rebase, fast-forward
	Force      bool   `json:"force"`
	AllowDirty bool   `json:"allow_dirty"`
}

// ResetOptions represents options for resetting repository state.
type ResetOptions struct {
	Mode   string `json:"mode"`   // soft, mixed, hard
	Target string `json:"target"` // commit hash, branch, tag
	Force  bool   `json:"force"`
}

// OperationResult represents the result of a Git operation.
type OperationResult struct {
	Success      bool              `json:"success"`
	Message      string            `json:"message"`
	Error        string            `json:"error,omitempty"`
	Duration     time.Duration     `json:"duration"`
	ChangedFiles []string          `json:"changed_files,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// GitClient defines the interface for Git operations.
type GitClient interface {
	// Repository operations
	Clone(ctx context.Context, options CloneOptions) (*OperationResult, error)
	Pull(ctx context.Context, repoPath string, options PullOptions) (*OperationResult, error)
	Fetch(ctx context.Context, repoPath string, remote string) (*OperationResult, error)
	Reset(ctx context.Context, repoPath string, options ResetOptions) (*OperationResult, error)

	// Repository status
	GetRepository(ctx context.Context, path string) (*Repository, error)
	IsRepository(ctx context.Context, path string) bool
	IsDirty(ctx context.Context, repoPath string) (bool, error)
	GetCurrentBranch(ctx context.Context, repoPath string) (string, error)
	GetDefaultBranch(ctx context.Context, repoPath string) (string, error)

	// Branch operations
	ListBranches(ctx context.Context, repoPath string) ([]string, error)
	CreateBranch(ctx context.Context, repoPath, branchName string) (*OperationResult, error)
	CheckoutBranch(ctx context.Context, repoPath, branchName string) (*OperationResult, error)
	DeleteBranch(ctx context.Context, repoPath, branchName string) (*OperationResult, error)

	// Remote operations
	ListRemotes(ctx context.Context, repoPath string) (map[string]string, error)
	AddRemote(ctx context.Context, repoPath, name, url string) (*OperationResult, error)
	RemoveRemote(ctx context.Context, repoPath, name string) (*OperationResult, error)
	SetRemoteURL(ctx context.Context, repoPath, remote, url string) (*OperationResult, error)

	// Commit operations
	GetLastCommit(ctx context.Context, repoPath string) (*Commit, error)
	GetCommitHistory(ctx context.Context, repoPath string, limit int) ([]Commit, error)

	// Utility operations
	ValidateRepository(ctx context.Context, path string) error
	GetStatus(ctx context.Context, repoPath string) (*StatusResult, error)
}

// StatusResult represents the status of a Git repository.
type StatusResult struct {
	Clean          bool     `json:"clean"`
	Branch         string   `json:"branch"`
	Ahead          int      `json:"ahead"`
	Behind         int      `json:"behind"`
	ModifiedFiles  []string `json:"modified_files"`
	StagedFiles    []string `json:"staged_files"`
	UntrackedFiles []string `json:"untracked_files"`
	ConflictFiles  []string `json:"conflict_files"`
}

// StrategyExecutor defines the interface for executing different Git strategies.
type StrategyExecutor interface {
	// Execute strategy on a repository
	ExecuteStrategy(ctx context.Context, repoPath, strategy string) (*OperationResult, error)

	// Get supported strategies
	GetSupportedStrategies() []string

	// Validate strategy
	IsValidStrategy(strategy string) bool

	// Get strategy description
	GetStrategyDescription(strategy string) string
}

// BulkOperator defines the interface for bulk Git operations.
type BulkOperator interface {
	// Execute operation on multiple repositories
	ExecuteBulkOperation(ctx context.Context, repoPaths []string, operation BulkOperation) ([]BulkResult, error)

	// Execute with concurrency control
	ExecuteBulkOperationWithOptions(ctx context.Context, repoPaths []string, operation BulkOperation, options BulkOptions) ([]BulkResult, error)

	// Get operation progress
	GetProgress() <-chan BulkProgress
}

// BulkOperation represents a bulk operation to execute.
type BulkOperation struct {
	Type     string                 `json:"type"` // clone, pull, fetch, reset
	Strategy string                 `json:"strategy,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// BulkOptions represents options for bulk operations.
type BulkOptions struct {
	Concurrency      int                `json:"concurrency"`
	Timeout          time.Duration      `json:"timeout"`
	ContinueOnError  bool               `json:"continue_on_error"`
	ProgressCallback func(BulkProgress) `json:"-"`
}

// BulkResult represents the result of a bulk operation on a repository.
type BulkResult struct {
	RepoPath string           `json:"repo_path"`
	Success  bool             `json:"success"`
	Result   *OperationResult `json:"result,omitempty"`
	Error    string           `json:"error,omitempty"`
	Duration time.Duration    `json:"duration"`
}

// BulkProgress represents progress information for bulk operations.
type BulkProgress struct {
	TotalRepos      int           `json:"total_repos"`
	CompletedRepos  int           `json:"completed_repos"`
	SuccessfulRepos int           `json:"successful_repos"`
	FailedRepos     int           `json:"failed_repos"`
	CurrentRepo     string        `json:"current_repo"`
	ElapsedTime     time.Duration `json:"elapsed_time"`
	EstimatedTime   time.Duration `json:"estimated_time"`
}

// AuthManager defines the interface for Git authentication.
type AuthManager interface {
	// Configure SSH authentication
	ConfigureSSHAuth(ctx context.Context, keyPath, passphrase string) error

	// Configure token authentication
	ConfigureTokenAuth(ctx context.Context, token string) error

	// Configure username/password authentication
	ConfigurePasswordAuth(ctx context.Context, username, password string) error

	// Get current authentication method
	GetAuthMethod() string

	// Validate authentication
	ValidateAuth(ctx context.Context, remoteURL string) error
}

// GitService provides a unified interface for all Git operations.
type GitService interface {
	GitClient
	StrategyExecutor
	BulkOperator
	AuthManager
}
