// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package devenv

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

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
			// 기존 명령 계약대로 발견한 Include를 읽지 못해도 전체 저장은 중단하지 않는다.
			fmt.Printf("Warning: failed to parse include file %s: %v\n", includeFile, err)
		}
	}
	return nil
}

// parseConfigFile parses a single SSH config file.
func (p *SSHConfigParser) parseConfigFile(configPath string, result *ParsedSSHConfig) ([]string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	includes := []string{}
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
			fmt.Printf("Warning: failed to parse include line '%s': %v\n", line, err)
		} else {
			includes = append(includes, includeResult.IncludeFiles...)
		}

		// Parse IdentityFile directives
		if err := p.parseIdentityFileLine(line, result); err != nil {
			fmt.Printf("Warning: failed to parse identity file line '%s': %v\n", line, err)
		}
	}

	return includes, scanner.Err()
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
		// Only include regular files
		if stat, err := os.Stat(match); err == nil && stat.Mode().IsRegular() {
			result.IncludeFiles = append(result.IncludeFiles, match)
		}
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
	if stat, err := os.Stat(keyPath); err == nil && stat.Mode().IsRegular() {
		result.PrivateKeys = append(result.PrivateKeys, keyPath)

		// Also check for corresponding public key
		pubKeyPath := keyPath + ".pub"
		if stat, err := os.Stat(pubKeyPath); err == nil && stat.Mode().IsRegular() {
			result.PublicKeys = append(result.PublicKeys, pubKeyPath)
		}
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
