// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package duplicates

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// BinEntry: 실행 파일 항목 정보
// name: 파일명, path: 경로, realPath: 심볼릭 링크 해제한 실경로, source: 어떤 매니저/소스에서 기인했는지
type BinEntry struct {
	Name     string
	Path     string
	RealPath string
	Source   string
}

// Paths/Sources: 모든 후보 집합(Primary 포함).
type Conflict struct {
	Binary        string
	PrimaryPath   string
	PrimarySource string
	Paths         []string
	Sources       []string
}

// Source: 각 패키지 매니저 또는 경로 제공자
// 설치되어 있지 않으면 빈 리스트를 반환하고 에러는 내지 않도록 구현한다.
type Source interface {
	Name() string
	ListBins(ctx context.Context) ([]BinEntry, error)
}

// ===== 공통 유틸 =====

// isExecutableFile는 주어진 경로가 실행 가능한 정규 파일인지 판별한다.
func isExecutableFile(info fs.FileInfo) bool {
	if info == nil {
		return false
	}
	mode := info.Mode()
	return mode.IsRegular() && (mode&0o111 != 0)
}

// listExecutablesInDir는 디렉토리 내 실행 파일들을 BinEntry로 반환한다.
func listExecutablesInDir(dir, source string) ([]BinEntry, error) {
	entries := []BinEntry{}
	if dir == "" {
		return entries, nil
	}
	fi, err := os.Stat(dir)
	if err != nil {
		return entries, err
	}
	if !fi.IsDir() {
		return entries, nil
	}
	dents, err := os.ReadDir(dir)
	if err != nil {
		return entries, err
	}
	for _, de := range dents {
		if de.IsDir() {
			continue
		}
		info, err := de.Info()
		if err != nil || !isExecutableFile(info) {
			continue
		}
		p := filepath.Join(dir, de.Name())
		realPath, err := filepath.EvalSymlinks(p)
		if err != nil {
			realPath = p
		}
		if realPath == "" {
			realPath = p
		}
		entries = append(entries, BinEntry{
			Name:     de.Name(),
			Path:     p,
			RealPath: realPath,
			Source:   source,
		})
	}
	return entries, nil
}

// pathPriorityIndex는 PATH 디렉토리 내 우선순위 인덱스를 반환한다(낮을수록 먼저 탐색됨).
func pathPriorityIndex(pathDirs []string, filePath string) int {
	dir := filepath.Dir(filePath)
	for i, d := range pathDirs {
		if filepath.Clean(d) == filepath.Clean(dir) {
			return i
		}
	}
	return 1<<30 - 1
}

// resolvePrimary는 PATH 우선순위상 먼저 잡히는 항목을 결정한다.
func resolvePrimary(pathDirs []string, entries []BinEntry) (BinEntry, bool) {
	if len(entries) == 0 {
		return BinEntry{}, false
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return pathPriorityIndex(pathDirs, entries[i].Path) < pathPriorityIndex(pathDirs, entries[j].Path)
	})
	return entries[0], true
}

// ===== 소스 구현들 =====

type brewSource struct{}

func (s brewSource) Name() string { return "brew" }
func (s brewSource) ListBins(ctx context.Context) ([]BinEntry, error) {
	// brew 미설치 시 빈 결과
	if err := exec.CommandContext(ctx, "brew", "--prefix").Run(); err != nil {
		return nil, err
	}
	out, err := exec.CommandContext(ctx, "brew", "--prefix").Output()
	if err != nil {
		return nil, nil
	}
	prefix := strings.TrimSpace(string(out))
	bins := []BinEntry{}
	b1, err := listExecutablesInDir(filepath.Join(prefix, "bin"), s.Name())
	if err != nil {
		return nil, err
	}
	bins = append(bins, b1...)
	b2, err := listExecutablesInDir(filepath.Join(prefix, "sbin"), s.Name())
	if err != nil {
		return nil, err
	}
	bins = append(bins, b2...)
	return bins, nil
}

type asdfSource struct{}

func (s asdfSource) Name() string { return "asdf" }
func (s asdfSource) ListBins(ctx context.Context) ([]BinEntry, error) {
	// asdf 미설치 시 빈 결과
	if err := exec.CommandContext(ctx, "asdf", "--version").Run(); err != nil {
		return nil, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return listExecutablesInDir(filepath.Join(home, ".asdf", "shims"), s.Name())
}

type npmSource struct{}

func (s npmSource) Name() string { return "npm" }
func (s npmSource) ListBins(ctx context.Context) ([]BinEntry, error) {
	if err := exec.CommandContext(ctx, "npm", "--version").Run(); err != nil {
		return nil, err
	}
	out, err := exec.CommandContext(ctx, "npm", "bin", "-g").Output()
	if err != nil {
		return nil, err
	}
	dir := strings.TrimSpace(string(out))
	return listExecutablesInDir(dir, s.Name())
}

type pipSource struct{}

func (s pipSource) Name() string { return "pip" }
func (s pipSource) ListBins(ctx context.Context) ([]BinEntry, error) {
	// 우선순위: pip -> pip3 -> python3 user base
	if exec.CommandContext(ctx, "pip", "--version").Run() == nil {
		// pip가 설치되어 있어도 scripts 경로를 안정적으로 찾기 어려움 → PATH를 통해 병합되므로 넘어간다
	}
	// python3 사용자 base/bin 추정
	out, err := exec.CommandContext(ctx, "python3", "-c", "import sysconfig; print(sysconfig.get_path('scripts'))").Output()
	if err == nil {
		dir := strings.TrimSpace(string(out))
		return listExecutablesInDir(dir, s.Name())
	}
	// ~/.local/bin 폴백
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return listExecutablesInDir(filepath.Join(home, ".local", "bin"), s.Name())
}

type gemSource struct{}

func (s gemSource) Name() string { return "gem" }
func (s gemSource) ListBins(ctx context.Context) ([]BinEntry, error) {
	if err := exec.CommandContext(ctx, "gem", "--version").Run(); err != nil {
		return nil, err
	}
	out, err := exec.CommandContext(ctx, "gem", "env").Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(out), "\n")
	var bindir string
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "EXECUTABLE DIRECTORY:") {
			bindir = strings.TrimSpace(strings.TrimPrefix(ln, "EXECUTABLE DIRECTORY:"))
			break
		}
	}
	if bindir == "" {
		return nil, nil
	}
	return listExecutablesInDir(bindir, s.Name())
}

// pathSource: PATH 환경 변수 기반으로 스캔(소스명은 "path").
type pathSource struct {
	pathDirs []string
}

func (s pathSource) Name() string { return "path" }
func (s pathSource) ListBins(_ context.Context) ([]BinEntry, error) {
	bins := []BinEntry{}
	seen := map[string]struct{}{}
	for _, d := range s.pathDirs {
		b, err := listExecutablesInDir(d, s.Name())
		if err != nil {
			continue
		}
		for _, e := range b {
			key := e.Path
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			bins = append(bins, e)
		}
	}
	return bins, nil
}

// ===== 감지기 =====

// CollectAndDetectConflicts는 주어진 소스들을 스캔하여 충돌을 반환한다.
// pathDirs는 PATH 분해 결과이며, Primary 경로 결정을 위해 필요하다.
func CollectAndDetectConflicts(ctx context.Context, sources []Source, pathDirs []string) ([]Conflict, error) {
	if len(sources) == 0 {
		return nil, errors.New("no sources provided")
	}
	nameToEntries := map[string][]BinEntry{}
	for _, s := range sources {
		entries, err := s.ListBins(ctx) // 설치 안 된 경우 등을 고려해 에러 무시
		if err != nil {
			continue
		}
		for _, e := range entries {
			nameToEntries[e.Name] = append(nameToEntries[e.Name], e)
		}
	}
	var conflicts []Conflict
	for name, entries := range nameToEntries {
		if len(entries) < 2 {
			continue
		}
		// 동일 realpath만 여러 개면 충돌 아님
		realSet := map[string]struct{}{}
		for _, e := range entries {
			realSet[e.RealPath] = struct{}{}
		}
		if len(realSet) < 2 {
			continue
		}
		primary, ok := resolvePrimary(pathDirs, append([]BinEntry{}, entries...))
		if !ok {
			continue
		}
		paths := make([]string, 0, len(entries))
		sourcesSet := map[string]struct{}{}
		for _, e := range entries {
			paths = append(paths, fmt.Sprintf("%s (%s)", e.Path, e.Source))
			sourcesSet[e.Source] = struct{}{}
		}
		mgrs := make([]string, 0, len(sourcesSet))
		for k := range sourcesSet {
			mgrs = append(mgrs, k)
		}
		sort.Strings(mgrs)
		conflicts = append(conflicts, Conflict{
			Binary:        name,
			PrimaryPath:   fmt.Sprintf("%s (%s)", primary.Path, primary.Source),
			PrimarySource: primary.Source,
			Paths:         paths,
			Sources:       mgrs,
		})
	}
	// 안정된 출력 위해 이름 순 정렬
	sort.SliceStable(conflicts, func(i, j int) bool { return conflicts[i].Binary < conflicts[j].Binary })
	return conflicts, nil
}

// BuildDefaultSources는 대표 매니저 소스 + PATH 스캐너를 구성해준다.
func BuildDefaultSources(pathDirs []string) []Source {
	return []Source{
		brewSource{},
		asdfSource{},
		npmSource{},
		pipSource{},
		gemSource{},
		pathSource{pathDirs: pathDirs},
	}
}

// SplitPATH는 PATH를 디렉토리 목록으로 쪼갠다.
func SplitPATH(envPath string) []string {
	parts := strings.Split(envPath, string(os.PathListSeparator))
	var out []string
	for _, p := range parts {
		pp := strings.TrimSpace(p)
		if pp != "" {
			out = append(out, pp)
		}
	}
	return out
}

// PrintConflictsSummary는 간단한 요약/상위 N개를 텍스트로 출력한다.
func PrintConflictsSummary(conflicts []Conflict, maxCount int) {
	if len(conflicts) == 0 {
		fmt.Println("중복 설치 의심 항목 없음")
		return
	}
	fmt.Printf("중복 설치 의심 %d개 발견\n", len(conflicts))
	limit := maxCount
	if limit <= 0 || limit > len(conflicts) {
		limit = len(conflicts)
	}
	for i := 0; i < limit; i++ {
		c := conflicts[i]
		fmt.Printf("- %s: PATH우선=%s, 후보=%s\n", c.Binary, c.PrimaryPath, strings.Join(c.Paths, ", "))
	}
	if limit < len(conflicts) {
		fmt.Printf("...and %d more\n", len(conflicts)-limit)
	}
}
