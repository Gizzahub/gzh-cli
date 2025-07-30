// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/cli"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
)

// DevEnvResult represents development environment check result
type DevEnvResult struct {
	Tool       string                 `json:"tool"`
	Status     string                 `json:"status"`
	Version    string                 `json:"version"`
	Path       string                 `json:"path"`
	Required   bool                   `json:"required"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Suggestion string                 `json:"suggestion,omitempty"`
}

// DevEnvReport represents the complete development environment report
type DevEnvReport struct {
	Timestamp       time.Time      `json:"timestamp"`
	Platform        string         `json:"platform"`
	GoVersion       string         `json:"go_version"`
	WorkingDir      string         `json:"working_dir"`
	Results         []DevEnvResult `json:"results"`
	Summary         string         `json:"summary"`
	IsReady         bool           `json:"is_ready"`
	MissingTools    []string       `json:"missing_tools"`
	Recommendations []string       `json:"recommendations"`
}

// newDevEnvCmd creates the dev-env subcommand for development environment validation
func newDevEnvCmd() *cobra.Command {
	ctx := context.Background()

	var (
		checkOnly   bool
		fixIssues   bool
		showDetails bool
	)

	cmd := cli.NewCommandBuilder(ctx, "dev-env", "Validate development environment setup").
		WithLongDescription(`Comprehensive development environment validation for GZH Manager contributors.

This command checks all required development tools and configurations:
- Go version and modules
- Build tools (make, golangci-lint, gofumpt, gci)
- Git configuration and hooks
- Testing frameworks
- Documentation tools
- IDE configurations

Examples:
  gz doctor dev-env                    # Check all development tools
  gz doctor dev-env --fix              # Check and attempt to fix issues
  gz doctor dev-env --details          # Show detailed tool information
  gz doctor dev-env --check-only       # Only check, don't suggest fixes`).
		WithExample("gz doctor dev-env --fix --details").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			return runDevEnvCheck(ctx, flags, devEnvOptions{
				checkOnly:   checkOnly,
				fixIssues:   fixIssues,
				showDetails: showDetails,
			})
		}).
		Build()

	cmd.Flags().BoolVar(&checkOnly, "check-only", false, "Only check tools, don't provide fix suggestions")
	cmd.Flags().BoolVar(&fixIssues, "fix", false, "Attempt to automatically fix issues")
	cmd.Flags().BoolVar(&showDetails, "details", false, "Show detailed information about each tool")

	return cmd
}

type devEnvOptions struct {
	checkOnly   bool
	fixIssues   bool
	showDetails bool
}

func runDevEnvCheck(ctx context.Context, flags *cli.CommonFlags, opts devEnvOptions) error {
	logger := logger.NewSimpleLogger("doctor-dev-env")

	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	logger.Info("Starting development environment validation")

	report := &DevEnvReport{
		Timestamp:       time.Now(),
		Platform:        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		GoVersion:       runtime.Version(),
		WorkingDir:      workingDir,
		Results:         []DevEnvResult{},
		MissingTools:    []string{},
		Recommendations: []string{},
	}

	// Define required and optional tools
	tools := []struct {
		name     string
		required bool
		checker  func(context.Context) DevEnvResult
	}{
		{"go", true, checkGoTool},
		{"git", true, checkGitTool},
		{"make", true, checkMakeTool},
		{"golangci-lint", true, checkGolangciLint},
		{"gofumpt", true, checkGofumpt},
		{"gci", true, checkGci},
		{"gomock", false, checkGomock},
		{"pre-commit", false, checkPreCommit},
		{"docker", false, checkDocker},
		{"goreleaser", false, checkGoReleaser},
	}

	// Run checks
	for _, tool := range tools {
		logger.Debug("Checking tool", "tool", tool.name, "required", tool.required)
		result := tool.checker(ctx)
		result.Required = tool.required

		report.Results = append(report.Results, result)

		if result.Status == "missing" && tool.required {
			report.MissingTools = append(report.MissingTools, tool.name)
		}
	}

	// Check project-specific configuration
	logger.Debug("Checking project configuration")
	projectChecks := []func(context.Context, *DevEnvReport){
		checkGoModules,
		checkMakefile,
		checkPreCommitConfig,
		checkGolangciConfig,
		checkGitHooks,
	}

	for _, check := range projectChecks {
		check(ctx, report)
	}

	// Generate summary and recommendations
	generateDevEnvSummary(report)

	// Attempt fixes if requested
	if opts.fixIssues {
		logger.Info("Attempting to fix development environment issues")
		attemptDevEnvFixes(ctx, report, logger)
	}

	// Generate output
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		return formatter.FormatOutput(report)
	default:
		return displayDevEnvResults(report, opts)
	}
}

func checkGoTool(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "go", Status: "missing"}

	goPath, err := exec.LookPath("go")
	if err != nil {
		result.Suggestion = "Install Go from https://golang.org/dl/"
		return result
	}

	result.Path = goPath
	result.Status = "found"

	// Check Go version
	cmd := exec.CommandContext(ctx, "go", "version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))

	// Check if version is supported (1.19+)
	if isGoVersionOutdated(result.Version) {
		result.Status = "outdated"
		result.Suggestion = "Update to Go 1.21+ for best compatibility"
	} else {
		result.Status = "ok"
	}

	// Get additional Go environment info
	result.Details = map[string]interface{}{
		"GOPATH":      os.Getenv("GOPATH"),
		"GOROOT":      os.Getenv("GOROOT"),
		"GOPROXY":     os.Getenv("GOPROXY"),
		"GOSUMDB":     os.Getenv("GOSUMDB"),
		"GO111MODULE": os.Getenv("GO111MODULE"),
	}

	return result
}

func checkGitTool(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "git", Status: "missing"}

	gitPath, err := exec.LookPath("git")
	if err != nil {
		result.Suggestion = "Install Git from https://git-scm.com/"
		return result
	}

	result.Path = gitPath
	result.Status = "found"

	// Check Git version
	cmd := exec.CommandContext(ctx, "git", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	// Check Git configuration
	gitConfig := make(map[string]string)
	configKeys := []string{"user.name", "user.email", "core.editor", "init.defaultBranch"}

	for _, key := range configKeys {
		cmd := exec.CommandContext(ctx, "git", "config", "--global", key)
		output, err := cmd.Output()
		if err == nil {
			gitConfig[key] = strings.TrimSpace(string(output))
		}
	}

	result.Details = map[string]interface{}{"config": gitConfig}

	if gitConfig["user.name"] == "" || gitConfig["user.email"] == "" {
		result.Status = "incomplete"
		result.Suggestion = "Configure Git with 'git config --global user.name' and 'git config --global user.email'"
	}

	return result
}

func checkMakeTool(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "make", Status: "missing"}

	makePath, err := exec.LookPath("make")
	if err != nil {
		// Check for alternatives on Windows
		if runtime.GOOS == "windows" {
			if _, err := exec.LookPath("mingw32-make"); err == nil {
				result.Status = "alternative"
				result.Version = "mingw32-make available"
				result.Suggestion = "Using mingw32-make as alternative to make"
				return result
			}
		}
		result.Suggestion = "Install make (build-essential on Ubuntu, Xcode tools on macOS)"
		return result
	}

	result.Path = makePath
	result.Status = "found"

	// Check make version
	cmd := exec.CommandContext(ctx, "make", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		result.Version = strings.TrimSpace(lines[0])
	}
	result.Status = "ok"

	return result
}

func checkGolangciLint(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "golangci-lint", Status: "missing"}

	lintPath, err := exec.LookPath("golangci-lint")
	if err != nil {
		result.Suggestion = "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
		return result
	}

	result.Path = lintPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "golangci-lint", "version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	return result
}

func checkGofumpt(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "gofumpt", Status: "missing"}

	gofumptPath, err := exec.LookPath("gofumpt")
	if err != nil {
		result.Suggestion = "Install with: go install mvdan.cc/gofumpt@latest"
		return result
	}

	result.Path = gofumptPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "gofumpt", "-version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	return result
}

func checkGci(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "gci", Status: "missing"}

	gciPath, err := exec.LookPath("gci")
	if err != nil {
		result.Suggestion = "Install with: go install github.com/daixiang0/gci@latest"
		return result
	}

	result.Path = gciPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "gci", "version")
	output, err := cmd.Output()
	if err != nil {
		// gci might not have version command, that's ok
		result.Version = "available"
	} else {
		result.Version = strings.TrimSpace(string(output))
	}

	result.Status = "ok"
	return result
}

func checkGomock(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "gomock", Status: "missing"}

	mockgenPath, err := exec.LookPath("mockgen")
	if err != nil {
		result.Suggestion = "Install with: go install go.uber.org/mock/mockgen@latest"
		return result
	}

	result.Path = mockgenPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "mockgen", "-version")
	output, err := cmd.Output()
	if err != nil {
		result.Version = "available"
	} else {
		result.Version = strings.TrimSpace(string(output))
	}

	result.Status = "ok"
	return result
}

func checkPreCommit(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "pre-commit", Status: "missing"}

	preCommitPath, err := exec.LookPath("pre-commit")
	if err != nil {
		result.Suggestion = "Install with: pip install pre-commit"
		return result
	}

	result.Path = preCommitPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "pre-commit", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	return result
}

func checkDocker(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "docker", Status: "missing"}

	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		result.Suggestion = "Install Docker from https://docker.com/"
		return result
	}

	result.Path = dockerPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	return result
}

func checkGoReleaser(ctx context.Context) DevEnvResult {
	result := DevEnvResult{Tool: "goreleaser", Status: "missing"}

	goreleaserPath, err := exec.LookPath("goreleaser")
	if err != nil {
		result.Suggestion = "Install with: go install github.com/goreleaser/goreleaser@latest"
		return result
	}

	result.Path = goreleaserPath
	result.Status = "found"

	// Check version
	cmd := exec.CommandContext(ctx, "goreleaser", "--version")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "error"
		result.Details = map[string]interface{}{"error": err.Error()}
		return result
	}

	result.Version = strings.TrimSpace(string(output))
	result.Status = "ok"

	return result
}

func checkGoModules(ctx context.Context, report *DevEnvReport) {
	result := DevEnvResult{Tool: "go.mod", Status: "missing", Required: true}

	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		result.Suggestion = "Initialize Go modules with: go mod init"
		report.Results = append(report.Results, result)
		return
	}

	result.Status = "found"

	// Check if modules are tidy
	cmd := exec.CommandContext(ctx, "go", "mod", "tidy", "-diff")
	output, err := cmd.Output()
	if err != nil {
		result.Status = "needs_tidy"
		result.Suggestion = "Run 'go mod tidy' to clean up dependencies"
		result.Details = map[string]interface{}{"error": err.Error()}
	} else if len(output) > 0 {
		result.Status = "needs_tidy"
		result.Suggestion = "Run 'go mod tidy' to clean up dependencies"
	} else {
		result.Status = "ok"
	}

	report.Results = append(report.Results, result)
}

func checkMakefile(ctx context.Context, report *DevEnvReport) {
	result := DevEnvResult{Tool: "Makefile", Status: "missing", Required: true}

	if _, err := os.Stat("Makefile"); os.IsNotExist(err) {
		result.Suggestion = "Project requires Makefile for build automation"
		report.Results = append(report.Results, result)
		return
	}

	result.Status = "found"
	result.Status = "ok"

	// Check if essential targets exist
	makeTargets := []string{"build", "test", "lint", "fmt"}
	cmd := exec.CommandContext(ctx, "make", "-pRrq", ":")
	output, err := cmd.Output()
	if err == nil {
		targets := make(map[string]bool)
		for _, target := range makeTargets {
			if strings.Contains(string(output), target+":") {
				targets[target] = true
			}
		}
		result.Details = map[string]interface{}{"targets": targets}
	}

	report.Results = append(report.Results, result)
}

func checkPreCommitConfig(ctx context.Context, report *DevEnvReport) {
	result := DevEnvResult{Tool: ".pre-commit-config.yaml", Status: "missing", Required: false}

	if _, err := os.Stat(".pre-commit-config.yaml"); os.IsNotExist(err) {
		result.Suggestion = "Add pre-commit configuration for automated code quality checks"
		report.Results = append(report.Results, result)
		return
	}

	result.Status = "found"
	result.Status = "ok"

	report.Results = append(report.Results, result)
}

func checkGolangciConfig(ctx context.Context, report *DevEnvReport) {
	result := DevEnvResult{Tool: ".golangci.yml", Status: "missing", Required: true}

	configFiles := []string{".golangci.yml", ".golangci.yaml", "golangci.yml", "golangci.yaml"}
	var foundConfig string

	for _, file := range configFiles {
		if _, err := os.Stat(file); err == nil {
			foundConfig = file
			break
		}
	}

	if foundConfig == "" {
		result.Suggestion = "Add golangci-lint configuration file"
		report.Results = append(report.Results, result)
		return
	}

	result.Status = "found"
	result.Status = "ok"
	result.Details = map[string]interface{}{"config_file": foundConfig}

	report.Results = append(report.Results, result)
}

func checkGitHooks(ctx context.Context, report *DevEnvReport) {
	result := DevEnvResult{Tool: "git-hooks", Status: "missing", Required: false}

	gitDir := ".git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		result.Suggestion = "Initialize git repository with: git init"
		report.Results = append(report.Results, result)
		return
	}

	hooksDir := filepath.Join(gitDir, "hooks")
	hooks := []string{"pre-commit", "pre-push"}

	var installedHooks []string
	for _, hook := range hooks {
		hookPath := filepath.Join(hooksDir, hook)
		if _, err := os.Stat(hookPath); err == nil {
			installedHooks = append(installedHooks, hook)
		}
	}

	if len(installedHooks) == 0 {
		result.Status = "missing"
		result.Suggestion = "Install git hooks with: make pre-commit-install"
	} else {
		result.Status = "ok"
		result.Details = map[string]interface{}{"installed_hooks": installedHooks}
	}

	report.Results = append(report.Results, result)
}

func generateDevEnvSummary(report *DevEnvReport) {
	requiredTools := 0
	availableTools := 0
	missingRequired := 0

	for _, result := range report.Results {
		if result.Required {
			requiredTools++
			if result.Status == "ok" || result.Status == "found" {
				availableTools++
			} else {
				missingRequired++
			}
		}
	}

	report.IsReady = missingRequired == 0

	if report.IsReady {
		report.Summary = fmt.Sprintf("Development environment is ready (%d/%d required tools available)",
			availableTools, requiredTools)
	} else {
		report.Summary = fmt.Sprintf("Development environment needs setup (%d/%d required tools missing)",
			missingRequired, requiredTools)
	}

	// Generate recommendations
	if missingRequired > 0 {
		report.Recommendations = append(report.Recommendations,
			"Install missing required tools before contributing")
	}

	if len(report.MissingTools) > 0 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Missing tools: %s", strings.Join(report.MissingTools, ", ")))
	}

	// Check for outdated tools
	outdatedTools := []string{}
	for _, result := range report.Results {
		if result.Status == "outdated" {
			outdatedTools = append(outdatedTools, result.Tool)
		}
	}

	if len(outdatedTools) > 0 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Update outdated tools: %s", strings.Join(outdatedTools, ", ")))
	}

	if report.IsReady {
		report.Recommendations = append(report.Recommendations,
			"Development environment is ready - you can start contributing!")
	}
}

func attemptDevEnvFixes(ctx context.Context, report *DevEnvReport, logger logger.CommonLogger) {
	logger.Info("Attempting to fix development environment issues")

	fixCount := 0

	for i := range report.Results {
		result := &report.Results[i]

		if result.Status == "missing" || result.Status == "needs_tidy" {
			if tryFixDevEnvIssue(ctx, result, logger) {
				fixCount++
			}
		}
	}

	logger.Info("Development environment fixes completed", "fixes_applied", fixCount)

	if fixCount > 0 {
		logger.Info("Re-run the command to verify fixes")
	}
}

func tryFixDevEnvIssue(ctx context.Context, result *DevEnvResult, logger logger.CommonLogger) bool {
	logger.Debug("Attempting to fix issue", "tool", result.Tool, "status", result.Status)

	switch result.Tool {
	case "go.mod":
		if result.Status == "needs_tidy" {
			cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
			if err := cmd.Run(); err != nil {
				logger.Warn("Failed to run go mod tidy", "error", err)
				return false
			}
			result.Status = "ok"
			result.Suggestion = ""
			logger.Info("Successfully ran go mod tidy")
			return true
		}

	case "git-hooks":
		if result.Status == "missing" {
			// Try to install pre-commit hooks if Makefile target exists
			if _, err := os.Stat("Makefile"); err == nil {
				cmd := exec.CommandContext(ctx, "make", "pre-commit-install")
				if err := cmd.Run(); err != nil {
					logger.Warn("Failed to install pre-commit hooks", "error", err)
					return false
				}
				result.Status = "ok"
				result.Suggestion = ""
				logger.Info("Successfully installed pre-commit hooks")
				return true
			}
		}
	}

	return false
}

func displayDevEnvResults(report *DevEnvReport, opts devEnvOptions) error {
	// Display header
	logger.SimpleInfo("🔧 Development Environment Report",
		"timestamp", report.Timestamp.Format("2006-01-02 15:04:05"),
		"platform", report.Platform,
		"go_version", report.GoVersion,
	)

	if !report.IsReady {
		logger.SimpleWarn("❌ Environment is not ready for development",
			"missing_required", len(report.MissingTools))
	} else {
		logger.SimpleInfo("✅ Environment is ready for development")
	}

	// Display tool results by category
	categories := map[string][]DevEnvResult{
		"Core Tools":     {},
		"Build Tools":    {},
		"Project Config": {},
		"Optional Tools": {},
	}

	for _, result := range report.Results {
		switch result.Tool {
		case "go", "git", "make":
			categories["Core Tools"] = append(categories["Core Tools"], result)
		case "golangci-lint", "gofumpt", "gci", "gomock":
			categories["Build Tools"] = append(categories["Build Tools"], result)
		case "go.mod", "Makefile", ".golangci.yml", ".pre-commit-config.yaml", "git-hooks":
			categories["Project Config"] = append(categories["Project Config"], result)
		default:
			categories["Optional Tools"] = append(categories["Optional Tools"], result)
		}
	}

	for category, results := range categories {
		if len(results) == 0 {
			continue
		}

		logger.SimpleInfo(fmt.Sprintf("📋 %s", category))

		for _, result := range results {
			statusIcon := getStatusIcon(result.Status)
			requiredText := ""
			if result.Required {
				requiredText = " (required)"
			}

			logger.SimpleInfo(fmt.Sprintf("  %s %s%s", statusIcon, result.Tool, requiredText),
				"status", result.Status,
				"version", result.Version,
			)

			if opts.showDetails && result.Details != nil {
				for key, value := range result.Details {
					logger.SimpleInfo(fmt.Sprintf("    %s: %v", key, value))
				}
			}

			if !opts.checkOnly && result.Suggestion != "" {
				logger.SimpleInfo(fmt.Sprintf("    💡 %s", result.Suggestion))
			}
		}
	}

	// Display summary
	logger.SimpleInfo("📊 Summary", "summary", report.Summary)

	// Display recommendations
	if len(report.Recommendations) > 0 {
		logger.SimpleInfo("💡 Recommendations:")
		for i, rec := range report.Recommendations {
			logger.SimpleInfo(fmt.Sprintf("  %d. %s", i+1, rec))
		}
	}

	return nil
}

func getStatusIcon(status string) string {
	switch status {
	case "ok", "found":
		return "✅"
	case "missing":
		return "❌"
	case "outdated", "needs_tidy", "incomplete":
		return "⚠️"
	case "error":
		return "🔴"
	case "alternative":
		return "🔄"
	default:
		return "❓"
	}
}

func isGoVersionOutdated(version string) bool {
	// Simple check - consider Go 1.19 and below as outdated
	outdatedVersions := []string{"go1.16", "go1.17", "go1.18", "go1.19"}
	for _, old := range outdatedVersions {
		if strings.Contains(version, old) {
			return true
		}
	}
	return false
}
