// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package tools

import "context"

// ToolType defines the type of quality tool
type ToolType int

const (
	FORMAT ToolType = iota
	LINT
	BOTH
)

func (t ToolType) String() string {
	switch t {
	case FORMAT:
		return "formatter"
	case LINT:
		return "linter"
	case BOTH:
		return "formatter+linter"
	default:
		return "unknown"
	}
}

// QualityTool represents a code quality tool (formatter or linter)
type QualityTool interface {
	// Name returns the tool name (e.g., "gofumpt", "eslint")
	Name() string

	// Language returns the programming language (e.g., "Go", "Python")
	Language() string

	// Type returns the tool type (FORMAT, LINT, or BOTH)
	Type() ToolType

	// IsAvailable checks if the tool is installed and available
	IsAvailable() bool

	// Install attempts to install the tool automatically
	Install() error

	// GetVersion returns the version of the installed tool
	GetVersion() (string, error)

	// Upgrade attempts to upgrade the tool to the latest version
	Upgrade() error

	// FindConfigFiles returns configuration files the tool would use
	FindConfigFiles(projectRoot string) []string

	// Execute runs the tool on the specified files
	Execute(ctx context.Context, files []string, options ExecuteOptions) (*Result, error)
}

// ExecuteOptions contains options for tool execution
type ExecuteOptions struct {
	// ProjectRoot is the root directory of the project
	ProjectRoot string

	// ConfigFile is the path to the tool's configuration file
	ConfigFile string

	// Fix indicates whether to auto-fix issues (if supported)
	Fix bool

	// FormatOnly runs only formatting (for tools that support both)
	FormatOnly bool

	// LintOnly runs only linting (for tools that support both)
	LintOnly bool

	// ExtraArgs are additional arguments to pass to the tool
	ExtraArgs []string

	// Env contains environment variables for the tool
	Env map[string]string
}

// Result contains the results of tool execution
type Result struct {
	// Tool is the name of the tool that was executed
	Tool string

	// Language is the programming language
	Language string

	// Success indicates whether the tool executed successfully
	Success bool

	// Error contains any execution error
	Error error

	// FilesProcessed is the number of files processed
	FilesProcessed int

	// Duration is how long the tool took to run
	Duration string

	// Issues contains any issues found by the tool
	Issues []Issue

	// Output contains the raw output from the tool
	Output string
}

// Issue represents a code quality issue found by a tool
type Issue struct {
	// File is the path to the file containing the issue
	File string

	// Line is the line number (1-based)
	Line int

	// Column is the column number (1-based)
	Column int

	// Severity is the issue severity (error, warning, info)
	Severity string

	// Rule is the rule that was violated
	Rule string

	// Message is the issue description
	Message string

	// Suggestion is an optional fix suggestion
	Suggestion string
}

// LanguageDetector detects programming languages in a project
type LanguageDetector interface {
	// DetectLanguages scans a directory and returns detected languages
	DetectLanguages(projectRoot string) ([]string, error)

	// GetFilesByLanguage returns files grouped by language
	GetFilesByLanguage(projectRoot string, languages []string) (map[string][]string, error)
}

// ToolRegistry manages available quality tools
type ToolRegistry interface {
	// Register adds a tool to the registry
	Register(tool QualityTool)

	// GetTools returns all registered tools
	GetTools() []QualityTool

	// GetToolsByLanguage returns tools for a specific language
	GetToolsByLanguage(language string) []QualityTool

	// GetToolsByType returns tools of a specific type
	GetToolsByType(toolType ToolType) []QualityTool

	// FindTool finds a tool by name
	FindTool(name string) QualityTool
}

// ConfigDetector finds configuration files for quality tools
type ConfigDetector interface {
	// FindConfigs searches for tool configuration files
	FindConfigs(projectRoot string, tools []QualityTool) map[string]string

	// ValidateConfig checks if a configuration file is valid
	ValidateConfig(toolName, configPath string) error
}

// ExecutionPlan represents a plan for executing quality tools
type ExecutionPlan struct {
	// Tasks are the individual tool execution tasks
	Tasks []Task

	// TotalFiles is the total number of files to be processed
	TotalFiles int

	// EstimatedDuration is the estimated time to complete all tasks
	EstimatedDuration string
}

// Task represents a single tool execution task
type Task struct {
	// Tool is the quality tool to execute
	Tool QualityTool

	// Files are the files to process
	Files []string

	// Options are the execution options
	Options ExecuteOptions

	// Priority affects execution order (higher = earlier)
	Priority int
}

// Executor runs quality tools according to an execution plan
type Executor interface {
	// Execute runs the execution plan
	Execute(ctx context.Context, plan *ExecutionPlan) ([]*Result, error)

	// ExecuteParallel runs the plan with parallel execution
	ExecuteParallel(ctx context.Context, plan *ExecutionPlan, workers int) ([]*Result, error)
}
