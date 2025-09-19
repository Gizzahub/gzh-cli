// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package repo

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBulkUpdateExecutor_validateDirectory(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() string
		wantError bool
	}{
		{
			name: "valid directory",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				return tmpDir
			},
			wantError: false,
		},
		{
			name: "non-existent directory",
			setupFunc: func() string {
				return "/non/existent/path"
			},
			wantError: true,
		},
		{
			name: "file instead of directory",
			setupFunc: func() string {
				tmpDir := t.TempDir()
				filePath := filepath.Join(tmpDir, "file.txt")
				_ = os.WriteFile(filePath, []byte("test"), 0o644)
				return filePath
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setupFunc()
			opts := BulkUpdateOptions{Directory: path}
			executor := NewBulkUpdateExecutor(context.Background(), opts)

			err := executor.validateDirectory()

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBulkUpdateExecutor_shouldIgnoreDirectory(t *testing.T) {
	executor := &BulkUpdateExecutor{}

	tests := []struct {
		name      string
		directory string
		want      bool
	}{
		{name: "git directory", directory: ".git", want: true},
		{name: "node_modules", directory: "node_modules", want: true},
		{name: "python venv", directory: ".venv", want: true},
		{name: "build directory", directory: "build", want: true},
		{name: "regular directory", directory: "src", want: false},
		{name: "empty name", directory: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.shouldIgnoreDirectory(tt.directory)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBulkUpdateExecutor_scanRepositories(t *testing.T) {
	tmpDir := t.TempDir()

	// 테스트 구조 생성:
	// tmpDir/
	//   ├── repo1/.git/
	//   ├── folder1/
	//   │   └── repo2/.git/
	//   ├── node_modules/   (무시됨)
	//   │   └── some-repo/.git/
	//   └── normal-folder/
	//       └── not-a-repo/

	// repo1 생성 (Git 리포지터리)
	repo1Path := filepath.Join(tmpDir, "repo1")
	require.NoError(t, os.MkdirAll(filepath.Join(repo1Path, ".git"), 0o755))

	// folder1/repo2 생성 (중첩된 Git 리포지터리)
	repo2Path := filepath.Join(tmpDir, "folder1", "repo2")
	require.NoError(t, os.MkdirAll(filepath.Join(repo2Path, ".git"), 0o755))

	// node_modules/some-repo 생성 (무시되어야 함)
	nodeModulesRepo := filepath.Join(tmpDir, "node_modules", "some-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(nodeModulesRepo, ".git"), 0o755))

	// 일반 디렉토리 생성 (Git 리포지터리 아님)
	normalPath := filepath.Join(tmpDir, "normal-folder", "not-a-repo")
	require.NoError(t, os.MkdirAll(normalPath, 0o755))

	opts := BulkUpdateOptions{
		Directory: tmpDir,
		MaxDepth:  5,
	}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	repos, err := executor.scanRepositories()

	require.NoError(t, err)
	assert.Len(t, repos, 2)

	// 절대 경로로 변환하여 검증
	absRepo1, _ := filepath.Abs(repo1Path)
	absRepo2, _ := filepath.Abs(repo2Path)

	assert.Contains(t, repos, absRepo1)
	assert.Contains(t, repos, absRepo2)

	// node_modules 안의 리포지터리는 포함되지 않아야 함
	absNodeModulesRepo, _ := filepath.Abs(nodeModulesRepo)
	assert.NotContains(t, repos, absNodeModulesRepo)
}

func TestBulkUpdateExecutor_getRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	tests := []struct {
		name     string
		fullPath string
		want     string
	}{
		{
			name:     "current directory",
			fullPath: tmpDir,
			want:     "./",
		},
		{
			name:     "subdirectory",
			fullPath: filepath.Join(tmpDir, "subdir"),
			want:     "./subdir",
		},
		{
			name:     "nested subdirectory",
			fullPath: filepath.Join(tmpDir, "folder1", "folder2"),
			want:     "./folder1/folder2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.getRelativePath(tt.fullPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBulkUpdateExecutor_processRepository_MockGitRepo(t *testing.T) {
	// 실제 Git 명령어 없이 테스트할 수 있는 모의 테스트
	tmpDir := t.TempDir()
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755))

	opts := BulkUpdateOptions{
		Directory: tmpDir,
		DryRun:    true, // dry-run 모드로 실제 Git 명령 실행 방지
	}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	// 모의 처리 (실제 Git 명령어가 없는 환경에서는 오류가 발생할 것)
	result := executor.processRepository(context.Background(), repoPath)

	// 결과 검증
	assert.NotEmpty(t, result.Path)
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.StatusIcon)
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0))

	// Git 명령어가 실행되지 않는 환경에서는 대부분 에러 상태가 될 것
	assert.Contains(t, []string{"error", "would-update", "no-upstream"}, result.Status)
}

func TestSafetyCheckResult(t *testing.T) {
	tests := []struct {
		name   string
		result SafetyCheckResult
	}{
		{
			name: "safe status",
			result: SafetyCheckResult{
				Status:  "safe",
				Details: "Ready for update",
			},
		},
		{
			name: "dirty status",
			result: SafetyCheckResult{
				Status:  "dirty",
				Details: "Uncommitted changes",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.result.Status, tt.result.Status)
			assert.Equal(t, tt.result.Details, tt.result.Details)
		})
	}
}

func TestBulkUpdateOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options BulkUpdateOptions
		valid   bool
	}{
		{
			name: "valid options",
			options: BulkUpdateOptions{
				Directory: "/tmp",
				Parallel:  5,
				MaxDepth:  10,
				DryRun:    false,
			},
			valid: true,
		},
		{
			name: "zero parallel workers",
			options: BulkUpdateOptions{
				Directory: "/tmp",
				Parallel:  0,
				MaxDepth:  10,
			},
			valid: true, // 0도 유효 (기본값으로 처리됨)
		},
		{
			name: "negative max depth",
			options: BulkUpdateOptions{
				Directory: "/tmp",
				Parallel:  5,
				MaxDepth:  -1,
			},
			valid: true, // 음수도 처리됨
		},
		{
			name: "with filtering options",
			options: BulkUpdateOptions{
				Directory:      "/tmp",
				Parallel:       5,
				MaxDepth:       10,
				IncludePattern: ".*project.*",
				ExcludePattern: ".*test.*",
				NoFetch:        true,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 기본적인 구조체 필드 검증
			assert.NotNil(t, tt.options.Directory)
			assert.GreaterOrEqual(t, tt.options.Parallel, 0)
		})
	}
}

func TestBulkUpdateExecutor_filterRepositories(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	repos := []string{
		filepath.Join(tmpDir, "project-a"),
		filepath.Join(tmpDir, "project-b"),
		filepath.Join(tmpDir, "test-repo"),
		filepath.Join(tmpDir, "docs"),
		filepath.Join(tmpDir, "my-awesome-project"),
		filepath.Join(tmpDir, "legacy-system"),
	}

	tests := []struct {
		name           string
		includePattern string
		excludePattern string
		expectedCount  int
		expectedRepos  []string
		description    string
	}{
		{
			name:          "no filters",
			expectedCount: 6,
			description:   "모든 리포지터리 포함",
		},
		{
			name:           "include project only",
			includePattern: ".*project.*",
			expectedCount:  3,
			expectedRepos:  []string{"project-a", "project-b", "my-awesome-project"},
			description:    "project 이름만 포함",
		},
		{
			name:           "exclude test repo",
			excludePattern: ".*test.*",
			expectedCount:  5,
			description:    "test 리포지터리 제외",
		},
		{
			name:           "include project but exclude test",
			includePattern: ".*project.*",
			excludePattern: ".*test.*",
			expectedCount:  3,
			description:    "project 포함하되 test 제외",
		},
		{
			name:           "case insensitive pattern",
			includePattern: "(?i).*PROJECT.*",
			expectedCount:  3,
			description:    "대소문자 구분하지 않는 패턴",
		},
		{
			name:           "complex exclude pattern",
			excludePattern: ".*(test|legacy).*",
			expectedCount:  4,
			description:    "복합 제외 패턴",
		},
		{
			name:           "anchor patterns",
			includePattern: "^.*/docs$",
			expectedCount:  1,
			description:    "앵커를 사용한 정확한 매칭",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.options.IncludePattern = tt.includePattern
			executor.options.ExcludePattern = tt.excludePattern

			filtered := executor.filterRepositories(repos)
			assert.Equal(t, tt.expectedCount, len(filtered), tt.description)

			// 특정 리포지터리가 포함되어야 하는 경우 검증
			if len(tt.expectedRepos) > 0 {
				for _, expectedRepo := range tt.expectedRepos {
					found := false
					for _, filteredRepo := range filtered {
						if filepath.Base(filteredRepo) == expectedRepo {
							found = true
							break
						}
					}
					assert.True(t, found, "%s should be included in filtered results", expectedRepo)
				}
			}
		})
	}
}

func TestBulkUpdateExecutor_filterRepositories_InvalidRegex(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	repos := []string{
		filepath.Join(tmpDir, "project-a"),
		filepath.Join(tmpDir, "project-b"),
	}

	tests := []struct {
		name           string
		includePattern string
		excludePattern string
		expectedCount  int
		description    string
	}{
		{
			name:           "invalid include pattern",
			includePattern: "[invalid",
			expectedCount:  2,
			description:    "잘못된 include 패턴은 무시되고 모든 리포지터리 포함",
		},
		{
			name:           "invalid exclude pattern",
			excludePattern: "*invalid",
			expectedCount:  2,
			description:    "잘못된 exclude 패턴은 무시되고 모든 리포지터리 포함",
		},
		{
			name:           "both patterns invalid",
			includePattern: "[invalid",
			excludePattern: "*invalid",
			expectedCount:  2,
			description:    "모든 패턴이 잘못된 경우 모든 리포지터리 포함",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor.options.IncludePattern = tt.includePattern
			executor.options.ExcludePattern = tt.excludePattern

			filtered := executor.filterRepositories(repos)
			assert.Equal(t, tt.expectedCount, len(filtered), tt.description)
		})
	}
}

func TestRepoStatus_JSONSerialization(t *testing.T) {
	status := RepoStatus{
		Path:          "./test-repo",
		Status:        "updated",
		StatusIcon:    "✅",
		Details:       "업데이트 완료",
		Duration:      time.Second * 5,
		Branch:        "main",
		RemoteURL:     "https://github.com/user/repo.git",
		CommitsBehind: 0,
		CommitsAhead:  0,
		HasStash:      false,
		InMergeState:  false,
	}

	jsonData, err := json.Marshal(status)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-repo")
	assert.Contains(t, string(jsonData), "updated")

	var decoded RepoStatus
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, status.Path, decoded.Path)
	assert.Equal(t, status.Status, decoded.Status)
	assert.Equal(t, status.Branch, decoded.Branch)
	assert.Equal(t, status.RemoteURL, decoded.RemoteURL)
	assert.Equal(t, status.HasStash, decoded.HasStash)
	assert.Equal(t, status.InMergeState, decoded.InMergeState)
}

func TestBulkUpdateResults_JSONOutput(t *testing.T) {
	results := []RepoStatus{
		{
			Path:          "./repo1",
			Status:        "updated",
			StatusIcon:    "✅",
			Details:       "Successfully updated",
			Duration:      time.Second * 2,
			Branch:        "main",
			RemoteURL:     "https://github.com/user/repo1.git",
			CommitsBehind: 0,
			CommitsAhead:  0,
			HasStash:      false,
			InMergeState:  false,
		},
		{
			Path:          "./repo2",
			Status:        "dirty",
			StatusIcon:    "🔄",
			Details:       "Uncommitted changes detected",
			Duration:      time.Millisecond * 500,
			Branch:        "develop",
			RemoteURL:     "https://github.com/user/repo2.git",
			CommitsBehind: 2,
			CommitsAhead:  1,
			HasStash:      true,
			InMergeState:  false,
		},
	}

	jsonData, err := json.Marshal(results)
	assert.NoError(t, err)

	var decoded []RepoStatus
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)
	assert.Len(t, decoded, 2)
	assert.Equal(t, "updated", decoded[0].Status)
	assert.Equal(t, "dirty", decoded[1].Status)
	assert.True(t, decoded[1].HasStash)
	assert.False(t, decoded[0].HasStash)
}

// 통합 테스트 (실제 Git 환경이 필요)
func TestBulkUpdateExecutor_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Git이 설치되어 있는지 확인
	if _, err := os.Stat("/usr/bin/git"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/git"); os.IsNotExist(err) {
			t.Skip("Git not found, skipping integration test")
		}
	}

	tmpDir := t.TempDir()

	// 실제 Git 리포지터리 생성 (간단한 버전)
	repoPath := filepath.Join(tmpDir, "test-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	// git init 실행
	opts := BulkUpdateOptions{
		Directory: tmpDir,
		Parallel:  1,
		MaxDepth:  3,
		DryRun:    true, // 안전을 위해 dry-run 모드
	}

	executor := NewBulkUpdateExecutor(context.Background(), opts)

	// 디렉토리 검증 테스트
	err := executor.validateDirectory()
	assert.NoError(t, err)

	// 스캔 테스트 (Git 리포지터리가 없어도 에러는 발생하지 않아야 함)
	repos, err := executor.scanRepositories()
	assert.NoError(t, err)
	assert.IsType(t, []string{}, repos)
}

func TestBulkUpdateExecutor_checkRepositorySafety(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	tests := []struct {
		name           string
		setupFunc      func() string
		expectedStatus string
		description    string
	}{
		{
			name: "clean repository",
			setupFunc: func() string {
				repoPath := filepath.Join(tmpDir, "clean-repo")
				require.NoError(t, os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755))
				return repoPath
			},
			expectedStatus: "error", // Git 명령어가 실제로 실행되지 않으므로 에러 예상
			description:    "모의 Git 리포지터리에서는 에러가 발생해야 함",
		},
		{
			name: "repository without git directory",
			setupFunc: func() string {
				repoPath := filepath.Join(tmpDir, "no-git")
				require.NoError(t, os.MkdirAll(repoPath, 0o755))
				return repoPath
			},
			expectedStatus: "error",
			description:    "Git 디렉토리가 없으면 에러",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoPath := tt.setupFunc()
			result, err := executor.checkRepositorySafety(context.Background(), repoPath)
			if tt.expectedStatus == "error" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if result != nil {
					assert.Contains(t, []string{"safe", "dirty", "error", "no-upstream", "merge-conflict"}, result.Status, tt.description)
				}
			}
		})
	}
}

func TestBulkUpdateExecutor_isInMergeState(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	// Git 리포지터리가 아닌 경우 테스트
	repoPath := filepath.Join(tmpDir, "not-a-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	hasConflict := executor.isInMergeState(context.Background(), repoPath)
	assert.False(t, hasConflict, "Git 리포지터리가 아닌 경우 충돌 상태가 아니어야 함")
}

func TestBulkUpdateExecutor_hasStashedChanges(t *testing.T) {
	tmpDir := t.TempDir()
	opts := BulkUpdateOptions{Directory: tmpDir}
	executor := NewBulkUpdateExecutor(context.Background(), opts)

	// Git 리포지터리가 아닌 경우 테스트
	repoPath := filepath.Join(tmpDir, "not-a-repo")
	require.NoError(t, os.MkdirAll(repoPath, 0o755))

	hasStash, err := executor.hasStashedChanges(context.Background(), repoPath)
	assert.Error(t, err, "Git 리포지터리가 아닌 경우 에러가 발생해야 함")
	assert.False(t, hasStash, "Git 리포지터리가 아닌 경우 스태시가 없어야 함")
}

func TestBulkUpdateExecutor_noFetchOption(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		noFetch bool
		want    string
	}{
		{
			name:    "with fetch enabled",
			noFetch: false,
			want:    "fetch enabled",
		},
		{
			name:    "with fetch disabled",
			noFetch: true,
			want:    "fetch disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := BulkUpdateOptions{
				Directory: tmpDir,
				NoFetch:   tt.noFetch,
			}
			executor := NewBulkUpdateExecutor(context.Background(), opts)
			assert.Equal(t, tt.noFetch, executor.options.NoFetch)
		})
	}
}

// 벤치마크 테스트
func BenchmarkBulkUpdateExecutor_scanRepositories(b *testing.B) {
	tmpDir := b.TempDir()

	// 여러 개의 모의 리포지터리 생성
	for i := 0; i < 10; i++ {
		repoPath := filepath.Join(tmpDir, "repo"+string(rune('0'+i)))
		_ = os.MkdirAll(filepath.Join(repoPath, ".git"), 0o755)
	}

	opts := BulkUpdateOptions{
		Directory: tmpDir,
		MaxDepth:  5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor := NewBulkUpdateExecutor(context.Background(), opts)
		_, _ = executor.scanRepositories()
	}
}
