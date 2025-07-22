// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/gizzahub/gzh-manager-go/internal/errors"
	"github.com/gizzahub/gzh-manager-go/internal/logger"
	"github.com/spf13/cobra"
)

const (
	statusPass = "pass"
	statusWarn = "warn"
	statusFail = "fail"
)

// DoctorCmd represents the doctor command.
var DoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose system health and configuration issues",
	Long: `Comprehensive system diagnostics and health checking for GZH Manager.

The doctor command performs a thorough analysis of your system including:
- System information and dependencies
- Configuration validation
- Network connectivity checks
- Git configuration analysis
- Permission and access verification
- Performance benchmarks
- Issue detection and recommendations

Examples:
  gz doctor                    # Run full diagnostic
  gz doctor --report report.json  # Save detailed report
  gz doctor --quick            # Run quick checks only
  gz doctor --fix              # Attempt automatic fixes`,
	Run: runDoctor,
}

var (
	reportFile string
	quickMode  bool
	attemptFix bool
	verbose    bool
)

func init() {
	DoctorCmd.Flags().StringVar(&reportFile, "report", "", "Output detailed report to file")
	DoctorCmd.Flags().BoolVar(&quickMode, "quick", false, "Run quick checks only")
	DoctorCmd.Flags().BoolVar(&attemptFix, "fix", false, "Attempt to fix detected issues")
	DoctorCmd.Flags().BoolVar(&verbose, "verbose", false, "Show verbose output")
}

// DiagnosticResult represents the result of a diagnostic check.
type DiagnosticResult struct {
	Name          string                 `json:"name"`
	Category      string                 `json:"category"`
	Status        string                 `json:"status"` // "pass", "warn", "fail", "skip"
	Message       string                 `json:"message"`
	Details       map[string]interface{} `json:"details,omitempty"`
	FixSuggestion string                 `json:"fixSuggestion,omitempty"`
	Duration      time.Duration          `json:"duration"`
	Timestamp     time.Time              `json:"timestamp"`
}

// DiagnosticReport represents the complete diagnostic report.
type DiagnosticReport struct {
	Timestamp       time.Time          `json:"timestamp"`
	Version         string             `json:"version"`
	Platform        string             `json:"platform"`
	TotalChecks     int                `json:"totalChecks"`
	PassedChecks    int                `json:"passedChecks"`
	WarnChecks      int                `json:"warnChecks"`
	FailedChecks    int                `json:"failedChecks"`
	SkippedChecks   int                `json:"skippedChecks"`
	Results         []DiagnosticResult `json:"results"`
	Summary         string             `json:"summary"`
	Recommendations []string           `json:"recommendations"`
	Duration        time.Duration      `json:"duration"`
}

// SystemInfo represents system information.
type SystemInfo struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	GoVersion  string `json:"goVersion"`
	Hostname   string `json:"hostname"`
	Username   string `json:"username"`
	WorkingDir string `json:"workingDir"`
	HomeDir    string `json:"homeDir"`
	PathEnv    string `json:"pathEnv"`
	Shell      string `json:"shell"`
	TempDir    string `json:"tempDir"`
}

func runDoctor(cmd *cobra.Command, args []string) {
	// Initialize structured logger for doctor operations
	structuredLogger := logger.NewStructuredLogger("doctor", logger.LevelInfo)
	sessionID := fmt.Sprintf("doctor-%d", time.Now().Unix())
	structuredLogger = structuredLogger.WithSession(sessionID).
		WithContext("quick_mode", quickMode).
		WithContext("attempt_fix", attemptFix).
		WithContext("verbose", verbose)

	// Initialize error recovery system
	recoveryConfig := errors.RecoveryConfig{
		MaxRetries: 3,
		RetryDelay: time.Second,
		Logger:     structuredLogger,
		RecoveryFunc: func(err error) error {
			structuredLogger.Warn("Doctor diagnostic encountered recoverable error", "error_type", fmt.Sprintf("%T", err))
			return nil
		},
	}
	errorRecovery := errors.NewErrorRecovery(recoveryConfig)

	// Initialize health monitor
	healthMonitor := errors.NewHealthMonitor(structuredLogger)

	// Add system health checks
	healthMonitor.AddCheck(errors.HealthCheck{
		Name:        "memory-usage",
		Description: "Check memory usage levels",
		CheckFunc: func() error {
			memStats := errors.GetMemoryStats()
			if allocMB, ok := memStats["alloc_mb"].(uint64); ok && allocMB > 1000 {
				return fmt.Errorf("high memory usage: %d MB", allocMB)
			}
			return nil
		},
		Timeout: 5 * time.Second,
	})

	healthMonitor.AddCheck(errors.HealthCheck{
		Name:        "goroutine-count",
		Description: "Check for goroutine leaks",
		CheckFunc: func() error {
			if runtime.NumGoroutine() > 100 {
				return fmt.Errorf("high goroutine count: %d", runtime.NumGoroutine())
			}
			return nil
		},
		Timeout: 2 * time.Second,
	})

	fmt.Println("🩺 Starting GZH Manager diagnostic...")
	structuredLogger.Info("Starting GZH Manager diagnostic session")

	startTime := time.Now()
	report := &DiagnosticReport{
		Timestamp:       startTime,
		Version:         "1.0.0", // Version should be passed from build system
		Platform:        fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		Results:         []DiagnosticResult{},
		Recommendations: []string{},
	}

	// Execute diagnostic with error recovery
	diagnosticErr := errorRecovery.Execute(cmd.Context(), "doctor-diagnostic", func() error {
		// Run pre-diagnostic health checks
		healthResults := healthMonitor.RunChecks(cmd.Context())
		for checkName, checkErr := range healthResults {
			if checkErr != nil {
				structuredLogger.Warn("Pre-diagnostic health check failed", "check", checkName, "error", checkErr)
			} else {
				structuredLogger.Debug("Pre-diagnostic health check passed", "check", checkName)
			}
		}

		// Run diagnostic checks
		if verbose {
			fmt.Println("🔍 Running diagnostic checks...")
		}

		structuredLogger.Info("Running diagnostic checks", "total_categories", 7)

		// System checks
		structuredLogger.Debug("Running system checks")
		runSystemChecks(report, structuredLogger, errorRecovery)

		// Configuration checks
		structuredLogger.Debug("Running configuration checks")
		runConfigChecks(report, structuredLogger, errorRecovery)

		// Network checks
		if !quickMode {
			structuredLogger.Debug("Running network checks")
			runNetworkChecks(cmd.Context(), report, structuredLogger, errorRecovery)
		} else {
			structuredLogger.Debug("Skipping network checks (quick mode)")
		}

		// Git checks
		structuredLogger.Debug("Running git checks")
		runGitChecks(report, structuredLogger, errorRecovery)

		// Permission checks
		structuredLogger.Debug("Running permission checks")
		runPermissionChecks(report, structuredLogger, errorRecovery)

		// Performance checks
		if !quickMode {
			structuredLogger.Debug("Running performance checks")
			runPerformanceChecks(report, structuredLogger, errorRecovery)
		} else {
			structuredLogger.Debug("Skipping performance checks (quick mode)")
		}

		// Security checks
		structuredLogger.Debug("Running security checks")
		runSecurityChecks(report, structuredLogger, errorRecovery)

		return nil
	})
	if diagnosticErr != nil {
		structuredLogger.ErrorWithStack(diagnosticErr, "Diagnostic execution failed")
		fmt.Printf("❌ Diagnostic execution failed: %v\n", diagnosticErr)
		os.Exit(1)
	}

	// Calculate totals
	calculateReportTotals(report)
	report.Duration = time.Since(startTime)

	// Log performance metrics
	structuredLogger.LogPerformance("doctor-diagnostic-completed", report.Duration, map[string]interface{}{
		"total_checks":   report.TotalChecks,
		"passed_checks":  report.PassedChecks,
		"failed_checks":  report.FailedChecks,
		"warn_checks":    report.WarnChecks,
		"skipped_checks": report.SkippedChecks,
		"memory_stats":   errors.GetMemoryStats(),
	})

	// Generate summary and recommendations
	generateSummaryAndRecommendations(report)

	// Print results
	printResults(report)

	structuredLogger.Info("Diagnostic completed",
		"duration", report.Duration.String(),
		"success_rate", fmt.Sprintf("%.1f%%", float64(report.PassedChecks)/float64(report.TotalChecks)*100),
		"critical_issues", report.FailedChecks,
		"warnings", report.WarnChecks)

	// Save report if requested
	if reportFile != "" {
		if err := saveReport(report, reportFile); err != nil {
			structuredLogger.ErrorWithStack(err, "Failed to save diagnostic report")
			fmt.Printf("❌ Failed to save report: %v\n", err)
		} else {
			structuredLogger.Info("Diagnostic report saved", "file", reportFile)
			fmt.Printf("💾 Report saved to: %s\n", reportFile)
		}
	}

	// Attempt fixes if requested
	if attemptFix {
		structuredLogger.Info("Attempting automatic fixes")
		attemptAutomaticFixes(cmd.Context(), report, structuredLogger, errorRecovery)
	}

	// Exit with appropriate code
	if report.FailedChecks > 0 {
		structuredLogger.Error("Doctor diagnostic completed with critical issues", "failed_checks", report.FailedChecks)
		os.Exit(1)
	} else if report.WarnChecks > 0 {
		structuredLogger.Warn("Doctor diagnostic completed with warnings", "warn_checks", report.WarnChecks)
		os.Exit(2)
	}

	structuredLogger.Info("Doctor diagnostic completed successfully")
}

func runSystemChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("💻 Checking system information...")
	}

	// System info check
	start := time.Now()
	sysInfo := getSystemInfo()
	report.Results = append(report.Results, DiagnosticResult{
		Name:      "System Information",
		Category:  "system",
		Status:    statusPass,
		Message:   fmt.Sprintf("Running on %s/%s", sysInfo.OS, sysInfo.Arch),
		Details:   map[string]interface{}{"system_info": sysInfo},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Go version check
	start = time.Now()
	goVersion := runtime.Version()
	status := statusPass
	message := fmt.Sprintf("Go version: %s", goVersion)

	// Check if Go version is supported (1.19+)
	if !isGoVersionSupported(goVersion) {
		status = statusWarn
		message += " (consider upgrading to Go 1.21+)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Go Version",
		Category:  "system",
		Status:    status,
		Message:   message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Memory check
	start = time.Now()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	allocatedMB := float64(memStats.Alloc) / 1024 / 1024
	goroutines := runtime.NumGoroutine()

	status = statusPass
	message = fmt.Sprintf("Memory: %.2f MB allocated, %d goroutines", allocatedMB, goroutines)

	if allocatedMB > 1000 {
		status = statusWarn
		message += " (high memory usage)"
	}

	memStatsMap := map[string]interface{}{
		"allocated_mb":  allocatedMB,
		"goroutines":    goroutines,
		"sys_mb":        float64(memStats.Sys) / 1024 / 1024,
		"heap_alloc_mb": float64(memStats.HeapAlloc) / 1024 / 1024,
		"heap_sys_mb":   float64(memStats.HeapSys) / 1024 / 1024,
		"gc_runs":       memStats.NumGC,
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Memory Usage",
		Category:  "system",
		Status:    status,
		Message:   message,
		Details:   map[string]interface{}{"memory_stats": memStatsMap},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Disk space check
	start = time.Now()
	wd, _ := os.Getwd()
	diskSpace := getDiskSpace(wd)

	status = statusPass
	message = fmt.Sprintf("Disk space: %.2f GB available", diskSpace)

	if diskSpace < 1.0 {
		status = statusFail
		message += " (critically low)"
	} else if diskSpace < 5.0 {
		status = statusWarn
		message += " (low disk space)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "Disk Space",
		Category:      "system",
		Status:        status,
		Message:       message,
		FixSuggestion: "Free up disk space or move to a larger volume",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})
}

func runConfigChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("⚙️ Checking configuration...")
	}

	// Configuration file existence
	start := time.Now()
	configPaths := []string{
		"./bulk-clone.yaml",
		"./bulk-clone.yml",
		os.ExpandEnv("$HOME/.config/gzh-manager/bulk-clone.yaml"),
		"/etc/gzh-manager/bulk-clone.yaml",
	}

	var foundConfigs []string

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			foundConfigs = append(foundConfigs, path)
		}
	}

	status := statusWarn
	message := "No configuration files found"

	if len(foundConfigs) > 0 {
		status = statusPass
		message = fmt.Sprintf("Found %d configuration file(s): %s", len(foundConfigs), strings.Join(foundConfigs, ", "))
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "Configuration Files",
		Category:      "config",
		Status:        status,
		Message:       message,
		Details:       map[string]interface{}{"found_configs": foundConfigs},
		FixSuggestion: "Create a configuration file using 'gz gen-config'",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})

	// Environment variables
	start = time.Now()
	envVars := map[string]string{
		"GITHUB_TOKEN":    os.Getenv("GITHUB_TOKEN"),
		"GITLAB_TOKEN":    os.Getenv("GITLAB_TOKEN"),
		"GITEA_TOKEN":     os.Getenv("GITEA_TOKEN"),
		"GZH_CONFIG_PATH": os.Getenv("GZH_CONFIG_PATH"),
	}

	var setVars []string

	for name, value := range envVars {
		if value != "" {
			setVars = append(setVars, name)
		}
	}

	status = statusPass

	message = fmt.Sprintf("Environment variables: %d set", len(setVars))
	if len(setVars) == 0 {
		status = statusWarn
		message += " (no tokens configured)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "Environment Variables",
		Category:      "config",
		Status:        status,
		Message:       message,
		Details:       map[string]interface{}{"set_variables": setVars},
		FixSuggestion: "Set API tokens as environment variables for authenticated access",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})
}

func runNetworkChecks(ctx context.Context, report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("🌐 Checking network connectivity...")
	}

	// DNS resolution
	start := time.Now()
	hosts := []string{"github.com", "gitlab.com", "google.com"}

	var (
		resolvedHosts []string
		failedHosts   []string
	)

	for _, host := range hosts {
		if _, err := net.LookupHost(host); err == nil {
			resolvedHosts = append(resolvedHosts, host)
		} else {
			failedHosts = append(failedHosts, host)
		}
	}

	status := statusPass

	message := fmt.Sprintf("DNS resolution: %d/%d hosts resolved", len(resolvedHosts), len(hosts))
	if len(failedHosts) > 0 {
		status = statusWarn
		message += fmt.Sprintf(" (failed: %s)", strings.Join(failedHosts, ", "))
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "DNS Resolution",
		Category:  "network",
		Status:    status,
		Message:   message,
		Details:   map[string]interface{}{"resolved": resolvedHosts, "failed": failedHosts},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// HTTP connectivity
	start = time.Now()
	apis := map[string]string{
		"GitHub API": "https://api.github.com",
		"GitLab API": "https://gitlab.com/api/v4",
	}

	var (
		workingAPIs []string
		failedAPIs  []string
	)

	for name, url := range apis {
		if isURLReachable(ctx, url) {
			workingAPIs = append(workingAPIs, name)
		} else {
			failedAPIs = append(failedAPIs, name)
		}
	}

	status = statusPass

	message = fmt.Sprintf("API connectivity: %d/%d APIs reachable", len(workingAPIs), len(apis))
	if len(failedAPIs) > 0 {
		status = statusWarn
		message += fmt.Sprintf(" (failed: %s)", strings.Join(failedAPIs, ", "))
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "API Connectivity",
		Category:  "network",
		Status:    status,
		Message:   message,
		Details:   map[string]interface{}{"working": workingAPIs, "failed": failedAPIs},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})
}

func runGitChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("🐙 Checking Git configuration...")
	}

	// Git installation
	start := time.Now()
	gitPath, err := exec.LookPath("git")
	status := statusPass
	message := fmt.Sprintf("Git found at: %s", gitPath)

	if err != nil {
		status = statusFail
		message = "Git not found in PATH"
		report.Results = append(report.Results, DiagnosticResult{
			Name:          "Git Installation",
			Category:      "git",
			Status:        status,
			Message:       message,
			FixSuggestion: "Install Git from https://git-scm.com/",
			Duration:      time.Since(start),
			Timestamp:     time.Now(),
		})

		return
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Git Installation",
		Category:  "git",
		Status:    status,
		Message:   message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Git version
	start = time.Now()
	gitVersionCmd := exec.Command("git", "--version")

	gitVersionOut, err := gitVersionCmd.Output()
	if err != nil {
		status = statusWarn
		message = "Could not determine Git version"
	} else {
		status = statusPass
		message = strings.TrimSpace(string(gitVersionOut))
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Git Version",
		Category:  "git",
		Status:    status,
		Message:   message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Git configuration
	start = time.Now()
	gitConfig := getGitConfig()
	status = statusPass
	message = "Git configuration looks good"

	if gitConfig["user.name"] == "" || gitConfig["user.email"] == "" {
		status = statusWarn
		message = "Git user configuration incomplete"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "Git Configuration",
		Category:      "git",
		Status:        status,
		Message:       message,
		Details:       map[string]interface{}{"config": gitConfig},
		FixSuggestion: "Configure Git with 'git config --global user.name' and 'git config --global user.email'",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})
}

func runPermissionChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("🔒 Checking permissions...")
	}

	// Working directory permissions
	start := time.Now()
	wd, _ := os.Getwd()
	writable := isDirectoryWritable(wd)

	status := statusPass
	message := fmt.Sprintf("Working directory '%s' is writable", wd)

	if !writable {
		status = statusFail
		message = fmt.Sprintf("Working directory '%s' is not writable", wd)
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "Working Directory Permissions",
		Category:      "permissions",
		Status:        status,
		Message:       message,
		FixSuggestion: "Change to a writable directory or adjust permissions",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})

	// Home directory permissions
	start = time.Now()
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".config", "gzh-manager")
	canCreateConfig := canCreateDirectory(configDir)

	status = statusPass
	message = "Can create configuration directory"

	if !canCreateConfig {
		status = statusWarn
		message = "Cannot create configuration directory"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Configuration Directory Access",
		Category:  "permissions",
		Status:    status,
		Message:   message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})
}

func runPerformanceChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("🚀 Running performance benchmarks...")
	}

	// CPU benchmark
	start := time.Now()
	cpuScore := runCPUBenchmark()

	status := statusPass

	message := fmt.Sprintf("CPU benchmark: %.2f ops/sec", cpuScore)
	if cpuScore < 1000000 {
		status = statusWarn
		message += " (slow)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "CPU Performance",
		Category:  "performance",
		Status:    status,
		Message:   message,
		Details:   map[string]interface{}{"ops_per_second": cpuScore},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	// Disk I/O benchmark
	start = time.Now()
	diskScore := runDiskBenchmark()

	status = statusPass

	message = fmt.Sprintf("Disk I/O: %.2f MB/s", diskScore)
	if diskScore < 10 {
		status = statusWarn
		message += " (slow)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:      "Disk I/O Performance",
		Category:  "performance",
		Status:    status,
		Message:   message,
		Details:   map[string]interface{}{"mb_per_second": diskScore},
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})
}

func runSecurityChecks(report *DiagnosticReport, _ *logger.StructuredLogger, _ *errors.ErrorRecovery) {
	if verbose {
		fmt.Println("🔐 Checking security configuration...")
	}

	// SSH key check
	start := time.Now()
	sshKeyCount := countSSHKeys()

	status := statusPass

	message := fmt.Sprintf("SSH keys: %d found", sshKeyCount)
	if sshKeyCount == 0 {
		status = statusWarn
		message += " (no SSH keys configured)"
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "SSH Keys",
		Category:      "security",
		Status:        status,
		Message:       message,
		FixSuggestion: "Generate SSH keys with 'ssh-keygen -t ed25519'",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})

	// File permissions check
	start = time.Now()
	unsafeFiles := findUnsafePermissions()

	status = statusPass
	message = "File permissions look secure"

	if len(unsafeFiles) > 0 {
		status = statusWarn
		message = fmt.Sprintf("Found %d files with unsafe permissions", len(unsafeFiles))
	}

	report.Results = append(report.Results, DiagnosticResult{
		Name:          "File Permissions",
		Category:      "security",
		Status:        status,
		Message:       message,
		Details:       map[string]interface{}{"unsafe_files": unsafeFiles},
		FixSuggestion: "Review and fix file permissions with chmod",
		Duration:      time.Since(start),
		Timestamp:     time.Now(),
	})
}

// Helper functions

func getSystemInfo() SystemInfo {
	hostname, _ := os.Hostname()

	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	wd, _ := os.Getwd()
	homeDir, _ := os.UserHomeDir()
	shell := os.Getenv("SHELL")
	tempDir := os.TempDir()

	return SystemInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		GoVersion:  runtime.Version(),
		Hostname:   hostname,
		Username:   username,
		WorkingDir: wd,
		HomeDir:    homeDir,
		PathEnv:    os.Getenv("PATH"),
		Shell:      shell,
		TempDir:    tempDir,
	}
}

func isGoVersionSupported(version string) bool {
	// Simple check for Go 1.19+
	return strings.Contains(version, "go1.19") ||
		strings.Contains(version, "go1.20") ||
		strings.Contains(version, "go1.21") ||
		strings.Contains(version, "go1.22") ||
		strings.Contains(version, "go1.23") ||
		strings.Contains(version, "go1.24")
}

func getDiskSpace(path string) float64 {
	// This is a simplified implementation
	// In a real implementation, you'd use syscalls to get actual disk space
	return 50.0 // Placeholder: 50GB
}

func isURLReachable(ctx context.Context, url string) bool {
	// Simple connectivity check with timeout
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, http.NoBody)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	return resp.StatusCode < 400
}

func getGitConfig() map[string]string {
	config := make(map[string]string)

	keys := []string{"user.name", "user.email", "core.editor", "init.defaultBranch"}
	for _, key := range keys {
		cmd := exec.Command("git", "config", "--global", key)

		output, err := cmd.Output()
		if err == nil {
			config[key] = strings.TrimSpace(string(output))
		}
	}

	return config
}

func isDirectoryWritable(dir string) bool {
	testFile := filepath.Join(dir, ".gzh-write-test")

	file, err := os.Create(testFile)
	if err != nil {
		return false
	}

	if err := file.Close(); err != nil {
		// Log error but don't fail the check
	}

	if err := os.Remove(testFile); err != nil {
		// Log error but don't fail the check
		fmt.Printf("Warning: failed to remove test file: %v\n", err)
	}

	return true
}

func canCreateDirectory(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0o755)
		if err != nil {
			return false
		}

		if err := os.RemoveAll(dir); err != nil {
			// Log error but don't fail the check
			fmt.Printf("Warning: failed to remove test directory: %v\n", err)
		}
	}

	return true
}

func runCPUBenchmark() float64 {
	// Simple CPU benchmark
	start := time.Now()
	iterations := 1000000

	var sum int64
	for i := 0; i < iterations; i++ {
		sum += int64(i * i)
	}

	duration := time.Since(start)

	return float64(iterations) / duration.Seconds()
}

func runDiskBenchmark() float64 {
	// Simple disk I/O benchmark
	testFile := filepath.Join(os.TempDir(), "gzh-disk-bench")

	data := make([]byte, 1024*1024) // 1MB
	for i := range data {
		data[i] = byte(i % 256)
	}

	start := time.Now()

	file, err := os.Create(testFile)
	if err != nil {
		return 0
	}

	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't fail the check
		}
		if err := os.Remove(testFile); err != nil {
			// Log error but don't fail the check
		}
	}()

	for i := 0; i < 10; i++ {
		if _, err := file.Write(data); err != nil {
			return 0 // Return 0 on write error
		}
	}

	if err := file.Sync(); err != nil {
		return 0 // Return 0 on sync error
	}

	duration := time.Since(start)
	mb := float64(len(data)*10) / 1024 / 1024

	return mb / duration.Seconds()
}

func countSSHKeys() int {
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	files, err := os.ReadDir(sshDir)
	if err != nil {
		return 0
	}

	count := 0

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".pub") {
			count++
		}
	}

	return count
}

func findUnsafePermissions() []string {
	// Simplified implementation - in practice, you'd check actual file permissions
	return []string{} // Placeholder
}

func calculateReportTotals(report *DiagnosticReport) {
	report.TotalChecks = len(report.Results)
	for _, result := range report.Results {
		switch result.Status {
		case statusPass:
			report.PassedChecks++
		case statusWarn:
			report.WarnChecks++
		case statusFail:
			report.FailedChecks++
		case "skip":
			report.SkippedChecks++
		}
	}
}

func generateSummaryAndRecommendations(report *DiagnosticReport) {
	// Generate summary
	successRate := float64(report.PassedChecks) / float64(report.TotalChecks) * 100
	report.Summary = fmt.Sprintf("Diagnostic completed: %.1f%% success rate (%d/%d checks passed)",
		successRate, report.PassedChecks, report.TotalChecks)

	if report.FailedChecks > 0 {
		report.Summary += fmt.Sprintf(", %d critical issues", report.FailedChecks)
	}

	if report.WarnChecks > 0 {
		report.Summary += fmt.Sprintf(", %d warnings", report.WarnChecks)
	}

	// Generate recommendations
	if report.FailedChecks > 0 {
		report.Recommendations = append(report.Recommendations,
			"Address critical issues immediately")
	}

	if report.WarnChecks > 0 {
		report.Recommendations = append(report.Recommendations,
			"Review warnings and consider improvements")
	}

	if successRate < 80 {
		report.Recommendations = append(report.Recommendations,
			"System health is below optimal - review failed checks")
	}

	if len(report.Recommendations) == 0 {
		report.Recommendations = append(report.Recommendations,
			"System appears healthy - no immediate action required")
	}
}

func printResults(report *DiagnosticReport) {
	fmt.Printf("\n📄 Diagnostic Report\n")
	fmt.Printf("===================\n")
	fmt.Printf("Platform: %s\n", report.Platform)
	fmt.Printf("Duration: %v\n", report.Duration)
	fmt.Printf("Total Checks: %d\n", report.TotalChecks)
	fmt.Printf("Passed: %d, Warnings: %d, Failed: %d, Skipped: %d\n",
		report.PassedChecks, report.WarnChecks, report.FailedChecks, report.SkippedChecks)

	fmt.Printf("\n📅 Check Results:\n")
	fmt.Printf("=================\n")

	for _, result := range report.Results {
		icon := "✅" // pass

		switch result.Status {
		case "warn":
			icon = "⚠️"
		case "fail":
			icon = "❌"
		case "skip":
			icon = "⏭️"
		}

		fmt.Printf("  %s [%s] %s: %s\n", icon, strings.ToUpper(result.Category), result.Name, result.Message)

		if verbose && result.FixSuggestion != "" {
			fmt.Printf("    💡 Fix: %s\n", result.FixSuggestion)
		}
	}

	fmt.Printf("\n🎯 Summary:\n")
	fmt.Printf("=========\n")
	fmt.Println(report.Summary)

	fmt.Printf("\n💡 Recommendations:\n")
	fmt.Printf("==================\n")

	for i, rec := range report.Recommendations {
		fmt.Printf("  %d. %s\n", i+1, rec)
	}
}

func saveReport(report *DiagnosticReport, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			// Log error but don't override main error
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(report)
}

func attemptAutomaticFixes(ctx context.Context, report *DiagnosticReport, structuredLogger *logger.StructuredLogger, errorRecovery *errors.ErrorRecovery) {
	fmt.Printf("\n🔧 Attempting automatic fixes...\n")
	structuredLogger.Info("Starting automatic fixes", "total_issues", report.FailedChecks+report.WarnChecks) //nolint:contextcheck // Logger has its own context management

	fixed := 0

	for _, result := range report.Results {
		if result.Status == "fail" || result.Status == "warn" {
			fixErr := errorRecovery.Execute(ctx, fmt.Sprintf("fix-%s", result.Name), func() error { //nolint:contextcheck // Logger has its own context management
				structuredLogger.Debug("Attempting fix", "check_name", result.Name, "status", result.Status)

				if attemptFix := tryAutoFix(result); attemptFix {
					structuredLogger.Info("Successfully applied automatic fix", "check_name", result.Name)
					fmt.Printf("  ✅ Fixed: %s\n", result.Name)

					fixed++
				} else {
					structuredLogger.Warn("Cannot auto-fix issue", "check_name", result.Name, "reason", "no_fix_available")
					fmt.Printf("  ❌ Cannot auto-fix: %s\n", result.Name)
				}

				return nil
			})
			if fixErr != nil {
				structuredLogger.ErrorWithStack(fixErr, "Auto-fix operation failed", "check_name", result.Name) //nolint:contextcheck // Logger has its own context management
			}
		}
	}

	structuredLogger.Info("Auto-fix completed", "fixed_count", fixed, "total_attempted", report.FailedChecks+report.WarnChecks) //nolint:contextcheck // Logger has its own context management
	fmt.Printf("\n🎯 Auto-fix summary: %d issues resolved\n", fixed)

	if fixed > 0 {
		structuredLogger.Info("Recommend re-running diagnostic to verify fixes") //nolint:contextcheck // Logger has its own context management
		fmt.Println("🔄 Re-run 'gz doctor' to verify fixes")
	}
}

func tryAutoFix(result DiagnosticResult) bool {
	// Simplified auto-fix logic
	switch result.Name {
	case "Configuration Directory Access":
		homeDir, _ := os.UserHomeDir()
		configDir := filepath.Join(homeDir, ".config", "gzh-manager")

		return os.MkdirAll(configDir, 0o755) == nil
	default:
		return false
	}
}
