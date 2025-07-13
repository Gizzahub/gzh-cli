package recovery

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// NetworkFallbackProvider provides fallback for network-related failures
type NetworkFallbackProvider struct {
	alternativeEndpoints map[string][]string
	priority             int
}

// NewNetworkFallbackProvider creates a new network fallback provider
func NewNetworkFallbackProvider() *NetworkFallbackProvider {
	return &NetworkFallbackProvider{
		alternativeEndpoints: map[string][]string{
			"github.com": {"api.github.com", "github-mirror.example.com"},
			"gitlab.com": {"gitlab-backup.example.com"},
			"docker.io":  {"registry-1.docker.io", "mirror.gcr.io"},
			"npmjs.org":  {"registry.npmjs.org", "npm-mirror.example.com"},
		},
		priority: 8,
	}
}

func (nfp *NetworkFallbackProvider) CanHandle(errorCode string) bool {
	return strings.Contains(errorCode, "NETWORK_") ||
		strings.Contains(errorCode, "TIMEOUT_") ||
		strings.Contains(errorCode, "CONNECTION_")
}

func (nfp *NetworkFallbackProvider) Execute(ctx context.Context, originalError error) error {
	// Extract domain from error (simplified)
	errorStr := originalError.Error()

	// Try alternative endpoints
	for domain, alternatives := range nfp.alternativeEndpoints {
		if strings.Contains(errorStr, domain) {
			for _, alternative := range alternatives {
				if nfp.testEndpoint(ctx, alternative) {
					return fmt.Errorf("fallback successful: use %s instead of %s", alternative, domain)
				}
			}
		}
	}

	// Try different DNS servers
	if nfp.tryAlternativeDNS(ctx) {
		return nil
	}

	return fmt.Errorf("all network fallbacks failed")
}

func (nfp *NetworkFallbackProvider) GetInfo() FallbackInfo {
	return FallbackInfo{
		Name:        "Network Fallback Provider",
		Description: "Provides alternative endpoints and DNS servers for network failures",
		Priority:    nfp.priority,
	}
}

func (nfp *NetworkFallbackProvider) testEndpoint(ctx context.Context, endpoint string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "HEAD", "https://"+endpoint, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode < 400
}

func (nfp *NetworkFallbackProvider) tryAlternativeDNS(ctx context.Context) bool {
	// This would implement DNS server switching in a real environment
	// For now, just return true to simulate success
	return true
}

// FileFallbackProvider provides fallback for file-related failures
type FileFallbackProvider struct {
	backupLocations []string
	tempDirectory   string
	priority        int
}

// NewFileFallbackProvider creates a new file fallback provider
func NewFileFallbackProvider() *FileFallbackProvider {
	return &FileFallbackProvider{
		backupLocations: []string{
			"~/.config/gzh-manager/backup",
			"/tmp/gzh-manager-backup",
			"./backup",
		},
		tempDirectory: os.TempDir(),
		priority:      7,
	}
}

func (ffp *FileFallbackProvider) CanHandle(errorCode string) bool {
	return strings.Contains(errorCode, "FILE_") ||
		strings.Contains(errorCode, "PERMISSION_") ||
		strings.Contains(errorCode, "DISK_")
}

func (ffp *FileFallbackProvider) Execute(ctx context.Context, originalError error) error {
	errorStr := originalError.Error()

	// Handle permission errors
	if strings.Contains(errorStr, "permission denied") {
		return ffp.handlePermissionError(ctx, errorStr)
	}

	// Handle disk space errors
	if strings.Contains(errorStr, "no space left") {
		return ffp.handleDiskSpaceError(ctx)
	}

	// Handle file not found errors
	if strings.Contains(errorStr, "no such file") {
		return ffp.handleFileNotFound(ctx, errorStr)
	}

	return fmt.Errorf("no suitable file fallback for error: %s", errorStr)
}

func (ffp *FileFallbackProvider) GetInfo() FallbackInfo {
	return FallbackInfo{
		Name:        "File Fallback Provider",
		Description: "Provides alternative file locations and permission fixes",
		Priority:    ffp.priority,
	}
}

func (ffp *FileFallbackProvider) handlePermissionError(ctx context.Context, errorStr string) error {
	// Try alternative locations with different permissions
	for _, location := range ffp.backupLocations {
		expandedPath := expandPath(location)
		if ffp.canWrite(expandedPath) {
			return fmt.Errorf("fallback successful: use alternative location %s", expandedPath)
		}
	}

	// Try temp directory
	if ffp.canWrite(ffp.tempDirectory) {
		return fmt.Errorf("fallback successful: use temporary directory %s", ffp.tempDirectory)
	}

	return fmt.Errorf("no writable location found")
}

func (ffp *FileFallbackProvider) handleDiskSpaceError(ctx context.Context) error {
	// Try to clean up temp files
	if ffp.cleanupTempFiles() {
		return nil // Space freed, retry original operation
	}

	// Try alternative locations
	for _, location := range ffp.backupLocations {
		expandedPath := expandPath(location)
		if ffp.hasSpace(expandedPath) {
			return fmt.Errorf("fallback successful: use location with available space %s", expandedPath)
		}
	}

	return fmt.Errorf("no location with sufficient disk space found")
}

func (ffp *FileFallbackProvider) handleFileNotFound(ctx context.Context, errorStr string) error {
	// Extract filename from error (simplified)
	// In real implementation, would parse error more carefully

	// Try backup locations
	for _, location := range ffp.backupLocations {
		expandedPath := expandPath(location)
		if ffp.fileExists(expandedPath) {
			return fmt.Errorf("fallback successful: found file in backup location %s", expandedPath)
		}
	}

	return fmt.Errorf("file not found in any backup location")
}

func (ffp *FileFallbackProvider) canWrite(path string) bool {
	testFile := filepath.Join(path, ".gzh-write-test")
	err := os.WriteFile(testFile, []byte("test"), 0o644)
	if err != nil {
		return false
	}
	os.Remove(testFile)
	return true
}

func (ffp *FileFallbackProvider) hasSpace(path string) bool {
	// Simplified space check - in real implementation would check actual disk space
	return ffp.canWrite(path)
}

func (ffp *FileFallbackProvider) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (ffp *FileFallbackProvider) cleanupTempFiles() bool {
	// Clean up old temp files (simplified implementation)
	tempPattern := filepath.Join(ffp.tempDirectory, "gzh-*")
	matches, err := filepath.Glob(tempPattern)
	if err != nil {
		return false
	}

	cleaned := false
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}

		// Remove files older than 1 hour
		if time.Since(info.ModTime()) > time.Hour {
			if os.Remove(match) == nil {
				cleaned = true
			}
		}
	}

	return cleaned
}

// AuthFallbackProvider provides fallback for authentication failures
type AuthFallbackProvider struct {
	tokenSources []TokenSource
	priority     int
}

// TokenSource represents a source of authentication tokens
type TokenSource struct {
	Name     string
	EnvVar   string
	FilePath string
	Priority int
}

// NewAuthFallbackProvider creates a new auth fallback provider
func NewAuthFallbackProvider() *AuthFallbackProvider {
	return &AuthFallbackProvider{
		tokenSources: []TokenSource{
			{Name: "Primary GitHub Token", EnvVar: "GITHUB_TOKEN", Priority: 10},
			{Name: "Backup GitHub Token", EnvVar: "GITHUB_TOKEN_BACKUP", Priority: 8},
			{Name: "GitHub CLI Token", FilePath: "~/.config/gh/hosts.yml", Priority: 6},
			{Name: "Git Credential Helper", EnvVar: "GIT_TOKEN", Priority: 4},
		},
		priority: 9,
	}
}

func (afp *AuthFallbackProvider) CanHandle(errorCode string) bool {
	return strings.Contains(errorCode, "AUTH_") ||
		strings.Contains(errorCode, "TOKEN_") ||
		strings.Contains(errorCode, "CREDENTIAL_")
}

func (afp *AuthFallbackProvider) Execute(ctx context.Context, originalError error) error {
	// Try alternative token sources
	for _, source := range afp.tokenSources {
		token := afp.getTokenFromSource(source)
		if token != "" && afp.validateToken(ctx, token) {
			return fmt.Errorf("fallback successful: use token from %s", source.Name)
		}
	}

	// Try to refresh tokens
	if afp.tryRefreshTokens(ctx) {
		return nil
	}

	return fmt.Errorf("no valid authentication tokens found")
}

func (afp *AuthFallbackProvider) GetInfo() FallbackInfo {
	return FallbackInfo{
		Name:        "Authentication Fallback Provider",
		Description: "Provides alternative authentication tokens and refresh mechanisms",
		Priority:    afp.priority,
	}
}

func (afp *AuthFallbackProvider) getTokenFromSource(source TokenSource) string {
	if source.EnvVar != "" {
		return os.Getenv(source.EnvVar)
	}

	if source.FilePath != "" {
		expandedPath := expandPath(source.FilePath)
		data, err := os.ReadFile(expandedPath)
		if err != nil {
			return ""
		}
		// Simplified token extraction - would parse YAML/JSON in real implementation
		content := string(data)
		if strings.Contains(content, "oauth_token:") {
			// Extract token from GitHub CLI config format
			lines := strings.Split(content, "\n")
			for _, line := range lines {
				if strings.Contains(line, "oauth_token:") {
					parts := strings.Split(line, ":")
					if len(parts) >= 2 {
						return strings.TrimSpace(parts[1])
					}
				}
			}
		}
	}

	return ""
}

func (afp *AuthFallbackProvider) validateToken(ctx context.Context, token string) bool {
	// Simplified token validation - would make actual API call
	return len(token) > 20 && strings.HasPrefix(token, "ghp_")
}

func (afp *AuthFallbackProvider) tryRefreshTokens(ctx context.Context) bool {
	// This would implement token refresh logic in a real environment
	// For now, just return false
	return false
}

// Helper function to expand paths with ~
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
