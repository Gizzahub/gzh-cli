// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// BulkUpdateOptions represents options for bulk repository updates.
type BulkUpdateOptions struct {
	Directory      string
	Parallel       int
	MaxDepth       int
	DryRun         bool
	JSON           bool
	Verbose        bool
	NoFetch        bool
	IncludePattern string
	ExcludePattern string
}

// RepoStatus represents the status of a repository after processing.
type RepoStatus struct {
	Path          string        `json:"path"`
	Status        string        `json:"status"`
	StatusIcon    string        `json:"statusIcon"`
	Details       string        `json:"details"`
	Error         error         `json:"error,omitempty"`
	Duration      time.Duration `json:"duration"`
	Branch        string        `json:"branch,omitempty"`
	RemoteURL     string        `json:"remoteUrl,omitempty"`
	CommitsBehind int           `json:"commitsBehind"`
	CommitsAhead  int           `json:"commitsAhead"`
	HasStash      bool          `json:"hasStash"`
	InMergeState  bool          `json:"inMergeState"`
}

// BulkUpdateExecutor handles the bulk update operation.
type BulkUpdateExecutor struct {
	options   BulkUpdateOptions
	ctx       context.Context
	results   []RepoStatus
	resultsMu sync.Mutex
}

// newRepoBulkUpdateCmd creates the bulk update command for repositories.
func newRepoBulkUpdateCmd() *cobra.Command {
	var opts BulkUpdateOptions

	cmd := &cobra.Command{
		Use:   "pull-all [directory]",
		Short: "Recursively update all Git repositories with pull --rebase",
		Long: `재귀적으로 하위 디렉토리의 모든 Git 리포지터리를 스캔하고 안전하게 업데이트합니다.

이 명령어는 다음 조건에서만 자동으로 pull --rebase를 실행합니다:
- 로컬 변경사항이 없는 경우 (clean working tree)
- 충돌이 예상되지 않는 경우
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
			// 디렉토리 인자 처리
			if len(args) > 0 {
				opts.Directory = args[0]
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return fmt.Errorf("현재 디렉토리를 가져올 수 없습니다: %w", err)
				}
				opts.Directory = cwd
			}

			// 절대 경로로 변환
			absPath, err := filepath.Abs(opts.Directory)
			if err != nil {
				return fmt.Errorf("절대 경로로 변환할 수 없습니다: %w", err)
			}
			opts.Directory = absPath

			executor := NewBulkUpdateExecutor(cmd.Context(), opts)
			return executor.Execute()
		},
	}

	// 플래그 설정
	cmd.Flags().IntVarP(&opts.Parallel, "parallel", "p", 5, "병렬 처리 워커 수")
	cmd.Flags().IntVar(&opts.MaxDepth, "max-depth", 10, "최대 스캔 깊이")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "실제 실행하지 않고 시뮬레이션만 수행")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "JSON 형식으로 결과 출력")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "상세 로그 출력")
	cmd.Flags().BoolVar(&opts.NoFetch, "no-fetch", false, "원격 저장소에서 변경사항을 가져오지 않음")
	cmd.Flags().StringVar(&opts.IncludePattern, "include-pattern", "", "포함할 리포지터리 패턴 (정규식)")
	cmd.Flags().StringVar(&opts.ExcludePattern, "exclude-pattern", "", "제외할 리포지터리 패턴 (정규식)")

	return cmd
}

// NewBulkUpdateExecutor creates a new bulk update executor.
func NewBulkUpdateExecutor(ctx context.Context, opts BulkUpdateOptions) *BulkUpdateExecutor {
	return &BulkUpdateExecutor{
		options: opts,
		ctx:     ctx,
		results: make([]RepoStatus, 0),
	}
}

// Execute runs the bulk update operation.
func (e *BulkUpdateExecutor) Execute() error {
	// 1. 디렉토리 검증
	if err := e.validateDirectory(); err != nil {
		return err
	}

	// 2. Git 리포지터리 스캔
	fmt.Printf("🔍 Git 리포지터리 스캔 중: %s\n", color.CyanString(e.options.Directory))
	repos, err := e.scanRepositories()
	if err != nil {
		return fmt.Errorf("리포지터리 스캔 실패: %w", err)
	}

	// 3. 리포지터리 필터링
	filteredRepos := e.filterRepositories(repos)

	if len(filteredRepos) == 0 {
		if len(repos) > 0 {
			fmt.Printf("필터링된 결과: Git 리포지터리 %d개 중 처리 대상 없음\n", len(repos))
		} else {
			fmt.Println("Git 리포지터리를 찾을 수 없습니다.")
		}
		return nil
	}

	if len(repos) != len(filteredRepos) {
		fmt.Printf("📦 발견된 리포지터리: %d개, 처리 대상: %s개\n\n", len(repos), color.GreenString(strconv.Itoa(len(filteredRepos))))
	} else {
		fmt.Printf("📦 발견된 리포지터리: %s개\n\n", color.GreenString(strconv.Itoa(len(filteredRepos))))
	}

	// 4. 병렬 처리로 업데이트 실행
	if err := e.processRepositories(filteredRepos); err != nil {
		return fmt.Errorf("리포지터리 처리 실패: %w", err)
	}

	// 5. 결과 출력
	e.renderResults()

	return nil
}

// validateDirectory validates the target directory.
func (e *BulkUpdateExecutor) validateDirectory() error {
	info, err := os.Stat(e.options.Directory)
	if err != nil {
		return fmt.Errorf("디렉토리 접근 불가: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("지정된 경로는 디렉토리가 아닙니다: %s", e.options.Directory)
	}

	return nil
}

// scanRepositories recursively scans for Git repositories.
func (e *BulkUpdateExecutor) scanRepositories() ([]string, error) {
	var repos []string

	err := e.walkDirectory(e.options.Directory, 0, &repos)
	if err != nil {
		return nil, err
	}

	// 경로 기준으로 정렬
	sort.Strings(repos)

	return repos, nil
}

// filterRepositories filters repositories based on include and exclude patterns.
func (e *BulkUpdateExecutor) filterRepositories(repos []string) []string {
	if e.options.IncludePattern == "" && e.options.ExcludePattern == "" {
		return repos
	}

	var includeRegex, excludeRegex *regexp.Regexp
	var err error

	// Include 패턴 컴파일
	if e.options.IncludePattern != "" {
		includeRegex, err = regexp.Compile(e.options.IncludePattern)
		if err != nil {
			if e.options.Verbose {
				fmt.Printf("⚠️  Include 패턴 오류 (무시됨): %v\n", err)
			}
			includeRegex = nil
		}
	}

	// Exclude 패턴 컴파일
	if e.options.ExcludePattern != "" {
		excludeRegex, err = regexp.Compile(e.options.ExcludePattern)
		if err != nil {
			if e.options.Verbose {
				fmt.Printf("⚠️  Exclude 패턴 오류 (무시됨): %v\n", err)
			}
			excludeRegex = nil
		}
	}

	var filtered []string
	for _, repo := range repos {
		// 상대 경로로 변환하여 패턴 매칭
		relPath := e.getRelativePath(repo)

		// Include 패턴 확인
		if includeRegex != nil {
			if !includeRegex.MatchString(relPath) && !includeRegex.MatchString(repo) {
				continue
			}
		}

		// Exclude 패턴 확인
		if excludeRegex != nil {
			if excludeRegex.MatchString(relPath) || excludeRegex.MatchString(repo) {
				if e.options.Verbose {
					fmt.Printf("⏭️  제외됨: %s (exclude 패턴)\n", relPath)
				}
				continue
			}
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// walkDirectory recursively walks directories to find Git repositories.
func (e *BulkUpdateExecutor) walkDirectory(dir string, depth int, repos *[]string) error {
	if depth > e.options.MaxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// 권한 없는 디렉토리는 무시
		if e.options.Verbose {
			fmt.Printf("⚠️  디렉토리 읽기 실패 (무시됨): %s\n", dir)
		}
		return nil
	}

	// 현재 디렉토리가 Git 리포지터리인지 확인
	gitDir := filepath.Join(dir, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		*repos = append(*repos, dir)
		// Git 리포지터리 내부는 더 이상 스캔하지 않음
		return nil
	}

	// 하위 디렉토리 탐색
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// 무시할 디렉토리들
		if e.shouldIgnoreDirectory(name) {
			continue
		}

		subPath := filepath.Join(dir, name)

		// 심볼릭 링크 무시
		if info, err := entry.Info(); err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				continue
			}
		}

		if err := e.walkDirectory(subPath, depth+1, repos); err != nil {
			// 개별 디렉토리 오류는 로그만 남기고 계속 진행
			if e.options.Verbose {
				fmt.Printf("⚠️  하위 디렉토리 스캔 실패: %s (%v)\n", subPath, err)
			}
		}
	}

	return nil
}

// shouldIgnoreDirectory checks if a directory should be ignored during scanning.
func (e *BulkUpdateExecutor) shouldIgnoreDirectory(name string) bool {
	ignorePatterns := []string{
		".git", "node_modules", ".venv", "venv", "__pycache__",
		"target", "build", "dist", ".gradle", ".idea", ".vscode",
		"vendor", "deps", ".next", ".nuxt", "coverage",
	}

	for _, pattern := range ignorePatterns {
		if name == pattern {
			return true
		}
	}

	return false
}

// processRepositories processes all repositories concurrently.
func (e *BulkUpdateExecutor) processRepositories(repos []string) error {
	// 동시 실행 제한을 위한 errgroup 사용
	g, ctx := errgroup.WithContext(e.ctx)
	g.SetLimit(e.options.Parallel)

	// Progress indicator
	if !e.options.JSON && !e.options.Verbose {
		fmt.Print("처리 중: ")
	}

	for _, repoPath := range repos {
		repoPath := repoPath // 클로저를 위한 복사
		g.Go(func() error {
			result := e.processRepository(ctx, repoPath)

			e.resultsMu.Lock()
			e.results = append(e.results, result)
			e.resultsMu.Unlock()

			// Progress indicator
			if !e.options.JSON && !e.options.Verbose {
				fmt.Print(".")
			}

			return nil // 개별 리포지터리 오류는 무시하고 계속 진행
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	if !e.options.JSON && !e.options.Verbose {
		fmt.Println() // Progress indicator 마무리
	}

	return nil
}

// processRepository processes a single repository.
func (e *BulkUpdateExecutor) processRepository(ctx context.Context, repoPath string) RepoStatus {
	start := time.Now()

	result := RepoStatus{
		Path:     e.getRelativePath(repoPath),
		Duration: 0,
	}

	defer func() {
		result.Duration = time.Since(start)
	}()

	// Git 정보 수집
	if branch, err := e.getCurrentBranch(ctx, repoPath); err == nil {
		result.Branch = branch
	}

	if remoteURL, err := e.getRemoteURL(ctx, repoPath); err == nil {
		result.RemoteURL = remoteURL
	}

	// 추가 상태 정보 수집
	result.InMergeState = e.isInMergeState(ctx, repoPath)
	if hasStash, err := e.hasStashedChanges(ctx, repoPath); err == nil {
		result.HasStash = hasStash
	}

	// 안전성 검사
	safetyCheck, err := e.checkRepositorySafety(ctx, repoPath)
	if err != nil {
		result.Status = "error"
		result.StatusIcon = "❌"
		result.Details = fmt.Sprintf("상태 확인 실패: %v", err)
		result.Error = err
		return result
	}

	// 상태에 따른 처리
	switch safetyCheck.Status {
	case "safe":
		if e.options.DryRun {
			result.Status = "would-update"
			result.StatusIcon = "🔍"
			result.Details = "업데이트 예정 (dry-run)"
		} else {
			pullResult := e.performPullRebase(ctx, repoPath)
			result.Status = pullResult.Status
			result.StatusIcon = pullResult.StatusIcon
			result.Details = pullResult.Details
			result.Error = pullResult.Error
		}

	case "uptodate":
		result.Status = "uptodate"
		result.StatusIcon = "⏭️"
		result.Details = "이미 최신 상태"

	case "dirty":
		result.Status = "dirty"
		result.StatusIcon = "⚠️"
		result.Details = safetyCheck.Details

	case "conflicts":
		result.Status = "conflicts"
		result.StatusIcon = "🔧"
		result.Details = safetyCheck.Details

	case "no-upstream":
		result.Status = "no-upstream"
		result.StatusIcon = "🚫"
		result.Details = "원격 브랜치가 설정되지 않음"

	case "merge-in-progress":
		result.Status = "merge-in-progress"
		result.StatusIcon = "🔀"
		result.Details = safetyCheck.Details

	default:
		result.Status = "unknown"
		result.StatusIcon = "❓"
		result.Details = safetyCheck.Details
	}

	if e.options.Verbose {
		fmt.Printf("✓ %s: %s %s\n", result.Path, result.StatusIcon, result.Details)
	}

	return result
}

// SafetyCheckResult represents the result of repository safety checks.
type SafetyCheckResult struct {
	Status  string // safe, dirty, conflicts, no-upstream, uptodate
	Details string
}

// checkRepositorySafety performs comprehensive safety checks on a repository.
func (e *BulkUpdateExecutor) checkRepositorySafety(ctx context.Context, repoPath string) (*SafetyCheckResult, error) {
	// 0. 병합/리베이스 진행 중인지 확인
	if e.isInMergeState(ctx, repoPath) {
		return &SafetyCheckResult{
			Status:  "merge-in-progress",
			Details: "병합 또는 리베이스가 진행 중임",
		}, nil
	}

	// 1. Working tree 상태 확인
	if dirty, err := e.hasUncommittedChanges(ctx, repoPath); err != nil {
		return nil, fmt.Errorf("working tree 상태 확인 실패: %w", err)
	} else if dirty {
		return &SafetyCheckResult{
			Status:  "dirty",
			Details: "커밋되지 않은 변경사항 있음",
		}, nil
	}

	// 2. upstream 브랜치 확인
	hasUpstream, err := e.hasUpstreamBranch(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("upstream 확인 실패: %w", err)
	}
	if !hasUpstream {
		return &SafetyCheckResult{
			Status:  "no-upstream",
			Details: "원격 브랜치가 설정되지 않음",
		}, nil
	}

	// 3. 원격 브랜치와의 상태 확인
	behind, ahead, err := e.getCommitComparison(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("커밋 비교 실패: %w", err)
	}

	if behind == 0 {
		return &SafetyCheckResult{
			Status:  "uptodate",
			Details: "이미 최신 상태",
		}, nil
	}

	// 4. pull 시 충돌 가능성 확인 (ahead가 있는 경우)
	if ahead > 0 {
		return &SafetyCheckResult{
			Status:  "conflicts",
			Details: fmt.Sprintf("로컬 커밋 %d개와 원격 커밋 %d개가 있어 충돌 가능", ahead, behind),
		}, nil
	}

	// 모든 검사 통과 - 안전하게 pull 가능
	return &SafetyCheckResult{
		Status:  "safe",
		Details: fmt.Sprintf("%d개 커밋 업데이트 가능", behind),
	}, nil
}

// Git 유틸리티 함수들
func (e *BulkUpdateExecutor) getCurrentBranch(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (e *BulkUpdateExecutor) getRemoteURL(ctx context.Context, repoPath string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func (e *BulkUpdateExecutor) hasUncommittedChanges(ctx context.Context, repoPath string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func (e *BulkUpdateExecutor) hasUpstreamBranch(ctx context.Context, repoPath string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}")
	cmd.Dir = repoPath
	_, err := cmd.Output()
	return err == nil, nil
}

func (e *BulkUpdateExecutor) getCommitComparison(ctx context.Context, repoPath string) (behind, ahead int, err error) {
	// NoFetch 옵션이 설정되지 않은 경우에만 fetch 실행
	if !e.options.NoFetch {
		// Context에 타임아웃 추가
		fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		fetchCmd := exec.CommandContext(fetchCtx, "git", "fetch", "--quiet")
		fetchCmd.Dir = repoPath
		if err := fetchCmd.Run(); err != nil {
			// fetch 실패는 무시하되 verbose 모드에서는 로그 남김
			if e.options.Verbose {
				fmt.Printf("⚠️  Fetch failed for %s: %v\n", repoPath, err)
			}
		}
	}

	// rev-list로 비교
	cmd := exec.CommandContext(ctx, "git", "rev-list", "--left-right", "--count", "HEAD...@{u}")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("unexpected rev-list output: %s", output)
	}

	ahead, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}

	behind, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}

	return behind, ahead, nil
}

// isInMergeState checks if the repository is in a merge or rebase state.
func (e *BulkUpdateExecutor) isInMergeState(ctx context.Context, repoPath string) bool {
	// 병합 진행 중 확인
	mergeHeadPath := filepath.Join(repoPath, ".git", "MERGE_HEAD")
	if _, err := os.Stat(mergeHeadPath); err == nil {
		return true
	}

	// 리베이스 진행 중 확인
	rebaseHeadPath := filepath.Join(repoPath, ".git", "rebase-merge")
	if _, err := os.Stat(rebaseHeadPath); err == nil {
		return true
	}

	rebaseApplyPath := filepath.Join(repoPath, ".git", "rebase-apply")
	if _, err := os.Stat(rebaseApplyPath); err == nil {
		return true
	}

	return false
}

// hasStashedChanges checks if there are any stashed changes.
func (e *BulkUpdateExecutor) hasStashedChanges(ctx context.Context, repoPath string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "stash", "list")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

// PullResult represents the result of a pull operation.
type PullResult struct {
	Status     string
	StatusIcon string
	Details    string
	Error      error
}

// performPullRebase performs git pull --rebase operation.
func (e *BulkUpdateExecutor) performPullRebase(ctx context.Context, repoPath string) PullResult {
	cmd := exec.CommandContext(ctx, "git", "pull", "--rebase")
	cmd.Dir = repoPath

	output, err := cmd.CombinedOutput()
	outputStr := strings.TrimSpace(string(output))

	if err != nil {
		return PullResult{
			Status:     "failed",
			StatusIcon: "❌",
			Details:    fmt.Sprintf("Pull 실패: %s", outputStr),
			Error:      err,
		}
	}

	if strings.Contains(outputStr, "Already up to date") ||
		strings.Contains(outputStr, "Current branch") && strings.Contains(outputStr, "is up to date") {
		return PullResult{
			Status:     "uptodate",
			StatusIcon: "⏭️",
			Details:    "이미 최신 상태",
		}
	}

	// 성공적인 업데이트
	lines := strings.Split(outputStr, "\n")
	details := "업데이트 완료"
	if len(lines) > 0 && lines[0] != "" {
		details = lines[0]
	}

	return PullResult{
		Status:     "updated",
		StatusIcon: "✅",
		Details:    details,
	}
}

// getRelativePath returns a relative path for display purposes.
func (e *BulkUpdateExecutor) getRelativePath(fullPath string) string {
	rel, err := filepath.Rel(e.options.Directory, fullPath)
	if err != nil {
		return fullPath
	}
	if rel == "." {
		return "./"
	}
	if !strings.HasPrefix(rel, "./") {
		return "./" + rel
	}
	return rel
}

// renderResults renders the final results in a table format.
func (e *BulkUpdateExecutor) renderResults() {
	if len(e.results) == 0 {
		return
	}

	// 결과 정렬 (경로 기준)
	sort.Slice(e.results, func(i, j int) bool {
		return e.results[i].Path < e.results[j].Path
	})

	if e.options.JSON {
		e.renderJSONResults()
		return
	}

	e.renderTableResults()
}

// renderTableResults renders results as a formatted table.
func (e *BulkUpdateExecutor) renderTableResults() {
	fmt.Println()

	table := tablewriter.NewWriter(os.Stdout)
	table.Header("Repository", "Status", "Details")

	// 상태별 카운터
	statusCounts := make(map[string]int)

	for _, result := range e.results {
		statusText := fmt.Sprintf("%s %s", result.StatusIcon, result.Status)

		// 컬러 적용
		switch result.Status {
		case "updated":
			statusText = color.GreenString(statusText)
		case "uptodate", "would-update":
			statusText = color.BlueString(statusText)
		case "dirty", "conflicts", "merge-in-progress":
			statusText = color.YellowString(statusText)
		case "failed", "error":
			statusText = color.RedString(statusText)
		case "no-upstream":
			statusText = color.MagentaString(statusText)
		}

		err := table.Append(
			result.Path,
			statusText,
			result.Details,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error adding table row: %v\n", err)
		}

		statusCounts[result.Status]++
	}

	err := table.Render()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering table: %v\n", err)
	}

	// 요약 출력
	fmt.Printf("\n📊 %s:\n", color.CyanString("요약"))
	for status, count := range statusCounts {
		var emoji string
		switch status {
		case "updated":
			emoji = "✅"
		case "uptodate", "would-update":
			emoji = "⏭️"
		case "dirty":
			emoji = "⚠️"
		case "conflicts":
			emoji = "🔧"
		case "failed", "error":
			emoji = "❌"
		case "no-upstream":
			emoji = "🚫"
		case "merge-in-progress":
			emoji = "🔀"
		default:
			emoji = "❓"
		}
		fmt.Printf("- %s %s: %d\n", emoji, status, count)
	}
	fmt.Println()
}

// JSONOutput represents the complete JSON output structure.
type JSONOutput struct {
	Directory    string            `json:"directory"`
	TotalRepos   int               `json:"totalRepos"`
	ProcessedAt  time.Time         `json:"processedAt"`
	Options      BulkUpdateOptions `json:"options"`
	Repositories []RepoStatus      `json:"repositories"`
	Summary      map[string]int    `json:"summary"`
}

// getSummary returns a summary of repository statuses.
func (e *BulkUpdateExecutor) getSummary() map[string]int {
	summary := make(map[string]int)
	for _, result := range e.results {
		summary[result.Status]++
	}
	return summary
}

// renderJSONResults renders results in JSON format.
func (e *BulkUpdateExecutor) renderJSONResults() {
	output := JSONOutput{
		Directory:    e.options.Directory,
		TotalRepos:   len(e.results),
		ProcessedAt:  time.Now(),
		Options:      e.options,
		Repositories: e.results,
		Summary:      e.getSummary(),
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "JSON 출력 오류: %v\n", err)
	}
}
