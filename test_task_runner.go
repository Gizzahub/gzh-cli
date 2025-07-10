package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type TodoItem struct {
	File        string
	Line        int
	Content     string
	IsCompleted bool
	IsBlocked   bool // [>] blocked items
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_task_runner.go [list|next|dry-run]")
		return
	}

	mode := os.Args[1]
	todoDir := "./tasks/todo"

	switch mode {
	case "list":
		listAllTasks(todoDir)
	case "next":
		showNextTask(todoDir)
	case "dry-run":
		showNextTask(todoDir)
	case "execute":
		executeNextTask(todoDir)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		fmt.Println("Available modes: list, next, dry-run, execute")
	}
}

func listAllTasks(dir string) {
	files, err := findTodoFiles(dir)
	if err != nil {
		fmt.Printf("Error finding files: %v\n", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("✅ No TODO files found!")
		return
	}

	allItems := []TodoItem{}
	for _, file := range files {
		items, err := parseTodoFile(file)
		if err != nil {
			fmt.Printf("⚠️  Warning: failed to parse %s: %v\n", file, err)
			continue
		}
		allItems = append(allItems, items...)
	}

	incompleteItems := []TodoItem{}
	for _, item := range allItems {
		if !item.IsCompleted && !item.IsBlocked {
			incompleteItems = append(incompleteItems, item)
		}
	}

	fmt.Printf("📋 Found %d incomplete TODO items:\n\n", len(incompleteItems))

	currentFile := ""
	for i, item := range incompleteItems {
		if item.File != currentFile {
			currentFile = item.File
			relPath := strings.TrimPrefix(item.File, "./")
			fmt.Printf("📁 %s\n", relPath)
		}

		status := "🔲"
		if i == 0 {
			status = "🎯" // Next task indicator
		}

		fmt.Printf("   %s Line %d: %s\n", status, item.Line, item.Content)
	}

	if len(incompleteItems) > 0 {
		fmt.Printf("\n🎯 Next task: %s:%d\n", incompleteItems[0].File, incompleteItems[0].Line)
	}
}

func showNextTask(dir string) {
	nextTask, err := findNextTask(dir)
	if err != nil {
		fmt.Printf("Error finding next task: %v\n", err)
		return
	}

	if nextTask == nil {
		fmt.Println("🎉 All tasks completed!")
		return
	}

	fmt.Printf("🎯 Next task to execute:\n")
	fmt.Printf("   File: %s:%d\n", nextTask.File, nextTask.Line)
	fmt.Printf("   Task: %s\n", nextTask.Content)

	// Analyze task
	analysis := analyzeTask(nextTask.Content)
	fmt.Printf("\n🔍 Task Analysis:\n")
	fmt.Printf("   • Type: %s\n", analysis.Type)
	fmt.Printf("   • Complexity: %s\n", analysis.Complexity)
	fmt.Printf("   • Dependencies: %v\n", analysis.Dependencies)

	fmt.Printf("\n🛠️  Implementation Steps:\n")
	for i, step := range analysis.Steps {
		fmt.Printf("   %d. %s\n", i+1, step)
	}
}

func executeNextTask(dir string) {
	nextTask, err := findNextTask(dir)
	if err != nil {
		fmt.Printf("Error finding next task: %v\n", err)
		return
	}

	if nextTask == nil {
		fmt.Println("🎉 All tasks completed!")
		return
	}

	fmt.Printf("🚀 Executing task from %s:%d\n", nextTask.File, nextTask.Line)
	fmt.Printf("📝 Task: %s\n", nextTask.Content)
	fmt.Println()

	// Mark task as completed and commit
	if err := markTaskCompleted(nextTask); err != nil {
		fmt.Printf("Error marking task completed: %v\n", err)
		return
	}

	fmt.Println("✅ Task marked as completed!")
	fmt.Println("🔄 Running git commit...")

	// Here you would run git commit
	fmt.Println("   (In real implementation, would run: git add . && git commit -m \"feat: complete task\")")
}

func findNextTask(dir string) (*TodoItem, error) {
	files, err := findTodoFiles(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		items, err := parseTodoFile(file)
		if err != nil {
			continue
		}

		for _, item := range items {
			if !item.IsCompleted && !item.IsBlocked {
				return &item, nil
			}
		}
	}

	return nil, nil
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

	sort.Strings(files)
	return files, nil
}

func parseTodoFile(filename string) ([]TodoItem, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var items []TodoItem
	scanner := bufio.NewScanner(file)
	lineNum := 0

	incompletePattern := regexp.MustCompile(`^\s*-\s+\[\s+\]\s+(.+)`)
	completedPattern := regexp.MustCompile(`^\s*-\s+\[x\]\s+(.+)`)
	blockedPattern := regexp.MustCompile(`^\s*-\s+\[>\]\s+(.+)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		var content string
		var isCompleted, isBlocked bool

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
			continue
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

func markTaskCompleted(task *TodoItem) error {
	// Read the file
	content, err := os.ReadFile(task.File)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	if task.Line <= 0 || task.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d", task.Line)
	}

	// Replace [ ] with [x] on the specific line
	line := lines[task.Line-1]
	updatedLine := regexp.MustCompile(`\[\s+\]`).ReplaceAllString(line, "[x]")
	lines[task.Line-1] = updatedLine

	// Write back to file
	updatedContent := strings.Join(lines, "\n")
	return os.WriteFile(task.File, []byte(updatedContent), 0o644)
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
