// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/cli"
	"github.com/gizzahub/gzh-manager-go/internal/pm/compat"
)

// 결과 JSON용 구조체
type UpdateRunMode struct {
	Compat string `json:"compat"`
}

type PluginResult struct {
	Name       string            `json:"name"`
	Actions    []string          `json:"actions,omitempty"`
	EnvApplied map[string]string `json:"envApplied,omitempty"`
	Warnings   []string          `json:"warnings,omitempty"`
	Conflicts  int               `json:"conflicts,omitempty"`
	Error      string            `json:"error,omitempty"`
}

type ManagerResult struct {
	Name    string         `json:"name"`
	Status  string         `json:"status"`
	Plugins []PluginResult `json:"plugins,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type Totals struct {
	Install   int `json:"install"`
	Skip      int `json:"skip"`
	Warnings  int `json:"warnings"`
	Conflicts int `json:"conflicts"`
}

type UpdateRunResult struct {
	RunID      string          `json:"runId"`
	Mode       UpdateRunMode   `json:"mode"`
	StartedAt  time.Time       `json:"startedAt"`
	FinishedAt time.Time       `json:"finishedAt"`
	Managers   []ManagerResult `json:"managers"`
	Totals     Totals          `json:"totals"`
}

func (r *UpdateRunResult) ensureManager(name string) *ManagerResult {
	for i := range r.Managers {
		if r.Managers[i].Name == name {
			return &r.Managers[i]
		}
	}
	r.Managers = append(r.Managers, ManagerResult{Name: name, Status: "success"})
	return &r.Managers[len(r.Managers)-1]
}

func (m *ManagerResult) addPluginResult(pr PluginResult) {
	m.Plugins = append(m.Plugins, pr)
	if pr.Error != "" {
		m.Status = "partial"
	}
}

func newUpdateCmd(ctx context.Context) *cobra.Command {
	var (
		allManagers  bool
		manager      string
		strategy     string
		compatMode   string
		managersCSV  string
		outputFormat string
	)

	builder := cli.NewCommandBuilder(ctx, "update", "Update packages based on version strategy").
		WithLongDescription(`Update packages for specified package managers based on configured version strategy.

Supports update strategies:
- latest: Update to the absolute latest version
- stable: Update to the latest stable version (default)
- minor: Update to latest patch/minor version only
- fixed: Keep exact versions (no updates)

Examples:
  # Update all package managers
  gz pm update --all

  # Update specific package manager
  gz pm update --manager brew

  # Update with specific strategy
  gz pm update --manager asdf --strategy latest

  # Dry run to see what would be updated
  gz pm update --all --dry-run`).
		WithDryRunFlag().
		WithCustomBoolFlag("all", false, "Update all package managers", &allManagers).
		WithCustomFlag("manager", "", "Package manager to update", &manager).
		WithCustomFlag("managers", "", "Comma-separated package managers to update (e.g., brew,asdf,pip)", &managersCSV).
		WithCustomFlag("strategy", "stable", "Update strategy: latest, stable, minor, fixed", &strategy).
		WithCustomFlag("compat", "auto", "Compatibility handling: auto, strict, off", &compatMode).
		WithCustomFlag("output", "text", "Output format: text, json", &outputFormat).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			res := &UpdateRunResult{
				RunID:     time.Now().UTC().Format("20060102T150405Z"),
				Mode:      UpdateRunMode{Compat: compatMode},
				StartedAt: time.Now().UTC(),
			}

			// Handle selected managers CSV first
			if managersCSV != "" {
				selected := parseCSVList(managersCSV)
				if len(selected) == 0 {
					return fmt.Errorf("no valid managers provided via --managers")
				}
				err := runUpdateSelected(ctx, selected, strategy, flags.DryRun, compatMode, res)
				res.FinishedAt = time.Now().UTC()
				if outputFormat == "json" {
					return printUpdateResultJSON(res)
				}
				return err
			}

			if manager != "" {
				err := runUpdateManager(ctx, manager, strategy, flags.DryRun, compatMode, res)
				res.FinishedAt = time.Now().UTC()
				if outputFormat == "json" {
					return printUpdateResultJSON(res)
				}
				return err
			}
			if allManagers {
				err := runUpdateAll(ctx, strategy, flags.DryRun, compatMode, res)
				res.FinishedAt = time.Now().UTC()
				if outputFormat == "json" {
					return printUpdateResultJSON(res)
				}
				return err
			}
			return fmt.Errorf("specify --manager, --managers, or --all")
		})

	return builder.Build()
}

func printUpdateResultJSON(res *UpdateRunResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

func parseCSVList(s string) []string {
	var list []string
	for _, p := range strings.Split(s, ",") {
		item := strings.TrimSpace(p)
		if item != "" {
			list = append(list, item)
		}
	}
	return list
}

func runUpdateSelected(ctx context.Context, managers []string, strategy string, dryRun bool, compatMode string, res *UpdateRunResult) error {
	fmt.Printf("Updating selected managers: %s\n", strings.Join(managers, ", "))
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}
	fmt.Println()

	for _, m := range managers {
		fmt.Printf("=== Updating %s ===\n", m)
		if err := runUpdateManager(ctx, m, strategy, dryRun, compatMode, res); err != nil {
			fmt.Printf("Warning: Failed to update %s: %v\n", m, err)
			continue
		}
		fmt.Println()
	}
	return nil
}

func runUpdateManager(ctx context.Context, manager, strategy string, dryRun bool, compatMode string, res *UpdateRunResult) error {
	fmt.Printf("Updating %s packages with strategy: %s\n", manager, strategy)
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}

	// TODO: Implement unified update logic
	// For now, this is a placeholder
	switch manager {
	case "brew":
		return updateBrew(ctx, strategy, dryRun, res)
	case "asdf":
		return updateAsdf(ctx, strategy, dryRun, compatMode, res)
	case "sdkman":
		return updateSdkman(ctx, strategy, dryRun, res)
	case "apt":
		return updateApt(ctx, strategy, dryRun, res)
	case "pip":
		return updatePip(ctx, strategy, dryRun, res)
	case "npm":
		return updateNpm(ctx, strategy, dryRun, res)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func runUpdateAll(ctx context.Context, strategy string, dryRun bool, compatMode string, res *UpdateRunResult) error {
	managers := []string{"brew", "asdf", "sdkman", "apt", "pip", "npm"}

	fmt.Println("Updating all package managers...")
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}
	fmt.Println()

	for _, manager := range managers {
		fmt.Printf("=== Updating %s ===\n", manager)
		if err := runUpdateManager(ctx, manager, strategy, dryRun, compatMode, res); err != nil {
			fmt.Printf("Warning: Failed to update %s: %v\n", manager, err)
			continue
		}
		fmt.Println()
	}

	return nil
}

// updateBrew updates Homebrew packages.
func updateBrew(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if brew is installed
	cmd := exec.CommandContext(ctx, "brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew is not installed or not in PATH")
	}

	fmt.Println("🍺 Updating Homebrew...")
	mgr := res.ensureManager("brew")

	// Update brew itself
	if !dryRun {
		cmd = exec.CommandContext(ctx, "brew", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update brew: %w", err)
		}
	} else {
		fmt.Println("Would run: brew update")
	}

	// Upgrade packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd = exec.CommandContext(ctx, "brew", "upgrade")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to upgrade brew packages: %w", err)
			}
		} else {
			fmt.Println("Would run: brew upgrade")
		}
	}

	// Cleanup old versions
	if !dryRun {
		cmd = exec.CommandContext(ctx, "brew", "cleanup")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: cleanup failed: %v\n", err)
		}
	} else {
		fmt.Println("Would run: brew cleanup")
	}

	_ = mgr // currently no per-plugin data for brew
	return nil
}

func updateAsdf(ctx context.Context, strategy string, dryRun bool, compatMode string, res *UpdateRunResult) error {
	// Check if asdf is installed
	cmd := exec.CommandContext(ctx, "asdf", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("asdf is not installed or not in PATH")
	}

	fmt.Println("🔄 Updating asdf plugins...")
	mgr := res.ensureManager("asdf")

	// Update asdf plugins
	if !dryRun {
		cmd = exec.CommandContext(ctx, "asdf", "plugin", "update", "--all")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = compat.MergeWithProcessEnv(nil)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update asdf plugins: %w", err)
		}
	} else {
		fmt.Println("Would run: asdf plugin update --all")
	}

	// Get list of installed plugins
	cmd = exec.CommandContext(ctx, "asdf", "plugin", "list")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list asdf plugins: %w", err)
	}

	plugins := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}
		fmt.Printf("Checking %s for updates...\n", plugin)

		// Build filter chain for this plugin based on mode
		var filters []compat.CompatibilityFilter
		if compatMode != "off" {
			filters = compat.BuildFilterChain("asdf", plugin)
		}
		warnings := compat.CollectWarnings(filters)
		for _, w := range warnings {
			fmt.Println(w)
		}

		conflicts := 0
		if compatMode == "strict" {
			conflicts = compat.CountConflicts(filters)
			if conflicts > 0 {
				mgr.addPluginResult(PluginResult{
					Name:      plugin,
					Warnings:  warnings,
					Conflicts: conflicts,
					Error:     fmt.Sprintf("compatibility conflicts detected for asdf plugin %s (mode=strict)", plugin),
				})
				return fmt.Errorf("compatibility conflicts detected for asdf plugin %s (mode=strict)", plugin)
			}
		}

		// Dry-run details: show env and post actions
		envPreview := compat.MergeEnvFromFilters(filters)
		post := compat.CollectPostActions(filters)
		if dryRun {
			if len(envPreview) > 0 {
				fmt.Println("Would apply environment variables:")
				for k, v := range envPreview {
					fmt.Printf("  %s=%s\n", k, v)
				}
			}
			if len(post) > 0 {
				fmt.Println("Would run post actions:")
				for _, a := range post {
					fmt.Printf("  - %s: %v\n", a.Description, a.Command)
				}
			}
		}

		pluginResult := PluginResult{
			Name:       plugin,
			Warnings:   warnings,
			Conflicts:  conflicts,
			EnvApplied: envPreview,
		}

		// Install latest version based on strategy
		if strategy == "latest" || strategy == "stable" {
			// Skip if already latest
			isLatest, checkErr := asdfIsLatestInstalled(ctx, plugin)
			if checkErr == nil && isLatest {
				fmt.Printf("version of %s is already latest; skipping install.\n", plugin)
				pluginResult.Actions = append(pluginResult.Actions, "skip:latest")
				mgr.addPluginResult(pluginResult)
				// Even if skipping, consider running post actions (idempotent)
				if !dryRun {
					for _, action := range post {
						postCmd := exec.CommandContext(ctx, action.Command[0], action.Command[1:]...)
						postCmd.Stdout = os.Stdout
						postCmd.Stderr = os.Stderr
						if action.Env != nil {
							postCmd.Env = compat.MergeEnvWithProcessEnv(action.Env)
						}
						if err := postCmd.Run(); err != nil && !action.IgnoreError {
							fmt.Printf("Warning: post action failed (%s): %v\n", action.Description, err)
						}
					}
				}
				continue
			}

			if !dryRun {
				cmd = exec.CommandContext(ctx, "asdf", "install", plugin, "latest")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Env = compat.MergeEnvWithProcessEnv(envPreview)
				if err := cmd.Run(); err != nil {
					fmt.Printf("Warning: failed to install latest %s: %v\n", plugin, err)
					pluginResult.Error = err.Error()
				} else {
					pluginResult.Actions = append(pluginResult.Actions, "install:latest")
					// Run post actions if install succeeded
					for _, action := range post {
						postCmd := exec.CommandContext(ctx, action.Command[0], action.Command[1:]...)
						postCmd.Stdout = os.Stdout
						postCmd.Stderr = os.Stderr
						if action.Env != nil {
							postCmd.Env = compat.MergeEnvWithProcessEnv(action.Env)
						}
						if err := postCmd.Run(); err != nil && !action.IgnoreError {
							fmt.Printf("Warning: post action failed (%s): %v\n", action.Description, err)
						}
					}
				}
			} else {
				fmt.Printf("Would run: asdf install %s latest\n", plugin)
				pluginResult.Actions = append(pluginResult.Actions, "would:install:latest")
			}
		}

		mgr.addPluginResult(pluginResult)
	}

	return nil
}

// asdfIsLatestInstalled checks if the latest version reported by asdf is already installed/current.
func asdfIsLatestInstalled(ctx context.Context, plugin string) (bool, error) {
	latestCmd := exec.CommandContext(ctx, "asdf", "latest", plugin)
	latestOut, err := latestCmd.Output()
	if err != nil {
		return false, err
	}
	latest := strings.TrimSpace(string(latestOut))
	if latest == "" {
		return false, fmt.Errorf("empty latest version for %s", plugin)
	}

	currentCmd := exec.CommandContext(ctx, "asdf", "current", plugin)
	currentOut, err := currentCmd.Output()
	if err != nil {
		return false, err
	}
	currentLine := strings.TrimSpace(string(currentOut))
	// Expected format: "plugin version (set by ...)"; extract version token(s)
	fields := strings.Fields(currentLine)
	if len(fields) < 2 {
		return false, nil
	}
	currentVersion := fields[1]
	return currentVersion == latest, nil
}

func updateSdkman(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if SDKMAN is installed
	sdkmanDir := os.Getenv("SDKMAN_DIR")
	if sdkmanDir == "" {
		sdkmanDir = os.Getenv("HOME") + "/.sdkman"
	}

	if _, err := os.Stat(sdkmanDir); os.IsNotExist(err) {
		return fmt.Errorf("sdkman is not installed")
	}

	fmt.Println("☕ Updating SDKMAN...")
	_ = res.ensureManager("sdkman")

	// Update SDKMAN itself
	if !dryRun {
		cmd := exec.CommandContext(ctx, "bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk selfupdate")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to update SDKMAN: %v\n", err)
		}
	} else {
		fmt.Println("Would run: sdk selfupdate")
	}

	// Update candidates based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd := exec.CommandContext(ctx, "bash", "-c", "source "+sdkmanDir+"/bin/sdkman-init.sh && sdk update")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: failed to update SDKMAN candidates: %v\n", err)
			}
		} else {
			fmt.Println("Would run: sdk update")
		}
	}

	return nil
}

func updateApt(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if apt is available
	cmd := exec.CommandContext(ctx, "apt", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("apt is not available on this system")
	}

	fmt.Println("📦 Updating APT packages...")
	_ = res.ensureManager("apt")

	// Update package lists
	if !dryRun {
		cmd = exec.CommandContext(ctx, "sudo", "apt", "update")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update apt package lists: %w", err)
		}
	} else {
		fmt.Println("Would run: sudo apt update")
	}

	// Upgrade packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		if !dryRun {
			cmd = exec.CommandContext(ctx, "sudo", "apt", "upgrade", "-y")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to upgrade apt packages: %w", err)
			}
		} else {
			fmt.Println("Would run: sudo apt upgrade -y")
		}
	}

	return nil
}

func updatePip(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if pip is installed
	pipCmd := findPipCommand(ctx)
	if pipCmd == "" {
		return fmt.Errorf("pip is not installed or not in PATH")
	}

	fmt.Println("🐍 Updating pip packages...")
	_ = res.ensureManager("pip")

	// Upgrade pip itself
	if !dryRun {
		cmd := exec.CommandContext(ctx, pipCmd, "install", "--upgrade", "pip")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: failed to upgrade pip: %v\n", err)
		}
	} else {
		fmt.Printf("Would run: %s install --upgrade pip\n", pipCmd)
	}

	// Update packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		return updateOutdatedPackages(ctx, pipCmd, dryRun, res)
	}

	return nil
}

// findPipCommand finds available pip command (python -m pip or pip3).
func findPipCommand(ctx context.Context) string {
	// Try python -m pip first
	cmd := exec.CommandContext(ctx, "python", "-m", "pip", "--version")
	if err := cmd.Run(); err == nil {
		return "python -m pip"
	}

	// Try pip3
	cmd = exec.CommandContext(ctx, "pip3", "--version")
	if err := cmd.Run(); err == nil {
		return "pip3"
	}

	return ""
}

// upgradePip upgrades pip itself.
func upgradePip(ctx context.Context, pipCmd string, dryRun bool) error {
	if !dryRun {
		args := strings.Split(pipCmd, " ")
		args = append(args, "install", "--upgrade", "pip")
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to upgrade pip: %w", err)
		}
	} else {
		fmt.Printf("Would run: %s install --upgrade pip\n", pipCmd)
	}
	return nil
}

// updateOutdatedPackages updates all outdated packages.
func updateOutdatedPackages(ctx context.Context, pipCmd string, dryRun bool, res *UpdateRunResult) error {
	fmt.Println("Checking for outdated packages...")

	cmd := exec.CommandContext(ctx, pipCmd, "list", "--outdated", "--format=freeze")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list outdated pip packages: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "==")
		if len(parts) < 1 {
			continue
		}
		pkg := parts[0]
		fmt.Printf("Upgrading %s...\n", pkg)

		if !dryRun {
			cmd = exec.CommandContext(ctx, pipCmd, "install", "--upgrade", pkg)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Warning: failed to upgrade %s: %v\n", pkg, err)
			}
		} else {
			fmt.Printf("Would run: %s install --upgrade %s\n", pipCmd, pkg)
		}
	}

	return nil
}

func updateNpm(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if npm is installed
	cmd := exec.CommandContext(ctx, "npm", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm is not installed or not in PATH")
	}

	fmt.Println("🧩 Updating npm global packages...")
	_ = res.ensureManager("npm")

	// Update global packages based on strategy
	if strategy == "latest" || strategy == "stable" {
		return updateGlobalNpmPackages(ctx, dryRun, res)
	}

	return nil
}

// updateGlobalNpmPackages updates globally installed npm packages.
func updateGlobalNpmPackages(ctx context.Context, dryRun bool, res *UpdateRunResult) error {
	fmt.Println("Checking for outdated global packages...")

	if !dryRun {
		cmd := exec.CommandContext(ctx, "npm", "update", "-g")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update global npm packages: %w", err)
		}
	} else {
		fmt.Println("Would run: npm update -g")
	}

	return nil
}
