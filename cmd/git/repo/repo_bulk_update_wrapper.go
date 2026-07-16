// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	gitrepo "github.com/gizzahub/gzh-cli-gitforge/pkg/repository"
)

// newRepoBulkUpdateCmd creates the bulk update command using gzh-cli-gitforge library
// This delegates repository scanning and updating to the external gzh-cli-gitforge package.
//
// NOTE: This function replaces the old implementation in repo_bulk_update.go
func newRepoBulkUpdateCmd() *cobra.Command {
	var opts bulkUpdateCmdOptions

	cmd := &cobra.Command{
		Use:   "pull-all [directory]",
		Short: "Recursively update all Git repositories with pull --rebase",
		Long: `재귀적으로 하위 디렉토리의 모든 Git 리포지터리를 스캔하고 안전하게 업데이트합니다.

이 명령어는 다음 조건에서만 자동으로 pull --rebase를 실행합니다:
- 로컬 변경사항이 없는 경우 (clean working tree)
- upstream 브랜치가 설정된 경우

그 외의 경우에는 알림만 표시하여 수동 처리를 유도합니다.

모든 스캔된 리포지터리의 처리 결과를 테이블 형식으로 출력합니다.`,
		Example: `
  # 현재 디렉토리부터 모든 Git 리포지터리 업데이트
  gz git repo pull-all

  # 특정 디렉토리 지정
  gz git repo pull-all /Users/example/repos

  # 병렬 처리 및 상세 출력
  gz git repo pull-all --parallel 5 --verbose

  # 실제 실행하지 않고 시뮬레이션
  gz git repo pull-all --dry-run`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run(cmd.Context(), args)
		},
	}

	// 플래그 설정
	cmd.Flags().IntVarP(&opts.parallel, "parallel", "p", 5, "병렬 처리 워커 수")
	cmd.Flags().IntVar(&opts.maxDepth, "max-depth", 5, "최대 스캔 깊이")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "실제 실행하지 않고 시뮬레이션만 수행")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "JSON 형식으로 결과 출력")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "상세 로그 출력")
	cmd.Flags().BoolVar(&opts.noFetch, "no-fetch", false, "원격 저장소에서 변경사항을 가져오지 않음")
	cmd.Flags().StringVar(&opts.includePattern, "include-pattern", "", "포함할 리포지터리 패턴 (정규식)")
	cmd.Flags().StringVar(&opts.excludePattern, "exclude-pattern", "", "제외할 리포지터리 패턴 (정규식)")

	return cmd
}

// bulkUpdateCmdOptions holds command options
type bulkUpdateCmdOptions struct {
	parallel       int
	maxDepth       int
	dryRun         bool
	jsonOutput     bool
	verbose        bool
	noFetch        bool
	includePattern string
	excludePattern string
}

// run executes the bulk update operation
func (opts *bulkUpdateCmdOptions) run(ctx context.Context, args []string) error {
	// Determine directory
	var directory string
	if len(args) > 0 {
		directory = args[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("현재 디렉토리를 가져올 수 없습니다: %w", err)
		}
		directory = cwd
	}

	// Create logger
	var logger gitrepo.Logger
	if opts.verbose {
		logger = gitrepo.NewWriterLogger(os.Stdout)
	} else {
		logger = gitrepo.NewNoopLogger()
	}

	// Create client
	client := gitrepo.NewClient(gitrepo.WithClientLogger(logger))

	// Prepare bulk update options
	bulkOpts := gitrepo.BulkUpdateOptions{
		Directory:      directory,
		Parallel:       opts.parallel,
		MaxDepth:       opts.maxDepth,
		DryRun:         opts.dryRun,
		Verbose:        opts.verbose,
		NoFetch:        opts.noFetch,
		IncludePattern: opts.includePattern,
		ExcludePattern: opts.excludePattern,
		Logger:         logger,
		ProgressCallback: func(current, total int, repo string) {
			if opts.verbose {
				fmt.Printf("[%d/%d] Processing: %s\n", current, total, repo)
			}
		},
	}

	// Print scan message
	fmt.Printf("🔍 Git 리포지터리 스캔 중: %s\n", color.CyanString(directory))

	// Execute bulk update
	result, err := client.BulkUpdate(ctx, bulkOpts)
	if err != nil {
		return fmt.Errorf("bulk update failed: %w", err)
	}

	// Print results
	if len(result.Repositories) == 0 {
		if result.TotalScanned > 0 {
			fmt.Printf("필터링된 결과: Git 리포지터리 %d개 중 처리 대상 없음\n", result.TotalScanned)
		} else {
			fmt.Println("Git 리포지터리를 찾을 수 없습니다.")
		}
		return nil
	}

	if result.TotalScanned != result.TotalProcessed {
		fmt.Printf("📦 발견된 리포지터리: %d개, 처리 대상: %s개\n\n", result.TotalScanned, color.GreenString(strconv.Itoa(result.TotalProcessed)))
	} else {
		fmt.Printf("📦 발견된 리포지터리: %s개\n\n", color.GreenString(strconv.Itoa(result.TotalProcessed)))
	}

	// Render results
	if opts.jsonOutput {
		renderJSONResults(result)
	} else {
		renderTableResults(result)
	}

	return nil
}

// renderTableResults renders results in table format
// Note: tablewriter v1.0.9+ API uses:
//   - table.Header(elements ...any) - no return value
//   - table.Append(rows ...interface{}) - for each row
//     Do NOT use SetHeader() or Row() methods (they don't exist)
func renderTableResults(result *gitrepo.BulkUpdateResult) {
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Repository", "Status", "Details")

	// Add rows
	for _, repo := range result.Repositories {
		statusIcon := getStatusIcon(repo.Status)
		statusDisplay := formatStatusDisplay(repo.Status)
		statusText := color.New(getStatusColor(repo.Status)).Sprint(statusIcon + " " + statusDisplay)

		table.Append(
			repo.RelativePath,
			statusText,
			repo.Message,
		)
	}

	table.Render()

	// Print summary
	fmt.Println()
	fmt.Printf("📊 Summary:\n")
	for status, count := range result.Summary {
		statusIcon := getStatusIcon(status)
		fmt.Printf("  %s %-15s: %d\n", statusIcon, status, count)
	}
	fmt.Printf("\n⏱️  Total duration: %s\n", result.Duration)
}

// renderJSONResults renders results in JSON format
func renderJSONResults(result *gitrepo.BulkUpdateResult) {
	output := map[string]any{
		"totalScanned":   result.TotalScanned,
		"totalProcessed": result.TotalProcessed,
		"duration":       result.Duration.String(),
		"summary":        result.Summary,
		"repositories":   result.Repositories,
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode JSON: %v\n", err)
	}
}

// getStatusIcon returns an icon for the status
func getStatusIcon(status string) string {
	switch status {
	case "updated":
		return "✅"
	case "up-to-date":
		return "✓"
	case "skipped":
		return "⏭️"
	case "would-update":
		return "🔄"
	case "no-upstream":
		return "⚠️"
	case "error":
		return "❌"
	default:
		return "•"
	}
}

// getStatusColor returns a color attribute for the status
func getStatusColor(status string) color.Attribute {
	switch status {
	case "updated":
		return color.FgGreen
	case "up-to-date":
		return color.FgGreen
	case "skipped":
		return color.FgYellow
	case "would-update":
		return color.FgCyan
	case "no-upstream":
		return color.FgYellow
	case "error":
		return color.FgRed
	default:
		return color.FgWhite
	}
}

// formatStatusDisplay formats status for display
func formatStatusDisplay(status string) string {
	switch status {
	case "up-to-date":
		return "최신 상태"
	case "updated":
		return "업데이트됨"
	case "skipped":
		return "건너뜀"
	case "would-update":
		return "업데이트 예정"
	case "no-upstream":
		return "upstream 없음"
	case "error":
		return "오류"
	default:
		return status
	}
}
