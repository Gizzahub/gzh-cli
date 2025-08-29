// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package doctor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/cmd/pm/utils"
	"github.com/Gizzahub/gzh-cli/internal/cli"
	"github.com/Gizzahub/gzh-cli/internal/pm/compat"
	"github.com/Gizzahub/gzh-cli/internal/pm/duplicates"
)

func NewDoctorCmd(ctx context.Context) *cobra.Command {
	var (
		managersCSV     string
		compatMode      string
		outputFormat    string
		checkConf       bool
		attemptFix      bool
		checkDuplicates bool
	)

	builder := cli.NewCommandBuilder(ctx, "doctor", "Diagnose package manager issues").
		WithLongDescription(`Run diagnostic checks for package manager configuration and conflicts.

Examples:
  gz pm doctor --check-conflicts
  gz pm doctor --managers asdf --compat strict --output json`).
		WithCustomFlag("managers", "", "Comma-separated managers (e.g., asdf,brew)", &managersCSV).
		WithCustomFlag("compat", "auto", "Compatibility handling: auto, strict, off", &compatMode).
		WithCustomFlag("output", "text", "Output format: text, json", &outputFormat).
		WithCustomBoolFlag("check-conflicts", true, "Check for known conflicts", &checkConf).
		WithCustomBoolFlag("fix", false, "Attempt to fix detected issues", &attemptFix).
		WithCustomBoolFlag("check-duplicates", true, "Check duplicate binaries across managers", &checkDuplicates).
		WithRunFuncE(func(ctx context.Context, flags *cli.CommonFlags, args []string) error {
			managers := managersCSV
			if managers == "" {
				managers = "asdf"
			}
			selected := utils.ParseCSVList(managers)
			report := DoctorReport{Managers: []DoctorManagerReport{}}

			for _, m := range selected {
				switch m {
				case "asdf":
					r, err := runAsdfDoctor(ctx, compatMode, checkConf, attemptFix)
					if err != nil {
						r.Error = err.Error()
					}
					report.Managers = append(report.Managers, r)
				default:
					report.Managers = append(report.Managers, DoctorManagerReport{Name: m, Status: "unsupported"})
				}
			}

			if outputFormat == "json" {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}

			// Text output
			for _, m := range report.Managers {
				fmt.Printf("=== %s ===\n", m.Name)
				if m.Error != "" {
					fmt.Printf("Error: %s\n", m.Error)
				}
				for _, p := range m.Plugins {
					fmt.Printf("- %s: conflicts=%d\n", p.Name, p.Conflicts)
					for _, w := range p.Warnings {
						fmt.Printf("  warn: %s\n", w)
					}
					if len(p.Suggestions) > 0 {
						fmt.Printf("  suggest:\n")
						for _, s := range p.Suggestions {
							fmt.Printf("    - %s\n", s)
						}
					}
				}
			}

			// 중복 설치 검사 요약
			if checkDuplicates {
				fmt.Println()
				fmt.Println("Duplicate installation check (experimental)")
				pathDirs := duplicates.SplitPATH(os.Getenv("PATH"))
				sources := duplicates.BuildDefaultSources(pathDirs)
				conflicts, _ := duplicates.CollectAndDetectConflicts(ctx, sources, pathDirs)
				duplicates.PrintConflictsSummary(conflicts, 10)
			}

			return nil
		})

	return builder.Build()
}

type DoctorReport struct {
	Managers []DoctorManagerReport `json:"managers"`
}

type DoctorManagerReport struct {
	Name    string               `json:"name"`
	Status  string               `json:"status"`
	Plugins []DoctorPluginReport `json:"plugins,omitempty"`
	Error   string               `json:"error,omitempty"`
}

type DoctorPluginReport struct {
	Name        string   `json:"name"`
	Conflicts   int      `json:"conflicts"`
	Warnings    []string `json:"warnings,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

func runAsdfDoctor(ctx context.Context, _ string, _, attemptFix bool) (DoctorManagerReport, error) {
	// Ensure asdf exists
	if err := exec.CommandContext(ctx, "asdf", "--version").Run(); err != nil {
		// asdf not installed is a valid state for the doctor report
		report := DoctorManagerReport{
			Name:   "asdf",
			Status: "not-installed",
			Error:  "asdf is not installed or not in PATH",
		}
		return report, nil //nolint:nilerr // asdf 미설치는 정상적인 상태이므로 오류가 아님
	}

	// List plugins
	out, err := exec.CommandContext(ctx, "asdf", "plugin", "list").Output()
	if err != nil {
		return DoctorManagerReport{Name: "asdf", Status: "error"}, fmt.Errorf("failed to list plugins: %w", err)
	}

	mgr := DoctorManagerReport{Name: "asdf", Status: "ok"}
	plugins := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, plugin := range plugins {
		if plugin == "" {
			continue
		}
		filters := compat.BuildFilterChain("asdf", plugin)
		warns := compat.CollectWarnings(filters)
		conflicts := compat.CountConflicts(filters)
		pr := DoctorPluginReport{
			Name:      plugin,
			Conflicts: conflicts,
			Warnings:  warns,
		}

		// Suggestions based on plugin
		switch plugin {
		case "rust":
			pr.Suggestions = append(pr.Suggestions, "rustup 단일 관리 권장 또는 PATH 정리")
		case "nodejs":
			pr.Suggestions = append(pr.Suggestions, "corepack enable 실행 권장")
		}

		// Optional fix
		if attemptFix {
			if plugin == "nodejs" {
				_ = exec.CommandContext(ctx, "bash", "-lc", "corepack enable").Run()
			}
		}

		mgr.Plugins = append(mgr.Plugins, pr)
	}

	return mgr, nil
}
