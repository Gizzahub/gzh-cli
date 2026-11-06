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
	"unicode"
	"unicode/utf8"
)

type (
	stateIdentity struct{ provider, organization string }
	stateRecord   struct {
		name  string
		state CloneState
	}
)

func stateFilename(provider, organization string) (string, error) {
	if err := validateStateIdentity(provider, organization); err != nil {
		return "", err
	}
	legacy := provider + "_" + organization + ".json"
	if portableStateToken(provider, false) && portableStateToken(organization, true) && len(legacy) <= 200 {
		return legacy, nil
	}
	// 길이로 identity 경계를 구분하고 고정 길이 이름을 사용한다. 실제 JSON도 항상 검증한다.
	digest := sha256.Sum256([]byte(fmt.Sprintf("%d:%s%d:%s", len(provider), provider, len(organization), organization)))
	return "state-v2-" + hex.EncodeToString(digest[:]) + ".json", nil
}

func validateStateIdentity(provider, organization string) error {
	if !utf8.ValidString(provider) || !utf8.ValidString(organization) {
		return errors.New("state identity must be valid UTF-8")
	}
	if err := rejectStateProvider(provider); err != nil {
		return err
	}
	if organization == "" {
		return errors.New("state organization must not be empty")
	}
	for _, component := range strings.Split(organization, "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("invalid state organization %q", organization)
		}
	}
	for _, c := range organization {
		if c == '\\' || c == ':' || unicode.IsControl(c) {
			return fmt.Errorf("invalid state organization %q", organization)
		}
	}
	return nil
}

func rejectStateProvider(provider string) error {
	if provider == "" || provider == "." || provider == ".." {
		return fmt.Errorf("invalid state provider %q", provider)
	}
	for _, c := range provider {
		if c == '/' || c == '\\' || c == ':' || unicode.IsControl(c) {
			return fmt.Errorf("invalid state provider %q", provider)
		}
	}
	return nil
}

func portableStateToken(value string, allowUnderscore bool) bool {
	for _, c := range value {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '.':
		case allowUnderscore && c == '_':
		default:
			return false
		}
	}
	return true
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
		info, present, err := inspectStateComponent(parent, component, name, optionalLegacy)
		if err != nil {
			return stateRecord{}, false, err
		}
		if !present {
			return stateRecord{}, false, nil
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

func inspectStateComponent(parent *os.Root, component, name string, optionalLegacy bool) (os.FileInfo, bool, error) {
	info, err := parent.Lstat(component)
	if err == nil {
		return info, true, nil
	}
	if os.IsNotExist(err) {
		return nil, false, nil
	}
	if optionalLegacy {
		missing, classifyErr := optionalLegacyComponentMissing(parent, component, err)
		if classifyErr != nil {
			return nil, false, fmt.Errorf("failed to inspect state component %s in %s: %w", component, name, classifyErr)
		}
		if missing {
			return nil, false, nil
		}
	}
	return nil, false, fmt.Errorf("failed to inspect state component %s in %s: %w", component, name, err)
}

func optionalLegacyComponentMissing(parent *os.Root, component string, lookupErr error) (bool, error) {
	tooLong := errors.Is(lookupErr, syscall.ENAMETOOLONG) ||
		runtime.GOOS == "windows" && errors.Is(lookupErr, syscall.Errno(206))
	if tooLong {
		return true, nil
	}
	tooLong, limitErr := legacyWindowsComponentTooLong(parent, component, lookupErr)
	if limitErr != nil {
		return false, errors.Join(lookupErr, fmt.Errorf("failed to verify state component length: %w", limitErr))
	}
	return tooLong, nil
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
