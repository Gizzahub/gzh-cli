// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package pm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-manager-go/internal/cli"
	"github.com/gizzahub/gzh-manager-go/internal/pm/compat"
	"github.com/gizzahub/gzh-manager-go/internal/pm/duplicates"
)

// 결과 JSON용 구조체
type UpdateRunMode struct {
	Compat             string `json:"compat"`
	PipAllowConda      bool   `json:"pipAllowConda,omitempty"`
	PacmanCleanOrphans bool   `json:"pacmanCleanOrphans,omitempty"`
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
		allManagers        bool
		manager            string
		strategy           string
		compatMode         string
		managersCSV        string
		outputFormat       string
		checkDuplicates    bool
		duplicatesMax      int
		pipAllowConda      bool
		pacmanCleanOrphans bool
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
		WithCustomBoolFlag("check-duplicates", true, "Check duplicate binaries across managers before update", &checkDuplicates).
		WithCustomIntFlag("duplicates-max", 10, "Max number of duplicate warnings to show", &duplicatesMax).
		WithCustomBoolFlag("pip-allow-conda", false, "Allow pip updates inside conda/mamba environment (use with caution)", &pipAllowConda).
		WithCustomBoolFlag("pacman-clean-orphans", false, "Also remove pacman orphan packages after upgrade (use with caution)", &pacmanCleanOrphans).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			res := &UpdateRunResult{
				RunID:     time.Now().UTC().Format("20060102T150405Z"),
				Mode:      UpdateRunMode{Compat: compatMode, PipAllowConda: pipAllowConda, PacmanCleanOrphans: pacmanCleanOrphans},
				StartedAt: time.Now().UTC(),
			}

			// Handle selected managers CSV first
			if managersCSV != "" {
				selected := parseCSVList(managersCSV)
				if len(selected) == 0 {
					return fmt.Errorf("no valid managers provided via --managers")
				}
				err := runUpdateSelected(ctx, selected, strategy, flags.DryRun, compatMode, res, checkDuplicates, duplicatesMax)
				res.FinishedAt = time.Now().UTC()
				if outputFormat == "json" {
					return printUpdateResultJSON(res)
				}
				return err
			}

			if manager != "" {
				// 단일 매니저 실행 전에도 중복 감지 요약을 보여준다
				if checkDuplicates {
					printSectionBanner("중복 설치 검사", "🧪")
					pathDirs := duplicates.SplitPATH(os.Getenv("PATH"))
					sources := duplicates.BuildDefaultSources(pathDirs)
					conflicts, _ := duplicates.CollectAndDetectConflicts(ctx, sources, pathDirs)
					duplicates.PrintConflictsSummary(conflicts, duplicatesMax)
					fmt.Println()
				}
				err := runUpdateManager(ctx, manager, strategy, flags.DryRun, compatMode, res)
				res.FinishedAt = time.Now().UTC()
				if outputFormat == "json" {
					return printUpdateResultJSON(res)
				}
				return err
			}
			if allManagers {
				err := runUpdateAll(ctx, strategy, flags.DryRun, compatMode, res, checkDuplicates, duplicatesMax)
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

// ===== 출력 하이라이트/개요 도우미 =====
// ANSI 컬러 상수
const (
	ansiReset  = "\x1b[0m"
	ansiBold   = "\x1b[1m"
	ansiCyan   = "\x1b[36m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
	ansiRed    = "\x1b[31m"
)

// 섹션 배너 출력
func printSectionBanner(title string, emoji string) {
	line := strings.Repeat("═", 10)
	fmt.Printf("\n%s%s%s %s %s %s%s\n", ansiBold, ansiCyan, line, emoji, title, line, ansiReset)
}

// 매니저 지원/설치 개요
type ManagerOverview struct {
	Name      string
	Supported bool
	Installed bool
	Reason    string // 미지원/미설치 사유
}

func detectManagerSupportOnOS(manager string) (bool, string) {
	goos := runtime.GOOS
	switch manager {
	case "brew":
		// macOS, Linux 모두 가능 (Linuxbrew)
		return goos == "darwin" || goos == "linux", ""
	case "apt":
		return goos == "linux", "apt는 Linux 전용"
	case "pacman", "yay":
		return goos == "linux", "Arch/Manjaro 계열 전용"
	case "sdkman":
		return goos == "darwin" || goos == "linux", ""
	case "asdf", "pip", "npm":
		return true, ""
	default:
		return true, ""
	}
}

func detectManagerInstalled(ctx context.Context, manager string) bool {
	switch manager {
	case "brew":
		return exec.CommandContext(ctx, "brew", "--version").Run() == nil
	case "asdf":
		return exec.CommandContext(ctx, "asdf", "--version").Run() == nil
	case "sdkman":
		sdkmanDir := os.Getenv("SDKMAN_DIR")
		if sdkmanDir == "" {
			sdkmanDir = os.Getenv("HOME") + "/.sdkman"
		}
		if _, err := os.Stat(sdkmanDir); err == nil {
			return true
		}
		return false
	case "apt":
		return exec.CommandContext(ctx, "apt", "--version").Run() == nil
	case "pacman":
		return exec.CommandContext(ctx, "pacman", "--version").Run() == nil
	case "yay":
		return exec.CommandContext(ctx, "yay", "--version").Run() == nil
	case "pip":
		if exec.CommandContext(ctx, "pip", "--version").Run() == nil {
			return true
		}
		return exec.CommandContext(ctx, "pip3", "--version").Run() == nil
	case "npm":
		return exec.CommandContext(ctx, "npm", "--version").Run() == nil
	default:
		_, err := exec.LookPath(manager)
		return err == nil
	}
}

func buildManagersOverview(ctx context.Context, managers []string) []ManagerOverview {
	var list []ManagerOverview
	for _, m := range managers {
		supported, reason := detectManagerSupportOnOS(m)
		installed := false
		if supported {
			installed = detectManagerInstalled(ctx, m)
		}
		mo := ManagerOverview{Name: m, Supported: supported, Installed: installed, Reason: reason}
		if !installed && reason == "" && supported {
			mo.Reason = "설치되어 있지 않음"
		}
		if !supported && mo.Reason == "" {
			mo.Reason = "현재 OS에서 지원되지 않음"
		}
		list = append(list, mo)
	}
	return list
}

func printManagersOverview(title string, overviews []ManagerOverview) {
	printSectionBanner(title, "📋")
	fmt.Printf("%-12s %-10s %-10s %s\n", "MANAGER", "SUPPORTED", "INSTALLED", "NOTE")
	fmt.Printf("%-12s %-10s %-10s %s\n", strings.Repeat("-", 12), strings.Repeat("-", 10), strings.Repeat("-", 10), strings.Repeat("-", 20))
	for _, m := range overviews {
		var sup, inst, note string
		if m.Supported {
			sup = "✅"
		} else {
			sup = "🚫"
		}
		if m.Installed {
			inst = "✅"
		} else {
			inst = "⛔"
		}
		note = m.Reason
		fmt.Printf("%-12s %-10s %-10s %s\n", m.Name, sup, inst, note)
	}
}

func runUpdateSelected(ctx context.Context, managers []string, strategy string, dryRun bool, compatMode string, res *UpdateRunResult, checkDuplicates bool, duplicatesMax int) error {
	fmt.Printf("Updating selected managers: %s\n", strings.Join(managers, ", "))
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}
	fmt.Println()

	// 개요 출력
	overview := buildManagersOverview(ctx, managers)
	printManagersOverview("지원 매니저 개요", overview)
	fmt.Println()

	// 중복 설치 검사 요약
	if checkDuplicates {
		printSectionBanner("중복 설치 검사", "🧪")
		pathDirs := duplicates.SplitPATH(os.Getenv("PATH"))
		sources := duplicates.BuildDefaultSources(pathDirs)
		conflicts, _ := duplicates.CollectAndDetectConflicts(ctx, sources, pathDirs)
		duplicates.PrintConflictsSummary(conflicts, duplicatesMax)
		fmt.Println()
	}

	// 순차 진행
	total := len(managers)
	for idx, m := range managers {
		ov := overview[idx]
		stepTitle := fmt.Sprintf("[%d/%d] %s", idx+1, total, m)
		if !ov.Supported {
			printSectionBanner(stepTitle+" — SKIP", "⚠️")
			fmt.Printf("%s%s이 매니저는 현재 OS에서 지원되지 않습니다: %s%s\n\n", ansiYellow, m, ov.Reason, ansiReset)
			continue
		}
		if !ov.Installed {
			printSectionBanner(stepTitle+" — SKIP", "⚠️")
			fmt.Printf("%s%s이(가) 설치되어 있지 않아 건너뜁니다. hint: 설치 후 다시 시도하세요.%s\n\n", ansiYellow, m, ansiReset)
			continue
		}

		printSectionBanner(stepTitle+" — Updating", "🚀")
		if err := runUpdateManager(ctx, m, strategy, dryRun, compatMode, res); err != nil {
			fmt.Printf("%sWarning: Failed to update %s: %v%s\n", ansiRed, m, err, ansiReset)
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
	case "pacman":
		return updatePacman(ctx, strategy, dryRun, res)
	case "yay":
		return updateYay(ctx, strategy, dryRun, res)
	case "pip":
		return updatePip(ctx, strategy, dryRun, res)
	case "npm":
		return updateNpm(ctx, strategy, dryRun, res)
	default:
		return fmt.Errorf("unsupported package manager: %s", manager)
	}
}

func runUpdateAll(ctx context.Context, strategy string, dryRun bool, compatMode string, res *UpdateRunResult, checkDuplicates bool, duplicatesMax int) error {
	managers := []string{"brew", "asdf", "sdkman", "apt", "pacman", "yay", "pip", "npm"}

	fmt.Println("Updating all package managers...")
	if dryRun {
		fmt.Println("(dry run - no changes will be made)")
	}
	fmt.Println()

	// 개요 출력
	overview := buildManagersOverview(ctx, managers)
	printManagersOverview("지원 매니저 개요", overview)
	fmt.Println()

	// 중복 설치 검사 요약
	if checkDuplicates {
		printSectionBanner("중복 설치 검사", "🧪")
		pathDirs := duplicates.SplitPATH(os.Getenv("PATH"))
		sources := duplicates.BuildDefaultSources(pathDirs)
		conflicts, _ := duplicates.CollectAndDetectConflicts(ctx, sources, pathDirs)
		duplicates.PrintConflictsSummary(conflicts, duplicatesMax)
		fmt.Println()
	}

	total := len(managers)
	for idx, manager := range managers {
		ov := overview[idx]
		stepTitle := fmt.Sprintf("[%d/%d] %s", idx+1, total, manager)
		if !ov.Supported {
			printSectionBanner(stepTitle+" — SKIP", "⚠️")
			fmt.Printf("%s%s이 매니저는 현재 OS에서 지원되지 않습니다: %s%s\n\n", ansiYellow, manager, ov.Reason, ansiReset)
			continue
		}
		if !ov.Installed {
			printSectionBanner(stepTitle+" — SKIP", "⚠️")
			fmt.Printf("%s%s이(가) 설치되어 있지 않아 건너뜁니다. hint: 설치 후 다시 시도하세요.%s\n\n", ansiYellow, manager, ansiReset)
			continue
		}

		printSectionBanner(stepTitle+" — Updating", "🚀")
		if err := runUpdateManager(ctx, manager, strategy, dryRun, compatMode, res); err != nil {
			fmt.Printf("%sWarning: Failed to update %s: %v%s\n", ansiRed, manager, err, ansiReset)
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

// conda/mamba 환경 감지: 활성화 여부와 종류 반환
func detectCondaOrMamba(ctx context.Context) (bool, string) {
	// 우선 환경변수로 확인
	if os.Getenv("CONDA_PREFIX") != "" || os.Getenv("CONDA_DEFAULT_ENV") != "" {
		// mamba 설치 여부로 구분
		if exec.CommandContext(ctx, "mamba", "--version").Run() == nil || exec.CommandContext(ctx, "micromamba", "--version").Run() == nil {
			return true, "mamba"
		}
		return true, "conda"
	}
	if os.Getenv("MAMBA_ROOT_PREFIX") != "" {
		return true, "mamba"
	}
	return false, ""
}

// pipCmd 문자열("python -m pip" 또는 "pip3")을 exec.Command 인자로 분해하여 Cmd를 생성한다.
func newPipExec(ctx context.Context, pipCmd string, moreArgs ...string) *exec.Cmd {
	parts := strings.Fields(pipCmd)
	if len(parts) == 0 {
		// 비정상 입력 방어
		parts = []string{"pip"}
	}
	args := append(parts[1:], moreArgs...)
	return exec.CommandContext(ctx, parts[0], args...)
}

func updatePip(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// Check if pip is installed
	pipCmd := findPipCommand(ctx)
	if pipCmd == "" {
		return fmt.Errorf("pip is not installed or not in PATH")
	}

	fmt.Println("🐍 Updating pip packages...")
	_ = res.ensureManager("pip")

	// conda/mamba 환경에서는 pip 업데이트가 충돌을 유발할 수 있어 기본적으로 차단
	if active, kind := detectCondaOrMamba(ctx); active && !res.Mode.PipAllowConda {
		fmt.Printf("%sConda/Mamba(%s) 환경이 감지되었습니다. 호환성 문제를 피하기 위해 pip 업데이트를 건너뜁니다.%s\n", ansiYellow, kind, ansiReset)
		fmt.Println("권장: conda/mamba 명령으로 환경을 업데이트하세요. 예) conda update --all 또는 mamba update --all")
		return nil
	}

	// Upgrade pip itself
	if !dryRun {
		cmd := newPipExec(ctx, pipCmd, "install", "--upgrade", "pip")
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

	cmd := newPipExec(ctx, pipCmd, "list", "--outdated", "--format=freeze")
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
			cmd = newPipExec(ctx, pipCmd, "install", "--upgrade", pkg)
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

// Arch/Manjaro pacman 업데이트
func updatePacman(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// pacman 존재 확인
	if err := exec.CommandContext(ctx, "pacman", "--version").Run(); err != nil {
		return fmt.Errorf("pacman is not installed or not in PATH")
	}

	fmt.Println("🐧 Updating pacman system packages...")
	_ = res.ensureManager("pacman")

	if !dryRun {
		// 시스템 업데이트 (비대화형)
		cmd := exec.CommandContext(ctx, "sudo", "-n", "pacman", "-Syu", "--noconfirm")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Warning: pacman -Syu failed (maybe sudo permission required): %v\n", err)
		}

		// 고아 패키지 제거는 기본 비활성화. 명시적 플래그가 있을 때만 수행한다.
		if res.Mode.PacmanCleanOrphans {
			// 중요 패키지 화이트리스트: 오탑재 방지
			critical := map[string]struct{}{
				"linux": {}, "linux-lts": {}, "systemd": {}, "glibc": {}, "bash": {},
				"zsh": {}, "coreutils": {}, "pacman": {}, "util-linux": {}, "filesystem": {},
				"shadow": {}, "iproute2": {}, "networkmanager": {}, "sudo": {},
			}
			listCmd := exec.CommandContext(ctx, "pacman", "-Qtdq")
			listOut, _ := listCmd.Output()
			lines := strings.Split(strings.TrimSpace(string(listOut)), "\n")
			var orphanPkgs []string
			for _, ln := range lines {
				ln = strings.TrimSpace(ln)
				if ln == "" {
					continue
				}
				if _, isCritical := critical[ln]; isCritical {
					fmt.Printf("Skipping critical package from orphan removal: %s\n", ln)
					continue
				}
				orphanPkgs = append(orphanPkgs, ln)
			}
			if len(orphanPkgs) > 0 {
				args := append([]string{"-n", "pacman", "-Rns", "--noconfirm"}, orphanPkgs...)
				rmCmd := exec.CommandContext(ctx, "sudo", args...)
				rmCmd.Stdout = os.Stdout
				rmCmd.Stderr = os.Stderr
				if err := rmCmd.Run(); err != nil {
					fmt.Printf("Warning: failed to remove orphan packages: %v\n", err)
				}
			} else {
				fmt.Println("No orphan packages to remove or all were critical/whitelisted.")
			}
		} else {
			// 안내만 출력
			fmt.Println("(info) pacman orphan cleanup is disabled by default. Use --pacman-clean-orphans to enable.")
		}
	} else {
		fmt.Println("Would run: sudo -n pacman -Syu --noconfirm")
		if res.Mode.PacmanCleanOrphans {
			fmt.Println("Would list/remove orphans: pacman -Qtdq | sudo -n pacman -Rns --noconfirm <orphans> (excluding critical packages)")
		} else {
			fmt.Println("(info) pacman orphan cleanup disabled; no removal will be attempted")
		}
	}

	return nil
}

// Arch/Manjaro yay(AUR) 업데이트
func updateYay(ctx context.Context, strategy string, dryRun bool, res *UpdateRunResult) error {
	// yay 존재 확인
	if err := exec.CommandContext(ctx, "yay", "--version").Run(); err != nil {
		return fmt.Errorf("yay is not installed or not in PATH")
	}

	fmt.Println("🧠 Updating yay (AUR) packages...")
	_ = res.ensureManager("yay")

	if !dryRun {
		cmd := exec.CommandContext(ctx, "yay", "-Syu", "--noconfirm", "--needed")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update yay packages: %w", err)
		}
		// 캐시/불필요 패키지 정리
		clean := exec.CommandContext(ctx, "yay", "-Yc", "--noconfirm")
		clean.Stdout = os.Stdout
		clean.Stderr = os.Stderr
		_ = clean.Run()
	} else {
		fmt.Println("Would run: yay -Syu --noconfirm --needed")
		fmt.Println("Would run: yay -Yc --noconfirm")
	}
	return nil
}
