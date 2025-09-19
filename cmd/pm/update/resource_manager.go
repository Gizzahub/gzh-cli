// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package update

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

// ResourceManager handles resource availability checks and management
// including disk space, network connectivity, and memory monitoring.
type ResourceManager struct {
	AvailableDiskGB        float64 `json:"availableDiskGB"`
	RequiredDiskGB         float64 `json:"requiredDiskGB"`
	EstimatedDownloadMB    float64 `json:"estimatedDownloadMB"`
	NetworkOK              bool    `json:"networkOK"`
	RepositoriesAccessible int     `json:"repositoriesAccessible"`
	MemoryMB               int     `json:"memoryMB"`
	formatter              *OutputFormatter
}

// ResourceCheckResult contains the results of resource availability checks
type ResourceCheckResult struct {
	DiskSpaceOK       bool     `json:"diskSpaceOK"`
	AvailableDiskGB   float64  `json:"availableDiskGB"`
	RequiredDiskGB    float64  `json:"requiredDiskGB"`
	NetworkOK         bool     `json:"networkOK"`
	RepositoriesOK    int      `json:"repositoriesOK"`
	TotalRepositories int      `json:"totalRepositories"`
	MemoryOK          bool     `json:"memoryOK"`
	AvailableMemoryMB int      `json:"availableMemoryMB"`
	Recommendations   []string `json:"recommendations"`
	Errors            []string `json:"errors"`
}

// Repository represents a package repository for network connectivity testing
type Repository struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Manager string `json:"manager"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(formatter *OutputFormatter) *ResourceManager {
	return &ResourceManager{
		formatter: formatter,
	}
}

// CheckResources performs comprehensive resource availability checks
func (rm *ResourceManager) CheckResources(ctx context.Context, managers []string, estimatedDownload float64) (*ResourceCheckResult, error) {
	result := &ResourceCheckResult{
		TotalRepositories: 4, // Default assumption
		Recommendations:   make([]string, 0),
		Errors:            make([]string, 0),
	}

	// Check disk space
	if err := rm.checkDiskSpace(result, estimatedDownload); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Disk space check failed: %v", err))
	}

	// Check network connectivity
	if err := rm.checkNetworkConnectivity(ctx, result, managers); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Network check failed: %v", err))
	}

	// Check memory availability
	if err := rm.checkMemoryAvailability(result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Memory check failed: %v", err))
	}

	// Generate recommendations based on results
	rm.generateRecommendations(result)

	// Print resource check results
	if rm.formatter != nil {
		rm.formatter.PrintResourceCheck(
			result.AvailableDiskGB,
			result.RequiredDiskGB,
			result.NetworkOK,
			result.RepositoriesOK,
		)
	}

	return result, nil
}

// checkDiskSpace checks available disk space and estimates requirements
func (rm *ResourceManager) checkDiskSpace(result *ResourceCheckResult, estimatedDownload float64) error {
	// Get current working directory or home directory for disk space check
	checkPath := "."
	if home, err := os.UserHomeDir(); err == nil {
		checkPath = home
	}

	availableBytes, err := getDiskSpace(checkPath)
	if err != nil {
		return fmt.Errorf("failed to get disk space: %w", err)
	}

	// Calculate available space in GB
	result.AvailableDiskGB = float64(availableBytes) / 1024 / 1024 / 1024

	// Estimate required space (download + temporary files + safety margin)
	requiredGB := (estimatedDownload / 1024) * 2.5 // 2.5x safety margin for temp files
	if requiredGB < 1.0 {
		requiredGB = 1.0 // Minimum 1GB requirement
	}
	result.RequiredDiskGB = requiredGB

	result.DiskSpaceOK = result.AvailableDiskGB >= result.RequiredDiskGB

	rm.AvailableDiskGB = result.AvailableDiskGB
	rm.RequiredDiskGB = result.RequiredDiskGB

	return nil
}

// checkNetworkConnectivity tests connectivity to package repositories
func (rm *ResourceManager) checkNetworkConnectivity(ctx context.Context, result *ResourceCheckResult, managers []string) error {
	repositories := rm.getRepositoriesForManagers(managers)
	result.TotalRepositories = len(repositories)

	if len(repositories) == 0 {
		result.NetworkOK = true
		result.RepositoriesOK = 0
		return nil
	}

	accessible := 0
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for _, repo := range repositories {
		if rm.testRepository(ctx, client, repo) {
			accessible++
		}
	}

	result.RepositoriesOK = accessible
	result.NetworkOK = accessible > len(repositories)/2 // At least half accessible
	rm.NetworkOK = result.NetworkOK
	rm.RepositoriesAccessible = accessible

	return nil
}

// getRepositoriesForManagers returns test repositories for given managers
func (rm *ResourceManager) getRepositoriesForManagers(managers []string) []Repository {
	allRepos := []Repository{
		{Name: "GitHub", URL: "https://api.github.com", Manager: "general"},
		{Name: "Homebrew", URL: "https://formulae.brew.sh", Manager: "brew"},
		{Name: "PyPI", URL: "https://pypi.org/simple/pip/", Manager: "pip"},
		{Name: "npm Registry", URL: "https://registry.npmjs.org", Manager: "npm"},
		{Name: "Ubuntu Archives", URL: "http://archive.ubuntu.com/ubuntu/", Manager: "apt"},
		{Name: "SDKMAN", URL: "https://api.sdkman.io", Manager: "sdkman"},
	}

	var selectedRepos []Repository
	managerSet := make(map[string]bool)
	for _, m := range managers {
		managerSet[m] = true
	}

	for _, repo := range allRepos {
		if repo.Manager == "general" || managerSet[repo.Manager] {
			selectedRepos = append(selectedRepos, repo)
		}
	}

	// Always include at least 4 repositories for testing
	if len(selectedRepos) < 4 {
		selectedRepos = allRepos[:4]
	}

	return selectedRepos
}

// testRepository tests connectivity to a single repository
func (rm *ResourceManager) testRepository(ctx context.Context, client *http.Client, repo Repository) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", repo.URL, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		// Try DNS resolution as fallback
		return rm.testDNSResolution(repo.URL)
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400
}

// testDNSResolution tests if we can resolve DNS for the repository
func (rm *ResourceManager) testDNSResolution(url string) bool {
	// Extract hostname from URL
	if len(url) < 8 { // Minimum for "https://"
		return false
	}

	start := 0
	if url[:7] == "http://" {
		start = 7
	} else if url[:8] == "https://" {
		start = 8
	}

	hostname := url[start:]
	if slashIndex := strings.Index(hostname, "/"); slashIndex != -1 {
		hostname = hostname[:slashIndex]
	}

	_, err := net.LookupHost(hostname)
	return err == nil
}

// checkMemoryAvailability checks system memory availability
func (rm *ResourceManager) checkMemoryAvailability(result *ResourceCheckResult) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get system memory information
	totalMemMB := int(memStats.Sys / 1024 / 1024)
	if totalMemMB < 512 {
		totalMemMB = 4096 // Default assumption if detection fails
	}

	// Estimate available memory (simplified approach)
	availableMemMB := totalMemMB - int(memStats.Alloc/1024/1024)
	if availableMemMB < 0 {
		availableMemMB = totalMemMB / 2 // Conservative estimate
	}

	result.AvailableMemoryMB = availableMemMB
	result.MemoryOK = availableMemMB >= 256 // Minimum 256MB required
	rm.MemoryMB = availableMemMB

	return nil
}

// generateRecommendations generates actionable recommendations based on resource checks
func (rm *ResourceManager) generateRecommendations(result *ResourceCheckResult) {
	// Disk space recommendations
	if !result.DiskSpaceOK {
		shortfall := result.RequiredDiskGB - result.AvailableDiskGB
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Free up %.1fGB disk space before proceeding", shortfall))

		if result.AvailableDiskGB < 5.0 {
			result.Recommendations = append(result.Recommendations,
				"Consider running package manager cleanup commands (brew cleanup, apt autoremove, etc.)")
		}
	}

	// Network connectivity recommendations
	if !result.NetworkOK {
		failedRepos := result.TotalRepositories - result.RepositoriesOK
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("%d/%d repositories unreachable - check network connectivity", failedRepos, result.TotalRepositories))

		if result.RepositoriesOK == 0 {
			result.Recommendations = append(result.Recommendations,
				"Check firewall settings, DNS configuration, or proxy settings")
		}
	}

	// Memory recommendations
	if !result.MemoryOK {
		result.Recommendations = append(result.Recommendations,
			fmt.Sprintf("Low memory detected (%dMB) - close unnecessary applications", result.AvailableMemoryMB))
	}
}

// EstimateDownloadSize estimates total download size for given managers and packages
func (rm *ResourceManager) EstimateDownloadSize(managers []string, packageCounts map[string]int) float64 {
	// Package size estimates in MB per package by manager
	averageSizes := map[string]float64{
		"brew":   8.5,  // Homebrew packages average size
		"asdf":   25.0, // Language runtime downloads
		"npm":    4.2,  // Node.js packages
		"pip":    3.1,  // Python packages
		"apt":    2.8,  // Debian packages
		"pacman": 3.5,  // Arch packages
		"yay":    4.0,  // AUR packages
		"sdkman": 35.0, // SDK downloads
	}

	totalMB := 0.0
	for _, manager := range managers {
		if avgSize, exists := averageSizes[manager]; exists {
			packageCount := packageCounts[manager]
			if packageCount == 0 {
				packageCount = rm.getEstimatedPackageCount(manager)
			}
			totalMB += avgSize * float64(packageCount)
		}
	}

	rm.EstimatedDownloadMB = totalMB
	return totalMB
}

// getEstimatedPackageCount returns estimated package count for a manager
func (rm *ResourceManager) getEstimatedPackageCount(manager string) int {
	// Conservative estimates for typical package counts
	estimates := map[string]int{
		"brew":   5,  // Typical brew upgrade
		"asdf":   2,  // Language versions
		"npm":    8,  // Global packages
		"pip":    6,  // Python packages
		"apt":    12, // System packages
		"pacman": 15, // Arch system update
		"yay":    5,  // AUR packages
		"sdkman": 1,  // SDK updates
	}

	if count, exists := estimates[manager]; exists {
		return count
	}
	return 3 // Default estimate
}

// CheckPrerequisites ensures all prerequisites are met before starting updates
func (rm *ResourceManager) CheckPrerequisites(ctx context.Context, managers []string) error {
	// Estimate download size
	packageCounts := make(map[string]int)
	estimatedMB := rm.EstimateDownloadSize(managers, packageCounts)

	// Perform resource checks
	result, err := rm.CheckResources(ctx, managers, estimatedMB)
	if err != nil {
		return fmt.Errorf("resource check failed: %w", err)
	}

	// Check for critical failures
	if !result.DiskSpaceOK {
		return fmt.Errorf("insufficient disk space: need %.1fGB, available %.1fGB",
			result.RequiredDiskGB, result.AvailableDiskGB)
	}

	if !result.NetworkOK && result.RepositoriesOK == 0 {
		return fmt.Errorf("no package repositories accessible - check network connectivity")
	}

	return nil
}

// GetResourceSummary returns a summary of current resource status
func (rm *ResourceManager) GetResourceSummary() map[string]interface{} {
	return map[string]interface{}{
		"diskSpaceGB":         rm.AvailableDiskGB,
		"estimatedDownloadMB": rm.EstimatedDownloadMB,
		"networkOK":           rm.NetworkOK,
		"repositoriesOK":      rm.RepositoriesAccessible,
		"memoryMB":            rm.MemoryMB,
	}
}

// MonitorResourceUsage monitors resource usage during update process
func (rm *ResourceManager) MonitorResourceUsage(ctx context.Context, interval time.Duration) <-chan map[string]interface{} {
	updates := make(chan map[string]interface{}, 10)

	go func() {
		defer close(updates)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Update memory stats
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)

				usage := map[string]interface{}{
					"timestamp":    time.Now(),
					"memoryUsedMB": memStats.Alloc / 1024 / 1024,
					"totalAllocMB": memStats.TotalAlloc / 1024 / 1024,
					"sysMemMB":     memStats.Sys / 1024 / 1024,
				}

				select {
				case updates <- usage:
				default:
					// Channel full, skip this update
				}
			}
		}
	}()

	return updates
}
