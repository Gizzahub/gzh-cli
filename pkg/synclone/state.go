// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package bulkclone

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"
)

// CloneState represents the state of a bulk clone operation.
type CloneState struct {
	// Operation metadata
	StartTime    time.Time `json:"startTime"`
	LastUpdated  time.Time `json:"lastUpdated"`
	Provider     string    `json:"provider"`
	Organization string    `json:"organization"`
	TargetPath   string    `json:"targetPath"`
	Strategy     string    `json:"strategy"`

	// Progress tracking
	TotalRepositories int                   `json:"totalRepositories"`
	CompletedRepos    []CompletedRepository `json:"completedRepos"`
	FailedRepos       []FailedRepository    `json:"failedRepos"`
	PendingRepos      []string              `json:"pendingRepos"`

	// Configuration
	Parallel   int `json:"parallel"`
	MaxRetries int `json:"maxRetries"`

	// Status
	Status string `json:"status"` // "in_progress", "completed", "failed", "canceled"
}

// CompletedRepository represents a successfully processed repository.
type CompletedRepository struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Operation   string    `json:"operation"` // "clone", "pull", "reset", "fetch"
	CompletedAt time.Time `json:"completedAt"`
	Message     string    `json:"message,omitempty"`
}

// FailedRepository represents a failed repository operation.
type FailedRepository struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Operation   string    `json:"operation"`
	Error       string    `json:"error"`
	Attempts    int       `json:"attempts"`
	LastAttempt time.Time `json:"lastAttempt"`
}

// StateManager handles saving and loading clone states.
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager.
func NewStateManager(stateDir string) *StateManager {
	if stateDir == "" {
		// Default to ~/.gzh/state
		if homeDir, err := os.UserHomeDir(); err == nil {
			stateDir = filepath.Join(homeDir, ".gzh", "state")
		} else {
			stateDir = ".gzh/state" // Fallback to current directory
		}
	}

	return &StateManager{
		stateDir: stateDir,
	}
}

// GetStateFilePath returns the existing validated legacy path or the new snapshot path.
// 검증할 수 없거나 중복된 경로는 빈 문자열을 반환한다. 상세 오류는 ResolveStateFilePath로 확인한다.
func (sm *StateManager) GetStateFilePath(provider, organization string) string {
	path, _ := sm.ResolveStateFilePath(provider, organization)
	return path
}

// ResolveStateFilePath returns the verified existing path or a new path for an absent state.
func (sm *StateManager) ResolveStateFilePath(provider, organization string) (string, error) {
	name, err := stateFilename(provider, organization)
	if err != nil {
		return "", err
	}
	root, err := os.OpenRoot(sm.stateDir)
	if os.IsNotExist(err) {
		return filepath.Join(sm.stateDir, name), nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	record, _, err := resolveStateName(root, provider, organization)
	if err != nil {
		return "", err
	}
	return filepath.Join(sm.stateDir, record.name), nil
}

// SaveState saves the clone state to disk.
func (sm *StateManager) SaveState(state *CloneState) (saveErr error) {
	if state == nil {
		return errors.New("cannot save a nil clone state")
	}
	if _, err := stateFilename(state.Provider, state.Organization); err != nil {
		return err
	}
	state.LastUpdated = time.Now()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// 사용자가 선택한 root 자체는 기존처럼 해석하고, 열린 후의 IO를 root 안으로 제한한다.
	if err := os.MkdirAll(sm.stateDir, 0o750); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	root, err := os.OpenRoot(sm.stateDir)
	if err != nil {
		return fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	record, _, err := resolveStateName(root, state.Provider, state.Organization)
	if err != nil {
		return err
	}
	stateName := record.name

	// root 상대 경로의 배타 생성과 rename으로 기존 스냅샷의 잘림을 방지한다.
	file, tempName, err := createStateTemp(root, filepath.Dir(stateName))
	if err != nil {
		return fmt.Errorf("failed to create temporary state file: %w", err)
	}
	cleanupTemp := true
	defer func() {
		if cleanupTemp {
			if err := root.Remove(tempName); err != nil && !os.IsNotExist(err) {
				saveErr = errors.Join(saveErr, fmt.Errorf("failed to remove temporary state file: %w", err))
			}
		}
	}()
	if _, err := file.Write(data); err != nil {
		return errors.Join(fmt.Errorf("failed to write state file: %w", err), file.Close())
	}
	if err := file.Sync(); err != nil {
		return errors.Join(fmt.Errorf("failed to sync state file: %w", err), file.Close())
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close state file: %w", err)
	}
	if err := root.Rename(tempName, stateName); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	cleanupTemp = false

	return nil
}

// LoadState loads the clone state from disk.
func (sm *StateManager) LoadState(provider, organization string) (*CloneState, error) {
	if _, err := stateFilename(provider, organization); err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(sm.stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	record, exists, err := resolveStateName(root, provider, organization)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("no state file found for %s/%s", provider, organization)
	}
	state := record.state
	return &state, nil
}

// DeleteState removes the state file.
func (sm *StateManager) DeleteState(provider, organization string) error {
	if _, err := stateFilename(provider, organization); err != nil {
		return err
	}
	root, err := os.OpenRoot(sm.stateDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	record, exists, err := resolveStateName(root, provider, organization)
	if err != nil || !exists {
		return err
	}
	if err := root.Remove(record.name); err != nil {
		return fmt.Errorf("failed to delete state file %s: %w", record.name, err)
	}
	return nil
}

// HasState checks if a state file exists for the given operation.
func (sm *StateManager) HasState(provider, organization string) bool {
	// bool API는 확실한 부재만 false로 반환하고 손상과 불확실성을 숨기지 않는다.
	if _, err := stateFilename(provider, organization); err != nil {
		return true
	}
	root, err := os.OpenRoot(sm.stateDir)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		return true
	}
	defer func() { _ = root.Close() }()
	_, exists, err := resolveStateName(root, provider, organization)
	return err != nil || exists
}

// ListStates returns all saved states.
func (sm *StateManager) ListStates() ([]CloneState, error) {
	root, err := os.OpenRoot(sm.stateDir)
	if os.IsNotExist(err) {
		return []CloneState{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	records, err := readStateInventory(root)
	if err != nil {
		return nil, err
	}
	states := make([]CloneState, 0, len(records))
	for _, record := range sortedStateRecords(records) {
		if _, _, err := resolveStateName(root, record.state.Provider, record.state.Organization); err != nil {
			return nil, err
		}
		states = append(states, record.state)
	}
	return states, nil
}

// DeleteAllStates validates the complete inventory before deleting its actual entries.
// It returns the number removed if a later filesystem operation fails.
func (sm *StateManager) DeleteAllStates() (int, error) {
	root, err := os.OpenRoot(sm.stateDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to open state directory: %w", err)
	}
	defer func() { _ = root.Close() }()
	records, err := readStateInventory(root)
	if err != nil {
		return 0, err
	}
	ordered := sortedStateRecords(records)
	for _, record := range ordered {
		if _, _, err := resolveStateName(root, record.state.Provider, record.state.Organization); err != nil {
			return 0, err
		}
	}
	for i, record := range ordered {
		if err := root.Remove(record.name); err != nil {
			return i, fmt.Errorf("failed to delete state file %s: %w", record.name, err)
		}
	}
	return len(ordered), nil
}

type (
	stateIdentity struct{ provider, organization string }
	stateRecord   struct {
		name  string
		state CloneState
	}
)

func stateFilename(provider, organization string) (string, error) {
	if !utf8.ValidString(provider) || !utf8.ValidString(organization) {
		return "", errors.New("state identity must be valid UTF-8")
	}
	if provider == "" || provider == "." || provider == ".." {
		return "", fmt.Errorf("invalid state provider %q", provider)
	}
	portableProvider := true
	for _, c := range provider {
		if c == '/' || c == '\\' || c == ':' || unicode.IsControl(c) {
			return "", fmt.Errorf("invalid state provider %q", provider)
		}
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.') {
			portableProvider = false
		}
	}
	if organization == "" {
		return "", errors.New("state organization must not be empty")
	}
	for _, component := range strings.Split(organization, "/") {
		if component == "" || component == "." || component == ".." {
			return "", fmt.Errorf("invalid state organization %q", organization)
		}
	}
	portable := portableProvider
	for _, c := range organization {
		if c == '\\' || c == ':' || unicode.IsControl(c) {
			return "", fmt.Errorf("invalid state organization %q", organization)
		}
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-' || c == '.') {
			portable = false
		}
	}
	legacy := provider + "_" + organization + ".json"
	if portable && len(legacy) <= 200 {
		return legacy, nil
	}
	// 길이로 identity 경계를 구분하고 고정 길이 이름을 사용한다. 실제 JSON도 항상 검증한다.
	digest := sha256.Sum256([]byte(fmt.Sprintf("%d:%s%d:%s", len(provider), provider, len(organization), organization)))
	return "state-v2-" + hex.EncodeToString(digest[:]) + ".json", nil
}

func legacyStateName(provider, organization string) string {
	name := filepath.FromSlash(provider + "_" + organization + ".json")
	if !filepath.IsLocal(name) {
		return ""
	}
	return name
}

// 열린 root 아래의 symlink를 따르지 않고 잘못되거나 중복된 데이터는 보존하며 거절한다.
func readStateInventory(root *os.Root) (map[stateIdentity]stateRecord, error) {
	records := make(map[stateIdentity]stateRecord)
	var walk func(string) error
	walk = func(dir string) error {
		file, err := root.Open(dir)
		if err != nil {
			return fmt.Errorf("failed to open state directory %s: %w", dir, err)
		}
		entries, readErr := file.ReadDir(-1)
		if err := errors.Join(readErr, file.Close()); err != nil {
			return fmt.Errorf("failed to read state directory %s: %w", dir, err)
		}
		slices.SortFunc(entries, func(a, b os.DirEntry) int { return strings.Compare(a.Name(), b.Name()) })
		for _, entry := range entries {
			name := filepath.Join(dir, entry.Name())
			if entry.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("state file is not a regular file (symbolic link): %s", name)
			}
			if entry.IsDir() {
				if err := walk(name); err != nil {
					return err
				}
				continue
			}
			if filepath.Ext(name) != ".json" {
				continue
			}
			record, err := readStateRecord(root, name)
			if err != nil {
				return err
			}
			identity := stateIdentity{record.state.Provider, record.state.Organization}
			if prior, exists := records[identity]; exists {
				return fmt.Errorf("duplicate state snapshots %s and %s; preserve both for recovery", prior.name, name)
			}
			records[identity] = record
		}
		return nil
	}
	if err := walk("."); err != nil {
		return nil, err
	}
	return records, nil
}

func readStateRecord(root *os.Root, name string) (stateRecord, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return stateRecord{}, fmt.Errorf("failed to inspect state file %s: %w", name, err)
	}
	if !info.Mode().IsRegular() {
		return stateRecord{}, fmt.Errorf("state file is not a regular file: %s", name)
	}
	data, err := root.ReadFile(name)
	if err != nil {
		return stateRecord{}, fmt.Errorf("failed to read state file %s: %w", name, err)
	}
	return decodeStateRecord(data, name)
}

func decodeStateRecord(data []byte, name string) (stateRecord, error) {
	var state CloneState
	if err := json.Unmarshal(data, &state); err != nil {
		return stateRecord{}, fmt.Errorf("failed to unmarshal state file %s; preserve it for recovery: %w", name, err)
	}
	canonical, err := stateFilename(state.Provider, state.Organization)
	if err != nil {
		return stateRecord{}, fmt.Errorf("invalid identity in state file %s: %w", name, err)
	}
	if name != canonical && name != legacyStateName(state.Provider, state.Organization) {
		return stateRecord{}, fmt.Errorf("state identity does not match filename %s; preserve it for recovery", name)
	}
	return stateRecord{name, state}, nil
}

func resolveStateName(root *os.Root, provider, organization string) (stateRecord, bool, error) {
	canonical, err := stateFilename(provider, organization)
	if err != nil {
		return stateRecord{}, false, err
	}
	record := stateRecord{name: canonical}
	exists := false
	for _, name := range []string{canonical, legacyStateName(provider, organization)} {
		if name == "" || exists && name == record.name {
			continue
		}
		candidate, present, err := inspectStateCandidate(root, name, name != canonical)
		if err != nil {
			return stateRecord{}, false, err
		}
		if !present {
			continue
		}
		if candidate.state.Provider != provider || candidate.state.Organization != organization {
			if name != canonical {
				continue
			}
			return stateRecord{}, false, fmt.Errorf("conflicting state file %s; preserve it for recovery", name)
		}
		if exists {
			return stateRecord{}, false, fmt.Errorf("duplicate state snapshots %s and %s; preserve both for recovery", record.name, name)
		}
		record, exists = candidate, true
	}
	return record, exists, nil
}

// 전체 경로가 아닌 열린 부모의 단일 component를 검사해 총 경로 길이 오류와 구별한다.
func inspectStateCandidate(root *os.Root, name string, optionalLegacy bool) (stateRecord, bool, error) {
	parent := root
	defer func() {
		if parent != root {
			_ = parent.Close()
		}
	}()
	components := strings.Split(name, string(filepath.Separator))
	for i, component := range components {
		info, err := parent.Lstat(component)
		if os.IsNotExist(err) {
			return stateRecord{}, false, nil
		}
		componentTooLong := errors.Is(err, syscall.ENAMETOOLONG) ||
			runtime.GOOS == "windows" && errors.Is(err, syscall.Errno(206))
		if optionalLegacy && componentTooLong {
			return stateRecord{}, false, nil
		}
		if err != nil {
			return stateRecord{}, false, fmt.Errorf("failed to inspect state component %s in %s: %w", component, name, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return stateRecord{}, false, fmt.Errorf("state file is not a regular file (symbolic link): %s", name)
		}
		if i == len(components)-1 {
			if !info.Mode().IsRegular() {
				return stateRecord{}, false, fmt.Errorf("state file is not a regular file: %s", name)
			}
			data, err := parent.ReadFile(component)
			if err != nil {
				return stateRecord{}, false, fmt.Errorf("failed to read state file %s: %w", name, err)
			}
			record, err := decodeStateRecord(data, name)
			return record, err == nil, err
		}
		if !info.IsDir() {
			return stateRecord{}, false, fmt.Errorf("state parent is not a directory: %s", name)
		}
		next, err := parent.OpenRoot(component)
		if err != nil {
			return stateRecord{}, false, fmt.Errorf("failed to open state parent %s: %w", name, err)
		}
		if parent != root {
			_ = parent.Close()
		}
		parent = next
	}
	return stateRecord{}, false, errors.New("empty state candidate path")
}

func sortedStateRecords(records map[stateIdentity]stateRecord) []stateRecord {
	ordered := make([]stateRecord, 0, len(records))
	for _, record := range records {
		ordered = append(ordered, record)
	}
	slices.SortFunc(ordered, func(a, b stateRecord) int { return strings.Compare(a.name, b.name) })
	return ordered
}

func createStateTemp(root *os.Root, dir string) (*os.File, string, error) {
	for range 10 {
		name := filepath.Join(dir, ".clone-state-"+rand.Text())
		file, err := root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if os.IsExist(err) {
			continue
		}
		return file, name, err
	}
	return nil, "", errors.New("temporary state filename collisions")
}

// NewCloneState creates a new clone state.
func NewCloneState(provider, organization, targetPath, strategy string, parallel, maxRetries int) *CloneState {
	return &CloneState{
		StartTime:         time.Now(),
		LastUpdated:       time.Now(),
		Provider:          provider,
		Organization:      organization,
		TargetPath:        targetPath,
		Strategy:          strategy,
		TotalRepositories: 0,
		CompletedRepos:    []CompletedRepository{},
		FailedRepos:       []FailedRepository{},
		PendingRepos:      []string{},
		Parallel:          parallel,
		MaxRetries:        maxRetries,
		Status:            "in_progress",
	}
}

// AddCompletedRepository adds a completed repository to the state.
func (cs *CloneState) AddCompletedRepository(name, path, operation, message string) {
	// 재시도 성공은 같은 저장소의 이전 실패를 대체한다.
	cs.FailedRepos = slices.DeleteFunc(cs.FailedRepos, func(repo FailedRepository) bool { return repo.Name == name })
	cs.CompletedRepos = slices.DeleteFunc(cs.CompletedRepos, func(repo CompletedRepository) bool { return repo.Name == name })
	cs.CompletedRepos = append(cs.CompletedRepos, CompletedRepository{
		Name:        name,
		Path:        path,
		Operation:   operation,
		CompletedAt: time.Now(),
		Message:     message,
	})

	// Remove from pending if it exists
	cs.removePendingRepo(name)

	// Update total repositories count
	cs.updateTotalRepositories()
}

// AddFailedRepository adds a failed repository to the state.
func (cs *CloneState) AddFailedRepository(name, path, operation, errorMsg string, attempts int) {
	// 한 저장소가 완료와 실패 양쪽에 남지 않도록 결과를 교체한다.
	cs.CompletedRepos = slices.DeleteFunc(cs.CompletedRepos, func(repo CompletedRepository) bool { return repo.Name == name })
	cs.removePendingRepo(name)
	// Check if this repo already failed and update it
	for i, failed := range cs.FailedRepos {
		if failed.Name == name {
			cs.FailedRepos[i].Error = errorMsg
			cs.FailedRepos[i].Attempts = attempts
			cs.FailedRepos[i].LastAttempt = time.Now()
			cs.FailedRepos[i].Path = path
			cs.FailedRepos[i].Operation = operation
			cs.updateTotalRepositories()

			return
		}
	}

	// Add new failed repo
	cs.FailedRepos = append(cs.FailedRepos, FailedRepository{
		Name:        name,
		Path:        path,
		Operation:   operation,
		Error:       errorMsg,
		Attempts:    attempts,
		LastAttempt: time.Now(),
	})

	// Remove from pending if it exists
	cs.removePendingRepo(name)

	// Update total repositories count
	cs.updateTotalRepositories()
}

// IsCompleted checks if a repository has been completed.
func (cs *CloneState) IsCompleted(name string) bool {
	for _, completed := range cs.CompletedRepos {
		if completed.Name == name {
			return true
		}
	}

	return false
}

// IsFailed checks if a repository has failed.
func (cs *CloneState) IsFailed(name string) bool {
	for _, failed := range cs.FailedRepos {
		if failed.Name == name {
			return true
		}
	}

	return false
}

// GetProgress returns the current progress statistics.
func (cs *CloneState) GetProgress() (completed, failed, pending int) {
	return len(cs.CompletedRepos), len(cs.FailedRepos), len(cs.PendingRepos)
}

// GetProgressPercent returns the progress as a percentage.
func (cs *CloneState) GetProgressPercent() float64 {
	if cs.TotalRepositories == 0 {
		return 0
	}

	completed, failed, _ := cs.GetProgress()
	processed := completed + failed

	return float64(processed) / float64(cs.TotalRepositories) * 100
}

// SetPendingRepositories sets the list of pending repositories.
func (cs *CloneState) SetPendingRepositories(repos []string) {
	cs.PendingRepos = repos
	cs.updateTotalRepositories()
}

// updateTotalRepositories updates the total repository count.
func (cs *CloneState) updateTotalRepositories() {
	cs.TotalRepositories = len(cs.PendingRepos) + len(cs.CompletedRepos) + len(cs.FailedRepos)
}

// removePendingRepo removes a repository from the pending list.
func (cs *CloneState) removePendingRepo(name string) {
	for i, repo := range cs.PendingRepos {
		if repo == name {
			cs.PendingRepos = append(cs.PendingRepos[:i], cs.PendingRepos[i+1:]...)
			break
		}
	}
}

// GetRemainingRepositories returns repositories that still need to be processed.
func (cs *CloneState) GetRemainingRepositories() []string {
	return cs.PendingRepos
}

// MarkCompleted marks the operation as completed.
func (cs *CloneState) MarkCompleted() {
	cs.Status = "completed"
	cs.LastUpdated = time.Now()
}

// MarkFailed marks the operation as failed.
func (cs *CloneState) MarkFailed() {
	cs.Status = "failed"
	cs.LastUpdated = time.Now()
}

// MarkCancelled marks the operation as canceled.
func (cs *CloneState) MarkCancelled() {
	cs.Status = "canceled"
	cs.LastUpdated = time.Now()
}
