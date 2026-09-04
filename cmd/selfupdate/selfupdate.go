// Copyright (c) 2025 Gizzahub
// SPDX-License-Identifier: MIT

package selfupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gizzahub/gzh-cli/internal/app"
	"github.com/gizzahub/gzh-cli/internal/logger"
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
	downloadClient *http.Client
}

type downloadFile interface {
	io.Writer
	Chmod(mode os.FileMode) error
	Sync() error
	Close() error
}

type replacementResult struct {
	cleanupWarning error
}

func NewUpdater(version string) *Updater {
	return &Updater{
		currentVersion: version,
		logger:         logger.NewStructuredLogger("selfupdate", logger.LevelInfo),
		downloadClient: &http.Client{},
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

// DownloadAsset는 호출자가 쓰기 권한을 가진 새 경로에만 자산을 내려받는다.
// 기존 파일, 심볼릭 링크, 그 밖의 충돌 경로는 덮어쓰지 않는다.
func (u *Updater) DownloadAsset(ctx context.Context, downloadURL, tempPath string) error {
	u.logger.Info("Downloading update", map[string]any{"url": downloadURL})

	//gosec:disable G304 -- AR-2026-001 호출자가 권한을 가진 새 destination만 배타 생성하며 내부 Update는 실행파일 디렉터리만 사용한다.
	tempFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("creating temporary file: %w", err)
	}

	if err := u.downloadAsset(ctx, downloadURL, tempFile); err != nil {
		return errors.Join(err, removeOwnedTemp(tempPath))
	}

	u.logger.Info("Download completed", map[string]any{"path": tempPath})
	return nil
}

func (u *Updater) downloadAsset(ctx context.Context, downloadURL string, tempFile downloadFile) (retErr error) {
	closed := false
	defer func() {
		if closed {
			return
		}
		if err := tempFile.Close(); err != nil {
			retErr = errors.Join(retErr, fmt.Errorf("closing downloaded file: %w", err))
		}
	}()

	ctx, cancel := context.WithTimeout(ctx, downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	client := u.downloadClient
	if client == nil {
		client = &http.Client{}
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading asset: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return fmt.Errorf("writing downloaded file: %w", err)
	}

	if runtime.GOOS != "windows" {
		//gosec:disable G302 -- AR-2026-002 릴리스 바이너리를 실행 가능하게 설치하는 기능 계약상 0755가 필요하다.
		if err := tempFile.Chmod(0o755); err != nil {
			return fmt.Errorf("setting executable permissions: %w", err)
		}
	}

	if err := tempFile.Sync(); err != nil {
		return fmt.Errorf("syncing downloaded file: %w", err)
	}

	closeErr := tempFile.Close()
	closed = true
	if closeErr != nil {
		return fmt.Errorf("closing downloaded file: %w", closeErr)
	}

	return nil
}

func (u *Updater) currentBinaryPath() (string, error) {
	currentPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting current executable path: %w", err)
	}

	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}
	return currentPath, nil
}

func currentBinaryIdentity(currentPath string) (os.FileInfo, error) {
	info, err := os.Stat(currentPath)
	if err != nil {
		return nil, fmt.Errorf("inspecting current binary: %w", err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("current binary must be a regular file: %s", currentPath)
	}
	return info, nil
}

func validateCurrentBinaryIdentity(currentPath string, expected os.FileInfo) error {
	current, err := currentBinaryIdentity(currentPath)
	if err != nil {
		return err
	}
	if !os.SameFile(expected, current) {
		return fmt.Errorf("current binary changed while downloading update: %s", currentPath)
	}
	return nil
}

func (u *Updater) downloadAssetForTarget(ctx context.Context, downloadURL, currentPath string) (string, error) {
	u.logger.Info("Downloading update", map[string]any{"url": downloadURL})

	pattern := ".gz-update-*"
	if runtime.GOOS == "windows" {
		pattern += ".exe"
	}

	tempFile, err := os.CreateTemp(filepath.Dir(currentPath), pattern)
	if err != nil {
		return "", fmt.Errorf("creating owned temporary file: %w", err)
	}
	tempPath := tempFile.Name()

	if err := u.downloadAsset(ctx, downloadURL, tempFile); err != nil {
		return "", errors.Join(err, removeOwnedTemp(tempPath))
	}

	u.logger.Info("Download completed", map[string]any{"path": tempPath})
	return tempPath, nil
}

// removeOwnedTemp는 이 프로세스가 배타적으로 생성한 임시 파일에만 사용한다.
func removeOwnedTemp(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("removing owned temporary file: %w", err)
	}
	return nil
}

func validateReplacementPaths(tempPath, currentPath string) error {
	tempPath = filepath.Clean(tempPath)
	currentPath = filepath.Clean(currentPath)
	if tempPath == currentPath {
		return errors.New("replacement stage must differ from current binary")
	}

	tempDir, err := filepath.Abs(filepath.Dir(tempPath))
	if err != nil {
		return fmt.Errorf("resolving temporary directory: %w", err)
	}
	currentDir, err := filepath.Abs(filepath.Dir(currentPath))
	if err != nil {
		return fmt.Errorf("resolving current binary directory: %w", err)
	}
	rel, err := filepath.Rel(currentDir, tempDir)
	if err != nil {
		return fmt.Errorf("comparing replacement directories: %w", err)
	}
	if rel != "." {
		return fmt.Errorf("replacement stage must be in current binary directory: %s", tempDir)
	}

	// 심볼릭 링크나 특수 파일을 실행 파일 위치로 이동하지 않는다.
	stageInfo, err := os.Lstat(tempPath)
	if err != nil {
		return fmt.Errorf("inspecting replacement stage: %w", err)
	}
	if !stageInfo.Mode().IsRegular() {
		return fmt.Errorf("replacement stage must be a regular file: %s", tempPath)
	}
	return nil
}

// ReplaceCurrentBinary는 현재 실행 파일과 같은 디렉터리의 일반 stage 파일만 교체한다.
func (u *Updater) ReplaceCurrentBinary(tempPath string) error {
	currentPath, err := u.currentBinaryPath()
	if err != nil {
		return err
	}
	return u.replaceCurrentBinary(tempPath, currentPath)
}

func (u *Updater) replaceCurrentBinary(tempPath, currentPath string) error {
	u.logger.Info("Replacing binary", map[string]any{
		"current": currentPath,
		"temp":    tempPath,
	})

	result, err := replaceBinary(tempPath, currentPath)
	if err != nil {
		return err
	}
	if result.cleanupWarning != nil {
		u.logger.Warn("Binary updated with cleanup pending", map[string]any{"error": result.cleanupWarning.Error()})
	}

	u.logger.Info("Binary updated successfully")
	return nil
}

func (u *Updater) Update(ctx context.Context, force bool) error {
	u.logger.Info("Checking for updates", map[string]any{"current_version": u.currentVersion})

	release, err := u.GetLatestRelease(ctx)
	if err != nil {
		return fmt.Errorf("getting latest release: %w", err)
	}

	if !force && !u.IsNewerVersion(release.TagName) {
		u.logger.Info("Already using the latest version", map[string]any{"version": u.currentVersion})
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

	currentPath, err := u.currentBinaryPath()
	if err != nil {
		return err
	}
	currentIdentity, err := currentBinaryIdentity(currentPath)
	if err != nil {
		return err
	}

	tempPath, err := u.downloadAssetForTarget(ctx, downloadURL, currentPath)
	if err != nil {
		return fmt.Errorf("downloading update: %w", err)
	}
	if err := validateCurrentBinaryIdentity(currentPath, currentIdentity); err != nil {
		return errors.Join(err, removeOwnedTemp(tempPath))
	}

	if err := u.replaceCurrentBinary(tempPath, currentPath); err != nil {
		return errors.Join(fmt.Errorf("replacing binary: %w", err), removeOwnedTemp(tempPath))
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
