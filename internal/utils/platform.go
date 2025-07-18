package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetCurrentUsername returns the current username in a cross-platform way.
func GetCurrentUsername() string {
	// Try USER first (Unix/Linux/macOS)
	if user := os.Getenv("USER"); user != "" {
		return user
	}

	// Try USERNAME (Windows)
	if user := os.Getenv("USERNAME"); user != "" {
		return user
	}

	// Fallback to "unknown"
	return "unknown"
}

// GetTempDir returns a cross-platform temporary directory.
func GetTempDir() string {
	return os.TempDir()
}

// GetHomeDir returns the user's home directory in a cross-platform way.
func GetHomeDir() (string, error) {
	return os.UserHomeDir()
}

// GetConfigDir returns the user's configuration directory.
func GetConfigDir() (string, error) {
	home, err := GetHomeDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "windows":
		// Use APPDATA on Windows
		if appData := os.Getenv("APPDATA"); appData != "" {
			return appData, nil
		}

		return filepath.Join(home, "AppData", "Roaming"), nil
	case "darwin":
		// Use Library/Application Support on macOS
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		// Use .config on Unix/Linux
		if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
			return configHome, nil
		}

		return filepath.Join(home, ".config"), nil
	}
}

// SetFilePermissions sets file permissions in a cross-platform way.
func SetFilePermissions(path string, mode os.FileMode) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't support Unix-style permissions, so we just ignore them
		return nil
	}

	return os.Chmod(path, mode)
}

// IsExecutableAvailable checks if an executable is available in PATH.
func IsExecutableAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// GetExecutableName returns the platform-appropriate executable name.
func GetExecutableName(name string) string {
	if runtime.GOOS == "windows" {
		if !strings.HasSuffix(name, ".exe") {
			return name + ".exe"
		}
	}

	return name
}

// GetShellCommand returns the appropriate shell command for the platform.
func GetShellCommand() (string, []string) {
	switch runtime.GOOS {
	case "windows":
		return "cmd", []string{"/C"}
	default:
		// Try to find a shell
		shells := []string{"bash", "sh", "zsh"}
		for _, shell := range shells {
			if IsExecutableAvailable(shell) {
				return shell, []string{"-c"}
			}
		}
		// Fallback to sh
		return "sh", []string{"-c"}
	}
}

// GetPathSeparator returns the path separator for the current platform.
func GetPathSeparator() string {
	return string(os.PathSeparator)
}

// GetListSeparator returns the path list separator for the current platform.
func GetListSeparator() string {
	return string(os.PathListSeparator)
}

// IsWindows returns true if running on Windows.
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMacOS returns true if running on macOS.
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if running on Linux.
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// GetPlatformSpecificConfig returns platform-specific configuration paths.
func GetPlatformSpecificConfig(appName string) ([]string, error) {
	var paths []string

	// User configuration directory
	configDir, err := GetConfigDir()
	if err != nil {
		return nil, err
	}

	paths = append(paths, filepath.Join(configDir, appName))

	// System configuration directory
	switch runtime.GOOS {
	case "windows":
		if programData := os.Getenv("PROGRAMDATA"); programData != "" {
			paths = append(paths, filepath.Join(programData, appName))
		}
	case "darwin":
		paths = append(paths, filepath.Join("/Library", "Application Support", appName))
	default:
		paths = append(paths, filepath.Join("/etc", appName))
	}

	return paths, nil
}

// CreateDirectoryIfNotExists creates a directory if it doesn't exist.
func CreateDirectoryIfNotExists(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, mode); err != nil {
			return err
		}
		// Set permissions (ignored on Windows)
		return SetFilePermissions(path, mode)
	}

	return nil
}

// GetBackupLocations returns platform-appropriate backup locations.
func GetBackupLocations(appName string) []string {
	var locations []string

	// User-specific backup directory
	if configDir, err := GetConfigDir(); err == nil {
		locations = append(locations, filepath.Join(configDir, appName, "backup"))
	}

	// Temporary directory backup
	locations = append(locations, filepath.Join(GetTempDir(), appName+"-backup"))

	// Current directory backup
	locations = append(locations, "./backup")

	return locations
}
