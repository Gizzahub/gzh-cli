// Copyright (c) 2025 Archmagece
// SPDX-License-Identifier: MIT

package ide

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// IDE represents an IDE installation
type IDE struct {
	Name          string    `json:"name"`
	Executable    string    `json:"executable"`
	Version       string    `json:"version"`
	Type          string    `json:"type"`
	InstallMethod string    `json:"install_method"`
	InstallPath   string    `json:"install_path"`
	LastUpdated   time.Time `json:"last_updated"`
	Aliases       []string  `json:"aliases"`
}

// IDECache represents cached IDE scan results
type IDECache struct {
	Timestamp time.Time `json:"timestamp"`
	IDEs      []IDE     `json:"ides"`
}

// IDEDetector handles IDE detection logic
type IDEDetector struct {
	cacheDir string
}

// NewIDEDetector creates a new IDE detector
func NewIDEDetector() *IDEDetector {
	homeDir, _ := os.UserHomeDir()
	cacheDir := filepath.Join(homeDir, ".gz", "cache")

	return &IDEDetector{
		cacheDir: cacheDir,
	}
}

// detectInstallMethod determines how an IDE was installed
func (d *IDEDetector) detectInstallMethod(execPath string) (method, installPath string) {
	// Check if it's an AppImage launcher script
	if method, path := d.detectAppImageLauncher(execPath); method != "" {
		return method, path
	}

	// Check package managers
	if method, path := d.detectPackageManager(execPath); method != "" {
		return method, path
	}

	// Check if it's a JetBrains Toolbox installation
	if strings.Contains(execPath, "JetBrains/Toolbox") {
		return "toolbox", execPath
	}

	// Default to direct installation
	return "direct", execPath
}

// detectAppImageLauncher checks if executable is a script launching an AppImage
func (d *IDEDetector) detectAppImageLauncher(execPath string) (string, string) {
	// Read the file to see if it's a script
	content, err := os.ReadFile(execPath)
	if err != nil {
		return "", ""
	}

	contentStr := string(content)

	// Check if it's a shell script
	if !strings.HasPrefix(contentStr, "#!/") {
		return "", ""
	}

	// Look for AppImage patterns in the script
	lines := strings.Split(contentStr, "\n")
	var appDir string

	// First pass: look for variable definitions (like APP_DIR="/home/user/Apps")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "APP_DIR=") {
			// Extract directory from APP_DIR="/path/to/dir"
			parts := strings.Split(line, "=")
			if len(parts) >= 2 {
				appDir = strings.Trim(parts[1], "\"")
			}
		}
	}

	// Second pass: look for AppImage usage
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".AppImage") {
			// Pattern: ls -1 "$APP_DIR"/Cursor-*.AppImage
			if strings.Contains(line, "*.AppImage") {
				if appDir != "" {
					return "appimage", appDir
				}
				// Try to extract from the line itself
				if dir := d.extractAppImageDir(line); dir != "" && dir != "VAR_REFERENCE" {
					return "appimage", dir
				}
			}
			// Direct AppImage execution: exec "$latest" "$@"
			if strings.Contains(line, "exec") && strings.Contains(line, "\"$") {
				if appDir != "" {
					return "appimage", appDir
				}
			}
		}
	}

	return "", ""
}

// extractAppImageDir extracts the AppImage directory from a launcher script line
func (d *IDEDetector) extractAppImageDir(line string) string {
	// Pattern: ls -1 "$APP_DIR"/Cursor-*.AppImage
	if strings.Contains(line, "\"$") && strings.Contains(line, "*.AppImage") {
		// Extract the pattern part: "$APP_DIR"/Cursor-*.AppImage
		start := strings.Index(line, "\"$")
		if start != -1 {
			end := strings.Index(line[start:], "*.AppImage")
			if end != -1 {
				pathPattern := line[start : start+end]
				// Remove quotes and extract the directory part
				pathPattern = strings.Trim(pathPattern, "\"")
				if strings.HasPrefix(pathPattern, "$") {
					// Return a marker that this is a variable, will be resolved later
					return "VAR_REFERENCE"
				}
			}
		}
	}

	// Pattern: /path/to/dir/App-*.AppImage (direct path)
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Contains(part, "*.AppImage") && strings.Contains(part, "/") {
			// Extract directory from pattern like "/home/user/Apps/Cursor-*.AppImage"
			cleaned := strings.Trim(part, "\"")
			return filepath.Dir(strings.Replace(cleaned, "*.AppImage", "", 1))
		}
	}

	return ""
}

// detectPackageManager checks which package manager installed the executable
func (d *IDEDetector) detectPackageManager(execPath string) (string, string) {
	// Try pacman first (Arch Linux) - check if file is owned by a package
	if d.isCommandAvailable("pacman") {
		if pkg := d.queryPackageManager("pacman", "-Qo", execPath); pkg != "" {
			return "pacman", pkg
		}
	}

	// Skip snap and flatpak for now as they can be slow
	// TODO: Add proper timeout handling for these package managers

	return "", ""
}

// isCommandAvailable checks if a command is available in PATH
func (d *IDEDetector) isCommandAvailable(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// queryPackageManager queries a package manager for package info
func (d *IDEDetector) queryPackageManager(manager string, args ...string) string {
	cmd := exec.Command(manager, args...)
	cmd.Stderr = nil // Suppress error output

	// Set a timeout to prevent hanging
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// queryFlatpak specifically handles flatpak queries
func (d *IDEDetector) queryFlatpak(appName string) string {
	cmd := exec.Command("flatpak", "list", "--columns=name,application")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(appName)) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}

// DetectIDEs scans the system for installed IDEs
func (d *IDEDetector) DetectIDEs(useCache bool) ([]IDE, error) {
	if useCache {
		if ides, err := d.loadFromCache(); err == nil {
			return ides, nil
		}
	}

	var allIDEs []IDE

	// Detect JetBrains IDEs
	jetbrainsIDEs, err := d.detectJetBrainsIDEs()
	if err == nil {
		allIDEs = append(allIDEs, jetbrainsIDEs...)
	}

	// Detect VS Code family
	vscodeIDEs, err := d.detectVSCodeFamily()
	if err == nil {
		allIDEs = append(allIDEs, vscodeIDEs...)
	}

	// Detect other IDEs
	otherIDEs, err := d.detectOtherIDEs()
	if err == nil {
		allIDEs = append(allIDEs, otherIDEs...)
	}

	// Save to cache
	if err := d.saveToCache(allIDEs); err != nil {
		// Don't fail if we can't save cache
		fmt.Printf("Warning: Failed to save IDE cache: %v\n", err)
	}

	return allIDEs, nil
}

// detectJetBrainsIDEs detects JetBrains IDE installations
func (d *IDEDetector) detectJetBrainsIDEs() ([]IDE, error) {
	var ides []IDE

	// Check JetBrains Toolbox installations
	toolboxPath := d.getJetBrainsToolboxPath()
	if _, err := os.Stat(toolboxPath); err == nil {
		entries, err := os.ReadDir(toolboxPath)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				ideName := entry.Name()
				idePath := filepath.Join(toolboxPath, ideName)

				if ide := d.createJetBrainsIDE(ideName, idePath); ide != nil {
					ides = append(ides, *ide)
				}
			}
		}
	}

	// Check system-wide installations
	systemPaths := d.getJetBrainsSystemPaths()
	for _, path := range systemPaths {
		if _, err := os.Stat(path); err == nil {
			if ide := d.createJetBrainsIDE(filepath.Base(path), path); ide != nil {
				ides = append(ides, *ide)
			}
		}
	}

	return ides, nil
}

// detectVSCodeFamily detects VS Code family IDEs
func (d *IDEDetector) detectVSCodeFamily() ([]IDE, error) {
	var ides []IDE

	vscodeTypes := []struct {
		name        string
		executable  string
		aliases     []string
		versionArgs []string
	}{
		{"Visual Studio Code", "code", []string{"code", "vscode"}, []string{"--version"}},
		{"Visual Studio Code Insiders", "code-insiders", []string{"code-insiders"}, []string{"--version"}},
		{"Cursor", "cursor", []string{"cursor"}, []string{"--version", "-v", "-V"}},
		{"VSCodium", "codium", []string{"codium"}, []string{"--version"}},
	}

	for _, vscode := range vscodeTypes {
		if path := d.findExecutable(vscode.executable); path != "" {
			// Detect installation method first
			installMethod, installPath := d.detectInstallMethod(path)

			// Enhanced version detection based on installation method
			version := d.getEnhancedVersion(path, vscode.versionArgs, installMethod, installPath, vscode.name)
			lastUpdated := d.getExecutableLastModified(path)

			ide := IDE{
				Name:          vscode.name,
				Executable:    path,
				Version:       version,
				Type:          "vscode",
				InstallMethod: installMethod,
				InstallPath:   installPath,
				LastUpdated:   lastUpdated,
				Aliases:       vscode.aliases,
			}
			ides = append(ides, ide)
		}
	}

	return ides, nil
}

// detectOtherIDEs detects other IDE installations
func (d *IDEDetector) detectOtherIDEs() ([]IDE, error) {
	var ides []IDE

	otherIDEs := []struct {
		name       string
		executable string
		aliases    []string
		versionArg string
	}{
		{"Sublime Text", "subl", []string{"subl", "sublime"}, "--version"},
		{"Neovim", "nvim", []string{"nvim", "neovim"}, "--version"},
		{"Vim", "vim", []string{"vim"}, "--version"},
		{"Emacs", "emacs", []string{"emacs"}, "--version"},
	}

	for _, other := range otherIDEs {
		if path := d.findExecutable(other.executable); path != "" {
			// Detect installation method
			installMethod, installPath := d.detectInstallMethod(path)

			version := d.getExecutableVersion(path, other.versionArg)
			lastUpdated := d.getExecutableLastModified(path)

			ide := IDE{
				Name:          other.name,
				Executable:    path,
				Version:       version,
				Type:          "other",
				InstallMethod: installMethod,
				InstallPath:   installPath,
				LastUpdated:   lastUpdated,
				Aliases:       other.aliases,
			}
			ides = append(ides, ide)
		}
	}

	return ides, nil
}

// createJetBrainsIDE creates an IDE instance for JetBrains products
func (d *IDEDetector) createJetBrainsIDE(ideName, idePath string) *IDE {
	// Map JetBrains product names
	jetbrainsProducts := map[string]struct {
		displayName string
		aliases     []string
		executable  string
	}{
		"intellij-idea-ultimate":  {"IntelliJ IDEA Ultimate", []string{"idea", "intellij"}, "idea.sh"},
		"intellij-idea-community": {"IntelliJ IDEA Community", []string{"idea", "intellij"}, "idea.sh"},
		"pycharm":                 {"PyCharm Professional", []string{"pycharm"}, "pycharm.sh"},
		"pycharm-community":       {"PyCharm Community", []string{"pycharm"}, "pycharm.sh"},
		"webstorm":                {"WebStorm", []string{"webstorm"}, "webstorm.sh"},
		"phpstorm":                {"PhpStorm", []string{"phpstorm"}, "phpstorm.sh"},
		"rubymine":                {"RubyMine", []string{"rubymine"}, "rubymine.sh"},
		"clion":                   {"CLion", []string{"clion"}, "clion.sh"},
		"goland":                  {"GoLand", []string{"goland"}, "goland.sh"},
		"datagrip":                {"DataGrip", []string{"datagrip"}, "datagrip.sh"},
		"dataspell":               {"DataSpell", []string{"dataspell"}, "dataspell.sh"},
		"rider":                   {"Rider", []string{"rider"}, "rider.sh"},
		"rustrover":               {"RustRover", []string{"rustrover"}, "rustrover.sh"},
		"android-studio":          {"Android Studio", []string{"android-studio"}, "studio.sh"},
	}

	// Normalize the product name
	normalizedName := strings.ToLower(ideName)
	for prefix, product := range jetbrainsProducts {
		if strings.HasPrefix(normalizedName, prefix) {
			// Find the executable
			execPath := d.findJetBrainsExecutable(idePath, product.executable)
			if execPath == "" {
				return nil
			}

			// Get version from build.txt file first, then fallback to other methods
			version := d.getJetBrainsVersion(idePath, execPath, ideName)
			lastUpdated := d.getExecutableLastModified(execPath)

			// Detect installation method
			installMethod, installPath := d.detectInstallMethod(execPath)

			return &IDE{
				Name:          product.displayName,
				Executable:    execPath,
				Version:       version,
				Type:          "jetbrains",
				InstallMethod: installMethod,
				InstallPath:   installPath,
				LastUpdated:   lastUpdated,
				Aliases:       product.aliases,
			}
		}
	}

	return nil
}

// findJetBrainsExecutable finds the executable for a JetBrains product
func (d *IDEDetector) findJetBrainsExecutable(productPath, executableName string) string {
	// For Toolbox installations, look in bin/ subdirectory
	binPath := filepath.Join(productPath, "bin", executableName)
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}

	// For Toolbox, also check in the latest version directory
	entries, err := os.ReadDir(productPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				versionBinPath := filepath.Join(productPath, entry.Name(), "bin", executableName)
				if _, err := os.Stat(versionBinPath); err == nil {
					return versionBinPath
				}
			}
		}
	}

	return ""
}

// getJetBrainsVersion gets version from multiple sources with priority
func (d *IDEDetector) getJetBrainsVersion(productPath, execPath, productDir string) string {
	// 1. Try to read from build.txt file (most accurate)
	if version := d.getJetBrainsVersionFromBuildFile(productPath); version != "unknown" {
		return version
	}

	// 2. Try to get version from executable --version command
	if version := d.getJetBrainsVersionFromCommand(execPath); version != "unknown" {
		return version
	}

	// 3. Fallback to directory name parsing
	return d.extractJetBrainsVersionFromDir(productDir)
}

// getJetBrainsVersionFromBuildFile reads version from build.txt file
func (d *IDEDetector) getJetBrainsVersionFromBuildFile(productPath string) string {
	buildFilePath := filepath.Join(productPath, "build.txt")

	data, err := os.ReadFile(buildFilePath)
	if err != nil {
		return "unknown"
	}

	buildNumber := strings.TrimSpace(string(data))
	if buildNumber == "" {
		return "unknown"
	}

	// Convert build number to user-friendly version
	return d.parseJetBrainsBuildNumber(buildNumber)
}

// getJetBrainsVersionFromCommand gets version from executable --version command
func (d *IDEDetector) getJetBrainsVersionFromCommand(execPath string) string {
	cmd := exec.Command(execPath, "--version")
	cmd.Stderr = nil // Suppress warnings

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Look for version patterns like "PyCharm 2025.2.0.1"
		if strings.Contains(line, "PyCharm") || strings.Contains(line, "IntelliJ") ||
			strings.Contains(line, "WebStorm") || strings.Contains(line, "GoLand") ||
			strings.Contains(line, "CLion") || strings.Contains(line, "DataGrip") ||
			strings.Contains(line, "PhpStorm") || strings.Contains(line, "RubyMine") ||
			strings.Contains(line, "Rider") || strings.Contains(line, "RustRover") ||
			strings.Contains(line, "DataSpell") {

			// Extract version number from line
			parts := strings.Fields(line)
			for _, part := range parts {
				if d.isVersionNumber(part) {
					return part
				}
			}
		}
	}

	return "unknown"
}

// parseJetBrainsBuildNumber converts build number to user-friendly version
func (d *IDEDetector) parseJetBrainsBuildNumber(buildNumber string) string {
	// Build number format: PY-252.23892.515, IU-252.23892.409, etc.
	// Convert to version like: 2025.2, 2025.2.0.1

	parts := strings.Split(buildNumber, "-")
	if len(parts) < 2 {
		return buildNumber // Return as-is if format is unexpected
	}

	versionPart := parts[1]
	versionSegments := strings.Split(versionPart, ".")

	if len(versionSegments) >= 2 {
		// Extract year and major version from first segment
		if len(versionSegments[0]) >= 3 {
			year := versionSegments[0][:3]  // "252" -> "252"
			major := versionSegments[0][3:] // remaining part

			// Convert 252 -> 2025.2
			if year == "252" {
				baseVersion := "2025.2"
				if major != "" {
					baseVersion += "." + major
				}

				// Add build number if available
				if len(versionSegments) >= 3 {
					baseVersion += "." + versionSegments[2]
				}

				return baseVersion
			}
		}
	}

	// If parsing fails, return the build number as-is
	return buildNumber
}

// extractJetBrainsVersionFromDir extracts version from JetBrains product directory name
func (d *IDEDetector) extractJetBrainsVersionFromDir(productDir string) string {
	// Extract version from directory name like "pycharm-2024.3" or similar
	parts := strings.Split(productDir, "-")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

// isVersionNumber checks if a string looks like a version number
func (d *IDEDetector) isVersionNumber(s string) bool {
	// Version pattern: X.Y.Z.W or X.Y.Z or X.Y
	if len(s) < 3 {
		return false
	}

	// Check for digit.digit pattern
	hasDigit := false
	hasDot := false

	for _, char := range s {
		if char >= '0' && char <= '9' {
			hasDigit = true
		} else if char == '.' {
			hasDot = true
		} else if char != '-' && char != '_' {
			return false // Invalid character for version
		}
	}

	return hasDigit && hasDot
}

// getJetBrainsToolboxPath returns the JetBrains Toolbox installation path
func (d *IDEDetector) getJetBrainsToolboxPath() string {
	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "linux":
		return filepath.Join(homeDir, ".local", "share", "JetBrains", "Toolbox", "apps")
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "JetBrains", "Toolbox", "apps")
	case "windows":
		return filepath.Join(homeDir, "AppData", "Local", "JetBrains", "Toolbox", "apps")
	default:
		return ""
	}
}

// getJetBrainsSystemPaths returns system-wide JetBrains installation paths
func (d *IDEDetector) getJetBrainsSystemPaths() []string {
	switch runtime.GOOS {
	case "linux":
		return []string{
			"/opt/idea",
			"/opt/pycharm",
			"/opt/webstorm",
			"/opt/phpstorm",
			"/opt/goland",
			"/opt/clion",
			"/usr/local/bin",
		}
	case "darwin":
		return []string{
			"/Applications/IntelliJ IDEA.app",
			"/Applications/PyCharm.app",
			"/Applications/WebStorm.app",
			"/Applications/PhpStorm.app",
			"/Applications/GoLand.app",
			"/Applications/CLion.app",
		}
	case "windows":
		return []string{
			"C:\\Program Files\\JetBrains",
			"C:\\Program Files (x86)\\JetBrains",
		}
	default:
		return []string{}
	}
}

// findExecutable searches for an executable in PATH
func (d *IDEDetector) findExecutable(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return ""
}

// getEnhancedVersion gets version using installation method-specific strategies
func (d *IDEDetector) getEnhancedVersion(execPath string, versionArgs []string, installMethod, installPath, appName string) string {
	// Try standard version detection first
	version := d.getVSCodeFamilyVersion(execPath, versionArgs)
	if version != "unknown" && version != "" {
		return version
	}

	// If standard detection failed, try method-specific approaches
	switch installMethod {
	case "appimage":
		if appImageVersion := d.getAppImageVersion(installPath, appName); appImageVersion != "unknown" {
			return appImageVersion
		}
		// Temporarily disable package manager version detection
		/*
			case "pacman":
				if pkgVersion := d.getPackageManagerVersion("pacman", "-Q", appName); pkgVersion != "unknown" {
					return pkgVersion
				}
			case "snap":
				if snapVersion := d.getPackageManagerVersion("snap", "list", appName); snapVersion != "unknown" {
					return snapVersion
				}
			case "flatpak":
				if flatpakVersion := d.getFlatpakVersion(appName); flatpakVersion != "unknown" {
					return flatpakVersion
				}
		*/
	}

	return "unknown"
}

// getAppImageVersion extracts version from AppImage installations
func (d *IDEDetector) getAppImageVersion(installPath, appName string) string {
	// For AppImage launchers, we need to find the actual AppImage files
	appDir := d.resolveAppImageDirectory(installPath, appName)
	if appDir == "" {
		return "unknown"
	}

	// Find AppImage files matching the app name
	pattern := filepath.Join(appDir, strings.Title(appName)+"-*.AppImage")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		// Try lowercase
		pattern = filepath.Join(appDir, strings.ToLower(appName)+"-*.AppImage")
		matches, err = filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			return "unknown"
		}
	}

	// Find the latest AppImage (highest version)
	var latestVersion string
	for _, match := range matches {
		if version := d.extractVersionFromAppImageName(match, appName); version != "" {
			if latestVersion == "" || d.compareVersions(version, latestVersion) > 0 {
				latestVersion = version
			}
		}
	}

	if latestVersion != "" {
		return latestVersion
	}

	return "unknown"
}

// resolveAppImageDirectory resolves the actual directory from launcher scripts
func (d *IDEDetector) resolveAppImageDirectory(installPath, appName string) string {
	// If we have a direct path from the script, use it
	if strings.Contains(installPath, "/") && installPath != "VAR_REFERENCE" {
		// Validate the directory exists
		if _, err := os.Stat(installPath); err == nil {
			return installPath
		}
	}

	// Fallback to common AppImage locations
	homeDir, _ := os.UserHomeDir()
	possibleDirs := []string{
		filepath.Join(homeDir, "Apps"),
		filepath.Join(homeDir, "Applications"),
		filepath.Join(homeDir, ".local", "share", "applications"),
		"/opt",
	}

	for _, dir := range possibleDirs {
		if _, err := os.Stat(dir); err == nil {
			return dir
		}
	}

	return ""
}

// extractVersionFromAppImageName extracts version from AppImage filename
func (d *IDEDetector) extractVersionFromAppImageName(filename, appName string) string {
	base := filepath.Base(filename)

	// Remove .AppImage extension
	name := strings.TrimSuffix(base, ".AppImage")

	// Pattern: AppName-Version-BuildInfo
	// e.g., Cursor-0.42.3-build-123456x64 -> 0.42.3
	parts := strings.Split(name, "-")
	if len(parts) >= 2 {
		// Look for version-like parts (starts with digit, contains dots)
		for i := 1; i < len(parts); i++ {
			if d.isVersionNumber(parts[i]) {
				return parts[i]
			}
		}
	}

	return ""
}

// compareVersions compares two version strings (simple comparison)
func (d *IDEDetector) compareVersions(v1, v2 string) int {
	// Simple lexicographic comparison for now
	// In a more sophisticated implementation, we'd parse semantic versions
	if v1 > v2 {
		return 1
	}
	if v1 < v2 {
		return -1
	}
	return 0
}

// getPackageManagerVersion gets version from package managers
func (d *IDEDetector) getPackageManagerVersion(manager, command, packageName string) string {
	var cmd *exec.Cmd

	switch manager {
	case "pacman":
		cmd = exec.Command("pacman", "-Q", packageName)
	case "snap":
		cmd = exec.Command("snap", "list", packageName)
	default:
		return "unknown"
	}

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse the output for version information
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		line := strings.TrimSpace(lines[0])
		// For pacman: "cursor 0.42.3-1"
		// For snap: "cursor 0.42.3 123 latest/stable"
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	return "unknown"
}

// getFlatpakVersion gets version from flatpak
func (d *IDEDetector) getFlatpakVersion(appName string) string {
	cmd := exec.Command("flatpak", "list", "--columns=name,version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), strings.ToLower(appName)) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return "unknown"
}

// getVSCodeFamilyVersion gets version for VS Code family with multiple version arguments
func (d *IDEDetector) getVSCodeFamilyVersion(execPath string, versionArgs []string) string {
	for _, arg := range versionArgs {
		if version := d.getExecutableVersion(execPath, arg); version != "unknown" && version != "" {
			// For VS Code family, extract just the version number
			return d.parseVSCodeVersion(version)
		}
	}
	return "unknown"
}

// parseVSCodeVersion extracts clean version from VS Code output
func (d *IDEDetector) parseVSCodeVersion(output string) string {
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])

		// VS Code output format: "1.103.1" (just version)
		// Cursor might have different format
		if d.isVersionNumber(firstLine) {
			return firstLine
		}

		// Try to extract version from line with app name
		parts := strings.Fields(firstLine)
		for _, part := range parts {
			if d.isVersionNumber(part) {
				return part
			}
		}
	}
	return "unknown"
}

// getExecutableVersion gets version information from an executable
func (d *IDEDetector) getExecutableVersion(execPath, versionArg string) string {
	cmd := exec.Command(execPath, versionArg)
	cmd.Stderr = nil // Suppress error output

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	// Parse version from output (first line, extract version numbers)
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0])
	}

	return "unknown"
}

// getExecutableLastModified gets the last modified time of an executable
func (d *IDEDetector) getExecutableLastModified(execPath string) time.Time {
	if info, err := os.Stat(execPath); err == nil {
		return info.ModTime()
	}
	return time.Time{}
}

// getCacheFilePath returns the path to the IDE cache file
func (d *IDEDetector) getCacheFilePath() string {
	return filepath.Join(d.cacheDir, "ide.json")
}

// loadFromCache loads IDE information from cache
func (d *IDEDetector) loadFromCache() ([]IDE, error) {
	cacheFile := d.getCacheFilePath()

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var cache IDECache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	// Check if cache is still valid (24 hours)
	if time.Since(cache.Timestamp) > 24*time.Hour {
		return nil, fmt.Errorf("cache expired")
	}

	return cache.IDEs, nil
}

// saveToCache saves IDE information to cache
func (d *IDEDetector) saveToCache(ides []IDE) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(d.cacheDir, 0o755); err != nil {
		return err
	}

	cache := IDECache{
		Timestamp: time.Now(),
		IDEs:      ides,
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	cacheFile := d.getCacheFilePath()
	return os.WriteFile(cacheFile, data, 0o644)
}

// FindIDEByAlias finds an IDE by its name or alias
func (d *IDEDetector) FindIDEByAlias(ides []IDE, nameOrAlias string) *IDE {
	nameOrAlias = strings.ToLower(nameOrAlias)

	for _, ide := range ides {
		// Check exact name match
		if strings.ToLower(ide.Name) == nameOrAlias {
			return &ide
		}

		// Check aliases
		for _, alias := range ide.Aliases {
			if strings.ToLower(alias) == nameOrAlias {
				return &ide
			}
		}
	}

	return nil
}
