// Package github provides comprehensive GitHub API integration and repository management.
//
// This package handles:
//   - GitHub API client implementation with rate limiting
//   - Repository and organization management
//   - Token validation and authentication
//   - Bulk cloning and synchronization operations
//   - Change logging and operation tracking
//   - User confirmation workflows for destructive operations
//   - Repository configuration management
//
// The package implements a sophisticated GitHub integration layer with multiple
// service interfaces for different concerns. It provides both low-level API
// access and high-level facade operations for common workflows.
//
// Main interfaces:
//   - GitHubService: Unified interface for all GitHub operations
//   - APIClient: GitHub API operations and rate limiting
//   - CloneService: Repository cloning and synchronization
//   - TokenValidatorInterface: Token validation and permission checking
//   - ChangeLoggerInterface: Operation logging and audit trails
//   - ConfirmationServiceInterface: User confirmation workflows
//
// Key types:
//   - GitHubManager: High-level facade for GitHub operations
//   - RepositoryInfo: Repository metadata and information
//   - BulkCloneRequest/Result: Bulk operation request/response structures
//   - TokenInfoRecord: Token information and validation results
//
// The package supports advanced features like:
//   - Comprehensive rate limiting with GitHub API limits
//   - Token scope validation for specific operations
//   - Bulk operations with filtering and concurrency control
//   - Operation logging with structured metadata
//   - Risk-based confirmation prompts for destructive operations
//   - Factory pattern for provider instantiation
package github
