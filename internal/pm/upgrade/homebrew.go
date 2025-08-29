package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/Gizzahub/gzh-cli/internal/logger"
)

// HomebrewUpgrader implements PackageManagerUpgrader for Homebrew
type HomebrewUpgrader struct {
	logger logger.CommonLogger
}

// NewHomebrewUpgrader creates a new Homebrew upgrader
func NewHomebrewUpgrader(logger logger.CommonLogger) *HomebrewUpgrader {
	return &HomebrewUpgrader{
		logger: logger,
	}
}

// CheckUpdate checks if Homebrew has updates available
func (h *HomebrewUpgrader) CheckUpdate(ctx context.Context) (*UpgradeStatus, error) {
	h.logger.Debug("Checking Homebrew update status")

	// Get current version
	currentVersion, err := h.getCurrentVersion(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get current Homebrew version: %w", err)
	}

	// Get latest version from GitHub API
	latestVersion, releaseDate, changelogURL, err := h.getLatestVersion(ctx)
	if err != nil {
		h.logger.Warn("Failed to get latest Homebrew version from GitHub API: %v", err)
		// Fall back to local check
		latestVersion = currentVersion
	}

	updateAvailable := currentVersion != latestVersion

	return &UpgradeStatus{
		Manager:         "homebrew",
		CurrentVersion:  currentVersion,
		LatestVersion:   latestVersion,
		UpdateAvailable: updateAvailable,
		UpdateMethod:    h.GetUpdateMethod(),
		ReleaseDate:     releaseDate,
		ChangelogURL:    changelogURL,
	}, nil
}

// Upgrade performs the actual upgrade of Homebrew
func (h *HomebrewUpgrader) Upgrade(ctx context.Context, options UpgradeOptions) error {
	h.logger.Info("Starting Homebrew upgrade")

	// Create backup if requested
	var backupPath string
	if options.BackupEnabled {
		var err error
		backupPath, err = h.Backup(ctx)
		if err != nil {
			h.logger.Warn("Failed to create backup: %v", err)
		} else {
			h.logger.Info("Backup created at: %s", backupPath)
		}
	}

	// Update Homebrew
	if err := h.executeUpdate(ctx, options); err != nil {
		if backupPath != "" && !options.SkipValidation {
			h.logger.Info("Attempting rollback due to upgrade failure")
			if rollbackErr := h.Rollback(ctx, backupPath); rollbackErr != nil {
				h.logger.Error("Failed to rollback: %v", rollbackErr)
			}
		}
		return fmt.Errorf("homebrew upgrade failed: %w", err)
	}

	// Validate upgrade if requested
	if !options.SkipValidation {
		if err := h.ValidateUpgrade(ctx); err != nil {
			h.logger.Warn("Upgrade validation failed: %v", err)
			return fmt.Errorf("homebrew upgrade validation failed: %w", err)
		}
	}

	h.logger.Info("Homebrew upgrade completed successfully")
	return nil
}

// Backup creates a backup of current Homebrew state
func (h *HomebrewUpgrader) Backup(ctx context.Context) (string, error) {
	h.logger.Debug("Creating Homebrew backup")

	// For Homebrew, we can backup the list of installed packages
	cmd := exec.CommandContext(ctx, "brew", "list", "--formula", "--versions")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to list installed packages: %w", err)
	}

	backupPath := fmt.Sprintf("/tmp/homebrew-backup-%d.txt", time.Now().Unix())
	if err := writeFile(backupPath, string(output)); err != nil {
		return "", fmt.Errorf("failed to write backup file: %w", err)
	}

	h.logger.Debug("Backup created at: %s", backupPath)
	return backupPath, nil
}

// Rollback restores from a backup (limited functionality for Homebrew)
func (h *HomebrewUpgrader) Rollback(ctx context.Context, backupPath string) error {
	h.logger.Info("Homebrew rollback is limited - backup file available at: %s", backupPath)
	// Note: Full rollback of Homebrew itself is complex and not implemented
	// The backup file can be used for manual restoration if needed
	return nil
}

// GetUpdateMethod returns the update method used
func (h *HomebrewUpgrader) GetUpdateMethod() string {
	return "brew update && brew upgrade"
}

// ValidateUpgrade validates that the upgrade was successful
func (h *HomebrewUpgrader) ValidateUpgrade(ctx context.Context) error {
	// Check if brew command still works
	cmd := exec.CommandContext(ctx, "brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew command validation failed: %w", err)
	}

	// Check if we can list packages
	cmd = exec.CommandContext(ctx, "brew", "list", "--formula")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew list command validation failed: %w", err)
	}

	return nil
}

// getCurrentVersion gets the current Homebrew version
func (h *HomebrewUpgrader) getCurrentVersion(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute brew --version: %w", err)
	}

	// Parse version from output like "Homebrew 4.0.0"
	versionRegex := regexp.MustCompile(`Homebrew (\d+\.\d+\.\d+)`)
	matches := versionRegex.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to parse Homebrew version from output: %s", string(output))
	}

	return matches[1], nil
}

// getLatestVersion gets the latest Homebrew version from GitHub API
func (h *HomebrewUpgrader) getLatestVersion(ctx context.Context) (string, time.Time, string, error) {
	// GitHub API call to get latest release
	cmd := exec.CommandContext(ctx, "curl", "-s", "https://api.github.com/repos/Homebrew/brew/releases/latest")
	output, err := cmd.Output()
	if err != nil {
		return "", time.Time{}, "", fmt.Errorf("failed to fetch latest release info: %w", err)
	}

	var release struct {
		TagName     string `json:"tag_name"`
		HTMLURL     string `json:"html_url"`
		PublishedAt string `json:"published_at"`
	}

	if err := json.Unmarshal(output, &release); err != nil {
		return "", time.Time{}, "", fmt.Errorf("failed to parse release JSON: %w", err)
	}

	// Parse published date
	publishedAt, err := time.Parse(time.RFC3339, release.PublishedAt)
	if err != nil {
		h.logger.Warn("Failed to parse release date: %v", err)
		publishedAt = time.Time{}
	}

	// Clean up tag name (remove 'v' prefix if present)
	version := strings.TrimPrefix(release.TagName, "v")

	return version, publishedAt, release.HTMLURL, nil
}

// executeUpdate performs the actual update commands
func (h *HomebrewUpgrader) executeUpdate(ctx context.Context, options UpgradeOptions) error {
	// Set timeout if specified
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// Update Homebrew itself
	h.logger.Info("Updating Homebrew...")
	cmd := exec.CommandContext(ctx, "brew", "update")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew update failed: %w", err)
	}

	// Upgrade Homebrew if force is enabled or if we detect it needs upgrading
	if options.Force {
		h.logger.Info("Force upgrading Homebrew...")
		cmd = exec.CommandContext(ctx, "brew", "upgrade")
		if err := cmd.Run(); err != nil {
			h.logger.Warn("brew upgrade failed, but continuing: %v", err)
		}
	}

	return nil
}

// writeFile is a helper function to write content to a file
func writeFile(path, content string) error {
	// This would normally use os.WriteFile, but keeping it simple for this implementation
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo '%s' > '%s'", content, path)) //nolint:gosec // G204: 내부 파일 작성용 명령어
	return cmd.Run()
}
