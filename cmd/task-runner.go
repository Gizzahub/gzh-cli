package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// NewTaskRunnerCmd creates the task runner command.
func NewTaskRunnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task-runner [directory]",
		Short: "🚀 TASK_RUNNER.todo - 자동 TODO 작업 실행기",
		Long: `TASK_RUNNER.todo 프롬프트 시스템

/tasks/todo/ 디렉터리의 미완료 TODO 파일을 순차적으로 읽어 
작업 → 커밋 → 완료 파일 이동을 자동화합니다.

사용법:
  gz task-runner                      # /tasks/todo 디렉터리 처리
  gz task-runner /tasks/todo/feature  # 특정 디렉터리 처리

절차:
1. 다음 [ ] 미완료 항목 하나 선택 (파일명 오름차순 → 항목 순서)
2. 분석 & 의존성 파악 
3. 구현 & 테스트 & 문서화
4. 포맷 & 커밋 ([x] 체크 후)
5. 모든 항목 완료시 파일을 /tasks/done/으로 이동`,
		RunE: runTaskRunner,
	}

	cmd.Flags().BoolP("dry-run", "n", false, "Show next task without executing")
	cmd.Flags().BoolP("list", "l", false, "List all incomplete tasks")

	return cmd
}

type TodoItem struct {
	File        string
	Line        int
	Content     string
	IsCompleted bool
	IsBlocked   bool // [>] blocked items
}

type TodoFile struct {
	Path  string
	Items []TodoItem
}

func runTaskRunner(cmd *cobra.Command, args []string) error {
	// Determine directory to process
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	dir := filepath.Join(wd, "tasks", "todo")
	if len(args) > 0 {
		dir = args[0]
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(wd, dir)
		}
	}

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	listMode, _ := cmd.Flags().GetBool("list")

	// Find all TODO files
	todoFiles, err := findTodoFiles(dir)
	if err != nil {
		return fmt.Errorf("failed to find TODO files: %w", err)
	}

	if len(todoFiles) == 0 {
		fmt.Println("✅ No TODO files found or all tasks completed!")
		return nil
	}

	// Parse all TODO items
	allItems := []TodoItem{}

	for _, file := range todoFiles {
		items, err := parseTodoFile(file)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse %s: %v\n", file, err)
			continue
		}

		allItems = append(allItems, items...)
	}

	// Filter incomplete items
	incompleteItems := []TodoItem{}

	for _, item := range allItems {
		if !item.IsCompleted && !item.IsBlocked {
			incompleteItems = append(incompleteItems, item)
		}
	}

	if len(incompleteItems) == 0 {
		fmt.Println("🎉 All tasks completed! Moving files to /tasks/done/")
		return moveCompletedFiles(todoFiles)
	}

	if listMode {
		return listIncompleteTasks(incompleteItems)
	}

	// Get next task (first incomplete item in first file)
	nextTask := incompleteItems[0]

	if dryRun {
		fmt.Printf("🎯 Next task to execute:\n")
		fmt.Printf("   File: %s:%d\n", nextTask.File, nextTask.Line)
		fmt.Printf("   Task: %s\n", nextTask.Content)

		return nil
	}

	// Execute the task
	return executeTask(nextTask)
}

func findTodoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort files alphabetically
	sort.Strings(files)

	return files, nil
}

func parseTodoFile(filename string) ([]TodoItem, error) {
	// Validate filename to prevent directory traversal
	if !filepath.IsAbs(filename) {
		return nil, fmt.Errorf("filename must be absolute: %s", filename)
	}
	file, err := os.Open(filepath.Clean(filename))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []TodoItem

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for different TODO item types
	incompletePattern := regexp.MustCompile(`^\s*-\s+\[\s+\]\s+(.+)`)
	completedPattern := regexp.MustCompile(`^\s*-\s+\[x\]\s+(.+)`)
	blockedPattern := regexp.MustCompile(`^\s*-\s+\[>\]\s+(.+)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var (
			content                string
			isCompleted, isBlocked bool
		)

		if match := incompletePattern.FindStringSubmatch(line); match != nil {
			content = strings.TrimSpace(match[1])
			isCompleted = false
			isBlocked = false
		} else if match := completedPattern.FindStringSubmatch(line); match != nil {
			content = strings.TrimSpace(match[1])
			isCompleted = true
			isBlocked = false
		} else if match := blockedPattern.FindStringSubmatch(line); match != nil {
			content = strings.TrimSpace(match[1])
			isCompleted = false
			isBlocked = true
		} else {
			continue // Not a TODO item
		}

		items = append(items, TodoItem{
			File:        filename,
			Line:        lineNum,
			Content:     content,
			IsCompleted: isCompleted,
			IsBlocked:   isBlocked,
		})
	}

	return items, scanner.Err()
}

func listIncompleteTasks(items []TodoItem) error {
	fmt.Println("📋 Incomplete TODO items:")
	fmt.Println()

	currentFile := ""
	for i, item := range items {
		if item.File != currentFile {
			currentFile = item.File
			relPath := strings.TrimPrefix(item.File, "/home/archmagece/myopen/Gizzahub/gzh-manager-go/")
			fmt.Printf("📁 %s\n", relPath)
		}

		status := "🔲"
		if i == 0 {
			status = "🎯" // Next task indicator
		}

		fmt.Printf("   %s Line %d: %s\n", status, item.Line, item.Content)
	}

	fmt.Printf("\n🎯 Next task: %s:%d\n", items[0].File, items[0].Line)

	return nil
}

func executeTask(task TodoItem) error {
	fmt.Printf("🚀 Executing task from %s:%d\n", task.File, task.Line)
	fmt.Printf("📝 Task: %s\n", task.Content)
	fmt.Println()

	// Here we would implement the actual task execution logic
	// For now, we'll provide guidance and mark it as completed

	fmt.Println("🔍 Task Analysis:")
	fmt.Printf("   • Task content: %s\n", task.Content)
	fmt.Printf("   • File location: %s:%d\n", task.File, task.Line)

	// Analyze task content to determine implementation approach
	taskAnalysis := analyzeTask(task.Content)
	fmt.Printf("   • Implementation type: %s\n", taskAnalysis.Type)
	fmt.Printf("   • Estimated complexity: %s\n", taskAnalysis.Complexity)
	fmt.Printf("   • Dependencies: %v\n", taskAnalysis.Dependencies)

	fmt.Println()
	fmt.Println("🛠️  Implementation Steps:")

	for i, step := range taskAnalysis.Steps {
		fmt.Printf("   %d. %s\n", i+1, step)
	}

	fmt.Println()
	fmt.Println("⚠️  Note: This is a preview implementation.")
	fmt.Println("   For actual task execution, integrate with specific implementation logic.")
	fmt.Println("   After implementation, mark task as [x] and commit changes.")

	return nil
}

type TaskAnalysis struct {
	Type         string
	Complexity   string
	Dependencies []string
	Steps        []string
}

func analyzeTask(content string) TaskAnalysis {
	content = strings.ToLower(content)

	analysis := TaskAnalysis{
		Type:         "general",
		Complexity:   "medium",
		Dependencies: []string{},
		Steps:        []string{},
	}

	// Analyze task type
	if strings.Contains(content, "api") || strings.Contains(content, "구현") {
		analysis.Type = "implementation"
		analysis.Complexity = "high"
		analysis.Steps = []string{
			"API 스키마 설계",
			"핸들러 함수 구현",
			"라우팅 설정",
			"테스트 코드 작성",
			"문서화 업데이트",
		}
	} else if strings.Contains(content, "테스트") || strings.Contains(content, "test") {
		analysis.Type = "testing"
		analysis.Complexity = "medium"
		analysis.Steps = []string{
			"테스트 케이스 설계",
			"테스트 코드 작성",
			"모킹 설정",
			"커버리지 확인",
		}
	} else if strings.Contains(content, "문서") || strings.Contains(content, "documentation") {
		analysis.Type = "documentation"
		analysis.Complexity = "low"
		analysis.Steps = []string{
			"문서 구조 설계",
			"내용 작성",
			"예제 코드 추가",
			"리뷰 및 검증",
		}
	} else if strings.Contains(content, "설정") || strings.Contains(content, "config") {
		analysis.Type = "configuration"
		analysis.Complexity = "medium"
		analysis.Steps = []string{
			"설정 스키마 정의",
			"기본값 설정",
			"검증 로직 구현",
			"마이그레이션 스크립트 작성",
		}
	}

	// Analyze dependencies
	if strings.Contains(content, "웹훅") {
		analysis.Dependencies = append(analysis.Dependencies, "GitHub API", "HTTP client")
	}

	if strings.Contains(content, "api") {
		analysis.Dependencies = append(analysis.Dependencies, "REST framework", "authentication")
	}

	return analysis
}

func moveCompletedFiles(files []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	doneDir := filepath.Join(wd, "tasks", "done")

	// Create done directory if it doesn't exist
	if err := os.MkdirAll(doneDir, 0o750); err != nil {
		return fmt.Errorf("failed to create done directory: %w", err)
	}

	timestamp := time.Now().Format("20060102")

	for _, file := range files {
		// Check if all tasks in file are completed
		items, err := parseTodoFile(file)
		if err != nil {
			continue
		}

		allCompleted := true

		for _, item := range items {
			if !item.IsCompleted && !item.IsBlocked {
				allCompleted = false
				break
			}
		}

		if !allCompleted {
			continue
		}

		// Move file to done directory with timestamp
		basename := filepath.Base(file)
		ext := filepath.Ext(basename)
		name := strings.TrimSuffix(basename, ext)
		newName := fmt.Sprintf("%s__DONE_%s%s", name, timestamp, ext)
		newPath := filepath.Join(doneDir, newName)

		if err := os.Rename(file, newPath); err != nil {
			fmt.Printf("⚠️  Failed to move %s: %v\n", file, err)
		} else {
			fmt.Printf("✅ Moved %s to %s\n", basename, newName)
		}
	}

	return nil
}
