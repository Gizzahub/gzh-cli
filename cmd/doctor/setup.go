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

	"github.com/gizzahub/gzh-cli/internal/cli"
	"github.com/gizzahub/gzh-cli/internal/logger"
)

// SetupResult represents a setup step result.
type SetupResult struct {
	Step     string        `json:"step"`
	Status   string        `json:"status"` // "success", "failed", "skipped"
	Message  string        `json:"message"`
	Duration time.Duration `json:"duration"`
	Command  string        `json:"command,omitempty"`
	Output   string        `json:"output,omitempty"`
	Error    string        `json:"error,omitempty"`
}

// SetupReport represents the complete setup report.
type SetupReport struct {
	Timestamp    time.Time     `json:"timestamp"`
	SetupType    string        `json:"setup_type"`
	Platform     string        `json:"platform"`
	Results      []SetupResult `json:"results"`
	Summary      string        `json:"summary"`
	Success      bool          `json:"success"`
	TotalSteps   int           `json:"total_steps"`
	SuccessSteps int           `json:"success_steps"`
	FailedSteps  int           `json:"failed_steps"`
	SkippedSteps int           `json:"skipped_steps"`
	Duration     time.Duration `json:"duration"`
}

// newSetupCmd creates the setup subcommand for automated development environment setup.
func newSetupCmd() *cobra.Command {
	ctx := context.Background()

	var (
		setupType    string
		force        bool
		skipOptional bool
		dryRun       bool
	)

	cmd := cli.NewCommandBuilder(ctx, "setup", "Automated development environment setup").
		WithLongDescription(`Automated setup for GZH Manager development environment.

This command provides one-click setup for different development scenarios:
- 'dev': Complete development environment setup
- 'contributor': New contributor onboarding setup
- 'ci': CI/CD environment setup
- 'tools': Install only development tools

Features:
- Automatic tool installation and configuration
- Platform-specific setup (Linux, macOS, Windows)
- Interactive setup wizard
- Dependency verification and validation
- Git hooks and pre-commit setup

Examples:
  gz doctor setup dev                      # Full development setup
  gz doctor setup contributor              # New contributor setup
  gz doctor setup --type tools            # Install development tools only
  gz doctor setup --dry-run               # Preview setup steps
  gz doctor setup --force                 # Force reinstall existing tools`).
		WithExample("gz doctor setup dev --force").
		WithFormatFlag("table", []string{"table", "json", "yaml"}).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			// Determine setup type from argument or flag
			selectedSetupType := setupType
			if len(args) > 0 {
				selectedSetupType = args[0]
			}
			if selectedSetupType == "" {
				selectedSetupType = "dev" // Default to full development setup
			}

			return runAutomatedSetup(ctx, flags, setupOptions{
				setupType:    selectedSetupType,
				force:        force,
				skipOptional: skipOptional,
				dryRun:       dryRun,
			})
		}).
		Build()

	cmd.Flags().StringVar(&setupType, "type", "", "Setup type: dev, contributor, ci, tools")
	cmd.Flags().BoolVar(&force, "force", false, "Force reinstall existing tools and configurations")
	cmd.Flags().BoolVar(&skipOptional, "skip-optional", false, "Skip optional tools and configurations")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview setup steps without executing")

	return cmd
}

type setupOptions struct {
	setupType    string
	force        bool
	skipOptional bool
	dryRun       bool
}

func runAutomatedSetup(ctx context.Context, flags *cli.CommonFlags, opts setupOptions) error {
	logger := logger.NewSimpleLogger("doctor-setup")

	startTime := time.Now()

	logger.Info("Starting automated development environment setup",
		"setup_type", opts.setupType,
		"platform", runtime.GOOS,
		"dry_run", opts.dryRun,
	)

	report := &SetupReport{
		Timestamp: startTime,
		SetupType: opts.setupType,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Results:   []SetupResult{},
	}

	// Define setup steps based on type
	var setupSteps []SetupStep

	switch opts.setupType {
	case "dev", "development":
		setupSteps = getDevSetupSteps()
	case "contributor":
		setupSteps = getContributorSetupSteps()
	case "ci":
		setupSteps = getCISetupSteps()
	case "tools":
		setupSteps = getToolsSetupSteps()
	default:
		return fmt.Errorf("unknown setup type: %s", opts.setupType)
	}

	if opts.dryRun {
		logger.Info("Dry run mode - showing setup steps without execution")
		return displaySetupSteps(setupSteps, flags)
	}

	// Execute setup steps
	logger.Info("Executing setup steps", "total_steps", len(setupSteps))

	for i, step := range setupSteps {
		logger.Info(fmt.Sprintf("Step %d/%d: %s", i+1, len(setupSteps), step.Name))

		result := executeSetupStep(ctx, step, opts, logger)
		report.Results = append(report.Results, result)

		// Update counters
		switch result.Status {
		case "success":
			report.SuccessSteps++
		case "failed":
			report.FailedSteps++
		case "skipped":
			report.SkippedSteps++
		}

		// Stop on critical failures unless force mode
		if result.Status == "failed" && step.Critical && !opts.force {
			logger.Error("Critical setup step failed, stopping", "step", step.Name)
			break
		}
	}

	// Generate summary
	report.TotalSteps = len(setupSteps)
	report.Duration = time.Since(startTime)
	report.Success = report.FailedSteps == 0

	if report.Success {
		report.Summary = fmt.Sprintf("Setup completed successfully (%d/%d steps)",
			report.SuccessSteps, report.TotalSteps)
	} else {
		report.Summary = fmt.Sprintf("Setup completed with issues (%d succeeded, %d failed, %d skipped)",
			report.SuccessSteps, report.FailedSteps, report.SkippedSteps)
	}

	// Generate output
	formatter := cli.NewOutputFormatter(flags.Format)

	switch flags.Format {
	case "json", "yaml":
		return formatter.FormatOutput(report)
	default:
		return displaySetupReport(report)
	}
}

// SetupStep represents a single setup step.
type SetupStep struct {
	Name        string
	Description string
	Command     string
	Args        []string
	Critical    bool
	Optional    bool
	Platform    string                     // "all", "linux", "darwin", "windows"
	Condition   func(context.Context) bool // Optional condition check
	PostCheck   func(context.Context) bool // Optional verification after execution
}

func getDevSetupSteps() []SetupStep {
	return []SetupStep{
		{
			Name:        "Install Go Tools",
			Description: "Install essential Go development tools",
			Command:     "go",
			Args:        []string{"install", "golang.org/x/tools/...@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install golangci-lint",
			Description: "Install golangci-lint for code linting",
			Command:     "go",
			Args:        []string{"install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
			Critical:    true,
			Platform:    "all",
			PostCheck: func(ctx context.Context) bool {
				_, err := exec.LookPath("golangci-lint")
				return err == nil
			},
		},
		{
			Name:        "Install gofumpt",
			Description: "Install gofumpt for code formatting",
			Command:     "go",
			Args:        []string{"install", "mvdan.cc/gofumpt@latest"},
			Critical:    true,
			Platform:    "all",
			PostCheck: func(ctx context.Context) bool {
				_, err := exec.LookPath("gofumpt")
				return err == nil
			},
		},
		{
			Name:        "Install gci",
			Description: "Install gci for import sorting",
			Command:     "go",
			Args:        []string{"install", "github.com/daixiang0/gci@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install gomock",
			Description: "Install gomock for generating mocks",
			Command:     "go",
			Args:        []string{"install", "go.uber.org/mock/mockgen@latest"},
			Critical:    false,
			Platform:    "all",
		},
		{
			Name:        "Setup Go Modules",
			Description: "Ensure Go modules are properly configured",
			Command:     "go",
			Args:        []string{"mod", "tidy"},
			Critical:    true,
			Platform:    "all",
			Condition: func(ctx context.Context) bool {
				_, err := os.Stat("go.mod")
				return err == nil
			},
		},
		{
			Name:        "Install Pre-commit",
			Description: "Install pre-commit hooks system",
			Command:     getPlatformPackageManager(),
			Args:        getPlatformPreCommitInstallArgs(),
			Critical:    false,
			Platform:    "all",
			Condition: func(ctx context.Context) bool {
				_, err := exec.LookPath("pre-commit")
				return err != nil // Install only if not present
			},
		},
		{
			Name:        "Setup Pre-commit Hooks",
			Description: "Install and configure pre-commit hooks",
			Command:     "make",
			Args:        []string{"pre-commit-install"},
			Critical:    false,
			Platform:    "all",
			Condition: func(ctx context.Context) bool {
				_, err := os.Stat("Makefile")
				return err == nil
			},
		},
		{
			Name:        "Verify Build System",
			Description: "Verify that the project builds successfully",
			Command:     "make",
			Args:        []string{"build"},
			Critical:    true,
			Platform:    "all",
			Condition: func(ctx context.Context) bool {
				_, err := os.Stat("Makefile")
				return err == nil
			},
		},
	}
}

func getContributorSetupSteps() []SetupStep {
	steps := getDevSetupSteps()

	// Add contributor-specific steps
	contributorSteps := []SetupStep{
		{
			Name:        "Configure Git",
			Description: "Configure Git for contributions",
			Command:     "git",
			Args:        []string{"config", "--global", "init.defaultBranch", "main"},
			Critical:    false,
			Platform:    "all",
		},
		{
			Name:        "Create Developer Config",
			Description: "Create local development configuration",
			Critical:    false,
			Platform:    "all",
			Condition: func(ctx context.Context) bool {
				homeDir, _ := os.UserHomeDir()
				configDir := filepath.Join(homeDir, ".config", "gzh-manager")
				_, err := os.Stat(configDir)
				return os.IsNotExist(err) // Create only if not exists
			},
		},
	}

	return append(steps, contributorSteps...)
}

func getCISetupSteps() []SetupStep {
	return []SetupStep{
		{
			Name:        "Install golangci-lint",
			Description: "Install golangci-lint for CI",
			Command:     "go",
			Args:        []string{"install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install build tools",
			Description: "Install essential build tools",
			Command:     "go",
			Args:        []string{"install", "mvdan.cc/gofumpt@latest", "github.com/daixiang0/gci@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Verify Dependencies",
			Description: "Download and verify all dependencies",
			Command:     "go",
			Args:        []string{"mod", "download"},
			Critical:    true,
			Platform:    "all",
		},
	}
}

func getToolsSetupSteps() []SetupStep {
	return []SetupStep{
		{
			Name:        "Install golangci-lint",
			Description: "Install golangci-lint",
			Command:     "go",
			Args:        []string{"install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install gofumpt",
			Description: "Install gofumpt",
			Command:     "go",
			Args:        []string{"install", "mvdan.cc/gofumpt@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install gci",
			Description: "Install gci",
			Command:     "go",
			Args:        []string{"install", "github.com/daixiang0/gci@latest"},
			Critical:    true,
			Platform:    "all",
		},
		{
			Name:        "Install gomock",
			Description: "Install gomock",
			Command:     "go",
			Args:        []string{"install", "go.uber.org/mock/mockgen@latest"},
			Critical:    false,
			Platform:    "all",
		},
	}
}

func executeSetupStep(ctx context.Context, step SetupStep, opts setupOptions, logger logger.CommonLogger) SetupResult {
	startTime := time.Now()

	result := SetupResult{
		Step:    step.Name,
		Status:  "success",
		Message: step.Description,
	}

	// Check platform compatibility
	if step.Platform != "all" && step.Platform != runtime.GOOS {
		result.Status = "skipped"
		result.Message = fmt.Sprintf("Skipped (platform %s not supported)", runtime.GOOS)
		result.Duration = time.Since(startTime)
		return result
	}

	// Check optional steps
	if step.Optional && opts.skipOptional {
		result.Status = "skipped"
		result.Message = "Skipped (optional step)"
		result.Duration = time.Since(startTime)
		return result
	}

	// Check condition
	if step.Condition != nil && !step.Condition(ctx) {
		result.Status = "skipped"
		result.Message = "Skipped (condition not met)"
		result.Duration = time.Since(startTime)
		return result
	}

	// Handle special cases
	if step.Command == "" {
		// Custom setup step
		return executeCustomSetupStep(ctx, step, opts, logger)
	}

	// Execute command
	result.Command = fmt.Sprintf("%s %s", step.Command, strings.Join(step.Args, " "))

	logger.Debug("Executing setup command", "command", result.Command)

	cmd := exec.CommandContext(ctx, step.Command, step.Args...)
	output, err := cmd.CombinedOutput()

	result.Output = string(output)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.Message = fmt.Sprintf("Failed: %s", err.Error())
		logger.Warn("Setup step failed", "step", step.Name, "error", err)
		return result
	}

	// Run post-check if available
	if step.PostCheck != nil && !step.PostCheck(ctx) {
		result.Status = "failed"
		result.Message = "Post-check failed"
		logger.Warn("Setup step post-check failed", "step", step.Name)
		return result
	}

	result.Message = "Completed successfully"
	logger.Info("Setup step completed", "step", step.Name, "duration", result.Duration)

	return result
}

func executeCustomSetupStep(ctx context.Context, step SetupStep, opts setupOptions, logger logger.CommonLogger) SetupResult {
	startTime := time.Now()

	result := SetupResult{
		Step:    step.Name,
		Status:  "success",
		Message: step.Description,
	}

	switch step.Name {
	case "Create Developer Config":
		homeDir, err := os.UserHomeDir()
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			result.Message = "Failed to get home directory"
			return result
		}

		configDir := filepath.Join(homeDir, ".config", "gzh-manager")
		if err := os.MkdirAll(configDir, 0o755); err != nil {
			result.Status = "failed"
			result.Error = err.Error()
			result.Message = "Failed to create config directory"
			return result
		}

		result.Message = fmt.Sprintf("Created config directory: %s", configDir)

	default:
		result.Status = "skipped"
		result.Message = "Unknown custom step"
	}

	result.Duration = time.Since(startTime)
	return result
}

func getPlatformPackageManager() string {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			return "brew"
		}
		return "pip3"
	case "linux":
		if _, err := exec.LookPath("apt"); err == nil {
			return "apt"
		} else if _, err := exec.LookPath("yum"); err == nil {
			return "yum"
		} else if _, err := exec.LookPath("pacman"); err == nil {
			return "pacman"
		}
		return "pip3"
	case "windows":
		if _, err := exec.LookPath("choco"); err == nil {
			return "choco"
		}
		return "pip"
	default:
		return "pip3"
	}
}

func getPlatformPreCommitInstallArgs() []string {
	packageManager := getPlatformPackageManager()

	switch packageManager {
	case "brew":
		return []string{"install", "pre-commit"}
	case "apt":
		return []string{"install", "-y", "pre-commit"}
	case "yum":
		return []string{"install", "-y", "pre-commit"}
	case "pacman":
		return []string{"-S", "--noconfirm", "pre-commit"}
	case "choco":
		return []string{"install", "pre-commit"}
	default:
		return []string{"install", "pre-commit"}
	}
}

func displaySetupSteps(steps []SetupStep, flags *cli.CommonFlags) error {
	logger.SimpleInfo("🔍 Setup Steps Preview")

	for i, step := range steps {
		statusIcon := "📋"
		if step.Critical {
			statusIcon = "⚠️"
		}
		if step.Optional {
			statusIcon = "⭐"
		}

		logger.SimpleInfo(fmt.Sprintf("  Step %d: %s %s", i+1, statusIcon, step.Name),
			"description", step.Description,
			"platform", step.Platform,
		)

		if step.Command != "" {
			logger.SimpleInfo(fmt.Sprintf("    Command: %s %s", step.Command, strings.Join(step.Args, " ")))
		}
	}

	logger.SimpleInfo("📊 Summary",
		"total_steps", len(steps),
		"critical_steps", countStepsByType(steps, true, false),
		"optional_steps", countStepsByType(steps, false, true),
	)

	return nil
}

func displaySetupReport(report *SetupReport) error {
	// Display header
	logger.SimpleInfo("🚀 Setup Report",
		"setup_type", report.SetupType,
		"platform", report.Platform,
		"duration", report.Duration.String(),
	)

	if report.Success {
		logger.SimpleInfo("✅ Setup completed successfully",
			"success_rate", fmt.Sprintf("%d/%d", report.SuccessSteps, report.TotalSteps))
	} else {
		logger.SimpleWarn("⚠️ Setup completed with issues",
			"succeeded", report.SuccessSteps,
			"failed", report.FailedSteps,
			"skipped", report.SkippedSteps,
		)
	}

	// Display step results
	logger.SimpleInfo("📋 Step Results:")

	for _, result := range report.Results {
		statusIcon := getSetupStatusIcon(result.Status)

		logger.SimpleInfo(fmt.Sprintf("  %s %s", statusIcon, result.Step),
			"status", result.Status,
			"duration", result.Duration.String(),
			"message", result.Message,
		)

		if result.Error != "" {
			logger.SimpleWarn(fmt.Sprintf("    Error: %s", result.Error))
		}
	}

	// Display summary
	logger.SimpleInfo("📊 Summary", "summary", report.Summary)

	// Display next steps
	if report.Success {
		logger.SimpleInfo("🎉 Next Steps:")
		logger.SimpleInfo("  1. Run 'gz doctor dev-env' to verify your setup")
		logger.SimpleInfo("  2. Try building the project with 'make build'")
		logger.SimpleInfo("  3. Run tests with 'make test'")
		logger.SimpleInfo("  4. Start contributing! 🚀")
	} else {
		logger.SimpleWarn("🔧 Recommended Actions:")
		logger.SimpleInfo("  1. Review failed steps above")
		logger.SimpleInfo("  2. Install missing dependencies manually")
		logger.SimpleInfo("  3. Re-run setup with --force flag")
		logger.SimpleInfo("  4. Check the project documentation")
	}

	return nil
}

func getSetupStatusIcon(status string) string {
	switch status {
	case "success":
		return "✅"
	case "failed":
		return "❌"
	case "skipped":
		return "⏭️"
	default:
		return "❓"
	}
}

func countStepsByType(steps []SetupStep, critical, optional bool) int {
	count := 0
	for _, step := range steps {
		if critical && step.Critical {
			count++
		} else if optional && step.Optional {
			count++
		}
	}
	return count
}
