package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetCurrentUsername(t *testing.T) {
	username := GetCurrentUsername()
	if username == "" {
		t.Error("GetCurrentUsername should not return empty string")
	}

	if username == "unknown" {
		t.Skip("Username detection returned fallback value 'unknown'")
	}
}

func TestGetTempDir(t *testing.T) {
	tempDir := GetTempDir()
	if tempDir == "" {
		t.Error("GetTempDir should not return empty string")
	}

	// Verify the directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Errorf("Temp directory %s does not exist", tempDir)
	}
}

func TestGetHomeDir(t *testing.T) {
	homeDir, err := GetHomeDir()
	if err != nil {
		t.Fatalf("GetHomeDir failed: %v", err)
	}

	if homeDir == "" {
		t.Error("GetHomeDir should not return empty string")
	}

	// Verify the directory exists
	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		t.Errorf("Home directory %s does not exist", homeDir)
	}
}

func TestGetConfigDir(t *testing.T) {
	configDir, err := GetConfigDir()
	if err != nil {
		t.Fatalf("GetConfigDir failed: %v", err)
	}

	if configDir == "" {
		t.Error("GetConfigDir should not return empty string")
	}

	// Verify platform-specific behavior
	switch runtime.GOOS {
	case "windows":
		if !strings.Contains(configDir, "AppData") && !strings.Contains(configDir, "APPDATA") {
			t.Errorf("Windows config dir should contain AppData, got: %s", configDir)
		}
	case "darwin":
		if !strings.Contains(configDir, "Library/Application Support") {
			t.Errorf("macOS config dir should contain Library/Application Support, got: %s", configDir)
		}
	default:
		if !strings.Contains(configDir, ".config") && os.Getenv("XDG_CONFIG_HOME") == "" {
			t.Errorf("Unix config dir should contain .config or use XDG_CONFIG_HOME, got: %s", configDir)
		}
	}
}

func TestSetFilePermissions(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(os.TempDir(), "test-permissions")
	defer os.Remove(tempFile)

	// Create the file
	file, err := os.Create(tempFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	file.Close()

	// Test setting permissions
	err = SetFilePermissions(tempFile, 0o644)
	if err != nil {
		t.Errorf("SetFilePermissions failed: %v", err)
	}

	// On Unix systems, verify permissions were set
	if runtime.GOOS != "windows" {
		info, err := os.Stat(tempFile)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}

		if info.Mode().Perm() != 0o644 {
			t.Errorf("Expected permissions 0644, got %o", info.Mode().Perm())
		}
	}
}

func TestIsExecutableAvailable(t *testing.T) {
	// Test with a command that should always be available
	var testCommand string

	switch runtime.GOOS {
	case "windows":
		testCommand = "cmd"
	default:
		testCommand = "sh"
	}

	if !IsExecutableAvailable(testCommand) {
		t.Errorf("Command %s should be available", testCommand)
	}

	// Test with a command that shouldn't exist
	if IsExecutableAvailable("definitely-does-not-exist-command-12345") {
		t.Error("Non-existent command should not be available")
	}
}

func TestGetExecutableName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"git", "git"},
		{"terraform", "terraform"},
	}

	for _, tc := range testCases {
		result := GetExecutableName(tc.input)

		if runtime.GOOS == "windows" {
			expectedWithExt := tc.expected + ".exe"
			if result != expectedWithExt {
				t.Errorf("On Windows, expected %s, got %s", expectedWithExt, result)
			}
		} else {
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		}
	}

	// Test with .exe already present
	if runtime.GOOS == "windows" {
		result := GetExecutableName("test.exe")
		if result != "test.exe" {
			t.Errorf("Should not double-add .exe extension, got %s", result)
		}
	}
}

func TestGetShellCommand(t *testing.T) {
	shell, args := GetShellCommand()

	if shell == "" {
		t.Error("Shell command should not be empty")
	}

	if len(args) == 0 {
		t.Error("Shell args should not be empty")
	}

	// Verify platform-specific behavior
	switch runtime.GOOS {
	case "windows":
		if shell != "cmd" {
			t.Errorf("On Windows, expected cmd, got %s", shell)
		}

		if len(args) != 1 || args[0] != "/C" {
			t.Errorf("On Windows, expected args [/C], got %v", args)
		}
	default:
		// Should be one of the Unix shells
		validShells := []string{"bash", "sh", "zsh"}
		isValid := false

		for _, validShell := range validShells {
			if shell == validShell {
				isValid = true
				break
			}
		}

		if !isValid {
			t.Errorf("Expected Unix shell (bash/sh/zsh), got %s", shell)
		}

		if len(args) != 1 || args[0] != "-c" {
			t.Errorf("On Unix, expected args [-c], got %v", args)
		}
	}
}

func TestGetPathSeparator(t *testing.T) {
	separator := GetPathSeparator()
	expected := string(os.PathSeparator)

	if separator != expected {
		t.Errorf("Expected path separator %q, got %q", expected, separator)
	}

	// Verify platform-specific behavior
	switch runtime.GOOS {
	case "windows":
		if separator != "\\" {
			t.Errorf("On Windows, expected \\, got %s", separator)
		}
	default:
		if separator != "/" {
			t.Errorf("On Unix, expected /, got %s", separator)
		}
	}
}

func TestPlatformDetection(t *testing.T) {
	// Only one should be true
	platforms := []bool{IsWindows(), IsMacOS(), IsLinux()}
	trueCount := 0

	for _, platform := range platforms {
		if platform {
			trueCount++
		}
	}

	if trueCount != 1 {
		t.Errorf("Exactly one platform should be detected as true, got %d", trueCount)
	}

	// Verify consistency with runtime.GOOS
	switch runtime.GOOS {
	case "windows":
		if !IsWindows() {
			t.Error("IsWindows() should return true on Windows")
		}
	case "darwin":
		if !IsMacOS() {
			t.Error("IsMacOS() should return true on macOS")
		}
	case "linux":
		if !IsLinux() {
			t.Error("IsLinux() should return true on Linux")
		}
	}
}

func TestGetPlatformSpecificConfig(t *testing.T) {
	appName := "test-app"

	paths, err := GetPlatformSpecificConfig(appName)
	if err != nil {
		t.Fatalf("GetPlatformSpecificConfig failed: %v", err)
	}

	if len(paths) == 0 {
		t.Error("Should return at least one config path")
	}

	// All paths should contain the app name
	for _, path := range paths {
		if !strings.Contains(path, appName) {
			t.Errorf("Config path should contain app name %s, got %s", appName, path)
		}
	}

	// Verify platform-specific paths
	switch runtime.GOOS {
	case "windows":
		hasAppData := false

		for _, path := range paths {
			if strings.Contains(path, "AppData") || strings.Contains(path, "PROGRAMDATA") {
				hasAppData = true
				break
			}
		}

		if !hasAppData {
			t.Error("Windows config paths should include AppData or PROGRAMDATA")
		}
	case "darwin":
		hasLibrary := false

		for _, path := range paths {
			if strings.Contains(path, "Library") {
				hasLibrary = true
				break
			}
		}

		if !hasLibrary {
			t.Error("macOS config paths should include Library")
		}
	default:
		hasEtc := false

		for _, path := range paths {
			if strings.Contains(path, "/etc") {
				hasEtc = true
				break
			}
		}

		if !hasEtc {
			t.Error("Unix config paths should include /etc")
		}
	}
}

func TestCreateDirectoryIfNotExists(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test-create-dir")
	defer os.RemoveAll(testDir)

	// Directory shouldn't exist initially
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Fatalf("Test directory already exists: %s", testDir)
	}

	// Create directory
	err := CreateDirectoryIfNotExists(testDir, 0o755)
	if err != nil {
		t.Fatalf("CreateDirectoryIfNotExists failed: %v", err)
	}

	// Verify directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Errorf("Directory was not created: %s", testDir)
	}

	// Should not fail if directory already exists
	err = CreateDirectoryIfNotExists(testDir, 0o755)
	if err != nil {
		t.Errorf("CreateDirectoryIfNotExists should not fail if directory exists: %v", err)
	}
}

func TestGetBackupLocations(t *testing.T) {
	appName := "test-backup-app"
	locations := GetBackupLocations(appName)

	if len(locations) == 0 {
		t.Error("Should return at least one backup location")
	}

	// All locations should contain the app name or "backup"
	for _, location := range locations {
		hasAppName := strings.Contains(location, appName)

		hasBackup := strings.Contains(location, "backup")
		if !hasAppName && !hasBackup {
			t.Errorf("Backup location should contain app name or 'backup', got %s", location)
		}
	}

	// Should include temp directory backup
	hasTempBackup := false

	tempDir := GetTempDir()
	for _, location := range locations {
		if strings.Contains(location, tempDir) {
			hasTempBackup = true
			break
		}
	}

	if !hasTempBackup {
		t.Error("Backup locations should include temp directory")
	}

	// Should include current directory backup
	hasCurrentDirBackup := false

	for _, location := range locations {
		if strings.Contains(location, "./backup") {
			hasCurrentDirBackup = true
			break
		}
	}

	if !hasCurrentDirBackup {
		t.Error("Backup locations should include current directory backup")
	}
}
