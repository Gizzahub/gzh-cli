// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Gizzahub/gzh-manager-go/cmd/quality/tools"
)

// ParallelExecutor executes quality tools in parallel.
type ParallelExecutor struct {
	maxWorkers int
	timeout    time.Duration
}

// NewParallelExecutor creates a new parallel executor.
func NewParallelExecutor(maxWorkers int, timeout time.Duration) *ParallelExecutor {
	if maxWorkers <= 0 {
		maxWorkers = 4 // Default to 4 workers
	}
	if timeout <= 0 {
		timeout = 5 * time.Minute // Default 5 minute timeout
	}

	return &ParallelExecutor{
		maxWorkers: maxWorkers,
		timeout:    timeout,
	}
}

// Execute runs the execution plan sequentially.
func (e *ParallelExecutor) Execute(ctx context.Context, plan *tools.ExecutionPlan) ([]*tools.Result, error) {
	return e.ExecuteParallel(ctx, plan, 1)
}

// ExecuteParallel runs the execution plan with parallel execution.
func (e *ParallelExecutor) ExecuteParallel(ctx context.Context, plan *tools.ExecutionPlan, workers int) ([]*tools.Result, error) {
	if workers <= 0 {
		workers = e.maxWorkers
	}

	// Sort tasks by priority (higher priority first)
	sortedTasks := make([]tools.Task, len(plan.Tasks))
	copy(sortedTasks, plan.Tasks)
	sort.Slice(sortedTasks, func(i, j int) bool {
		return sortedTasks[i].Priority > sortedTasks[j].Priority
	})

	// Create worker pool
	taskChan := make(chan tools.Task, len(sortedTasks))
	resultChan := make(chan *tools.Result, len(sortedTasks))
	errorChan := make(chan error, len(sortedTasks))

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go e.worker(timeoutCtx, &wg, taskChan, resultChan, errorChan)
	}

	// Send tasks to workers
	go func() {
		defer close(taskChan)
		for _, task := range sortedTasks {
			select {
			case taskChan <- task:
			case <-timeoutCtx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect results
	var results []*tools.Result
	var errors []error

	for {
		select {
		case result, ok := <-resultChan:
			if !ok {
				resultChan = nil
			} else {
				results = append(results, result)
			}
		case err, ok := <-errorChan:
			if !ok {
				errorChan = nil
			} else if err != nil {
				errors = append(errors, err)
			}
		case <-timeoutCtx.Done():
			return results, fmt.Errorf("execution timed out after %v", e.timeout)
		}

		if resultChan == nil && errorChan == nil {
			break
		}
	}

	// Return first error if any occurred
	if len(errors) > 0 {
		return results, errors[0]
	}

	return results, nil
}

// worker processes tasks from the task channel.
func (e *ParallelExecutor) worker(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan tools.Task, resultChan chan<- *tools.Result, errorChan chan<- error) {
	defer wg.Done()

	for {
		select {
		case task, ok := <-taskChan:
			if !ok {
				return
			}

			// Execute the task
			result, err := task.Tool.Execute(ctx, task.Files, task.Options)

			// Send result
			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}

			// Send error if any
			select {
			case errorChan <- err:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// ExecutionPlanner creates execution plans.
type ExecutionPlanner struct {
	analyzer ProjectAnalyzer
}

// NewExecutionPlanner creates a new execution planner.
func NewExecutionPlanner(analyzer ProjectAnalyzer) *ExecutionPlanner {
	return &ExecutionPlanner{
		analyzer: analyzer,
	}
}

// CreatePlan creates an execution plan for the given options.
func (p *ExecutionPlanner) CreatePlan(projectRoot string, registry tools.ToolRegistry, options PlanOptions) (*tools.ExecutionPlan, error) {
	// Analyze the project
	analysis, err := p.analyzer.AnalyzeProject(projectRoot, registry)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze project: %w", err)
	}

	// Get optimal tool selection
	selection := p.analyzer.GetOptimalToolSelection(analysis, registry)

	var tasks []tools.Task
	totalFiles := 0

	// Create tasks for each language
	for language, toolList := range selection {
		files := analysis.Languages[language]
		if len(files) == 0 {
			continue
		}

		// Apply file filtering based on options
		files, err = p.applyFileFilters(projectRoot, files, options)
		if err != nil {
			return nil, fmt.Errorf("failed to apply file filters: %w", err)
		}

		if len(files) == 0 {
			continue
		}

		// Create tasks for each tool
		for _, tool := range toolList {
			// Skip if tool type doesn't match options
			if !matchesToolType(tool, options) {
				continue
			}

			// Skip if language doesn't match filter
			if !matchesLanguageFilter(tool, options) {
				continue
			}

			// Skip if tool name doesn't match filter
			if !matchesToolFilter(tool, options) {
				continue
			}

			// Create execution options
			execOptions := tools.ExecuteOptions{
				ProjectRoot: projectRoot,
				Fix:         options.Fix,
				FormatOnly:  options.FormatOnly,
				LintOnly:    options.LintOnly,
				ExtraArgs:   options.ExtraArgs,
				Env:         options.Env,
			}

			// Set config file if found
			if configFile, exists := analysis.ConfigFiles[tool.Name()]; exists {
				execOptions.ConfigFile = configFile
			}

			// Determine priority
			priority := 0
			if tool.Type() == tools.FORMAT {
				priority = 10 // Formatters run first
			} else if tool.Type() == tools.LINT {
				priority = 5 // Linters run second
			} else {
				priority = 7 // BOTH tools run in between
			}

			task := tools.Task{
				Tool:     tool,
				Files:    files,
				Options:  execOptions,
				Priority: priority,
			}

			tasks = append(tasks, task)
			totalFiles += len(files)
		}
	}

	// Estimate duration (rough estimate: 100ms per file per tool)
	estimatedSeconds := len(tasks) * totalFiles / 10
	estimatedDuration := fmt.Sprintf("%ds", estimatedSeconds)

	return &tools.ExecutionPlan{
		Tasks:             tasks,
		TotalFiles:        totalFiles,
		EstimatedDuration: estimatedDuration,
	}, nil
}

// PlanOptions contains options for creating execution plans.
type PlanOptions struct {
	Files      []string          // Specific files to process
	Fix        bool              // Auto-fix issues if supported
	FormatOnly bool              // Run only formatters
	LintOnly   bool              // Run only linters
	ExtraArgs  []string          // Extra arguments to pass to tools
	Env        map[string]string // Environment variables
	Language   string            // Filter by specific language
	ToolFilter []string          // Filter by specific tool names
	// Git-based options
	Since   string // Process files changed since this commit
	Staged  bool   // Process only staged files
	Changed bool   // Process only changed files (staged + modified + untracked)
}

// matchesToolType checks if a tool matches the requested type options.
func matchesToolType(tool tools.QualityTool, options PlanOptions) bool {
	toolType := tool.Type()

	if options.FormatOnly {
		return toolType == tools.FORMAT || toolType == tools.BOTH
	}

	if options.LintOnly {
		return toolType == tools.LINT || toolType == tools.BOTH
	}

	// If neither FormatOnly nor LintOnly, include all tools
	return true
}

// matchesLanguageFilter checks if a tool matches the language filter.
func matchesLanguageFilter(tool tools.QualityTool, options PlanOptions) bool {
	if options.Language == "" {
		return true // No language filter
	}
	return tool.Language() == options.Language
}

// matchesToolFilter checks if a tool matches the tool name filter.
func matchesToolFilter(tool tools.QualityTool, options PlanOptions) bool {
	if len(options.ToolFilter) == 0 {
		return true // No tool filter
	}

	toolName := tool.Name()
	for _, filterName := range options.ToolFilter {
		if toolName == filterName {
			return true
		}
	}
	return false
}

// applyFileFilters applies various file filtering options.
func (p *ExecutionPlanner) applyFileFilters(projectRoot string, files []string, options PlanOptions) ([]string, error) {
	var filteredFiles []string

	// Handle Git-based filtering
	if options.Since != "" || options.Staged || options.Changed {
		gitFiles, err := p.getGitFilteredFiles(projectRoot, options)
		if err != nil {
			return nil, err
		}

		// Intersect with existing files
		filteredFiles = intersectFiles(files, gitFiles)
	} else {
		filteredFiles = files
	}

	// Apply specific file filtering if requested
	if len(options.Files) > 0 {
		filteredFiles = intersectFiles(filteredFiles, options.Files)
	}

	return filteredFiles, nil
}

// getGitFilteredFiles returns files based on Git filtering options.
func (p *ExecutionPlanner) getGitFilteredFiles(projectRoot string, options PlanOptions) ([]string, error) {
	// Lazy import to avoid dependency issues
	gitUtils := &GitUtils{projectRoot: projectRoot}

	if !gitUtils.IsGitRepository() {
		return nil, fmt.Errorf("git filtering requested but not in a git repository")
	}

	var gitFiles []string
	var err error

	switch {
	case options.Since != "":
		// Validate commit reference first
		if err := gitUtils.ValidateCommitish(options.Since); err != nil {
			return nil, err
		}
		gitFiles, err = gitUtils.GetChangedFiles(options.Since)
	case options.Staged:
		gitFiles, err = gitUtils.GetStagedFiles()
	case options.Changed:
		gitFiles, err = gitUtils.GetAllChangedFiles()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get git files: %w", err)
	}

	return gitFiles, nil
}

// GitUtils provides Git-related utilities (embedded for simplicity).
type GitUtils struct {
	projectRoot string
}

func (g *GitUtils) IsGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.projectRoot
	return cmd.Run() == nil
}

func (g *GitUtils) ValidateCommitish(commitish string) error {
	cmd := exec.Command("git", "rev-parse", "--verify", commitish+"^{commit}")
	cmd.Dir = g.projectRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("invalid commit reference '%s': %w", commitish, err)
	}
	return nil
}

func (g *GitUtils) GetChangedFiles(since string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", since)
	cmd.Dir = g.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get git diff: %w", err)
	}
	return g.parseFileList(string(output)), nil
}

func (g *GitUtils) GetStagedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--cached", "--name-only")
	cmd.Dir = g.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	return g.parseFileList(string(output)), nil
}

func (g *GitUtils) GetAllChangedFiles() ([]string, error) {
	var allFiles []string

	// Get staged files
	staged, err := g.GetStagedFiles()
	if err != nil {
		return nil, err
	}
	allFiles = append(allFiles, staged...)

	// Get modified files
	cmd := exec.Command("git", "diff", "--name-only")
	cmd.Dir = g.projectRoot
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get modified files: %w", err)
	}
	modified := g.parseFileList(string(output))
	allFiles = append(allFiles, modified...)

	// Get untracked files
	cmd = exec.Command("git", "ls-files", "--others", "--exclude-standard")
	cmd.Dir = g.projectRoot
	output, err = cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get untracked files: %w", err)
	}
	untracked := g.parseFileList(string(output))
	allFiles = append(allFiles, untracked...)

	return g.deduplicateAndMakeAbsolute(allFiles), nil
}

func (g *GitUtils) parseFileList(output string) []string {
	var files []string
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func (g *GitUtils) deduplicateAndMakeAbsolute(files []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, file := range files {
		if seen[file] {
			continue
		}
		seen[file] = true

		// Convert to absolute path
		absPath := filepath.Join(g.projectRoot, file)
		if _, err := os.Stat(absPath); err == nil {
			result = append(result, absPath)
		}
	}

	return result
}

// intersectFiles returns the intersection of two file slices.
func intersectFiles(files1, files2 []string) []string {
	fileSet := make(map[string]bool)
	for _, file := range files2 {
		fileSet[file] = true
	}

	var result []string
	for _, file := range files1 {
		if fileSet[file] {
			result = append(result, file)
		}
	}

	return result
}

// ProjectAnalyzer is an alias to avoid circular import.
type ProjectAnalyzer interface {
	AnalyzeProject(projectRoot string, registry tools.ToolRegistry) (*AnalysisResult, error)
	GetOptimalToolSelection(result *AnalysisResult, registry tools.ToolRegistry) map[string][]tools.QualityTool
}

// AnalysisResult contains the results of project analysis.
type AnalysisResult struct {
	ProjectRoot      string
	Languages        map[string][]string
	AvailableTools   []string
	RecommendedTools map[string][]string
	ConfigFiles      map[string]string
	Issues           []string
}

// Ensure ParallelExecutor implements Executor.
var _ tools.Executor = (*ParallelExecutor)(nil)
