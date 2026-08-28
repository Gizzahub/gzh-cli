// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var sshParserOpen = os.Open

// SSHConfigParser handles parsing SSH config files and extracting includes and keys.
type SSHConfigParser struct {
	configPath string
	sshDir     string
}

// NewSSHConfigParser creates a new SSH config parser.
func NewSSHConfigParser(configPath string) *SSHConfigParser {
	return &SSHConfigParser{
		configPath: configPath,
		sshDir:     filepath.Dir(configPath),
	}
}

// ParsedSSHConfig represents a parsed SSH configuration.
type ParsedSSHConfig struct {
	MainConfigPath string
	IncludeFiles   []string
	PrivateKeys    []string
	PublicKeys     []string
}

// Parse parses the SSH config and returns all related files.
func (p *SSHConfigParser) Parse() (*ParsedSSHConfig, error) {
	result := &ParsedSSHConfig{
		MainConfigPath: p.configPath,
		IncludeFiles:   []string{},
		PrivateKeys:    []string{},
		PublicKeys:     []string{},
	}

	seen := make(map[string]bool)
	if err := p.parseConfigTree(p.configPath, result, seen); err != nil {
		return nil, fmt.Errorf("failed to parse main config file: %w", err)
	}

	// Remove duplicates
	result.IncludeFiles = removeDuplicates(result.IncludeFiles)
	result.PrivateKeys = removeDuplicates(result.PrivateKeys)
	result.PublicKeys = removeDuplicates(result.PublicKeys)

	return result, nil
}

// parseConfigTree는 Include를 이름순으로 재귀 순회한다. 현재 지원하는 단순 Include
// 문법 범위에서 이미 방문한 실제 파일을 기록해 순환 Include를 건너뛴다.
func (p *SSHConfigParser) parseConfigTree(configPath string, result *ParsedSSHConfig, seen map[string]bool) error {
	canonical, err := filepath.EvalSymlinks(configPath)
	if err != nil {
		return err
	}
	if seen[canonical] {
		return nil
	}
	seen[canonical] = true

	includes, err := p.parseConfigFile(configPath, result)
	if err != nil {
		return err
	}
	for _, includeFile := range includes {
		result.IncludeFiles = append(result.IncludeFiles, includeFile)
		if err := p.parseConfigTree(includeFile, result, seen); err != nil {
			return fmt.Errorf("failed to parse discovered include %s: %w", includeFile, err)
		}
	}
	return nil
}

// parseConfigFile parses a single SSH config file.
func (p *SSHConfigParser) parseConfigFile(configPath string, result *ParsedSSHConfig) (includes []string, err error) {
	file, err := sshParserOpen(configPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close %s: %w", configPath, closeErr))
		}
	}()

	includes = []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse Include directives
		includeResult := &ParsedSSHConfig{}
		if err := p.parseIncludeLine(line, includeResult); err != nil {
			return nil, fmt.Errorf("invalid Include in %s: %w", configPath, err)
		}
		includes = append(includes, includeResult.IncludeFiles...)

		// Parse IdentityFile directives
		if err := p.parseIdentityFileLine(line, result); err != nil {
			return nil, fmt.Errorf("invalid IdentityFile in %s: %w", configPath, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", configPath, err)
	}
	return includes, nil
}

// parseIncludeLine parses Include directives.
func (p *SSHConfigParser) parseIncludeLine(line string, result *ParsedSSHConfig) error {
	// Case-insensitive match for Include
	includeRegex := regexp.MustCompile(`(?i)^\s*include\s+(.+)$`)
	matches := includeRegex.FindStringSubmatch(line)

	if len(matches) != 2 {
		return nil // Not an include line
	}

	includePath := strings.TrimSpace(matches[1])

	// Expand ~ to home directory
	if strings.HasPrefix(includePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		includePath = filepath.Join(homeDir, includePath[2:])
	}

	// Handle relative paths
	if !filepath.IsAbs(includePath) {
		includePath = filepath.Join(p.sshDir, includePath)
	}

	// Expand glob patterns
	globMatches, err := filepath.Glob(includePath)
	if err != nil {
		return fmt.Errorf("failed to expand glob pattern %s: %w", includePath, err)
	}

	for _, match := range globMatches {
		stat, err := os.Stat(match)
		if err != nil {
			return fmt.Errorf("failed to stat discovered include %s: %w", match, err)
		}
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("discovered include is not regular: %s", match)
		}
		result.IncludeFiles = append(result.IncludeFiles, match)
	}
	sort.Strings(result.IncludeFiles)

	return nil
}

// parseIdentityFileLine parses IdentityFile directives.
func (p *SSHConfigParser) parseIdentityFileLine(line string, result *ParsedSSHConfig) error {
	// Case-insensitive match for IdentityFile
	identityRegex := regexp.MustCompile(`(?i)^\s*identityfile\s+(.+)$`)
	matches := identityRegex.FindStringSubmatch(line)

	if len(matches) != 2 {
		return nil // Not an identity file line
	}

	keyPath := strings.TrimSpace(matches[1])

	// Expand ~ to home directory
	if strings.HasPrefix(keyPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		keyPath = filepath.Join(homeDir, keyPath[2:])
	}

	// Handle relative paths
	if !filepath.IsAbs(keyPath) {
		keyPath = filepath.Join(p.sshDir, keyPath)
	}

	// Check if private key file exists
	if stat, err := os.Stat(keyPath); err == nil {
		if !stat.Mode().IsRegular() {
			return fmt.Errorf("IdentityFile is not regular: %s", keyPath)
		}
		result.PrivateKeys = append(result.PrivateKeys, keyPath)

		// Also check for corresponding public key
		pubKeyPath := keyPath + ".pub"
		if stat, err := os.Stat(pubKeyPath); err == nil {
			if !stat.Mode().IsRegular() {
				return fmt.Errorf("public key is not regular: %s", pubKeyPath)
			}
			result.PublicKeys = append(result.PublicKeys, pubKeyPath)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to stat public key %s: %w", pubKeyPath, err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat IdentityFile %s: %w", keyPath, err)
	}

	return nil
}

// removeDuplicates removes duplicate strings from a slice.
func removeDuplicates(slice []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
