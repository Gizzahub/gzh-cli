package git

import (
	"context"
	"fmt"
	"time"
)

// Logger interface for dependency injection
type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

// CommandExecutor interface for dependency injection
type CommandExecutor interface {
	Execute(ctx context.Context, command string, args ...string) ([]byte, error)
	ExecuteInDir(ctx context.Context, dir, command string, args ...string) ([]byte, error)
}

// GitClientImpl implements the GitClient interface
type GitClientImpl struct {
	executor CommandExecutor
	logger   Logger
}

// GitClientConfig holds configuration for Git client
type GitClientConfig struct {
	Timeout       time.Duration
	RetryCount    int
	RetryDelay    time.Duration
	DefaultBranch string
}

// DefaultGitClientConfig returns default configuration
func DefaultGitClientConfig() *GitClientConfig {
	return &GitClientConfig{
		Timeout:       30 * time.Second,
		RetryCount:    3,
		RetryDelay:    time.Second,
		DefaultBranch: "main",
	}
}

// NewGitClient creates a new Git client with dependencies
func NewGitClient(config *GitClientConfig, executor CommandExecutor, logger Logger) GitClient {
	if config == nil {
		config = DefaultGitClientConfig()
	}

	return &GitClientImpl{
		executor: executor,
		logger:   logger,
	}
}

// Clone implements GitClient interface
func (g *GitClientImpl) Clone(ctx context.Context, url, path string) error {
	g.logger.Info("Cloning repository", "url", url, "path", path)

	_, err := g.executor.Execute(ctx, "git", "clone", url, path)
	if err != nil {
		g.logger.Error("Failed to clone repository", "url", url, "path", path, "error", err)
		return err
	}

	return nil
}

// Pull implements GitClient interface
func (g *GitClientImpl) Pull(ctx context.Context, path string) error {
	g.logger.Debug("Pulling repository", "path", path)

	_, err := g.executor.ExecuteInDir(ctx, path, "git", "pull")
	if err != nil {
		g.logger.Error("Failed to pull repository", "path", path, "error", err)
		return err
	}

	return nil
}

// Fetch implements GitClient interface
func (g *GitClientImpl) Fetch(ctx context.Context, path string) error {
	g.logger.Debug("Fetching repository", "path", path)

	_, err := g.executor.ExecuteInDir(ctx, path, "git", "fetch")
	if err != nil {
		g.logger.Error("Failed to fetch repository", "path", path, "error", err)
		return err
	}

	return nil
}

// Reset implements GitClient interface
func (g *GitClientImpl) Reset(ctx context.Context, path string, hard bool) error {
	g.logger.Debug("Resetting repository", "path", path, "hard", hard)

	args := []string{"reset"}
	if hard {
		args = append(args, "--hard", "HEAD")
	}

	_, err := g.executor.ExecuteInDir(ctx, path, "git", args...)
	if err != nil {
		g.logger.Error("Failed to reset repository", "path", path, "error", err)
		return err
	}

	return nil
}

// GetStatus implements GitClient interface
func (g *GitClientImpl) GetStatus(ctx context.Context, path string) (*GitStatus, error) {
	g.logger.Debug("Getting repository status", "path", path)

	output, err := g.executor.ExecuteInDir(ctx, path, "git", "status", "--porcelain")
	if err != nil {
		g.logger.Error("Failed to get repository status", "path", path, "error", err)
		return nil, err
	}

	// Parse git status output
	return &GitStatus{
		Clean: len(output) == 0,
		Files: string(output),
	}, nil
}

// GetCurrentBranch implements GitClient interface
func (g *GitClientImpl) GetCurrentBranch(ctx context.Context, path string) (string, error) {
	g.logger.Debug("Getting current branch", "path", path)

	output, err := g.executor.ExecuteInDir(ctx, path, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		g.logger.Error("Failed to get current branch", "path", path, "error", err)
		return "", err
	}

	return string(output), nil
}

// IsRepository implements GitClient interface
func (g *GitClientImpl) IsRepository(ctx context.Context, path string) bool {
	_, err := g.executor.ExecuteInDir(ctx, path, "git", "rev-parse", "--git-dir")
	return err == nil
}

// StrategyExecutorImpl implements the StrategyExecutor interface
type StrategyExecutorImpl struct {
	gitClient GitClient
	logger    Logger
}

// NewStrategyExecutor creates a new strategy executor with dependencies
func NewStrategyExecutor(gitClient GitClient, logger Logger) StrategyExecutor {
	return &StrategyExecutorImpl{
		gitClient: gitClient,
		logger:    logger,
	}
}

// ExecuteStrategy implements StrategyExecutor interface
func (s *StrategyExecutorImpl) ExecuteStrategy(ctx context.Context, strategy, path string) error {
	s.logger.Debug("Executing strategy", "strategy", strategy, "path", path)

	switch strategy {
	case "reset":
		if err := s.gitClient.Reset(ctx, path, true); err != nil {
			return err
		}
		return s.gitClient.Pull(ctx, path)
	case "pull":
		return s.gitClient.Pull(ctx, path)
	case "fetch":
		return s.gitClient.Fetch(ctx, path)
	default:
		s.logger.Warn("Unknown strategy, using default", "strategy", strategy)
		return s.gitClient.Pull(ctx, path)
	}
}

// GetSupportedStrategies implements StrategyExecutor interface
func (s *StrategyExecutorImpl) GetSupportedStrategies() []string {
	return []string{"reset", "pull", "fetch"}
}

// ValidateStrategy implements StrategyExecutor interface
func (s *StrategyExecutorImpl) ValidateStrategy(strategy string) error {
	supported := s.GetSupportedStrategies()
	for _, supportedStrategy := range supported {
		if strategy == supportedStrategy {
			return nil
		}
	}
	return fmt.Errorf("unsupported strategy: %s, supported: %v", strategy, supported)
}

// BulkOperatorImpl implements the BulkOperator interface
type BulkOperatorImpl struct {
	gitClient        GitClient
	strategyExecutor StrategyExecutor
	logger           Logger
}

// BulkOperatorConfig holds configuration for bulk operations
type BulkOperatorConfig struct {
	Concurrency int
	Timeout     time.Duration
}

// DefaultBulkOperatorConfig returns default configuration
func DefaultBulkOperatorConfig() *BulkOperatorConfig {
	return &BulkOperatorConfig{
		Concurrency: 5,
		Timeout:     10 * time.Minute,
	}
}

// NewBulkOperator creates a new bulk operator with dependencies
func NewBulkOperator(
	config *BulkOperatorConfig,
	gitClient GitClient,
	strategyExecutor StrategyExecutor,
	logger Logger,
) BulkOperator {
	if config == nil {
		config = DefaultBulkOperatorConfig()
	}

	return &BulkOperatorImpl{
		gitClient:        gitClient,
		strategyExecutor: strategyExecutor,
		logger:           logger,
	}
}

// ProcessRepositories implements BulkOperator interface
func (b *BulkOperatorImpl) ProcessRepositories(ctx context.Context, operations []GitOperation) (*BulkOperationResult, error) {
	b.logger.Info("Processing repositories", "count", len(operations))

	result := &BulkOperationResult{
		TotalOperations: len(operations),
		Results:         make([]OperationResult, 0, len(operations)),
	}

	for _, op := range operations {
		opResult := b.processOperation(ctx, op)
		result.Results = append(result.Results, opResult)

		if opResult.Success {
			result.SuccessfulOperations++
		} else {
			result.FailedOperations++
		}
	}

	return result, nil
}

// processOperation processes a single Git operation
func (b *BulkOperatorImpl) processOperation(ctx context.Context, op GitOperation) OperationResult {
	result := OperationResult{
		Repository: op.Repository,
		Operation:  op.Operation,
		Strategy:   op.Strategy,
	}

	switch op.Operation {
	case "clone":
		err := b.gitClient.Clone(ctx, op.Repository.URL, op.Repository.Path)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
	case "update":
		err := b.strategyExecutor.ExecuteStrategy(ctx, op.Strategy, op.Repository.Path)
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
	default:
		result.Error = fmt.Sprintf("unknown operation: %s", op.Operation)
	}

	return result
}

// GetOperationProgress implements BulkOperator interface
func (b *BulkOperatorImpl) GetOperationProgress(ctx context.Context) (*OperationProgress, error) {
	// Implementation would track progress across operations
	return &OperationProgress{
		Completed: 0,
		Total:     0,
		Current:   "",
	}, nil
}

// CancelOperations implements BulkOperator interface
func (b *BulkOperatorImpl) CancelOperations(ctx context.Context) error {
	b.logger.Info("Cancelling operations")
	// Implementation would cancel ongoing operations
	return nil
}

// AuthManagerImpl implements the AuthManager interface
type AuthManagerImpl struct {
	logger Logger
}

// NewAuthManager creates a new auth manager with dependencies
func NewAuthManager(logger Logger) AuthManager {
	return &AuthManagerImpl{
		logger: logger,
	}
}

// ConfigureAuth implements AuthManager interface
func (a *AuthManagerImpl) ConfigureAuth(ctx context.Context, repoPath string, authConfig *AuthConfig) error {
	a.logger.Debug("Configuring authentication", "path", repoPath)

	// Implementation would configure Git authentication
	return nil
}

// ValidateAuth implements AuthManager interface
func (a *AuthManagerImpl) ValidateAuth(ctx context.Context, repoPath string) error {
	a.logger.Debug("Validating authentication", "path", repoPath)

	// Implementation would validate Git authentication
	return nil
}

// GetAuthMethods implements AuthManager interface
func (a *AuthManagerImpl) GetAuthMethods(ctx context.Context) ([]string, error) {
	return []string{"ssh", "https", "token"}, nil
}

// RefreshCredentials implements AuthManager interface
func (a *AuthManagerImpl) RefreshCredentials(ctx context.Context, method string) error {
	a.logger.Debug("Refreshing credentials", "method", method)

	// Implementation would refresh credentials
	return nil
}

// GitService implements the unified Git service interface
type GitService struct {
	GitClient
	StrategyExecutor
	BulkOperator
	AuthManager
}

// GitServiceConfig holds configuration for the Git service
type GitServiceConfig struct {
	Client     *GitClientConfig
	BulkOp     *BulkOperatorConfig
	EnableAuth bool
}

// DefaultGitServiceConfig returns default configuration
func DefaultGitServiceConfig() *GitServiceConfig {
	return &GitServiceConfig{
		Client:     DefaultGitClientConfig(),
		BulkOp:     DefaultBulkOperatorConfig(),
		EnableAuth: true,
	}
}

// NewGitService creates a new Git service with all dependencies
func NewGitService(
	config *GitServiceConfig,
	executor CommandExecutor,
	logger Logger,
) GitService {
	if config == nil {
		config = DefaultGitServiceConfig()
	}

	gitClient := NewGitClient(config.Client, executor, logger)
	strategyExecutor := NewStrategyExecutor(gitClient, logger)
	bulkOperator := NewBulkOperator(config.BulkOp, gitClient, strategyExecutor, logger)
	authManager := NewAuthManager(logger)

	return &GitService{
		GitClient:        gitClient,
		StrategyExecutor: strategyExecutor,
		BulkOperator:     bulkOperator,
		AuthManager:      authManager,
	}
}

// ServiceDependencies holds all the dependencies needed for Git services
type ServiceDependencies struct {
	Executor CommandExecutor
	Logger   Logger
}

// NewServiceDependencies creates a default set of service dependencies
func NewServiceDependencies(executor CommandExecutor, logger Logger) *ServiceDependencies {
	return &ServiceDependencies{
		Executor: executor,
		Logger:   logger,
	}
}
