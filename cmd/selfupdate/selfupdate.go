// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/Gizzahub/gzh-cli/internal/app"
	"github.com/Gizzahub/gzh-cli/internal/logger"
)

const (
	githubAPIURL    = "https://api.github.com"
	githubRepo      = "Gizzahub/gzh-cli"
	downloadTimeout = 5 * time.Minute
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	Prerelease bool `json:"prerelease"`
}

type Updater struct {
	currentVersion string
	logger         *logger.StructuredLogger
}

func NewUpdater(version string) *Updater {
	return &Updater{
		currentVersion: version,
		logger:         logger.NewStructuredLogger("selfupdate", logger.LevelInfo),
	}
}

func (u *Updater) GetLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", githubAPIURL, githubRepo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("decoding release response: %w", err)
	}

	return &release, nil
}

func (u *Updater) IsNewerVersion(remoteVersion string) bool {
	if u.currentVersion == "" || u.currentVersion == "dev" {
		return true
	}

	// Remove 'v' prefix if present
	current := strings.TrimPrefix(u.currentVersion, "v")
	remote := strings.TrimPrefix(remoteVersion, "v")

	return current != remote
}

func (u *Updater) GetAssetName() string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go architecture names to release naming convention
	switch arch {
	case "amd64":
		arch = "x86_64"
	case "386":
		arch = "i386"
	}

	var suffix string
	if os == "windows" {
		suffix = ".exe"
	}

	return fmt.Sprintf("gz_%s_%s%s", os, arch, suffix)
}

func (u *Updater) DownloadAsset(ctx context.Context, downloadURL, tempPath string) error {
	u.logger.Info("Downloading update", map[string]interface{}{"url": downloadURL})

	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}
	defer tempFile.Close()

	// Copy downloaded content to temporary file
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("writing downloaded file: %w", err)
	}

	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(tempPath, 0o755); err != nil {
			return fmt.Errorf("setting executable permissions: %w", err)
		}
	}

	u.logger.Info("Download completed", map[string]interface{}{"path": tempPath})
	return nil
}

func (u *Updater) ReplaceCurrentBinary(tempPath string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting current executable path: %w", err)
	}

	// Resolve any symlinks
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	u.logger.Info("Replacing binary", map[string]interface{}{
		"current": currentPath,
		"temp":    tempPath,
	})

	// On Windows, we might need to rename the old file first
	if runtime.GOOS == "windows" {
		backupPath := currentPath + ".old"
		if err := os.Rename(currentPath, backupPath); err != nil {
			return fmt.Errorf("backing up current binary: %w", err)
		}

		if err := os.Rename(tempPath, currentPath); err != nil {
			// Try to restore backup
			os.Rename(backupPath, currentPath)
			return fmt.Errorf("replacing binary: %w", err)
		}

		// Remove backup
		os.Remove(backupPath)
	} else {
		// On Unix systems, we can replace directly
		if err := os.Rename(tempPath, currentPath); err != nil {
			return fmt.Errorf("replacing binary: %w", err)
		}
	}

	u.logger.Info("Binary updated successfully")
	return nil
}

func (u *Updater) Update(ctx context.Context, force bool) error {
	u.logger.Info("Checking for updates", map[string]interface{}{"current_version": u.currentVersion})

	release, err := u.GetLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("getting latest release: %w", err)
	}

	if !force && !u.IsNewerVersion(release.TagName) {
		u.logger.Info("Already using the latest version", map[string]interface{}{"version": u.currentVersion})
		fmt.Printf("gz is already up to date (version %s)\n", u.currentVersion)
		return nil
	}

	// Find the appropriate asset for current platform
	assetName := u.GetAssetName()
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.Name == assetName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("no asset found for platform %s/%s (looking for %s)", runtime.GOOS, runtime.GOARCH, assetName)
	}

	// Create temporary file for download
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, "gz_update_"+time.Now().Format("20060102_150405"))
	if runtime.GOOS == "windows" {
		tempPath += ".exe"
	}

	// Download the new version
	if err := u.DownloadAsset(ctx, downloadURL, tempPath); err != nil {
		os.Remove(tempPath) // Clean up on error
		return fmt.Errorf("downloading update: %w", err)
	}

	// Replace current binary
	if err := u.ReplaceCurrentBinary(tempPath); err != nil {
		os.Remove(tempPath) // Clean up on error
		return fmt.Errorf("replacing binary: %w", err)
	}

	fmt.Printf("✅ Successfully updated gz to version %s\n", release.TagName)
	return nil
}

func NewSelfUpdateCmd(appCtx *app.AppContext) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "selfupdate",
		Short: "Update gz binary to the latest version",
		Long: `Download and install the latest version of gz from GitHub releases.

This command checks GitHub for the latest release and automatically downloads
and replaces the current gz binary with the updated version.

Examples:
  gz selfupdate           # Check and update to latest version
  gz selfupdate --force   # Force update even if already latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// Get current version from root command
			version := "dev"
			if rootCmd := cmd.Root(); rootCmd != nil && rootCmd.Version != "" {
				version = rootCmd.Version
			}

			updater := NewUpdater(version)
			return updater.Update(ctx, force)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already using latest version")

	return cmd
}
